// Package configurator implements the Configurator interface that provides APIs to retrieve OSM control plane configurations.
package configurator

import (
	"time"

	corev1 "k8s.io/api/core/v1"

	configv1alpha2 "github.com/openservicemesh/osm/pkg/apis/config/v1alpha2"
	"github.com/openservicemesh/osm/pkg/k8s/informers"

	"github.com/openservicemesh/osm/pkg/auth"
	"github.com/openservicemesh/osm/pkg/logger"
)

var (
	log = logger.New("configurator")
)

// Client is the type used to represent the Kubernetes Client for the config.openservicemesh.io API group
type Client struct {
	osmNamespace   string
	informers      *informers.InformerCollection
	meshConfigName string
}

// Configurator is the controller interface for K8s namespaces
type Configurator interface {
	// GetMeshConfig returns the MeshConfig resource corresponding to the control plane
	GetMeshConfig() configv1alpha2.MeshConfig

	// GetOSMNamespace returns the namespace in which OSM controller pod resides
	GetOSMNamespace() string

	// GetMeshConfigJSON returns the MeshConfig in pretty JSON (human readable)
	GetMeshConfigJSON() (string, error)

	// IsPermissiveTrafficPolicyMode determines whether we are in "allow-all" mode or SMI policy (block by default) mode
	IsPermissiveTrafficPolicyMode() bool

	// IsEgressEnabled determines whether egress is globally enabled in the mesh or not
	IsEgressEnabled() bool

	// IsDebugServerEnabled determines whether osm debug HTTP server is enabled
	IsDebugServerEnabled() bool

	// IsTracingEnabled returns whether tracing is enabled
	IsTracingEnabled() bool

	// GetTracingHost is the host to which we send tracing spans
	GetTracingHost() string

	// GetTracingPort returns the tracing listener port
	GetTracingPort() uint32

	// GetTracingEndpoint returns the collector endpoint
	GetTracingEndpoint() string

	// GetMaxDataPlaneConnections returns the max data plane connections allowed, 0 if disabled
	GetMaxDataPlaneConnections() int

	// GetOsmLogLevel returns the configured OSM log level
	GetOSMLogLevel() string

	// GetSidecarLogLevel returns the sidecar log level
	GetSidecarLogLevel() string

	// GetSidecarClass returns the sidecar class
	GetSidecarClass() string

	// GetSidecarImage returns the sidecar image
	GetSidecarImage() string

	// GetSidecarWindowsImage returns the sidecar image
	GetSidecarWindowsImage() string

	// GetInitContainerImage returns the init container image
	GetInitContainerImage() string

	// GetProxyServerPort returns the port on which the Discovery Service listens for new connections from Sidecars
	GetProxyServerPort() uint32

	// GetSidecarDisabledMTLS returns the status of mTLS
	GetSidecarDisabledMTLS() bool

	// GetServiceCertValidityPeriod returns the validity duration for service certificates
	GetServiceCertValidityPeriod() time.Duration

	// GetIngressGatewayCertValidityPeriod returns the validity duration for the Ingress
	// Gateway certificate, default value if not specified
	GetIngressGatewayCertValidityPeriod() time.Duration

	// GetCertKeyBitSize returns the certificate key bit size
	GetCertKeyBitSize() int

	// IsPrivilegedInitContainer determines whether init containers should be privileged
	IsPrivilegedInitContainer() bool

	// GetConfigResyncInterval returns the duration for resync interval.
	// If error or non-parsable value, returns 0 duration
	GetConfigResyncInterval() time.Duration

	// GetProxyResources returns the `Resources` configured for proxies, if any
	GetProxyResources() corev1.ResourceRequirements

	// GetInboundExternalAuthConfig returns the External Authentication configuration for incoming traffic, if any
	GetInboundExternalAuthConfig() auth.ExtAuthConfig

	// GetFeatureFlags returns OSM's feature flags
	GetFeatureFlags() configv1alpha2.FeatureFlags
}
