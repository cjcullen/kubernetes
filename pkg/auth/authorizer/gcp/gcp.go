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

	"github.com/golang/glog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"

	kerrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/auth/authorizer"
	clientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util"
)

type gcpAuthz struct {
	kubeClient  clientset.Interface
	oauthClient *http.Client
	authzURL    string
	cache       *util.ExpireCache
}

func New(kc clientset.Interface, authzURL string) *gcpAuthz {
	oc := oauth2.NewClient(oauth2.NoContext, google.ComputeTokenSource(""))
	return &gcpAuthz{
		kubeClient:  kc,
		oauthClient: oc,
		authzURL:    authzURL,
		cache:       util.NewExpireCache(1024),
	}
}

type gcpAuthAttributes struct {
	User              string `json:"user,omitempty"`
	Verb              string `json:"verb,omitempty"`
	Namespace         string `json:"namespace,omitempty"`
	NamespaceID       string `json:"namespaceId,omitempty"`
	Resource          string `json:"resource,omitempty"`
	ResourceName      string `json:"resourceName,omitempty"`
	ResourceID        string `json:"resourceId,omitempty"`
	IsResourceRequest bool   `json:"isResourceRequest,omitempty"`
}

func (g *gcpAuthz) getNamespaceUID(nsName string) string {
	if len(nsName) == 0 {
		return ""
	}
	ns, err := g.kubeClient.Core().Namespaces().Get(nsName)
	if err != nil {
		if kerrors.IsNotFound(err) {
			glog.V(4).Infof("Namespace %q not found.", nsName)
			return ""
		}
		glog.Errorf("Error looking up namespace %q: %v", nsName, err)
		return ""
	}
	return string(ns.UID)
}

func (g *gcpAuthz) getResourceUID(nsName, resource, resName string) string {
	if len(resName) == 0 {
		return ""
	}
	var res runtime.Object
	var err error
	switch resource {
	case "namespaces":
		res, err = g.kubeClient.Core().Namespaces().Get(resName)
	case "replicationcontrollers":
		res, err = g.kubeClient.Core().ReplicationControllers(nsName).Get(resName)
	case "nodes":
		res, err = g.kubeClient.Core().Nodes().Get(resName)
	case "events":
		res, err = g.kubeClient.Core().Events(nsName).Get(resName)
	case "endpoints":
		res, err = g.kubeClient.Core().Endpoints(nsName).Get(resName)
	case "pods":
		res, err = g.kubeClient.Core().Pods(nsName).Get(resName)
	case "podtemplates":
		res, err = g.kubeClient.Core().PodTemplates(nsName).Get(resName)
	case "services":
		res, err = g.kubeClient.Core().Services(nsName).Get(resName)
	case "limitranges":
		res, err = g.kubeClient.Core().LimitRanges(nsName).Get(resName)
	case "resourcequotas":
		res, err = g.kubeClient.Core().ResourceQuotas(nsName).Get(resName)
	case "serviceaccounts":
		res, err = g.kubeClient.Core().ServiceAccounts(nsName).Get(resName)
	case "secrets":
		res, err = g.kubeClient.Core().Secrets(nsName).Get(resName)
	case "persistentvolumes":
		res, err = g.kubeClient.Core().PersistentVolumes().Get(resName)
	case "persistentvolumeclaims":
		res, err = g.kubeClient.Core().PersistentVolumeClaims(nsName).Get(resName)
	case "configmaps":
		res, err = g.kubeClient.Core().ConfigMaps(nsName).Get(resName)
	default:
		glog.Infof("Don't know how to look up %q", resource)
		return ""
	}
	if err != nil {
		if kerrors.IsNotFound(err) {
			glog.V(4).Infof("%s %q in namespace %q not found.", resource, resName, nsName)
			return ""
		}
		glog.Errorf("Error looking up %s %q in namespace %q: %v", resource, resName, nsName, err)
		return ""
	}
	uid, err := meta.NewAccessor().UID(res)
	if err != nil {
		glog.Errorf("Error looking up resource %q UID: %v", resName, err)
		return ""
	}
	return string(uid)
}

type authzResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	ExpireTime time.Time
}

func (a *authzResponse) Expiration() time.Time {
	return a.ExpireTime
}

// Authorizer implements authorizer.Authorize
func (g *gcpAuthz) Authorize(a authorizer.Attributes) error {
	nsName := a.GetNamespace()
	nsUID := g.getNamespaceUID(nsName)
	rName := a.GetResourceName()
	rType := a.GetResource()
	rUID := g.getResourceUID(nsName, rType, rName)
	gaa := gcpAuthAttributes{
		User:              a.GetUserName(),
		Verb:              a.GetVerb(),
		Namespace:         a.GetNamespace(),
		NamespaceID:       nsUID,
		Resource:          a.GetResource(),
		ResourceName:      a.GetResourceName(),
		ResourceID:        rUID,
		IsResourceRequest: a.IsResourceRequest(),
	}
	if e, ok := g.cache.Get(gaa); ok && e.Expiration().After(time.Now()) {
		if !e.(*authzResponse).Success {
			return errors.New(e.(*authzResponse).Message)
		}
		return nil
	}
	body, err := json.Marshal(gaa)
	if err != nil {
		return errors.New(fmt.Sprintf("Error marshaling GCP authorization attributes: %#v, %v", gaa, err))
	}
	req, err := http.NewRequest("POST", g.authzURL, bytes.NewReader(body))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create GCP authorization request: %v", err))
	}
	res, err := g.oauthClient.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("GCP Authorization request failed: %v", err))
	}
	defer res.Body.Close()
	if err := googleapi.CheckResponse(res); err != nil {
		return errors.New(fmt.Sprintf("GCP Authorization request failed: %v", err))
	}
	var resp authzResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return errors.New(fmt.Sprintf("Error decoding authorization response: %v", err))
	}
	if resp.Success {
		resp.ExpireTime = time.Now().Add(10 * time.Minute)
	} else {
		resp.ExpireTime = time.Now().Add(1 * time.Minute)
	}
	g.cache.Add(gaa, &resp)
	if resp.Success {
		return nil
	}
	return errors.New(resp.Message)
}
