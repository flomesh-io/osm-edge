package registry

import (
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
)

// NewProxyRegistry initializes a new empty *ProxyRegistry.
func NewProxyRegistry(mapper ProxyServiceMapper, msgBroker *messaging.Broker) *ProxyRegistry {
	return &ProxyRegistry{
		ProxyServiceMapper: mapper,
		msgBroker:          msgBroker,
	}
}

// RegisterProxy registers a newly connected proxy.
func (pr *ProxyRegistry) RegisterProxy(proxy *pipy.Proxy) {
	pr.connectedProxies.Store(proxy.UUID.String(), proxy)
	log.Debug().Str("proxy", proxy.String()).Msg("Registered new proxy")
}

// GetConnectedProxy loads a connected proxy from the registry.
func (pr *ProxyRegistry) GetConnectedProxy(uuid string) *pipy.Proxy {
	p, ok := pr.connectedProxies.Load(uuid)
	if !ok {
		return nil
	}
	return p.(*pipy.Proxy)
}

// RangeConnectedProxy calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (pr *ProxyRegistry) RangeConnectedProxy(f func(key, value interface{}) bool) {
	pr.connectedProxies.Range(f)
}

// UnregisterProxy unregisters the given proxy from the catalog.
func (pr *ProxyRegistry) UnregisterProxy(p *pipy.Proxy) {
	pr.connectedProxies.Delete(p.UUID.String())
	log.Debug().Msgf("Unregistered proxy %s", p.String())
}

// GetConnectedProxyCount counts the number of connected proxies
// TODO(steeling): switch to a regular map with mutex so we can get the count without iterating the entire list.
func (pr *ProxyRegistry) GetConnectedProxyCount() int {
	return len(pr.ListConnectedProxies())
}
