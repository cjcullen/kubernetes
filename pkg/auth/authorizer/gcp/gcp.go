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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"

	"k8s.io/kubernetes/pkg/auth/authorizer"
)

type gcpAuthz struct {
	oauthClient *http.Client
}

func New() *gcpAuthz {
	client := oauth2.NewClient(oauth2.NoContext, google.ComputeTokenSource(""))
	return &gcpAuthz{client}
}

// Authorizer implements authorizer.Authorize
func (g *gcpAuthz) Authorize(a authorizer.Attributes) error {
	glog.Infof("gcpauthz Authorizing user: %q, verb: %q, resource: %q", a.GetUserName(), a.GetVerb(), a.GetResource())
	url := "https://test-container.sandbox.googleapis.com/v1/masterProjects/518602579940/zones/us-central1-f/authorize"
	body := fmt.Sprintf(
		`{"projectNumber":486425062668,"clusterId":"gcloud","user":%q,"verb":%q,"namespace":%q,"resource":%q}`,
		a.GetUserName(), a.GetVerb(), a.GetNamespace(), a.GetResource())
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return err
	}
	res, err := g.oauthClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if err := googleapi.CheckResponse(res); err != nil {
		return err
	}
	var resp struct {
		Success bool `json:"success"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return err
	}

	if resp.Success {
		glog.Infof("Approving attributes: %#v", a)
		return nil
	}
	return errors.New("No policy matched.")
}
