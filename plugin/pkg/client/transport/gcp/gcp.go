/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gcp

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"k8s.io/kubernetes/pkg/client/transport"
)

func init() {
	transport.RegisterRoundTripperPlugin("gcp", &gcpRoundTripperPlugin{})
}

type gcpRoundTripperPlugin struct{}

func (*gcpRoundTripperPlugin) RoundTripper(rt http.RoundTripper) (http.RoundTripper, error) {
	ts, err := google.DefaultTokenSource(context.TODO(), "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}
	return &oauth2.Transport{
		Source: &cachedTokenSource{ts, path.Join(os.Getenv("HOME"), "/.kube/access_token")},
		Base:   rt,
	}, nil
}

type cachedTokenSource struct {
	source    oauth2.TokenSource
	tokenFile string
}

// Token returns an OAuth2 access token, either from a file, or retrived from
// the original token source.
func (c *cachedTokenSource) Token() (*oauth2.Token, error) {
	// If we have a valid, cached access token, use it.
	if tok, err := parseTokenFromFile(c.tokenFile); err == nil && isValid(tok) {
		return tok, nil
	}
	tok, err := c.source.Token()
	if err != nil {
		return nil, err
	}
	// Cache the token on disk.
	if err := saveTokenToFile(tok, c.tokenFile); err != nil {
		glog.Warningf("Failed to save token to file: %v", err)
	}
	return tok, nil
}

func isValid(tok *oauth2.Token) bool {
	return tok.Valid() && time.Now().Before(tok.Expiry.Add(-30*time.Second))
}

func parseTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	var t oauth2.Token
	if err := json.NewDecoder(f).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func saveTokenToFile(token *oauth2.Token, file string) error {
	tok := *token
	// Don't try to persist the long-term credential, even if it exists.
	tok.RefreshToken = ""
	dir := path.Dir(file)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(&tok); err != nil {
		return err
	}
	return nil
}
