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
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/golang/glog"
	googleoauth2 "google.golang.org/api/oauth2/v2"

	"k8s.io/kubernetes/pkg/auth/user"
)

type GCPAuthenticator struct{
	tokenService *googleoauth2.Service
}

// NewCSV returns a TokenAuthenticator, populated from a CSV file.
// The CSV file must contain records in the format "token,username,useruid"
func New() (*GCPAuthenticator, error) {
	return &GCPAuthenticator{googleoauth2.New(client)}
}

func (a *GCPAuthenticator) AuthenticateToken(value string) (user.Info, bool, error) {
	glog.Infof("trying to validate token w/ GCP: %q", value)
	return "asdf", nil
}
