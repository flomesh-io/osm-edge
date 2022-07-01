package registry

import (
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/sidecar"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
)

// ListConnectedProxies lists the Sidecar proxies already connected and the time they first connected.
func (pr *ProxyRegistry) ListConnectedProxies() map[certificate.CommonName]sidecar.Proxy {
	proxies := make(map[certificate.CommonName]sidecar.Proxy)
	pr.PodCNtoProxy.Range(func(cnIface, propsIface interface{}) bool {
		cn := cnIface.(certificate.CommonName)
		proxy := propsIface.(*pipy.Proxy)
		proxies[cn] = proxy
		return true // continue the iteration
	})
	return proxies
}
