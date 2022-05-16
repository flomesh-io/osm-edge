package registry

import (
	"github.com/openservicemesh/osm/pkg/certificate"
	"k8s.io/apimachinery/pkg/types"
	"sync"
	"time"

	"github.com/openservicemesh/osm/pkg/logger"
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/pipy"
)

var log = logger.New("proxy-registry")

// ProxyRegistry keeps track of Sidecar proxies as they connect and disconnect
// from the control plane.
type ProxyRegistry struct {
	ProxyServiceMapper

	connectedProxies sync.Map

	// Maintain a mapping of pod UID to CN of the Sidecar on the given pod
	podUIDToCN sync.Map

	// Maintain a mapping of pod UID to certificate SerialNumber of the Sidecar on the given pod
	podUIDToCertificateSerialNumber sync.Map

	releaseCertificateCallback func(podUID types.UID, endpointCN certificate.CommonName)

	msgBroker *messaging.Broker
}

type connectedProxy struct {
	// Proxy which connected to the XDS control plane
	proxy *pipy.Proxy

	// When the proxy connected to the XDS control plane
	connectedAt time.Time
}
