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
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	googleoauth2 "google.golang.org/api/oauth2/v2"

	"k8s.io/kubernetes/pkg/auth/user"
)

type GCPAuthenticator struct{
	tokenService *googleoauth2.Service
}

// New returns a GCPAuthenticator.
func New() (*GCPAuthenticator, error) {
	oauthClient, err := googleoauth2.New(&http.Client{})
	if err != nil {
		return nil, err
	}
	return &GCPAuthenticator{oauthClient}, nil
}

func (a *GCPAuthenticator) AuthenticateToken(value string) (user.Info, bool, error) {
	glog.Infof("trying to validate token w/ GCP: %q", value)
	info, err := a.tokenService.Tokeninfo().AccessToken(value).Do()
	if err != nil {
		return nil, false, err
	}
	glog.Infof("TokenInfo: %#v", info)
	if !strings.Contains(info.Scope, "https://www.googleapis.com/auth/cloud-platform") {
		return nil, false, fmt.Errorf("Token does not contain cloud-platform scope")
	}
	name := info.IssuedTo
	if info.Email != "" {
		name = info.Email
	}
	glog.Infof("Returning Name: %q", name)
	return &user.DefaultInfo{Name: name, UID: name}, true, nil
}
