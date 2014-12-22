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
	"flag"
	"fmt"
	"net"
	"os/exec"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/golang/glog"
)

var cbrCIDR util.IPNet

func init() {
	flag.Var(&cbrCIDR, "cbr_cidr", "A CIDR notation IP range for the cbr0 bridge.")
}
func Start() {
	ip := cbrCIDR.IP.To4()
	cbrCIDR := net.IPNet{net.IPv4(10, 132, ip[2], ip[3]), cbrCIDR.Mask}
	cmd := exec.Command("sudo ip link set dev cbr0 down")
	glog.Infof("Running '%v'", cmd)
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}

	cmd = exec.Command("sudo brctl delbr cbr0")
	glog.Infof("Running '%v'", cmd)
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}

	cmd = exec.Command("sudo brctl addbr cbr0")
	glog.Infof("Running '%v'", cmd)
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}

	cmd = exec.Command(fmt.Sprintf("sudo ip addr add %s dev cbr0"), cbrCIDR.String())
	glog.Infof("Running '%v'", cmd)
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}

	cmd = exec.Command("sudo ip link set dev cbr0 down")
	glog.Infof("Running '%v'", cmd)
	if err := cmd.Run(); err != nil {
		glog.Errorf("err: %v", err)
	}
}
