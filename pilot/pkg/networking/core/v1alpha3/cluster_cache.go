// Copyright Istio Authors. All Rights Reserved.
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

package v1alpha3

import (
	"strconv"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/util"
	"istio.io/istio/pkg/config/schema/kind"
	"istio.io/istio/pkg/util/hash"
)

var (
	Separator = []byte{'~'}
	Slash     = []byte{'/'}
)

// clusterCache includes the variables that can influence a Cluster Configuration.
// Implements XdsCacheEntry interface.
type clusterCache struct {
	clusterName string

	// proxy related cache fields
	proxyVersion   string         // will be matched by envoyfilter patches
	locality       *core.Locality // identifies the locality the cluster is generated for
	proxyClusterID string         // identifies the kubernetes cluster a proxy is in
	proxySidecar   bool           // identifies if this proxy is a Sidecar
	proxyView      model.ProxyView
	metadataCerts  *metadataCerts // metadata certificates of proxy

	// service attributes
	http2          bool // http2 identifies if the cluster is for an http2 service
	downstreamAuto bool
	supportsIPv4   bool

	// Dependent configs
	service         *model.Service
	destinationRule *model.ConsolidatedDestRule
	envoyFilterKeys []string
	peerAuthVersion string   // identifies the versions of all peer authentications
	serviceAccounts []string // contains all the service accounts associated with the service
}

func (t *clusterCache) Key() string {
	// nolint: gosec
	// Not security sensitive code
	h := hash.New()
	h.Write([]byte(t.clusterName))
	h.Write(Separator)
	h.Write([]byte(t.proxyVersion))
	h.Write(Separator)
	h.Write([]byte(util.LocalityToString(t.locality)))
	h.Write(Separator)
	h.Write([]byte(t.proxyClusterID))
	h.Write(Separator)
	h.Write([]byte(strconv.FormatBool(t.proxySidecar)))
	h.Write(Separator)
	h.Write([]byte(strconv.FormatBool(t.http2)))
	h.Write(Separator)
	h.Write([]byte(strconv.FormatBool(t.downstreamAuto)))
	h.Write(Separator)
	h.Write([]byte(strconv.FormatBool(t.supportsIPv4)))
	h.Write(Separator)

	if t.proxyView != nil {
		h.Write([]byte(t.proxyView.String()))
	}
	h.Write(Separator)

	if t.metadataCerts != nil {
		h.Write([]byte(t.metadataCerts.String()))
	}
	h.Write(Separator)

	if t.service != nil {
		h.Write([]byte(t.service.Hostname))
		h.Write(Slash)
		h.Write([]byte(t.service.Attributes.Namespace))
	}
	h.Write(Separator)

	for _, dr := range t.destinationRule.GetFrom() {
		h.Write([]byte(dr.Name))
		h.Write(Slash)
		h.Write([]byte(dr.Namespace))
	}
	h.Write(Separator)

	for _, efk := range t.envoyFilterKeys {
		h.Write([]byte(efk))
		h.Write(Separator)
	}
	h.Write(Separator)

	h.Write([]byte(t.peerAuthVersion))
	h.Write(Separator)

	for _, sa := range t.serviceAccounts {
		h.Write([]byte(sa))
		h.Write(Separator)
	}
	h.Write(Separator)

	return h.Sum()
}

func (t clusterCache) DependentConfigs() []model.ConfigHash {
	drs := t.destinationRule.GetFrom()
	configs := make([]model.ConfigHash, 0, len(drs)+1+len(t.envoyFilterKeys))
	if t.destinationRule != nil {
		for _, dr := range drs {
			configs = append(configs, model.ConfigKey{Kind: kind.DestinationRule, Name: dr.Name, Namespace: dr.Namespace}.HashCode())
		}
	}
	if t.service != nil {
		configs = append(configs, model.ConfigKey{Kind: kind.ServiceEntry, Name: string(t.service.Hostname), Namespace: t.service.Attributes.Namespace}.HashCode())
	}
	for _, efKey := range t.envoyFilterKeys {
		items := strings.Split(efKey, "/")
		configs = append(configs, model.ConfigKey{Kind: kind.EnvoyFilter, Name: items[1], Namespace: items[0]}.HashCode())
	}
	return configs
}

func (t clusterCache) Cacheable() bool {
	return true
}

// cacheStats keeps track of cache usage stats.
type cacheStats struct {
	hits, miss int
}

func (c cacheStats) empty() bool {
	return c.hits == 0 && c.miss == 0
}

func (c cacheStats) merge(other cacheStats) cacheStats {
	return cacheStats{
		hits: c.hits + other.hits,
		miss: c.miss + other.miss,
	}
}

func buildClusterKey(service *model.Service, port *model.Port, cb *ClusterBuilder, proxy *model.Proxy, efKeys []string) *clusterCache {
	clusterName := model.BuildSubsetKey(model.TrafficDirectionOutbound, "", service.Hostname, port.Port)
	clusterKey := &clusterCache{
		clusterName:     clusterName,
		proxyVersion:    cb.proxyVersion,
		locality:        cb.locality,
		proxyClusterID:  cb.clusterID,
		proxySidecar:    cb.sidecarProxy(),
		proxyView:       cb.proxyView,
		http2:           port.Protocol.IsHTTP2(),
		downstreamAuto:  cb.sidecarProxy() && util.IsProtocolSniffingEnabledForOutboundPort(port),
		supportsIPv4:    cb.supportsIPv4,
		service:         service,
		destinationRule: proxy.SidecarScope.DestinationRule(model.TrafficDirectionOutbound, proxy, service.Hostname),
		envoyFilterKeys: efKeys,
		metadataCerts:   cb.metadataCerts,
		peerAuthVersion: cb.req.Push.AuthnPolicies.GetVersion(),
		serviceAccounts: cb.req.Push.ServiceAccounts(service.Hostname, service.Attributes.Namespace, port.Port),
	}
	return clusterKey
}
