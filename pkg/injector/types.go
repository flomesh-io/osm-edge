// Package injector implements OSM's automatic sidecar injection facility. The sidecar injector's mutating webhook
// admission controller intercepts pod creation requests to mutate the pod spec to inject the sidecar proxy.
package injector

import (
	mapset "github.com/deckarep/golang-set"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/logger"
)

const (
	sidecarBootstrapConfigVolume = "sidecar-bootstrap-config-volume"
)

var log = logger.New("sidecar-injector")

// mutatingWebhook is the type used to represent the webhook for sidecar injection
type mutatingWebhook struct {
	config                 Config
	kubeClient             kubernetes.Interface
	certManager            certificate.Manager
	kubeController         k8s.Controller
	osmNamespace           string
	meshName               string
	cert                   *certificate.Certificate
	configurator           configurator.Configurator
	osmContainerPullPolicy corev1.PullPolicy

	nonInjectNamespaces mapset.Set
}

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
	OriginalHealthProbes healthProbes

	// Sidecar TLS configuration
	TLSMinProtocolVersion string
	TLSMaxProtocolVersion string
	CipherSuites          []string
	ECDHCurves            []string
}
