/*
Copyright 2014 Google Inc. All rights reserved.

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

package node

import (
	"net"
	"os/exec"
	"strings"

	"github.com/golang/glog"
)

func EnsureDocker() {
	cmd := exec.Command("ip", "addr", "flush", "dev", "docker0")
	glog.Infof("Running '%v'", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}

	cmd = exec.Command("/etc/init.d/docker", "restart")
	glog.Infof("Running '%v'", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}
}

func EnsureCBR0(cbrCIDR *net.IPNet) error {
	ip := cbrCIDR.IP.Mask(cbrCIDR.Mask).To4()
	// Grab the ip at the start of the range for the gateway (e.g. x.x.x.1 for a /24)
	cbrGatewayCIDR := net.IPNet{net.IPv4(ip[0], ip[1], ip[2], ip[3]+1), cbrCIDR.Mask}

	cmd := exec.Command("brctl", "addbr", "cbr0")
	glog.Infof("Running '%v'", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}

	cmd = exec.Command("ip", "link", "set", "dev", "cbr0", "mtu", "1460")
	glog.Infof("Running '%v'", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}

	cmd = exec.Command("ip", "addr", "add", cbrGatewayCIDR.String(), "dev", "cbr0")
	glog.Infof("Running '%v'", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}

	cmd = exec.Command("ip", "link", "set", "dev", "cbr0", "up")
	glog.Infof("Running '%v'", strings.Join(cmd.Args, " "))
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}
	return nil
}
