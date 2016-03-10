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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"

	"k8s.io/kubernetes/pkg/auth/user"
)

type gcpAuthn struct {
	oauthClient *http.Client
	authnURL    string
}

// New creates a new GCP Authentication plugin from the specified URL.
func New(authnURL string) *gcpAuthn {
	oc := oauth2.NewClient(oauth2.NoContext, google.ComputeTokenSource(""))
	return &gcpAuthn{
		oauthClient: oc,
		authnURL:    authnURL,
	}
}

// AuthenticateToken implements authenticator.Token
func (g *gcpAuthn) AuthenticateToken(token string) (user.Info, bool, error) {
	tok := struct {
		AccessToken string `json:"accessToken"`
	}{
		AccessToken: token,
	}
	body, err := json.Marshal(tok)
	if err != nil {
		return nil, false, errors.New(fmt.Sprintf("Error marshaling GCP authentication request: %#v, %v", tok, err))
	}
	req, err := http.NewRequest("POST", g.authnURL, bytes.NewReader(body))
	if err != nil {
		return nil, false, errors.New(fmt.Sprintf("Failed to create GCP authentication request: %v", err))
	}
	res, err := g.oauthClient.Do(req)
	if err != nil {
		return nil, false, errors.New(fmt.Sprintf("GCP Authentication request failed: %v", err))
	}
	defer res.Body.Close()
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, false, errors.New(fmt.Sprintf("GCP Authentication request failed: %v", err))
	}
	var resp struct {
		Email      string    `json:"email"`
		ExpireTime time.Time `json:"expireTime"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, false, errors.New(fmt.Sprintf("Error decoding authentication response: %v", err))
	}
	return &user.DefaultInfo{Name: resp.Email, UID: resp.Email}, true, nil
}
