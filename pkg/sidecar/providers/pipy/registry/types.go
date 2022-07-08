package registry

import (
	"sync"

	"github.com/openservicemesh/osm/pkg/logger"
	"github.com/openservicemesh/osm/pkg/messaging"
)

var log = logger.New("proxy-registry")

// ProxyRegistry keeps track of Sidecar proxies as they connect and disconnect
// from the control plane.
type ProxyRegistry struct {
	ProxyServiceMapper

	// Maintain a mapping of pod CN to Proxy of the Sidecar on the given pod
	PodCNtoProxy sync.Map

	// Maintain a mapping of pod CN to Pipy Repo Codebase ETag
	PodCNtoETag sync.Map

	// Maintain a mapping of pod UID to CN of the Sidecar on the given pod
	PodUIDToCN sync.Map

	// Maintain a mapping of pod UID to certificate SerialNumber of the Sidecar on the given pod
	PodUIDToCertificateSerialNumber sync.Map

	// Fire a inform to update proxies
	UpdateProxies func()

	msgBroker *messaging.Broker
}
