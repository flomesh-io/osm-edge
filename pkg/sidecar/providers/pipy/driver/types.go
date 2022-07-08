package driver

import (
	"github.com/openservicemesh/osm/pkg/logger"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
)

var log = logger.New("sidecar-pipy")

// Context needed to compose the Sidecar bootstrap YAML.
type sidecarBootstrapConfigMeta struct {
	SidecarAdminPort uint32
	OSMClusterName   string
	NodeID           string
	RootCert         []byte
	Cert             []byte
	Key              []byte

	// Host and port of the Sidecar xDS server
	ProxyServerHost string
	ProxyServerPort uint32

	// The bootstrap Sidecar config will be affected by the liveness, readiness, startup probes set on
	// the pod this Sidecar is fronting.
	OriginalHealthProbes driver.HealthProbes

	// Sidecar TLS configuration
	TLSMinProtocolVersion string
	TLSMaxProtocolVersion string
	CipherSuites          []string
	ECDHCurves            []string
}
