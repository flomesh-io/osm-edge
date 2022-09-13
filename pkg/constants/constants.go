// Package constants defines the constants that are used by multiple other packages within OSM.
package constants

import "time"

const (
	// WildcardIPAddr is a string constant.
	WildcardIPAddr = "0.0.0.0"

	// SidecarAdminPort is Sidecar's admin port
	SidecarAdminPort = 15000

	// SidecarAdminPortName is Sidecar's admin port name
	SidecarAdminPortName = "proxy-admin"

	// SidecarTCPInboundListenerPort is Sidecar's tcp inbound listener port number.
	SidecarTCPInboundListenerPort = 15003

	// SidecarUDPInboundListenerPort is Sidecar's udp inbound listener port number.
	SidecarUDPInboundListenerPort = 15004

	// SidecarTCPInboundListenerPortName is Sidecar's tcp inbound listener port name.
	SidecarTCPInboundListenerPortName = "pxy-tcp-inbound"

	// SidecarUDPInboundListenerPortName is Sidecar's udp inbound listener port name.
	SidecarUDPInboundListenerPortName = "pxy-udp-inbound"

	// SidecarInboundPrometheusListenerPortName is Sidecar's inbound listener port name for prometheus.
	SidecarInboundPrometheusListenerPortName = "proxy-metrics"

	// SidecarTCPOutboundListenerPort is Sidecar's tcp outbound listener port number.
	SidecarTCPOutboundListenerPort = 15001

	// SidecarUDPOutboundListenerPort is Sidecar's udp outbound listener port number.
	SidecarUDPOutboundListenerPort = 15002

	// SidecarTCPOutboundListenerPortName is Sidecar's tcp outbound listener port name.
	SidecarTCPOutboundListenerPortName = "proxy-tcp-outbound"

	// SidecarUDPOutboundListenerPortName is Sidecar's udp outbound listener port name.
	SidecarUDPOutboundListenerPortName = "proxy-udp-outbound"

	// SidecarUID is the Sidecar's User ID
	SidecarUID int64 = 1500

	// SidecarWindowsUser is the Sidecar's User name on Windows.
	SidecarWindowsUser string = "SidecarUser"

	// LocalhostIPAddress is the local host address.
	LocalhostIPAddress = "127.0.0.1"

	// SidecarMetricsCluster is the cluster name of the Prometheus metrics cluster
	SidecarMetricsCluster = "sidecar-metrics-cluster"

	// SidecarTracingCluster is the default name to refer to the tracing cluster.
	SidecarTracingCluster = "sidecar-tracing-cluster"

	// DefaultTracingEndpoint is the default endpoint route.
	DefaultTracingEndpoint = "/api/v2/spans"

	// DefaultTracingHost is the default tracing server name.
	DefaultTracingHost = "jaeger"

	// DefaultTracingPort is the tracing listener port.
	DefaultTracingPort = uint32(9411)

	// DefaultSidecarLogLevel is the default sidecar log level if not defined in the osm MeshConfig
	DefaultSidecarLogLevel = "error"

	// DefaultOSMLogLevel is the default OSM log level if none is specified
	DefaultOSMLogLevel = "info"

	// SidecarPrometheusInboundListenerPort is Sidecar's inbound listener port number for prometheus
	SidecarPrometheusInboundListenerPort = 15010

	// InjectorWebhookPort is the port on which the sidecar injection webhook listens
	InjectorWebhookPort = 9090

	// OSMHTTPServerPort is the port on which osm-controller and osm-injector serve HTTP requests for metrics, health probes etc.
	OSMHTTPServerPort = 9091

	// DebugPort is the port on which OSM exposes its debug server
	DebugPort = 9092

	// ValidatorWebhookPort is the port on which the resource validator webhook listens
	ValidatorWebhookPort = 9093

	// OSMControllerName is the name of the OSM Controller (formerly ADS service).
	OSMControllerName = "osm-controller"

	// OSMInjectorName is the name of the OSM Injector.
	OSMInjectorName = "osm-injector"

	// OSMBootstrapName is the name of the OSM Bootstrap.
	OSMBootstrapName = "osm-bootstrap"

	// ProxyServerPort is the port on which the Aggregated Discovery Service (ADS) listens for new gRPC connections from Envoy proxies
	ProxyServerPort = 15128

	// PrometheusScrapePath is the path for prometheus to scrap sidecar metrics from
	PrometheusScrapePath = "/stats/prometheus"

	// CertificationAuthorityCommonName is the CN used for the root certificate for OSM.
	CertificationAuthorityCommonName = "osm-ca.openservicemesh.io"

	// CertificationAuthorityRootValidityPeriod is when the root certificate expires
	CertificationAuthorityRootValidityPeriod = 87600 * time.Hour // a decade

	// OSMCertificateValidityPeriod is the TTL of the certificates used in the OSM control plane or for Envoy to xDS communication.
	OSMCertificateValidityPeriod = 87600 * time.Hour // a decade

	// DefaultCABundleSecretName is the default name of the secret for the OSM CA bundle
	DefaultCABundleSecretName = "osm-ca-bundle" // #nosec G101: Potential hardcoded credentials

	// RegexMatchAll is a regex pattern match for all
	RegexMatchAll = ".*"

	// WildcardHTTPMethod is a wildcard for all HTTP methods
	WildcardHTTPMethod = "*"

	// OSMKubeResourceMonitorAnnotation is the key of the annotation used to monitor a K8s resource
	OSMKubeResourceMonitorAnnotation = "openservicemesh.io/monitored-by"

	// KubernetesOpaqueSecretCAKey is the key which holds the CA bundle in a Kubernetes secret.
	KubernetesOpaqueSecretCAKey = "ca.crt"

	// KubernetesOpaqueSecretRootPrivateKeyKey is the key which holds the CA's private key in a Kubernetes secret.
	KubernetesOpaqueSecretRootPrivateKeyKey = "private.key"

	// SidecarUniqueIDLabelName is the label applied to pods with the unique ID of the Envoy sidecar.
	SidecarUniqueIDLabelName = "osm-proxy-uuid"

	// ----- Environment Variables

	// EnvVarLogKubernetesEvents is the name of the env var instructing the event handlers whether to log at all (true/false)
	EnvVarLogKubernetesEvents = "OSM_LOG_KUBERNETES_EVENTS"

	// EnvVarHumanReadableLogMessages is an environment variable, which when set to "true" enables colorful human-readable log messages.
	EnvVarHumanReadableLogMessages = "OSM_HUMAN_DEBUG_LOG"

	// ClusterWeightAcceptAll is the weight for a cluster that accepts 100 percent of traffic sent to it
	ClusterWeightAcceptAll = 100

	// PrometheusDefaultRetentionTime is the default days for which data is retained in prometheus
	PrometheusDefaultRetentionTime = "15d"

	// DomainDelimiter is a delimiter used in representing domains
	DomainDelimiter = "."

	// SidecarContainerName is the name used to identify the sidecar container added on mesh-enabled deployments
	SidecarContainerName = "sidecar"

	// InitContainerName is the name of the init container
	InitContainerName = "osm-init"

	// SidecarServiceNodeSeparator is the character separating the strings used to create an Sidecar service node parameter.
	// Example use: envoy --service-node 52883c80-6e0d-4c64-b901-cbcb75134949/bookstore/10.144.2.91/bookstore-v1/bookstore-v1
	SidecarServiceNodeSeparator = "/"

	// OSMMeshConfig is the name of the OSM MeshConfig
	OSMMeshConfig = "osm-mesh-config"
)

// HealthProbe constants
const (
	// LivenessProbePort is the port to use for liveness probe
	LivenessProbePort = int32(15901)

	// ReadinessProbePort is the port to use for readiness probe
	ReadinessProbePort = int32(15902)

	// StartupProbePort is the port to use for startup probe
	StartupProbePort = int32(15903)

	// HealthcheckPort is the port to use for healthcheck probe
	HealthcheckPort = int32(15904)

	// LivenessProbePath is the path to use for liveness probe
	LivenessProbePath = "/osm-liveness-probe"

	// ReadinessProbePath is the path to use for readiness probe
	ReadinessProbePath = "/osm-readiness-probe"

	// StartupProbePath is the path to use for startup probe
	StartupProbePath = "/osm-startup-probe"

	// HealthcheckPath is the path to use for healthcheck probe
	HealthcheckPath = "/osm-healthcheck"
)

// Annotations used by the control plane
const (
	// SidecarInjectionAnnotation is the annotation used for sidecar injection
	SidecarInjectionAnnotation = "openservicemesh.io/sidecar-injection"

	// MetricsAnnotation is the annotation used for enabling/disabling metrics
	MetricsAnnotation = "openservicemesh.io/metrics"
)

// Annotations and labels used by the MeshRootCertificate
const (
	// MRCStateValidatingRollout is the validating rollout status option for the State of the MeshRootCertificate
	MRCStateValidatingRollout = "validatingRollout"

	// MRCStateIssuingRollout is the issuing rollout status option for the State of the MeshRootCertificate
	MRCStateIssuingRollout = "issuingRollout"

	// MRCStateActive is the active status option for the State of the MeshRootCertificate
	MRCStateActive = "active"

	// MRCStateIssuingRollback is the issuing rollback status option for the State of the MeshRootCertificate
	MRCStateIssuingRollback = "issuingRollback"

	// MRCStateValidatingRollback is the validating rollback status option for the State of the MeshRootCertificate
	MRCStateValidatingRollback = "validatingRollback"

	// MRCStateInactive is the inactive status option for the State of the MeshRootCertificate
	MRCStateInactive = "inactive"

	// MRCStateError is the error status option for the State of the MeshRootCertificate
	MRCStateError = "error"
)

// Labels used by the control plane
const (
	// IgnoreLabel is the label used to ignore a resource
	IgnoreLabel = "openservicemesh.io/ignore"

	// ReconcileLabel is the label used to reconcile a resource
	ReconcileLabel = "openservicemesh.io/reconcile"

	// AppLabel is the label used to identify the app
	AppLabel = "app"
)

// Annotations used for Metrics
const (
	// PrometheusScrapeAnnotation is the annotation used to configure prometheus scraping
	PrometheusScrapeAnnotation = "prometheus.io/scrape"

	// PrometheusPortAnnotation is the annotation used to configure the port to scrape on
	PrometheusPortAnnotation = "prometheus.io/port"

	// PrometheusPathAnnotation is the annotation used to configure the path to scrape on
	PrometheusPathAnnotation = "prometheus.io/path"
)

// App labels as defined in the "osm.labels" template in _helpers.tpl of the Helm chart.
const (
	OSMAppNameLabelKey     = "app.kubernetes.io/name"
	OSMAppNameLabelValue   = "openservicemesh.io"
	OSMAppInstanceLabelKey = "app.kubernetes.io/instance"
	OSMAppVersionLabelKey  = "app.kubernetes.io/version"
)

// Application protocols
const (
	// HTTP protocol
	ProtocolHTTP = "http"

	// HTTPS protocol
	ProtocolHTTPS = "https"

	// TCP protocol
	ProtocolTCP = "tcp"

	// gRPC protocol
	ProtocolGRPC = "grpc"

	// ProtocolTCPServerFirst implies TCP based server first protocols
	// Ex. MySQL, SMTP, PostgreSQL etc. where the server initiates the first
	// byte in a TCP connection.
	ProtocolTCPServerFirst = "tcp-server-first"

	// UDP protocol
	ProtocolUDP = "udp"
)

// Operating systems.
const (
	// OSWindows is the name for Windows operating system.
	OSWindows string = "windows"

	// OSLinux is the name for Linux operating system.
	OSLinux string = "linux"
)

// Logging contexts
const (
	// LogFieldContext is the key used to specify the logging context
	LogFieldContext = "context"
)

// Control plane HTTP server paths
const (
	// OSMControllerReadinessPath is the path at which OSM controller serves readiness probes
	OSMControllerReadinessPath = "/health/ready"

	// OSMControllerLivenessPath is the path at which OSM controller serves liveness probes
	OSMControllerLivenessPath = "/health/alive"

	// OSMControllerSMIVersionPath is the path at which OSM controller servers SMI version info
	OSMControllerSMIVersionPath = "/smi/version"

	// MetricsPath is the path at which OSM controller serves metrics
	MetricsPath = "/metrics"

	// VersionPath is the path at which OSM controller serves version info
	VersionPath = "/version"

	// WebhookHealthPath is the path at which the webooks serve health probes
	WebhookHealthPath = "/healthz"
)

// OSM HTTP Server Responses
const (
	// ServiceReadyResponse is the response returned by the server to indicate it is ready
	ServiceReadyResponse = "Service is ready"

	// ServiceAliveResponse is the response returned by the server to indicate it is alive
	ServiceAliveResponse = "Service is alive"
)

var (
	// SupportedProtocolsInMesh is a list of the protocols OSM supports for in-mesh traffic
	SupportedProtocolsInMesh = []string{ProtocolTCPServerFirst, ProtocolHTTP, ProtocolTCP, ProtocolGRPC, ProtocolUDP}
)

const (
	// SidecarClassEnvoy is the SidecarClass field value for context field.
	SidecarClassEnvoy = "envoy"

	// SidecarClassPipy is the SidecarClass field value for context field.
	SidecarClassPipy = "pipy"
)
