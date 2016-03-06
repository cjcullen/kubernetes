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

package transport

import (
	"net/http"
	"sync"

	"github.com/golang/glog"
)

type RoundTripperPlugin interface {
	RoundTripper(http.RoundTripper) (http.RoundTripper, error)
}

// All registered transport plugins.
var pluginsLock sync.Mutex
var plugins = make(map[string]RoundTripperPlugin)

func RegisterRoundTripperPlugin(name string, plugin RoundTripperPlugin) {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()
	if _, found := plugins[name]; found {
		glog.Fatalf("Round tripper plugin %q was registered twice", name)
	}
	glog.V(1).Infof("Registered round tripper plugin %q", name)
	plugins[name] = plugin
}

func GetRoundTripperPlugin(name string) RoundTripperPlugin {
	pluginsLock.Lock()
	defer pluginsLock.Unlock()
	return plugins[name]
}
