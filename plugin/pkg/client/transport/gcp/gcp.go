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
	"net/http"

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
		Source: ts,
		Base:   rt,
	}, nil
}
