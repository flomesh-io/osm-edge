package registry

import (
	"github.com/openservicemesh/osm/pkg/models"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy"
)

// ListConnectedProxies lists the Envoy proxies already connected and the time they first connected.
func (pr *ProxyRegistry) ListConnectedProxies() map[string]models.Proxy {
	proxies := make(map[string]models.Proxy)
	pr.connectedProxies.Range(func(keyIface, propsIface interface{}) bool {
		uuid := keyIface.(string)
		proxies[uuid] = propsIface.(*envoy.Proxy)
		return true // continue the iteration
	})
	return proxies
}
