//go:build !linux
// +build !linux

// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ambient

import (
	"errors"

	"github.com/vishvananda/netlink"
	corev1 "k8s.io/api/core/v1"
)

var ErrNotImplemented = errors.New("not implemented")

func IsPodInIpset(pod *corev1.Pod) bool {
	return false
}

func getLinkWithDestinationOf(ip string) (netlink.Link, error) {
	return nil, ErrNotImplemented
}

func getVethWithDestinationOf(ip string) (*netlink.Veth, error) {
	return nil, ErrNotImplemented
}

func getDeviceWithDestinationOf(ip string) (string, error) {
	return "", ErrNotImplemented
}

func GetIndexAndPeerMac(podIfName, ns string) (int, net.HardwareAddr, error) {
	return 0, nil, ErrNotImplemented
}

func getMacFromNsIdx(ns string, ifIndex int) (net.HardwareAddr, error) {
	return nil, ErrNotImplemented
}

func getNsNameFromNsID(nsid int) (string, error) {
	return "", ErrNotImplemented
}

func getPeerIndex(veth *netlink.Veth) (int, error) {
	return 0, ErrNotImplemented
}

// CreateRulesOnNode initializes the routing, firewall and ipset rules on the node.
func (s *Server) CreateRulesOnNode(ztunnelVeth, ztunnelIP string, captureDNS bool) error {
	return ErrNotImplemented
}

func (s *Server) cleanup() {
}

func routeFlushTable(table int) error {
	return ErrNotImplemented
}

func routesDelete(routes []netlink.Route) error {
	return ErrNotImplemented
}
