package registry

import (
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/sidecar"
)

// ListConnectedProxies lists the Sidecar proxies already connected and the time they first connected.
func (pr *ProxyRegistry) ListConnectedProxies() map[certificate.CommonName]sidecar.Proxy {
	proxies := make(map[certificate.CommonName]sidecar.Proxy)
	pr.PodCNtoProxy.Range(func(cnIface, propsIface interface{}) bool {
		cn := cnIface.(certificate.CommonName)
		proxies[cn] = *propsIface.(*sidecar.Proxy)
		return true // continue the iteration
	})
	return proxies
}
