package registry

import (
	"k8s.io/apimachinery/pkg/types"

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

// RegisterProxy implements MeshCatalog and registers a newly connected proxy.
func (pr *ProxyRegistry) RegisterProxy(proxy *pipy.Proxy) {
	pr.PodCNtoProxy.Store(proxy.GetCertificateCommonName(), proxy)

	// If this proxy object is on a Kubernetes Pod - it will have an UID
	if proxy.HasPodMetadata() {
		podUID := types.UID(proxy.PodMetadata.UID)

		// Create a PodUID to Certificate CN map so we can easily determine the CN from the PodUID
		pr.PodUIDToCN.Store(podUID, proxy.GetCertificateCommonName())

		// Create a PodUID to Cert Serial Number so we can easily look-up the SerialNumber of the cert issued to a proxy for a given Pod.
		pr.PodUIDToCertificateSerialNumber.Store(podUID, proxy.GetCertificateSerialNumber())
	}
	log.Debug().Str("proxy", proxy.String()).Msg("Registered new proxy")
}

// UnregisterProxy unregisters the given proxy from the catalog.
func (pr *ProxyRegistry) UnregisterProxy(p *pipy.Proxy) {
	pr.PodCNtoProxy.Delete(p.GetCertificateCommonName())
	if p.HasPodMetadata() {
		podUID := types.UID(p.PodMetadata.UID)
		pr.PodUIDToCN.Delete(podUID)
		pr.PodUIDToCertificateSerialNumber.Delete(podUID)
	}
	log.Debug().Msgf("Unregistered proxy %s", p.String())
}

// GetConnectedProxyCount counts the number of connected proxies
func (pr *ProxyRegistry) GetConnectedProxyCount() int {
	return len(pr.ListConnectedProxies())
}
