package driver

import (
	"github.com/openservicemesh/osm/pkg/logger"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
)

const (
	sidecarBootstrapConfigVolume = "sidecar-bootstrap-config-volume"
)

var log = logger.New("sidecar-injector")

// Config is the type used to represent the config options for the sidecar injection
type Config struct {
	// ListenPort defines the port on which the sidecar injector listens
	ListenPort int
}

// Context needed to compose the Sidecar bootstrap YAML.
type sidecarBootstrapConfigMeta struct {
	SidecarAdminPort uint32
	XDSClusterName   string
	NodeID           string
	RootCert         []byte
	Cert             []byte
	Key              []byte

	// Host and port of the Sidecar xDS server
	XDSHost string
	XDSPort uint32

	// The bootstrap Sidecar config will be affected by the liveness, readiness, startup probes set on
	// the pod this Sidecar is fronting.
	OriginalHealthProbes driver.HealthProbes

	// Sidecar TLS configuration
	TLSMinProtocolVersion string
	TLSMaxProtocolVersion string
	CipherSuites          []string
	ECDHCurves            []string
}
