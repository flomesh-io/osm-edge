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

	// SidecarInboundListenerPort is Sidecar's inbound listener port number.
	SidecarInboundListenerPort = 15003

	// SidecarInboundListenerPortName is Sidecar's inbound listener port name.
	SidecarInboundListenerPortName = "proxy-inbound"

	// SidecarInboundPrometheusListenerPortName is Sidecar's inbound listener port name for prometheus.
	SidecarInboundPrometheusListenerPortName = "proxy-metrics"

	// SidecarOutboundListenerPort is Sidecar's outbound listener port number.
	SidecarOutboundListenerPort = 15001

	// SidecarOutboundListenerPortName is Sidecar's outbound listener port name.
	SidecarOutboundListenerPortName = "proxy-outbound"

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

	// ADSServerPort is the port on which the Aggregated Discovery Service (ADS) listens for new gRPC connections from Sidecar proxies
	ADSServerPort = 15128

	// PrometheusScrapePath is the path for prometheus to scrap sidecar metrics from
	PrometheusScrapePath = "/stats/prometheus"

	// CertificationAuthorityCommonName is the CN used for the root certificate for OSM.
	CertificationAuthorityCommonName = "osm-ca.openservicemesh.io"

	// CertificationAuthorityRootValidityPeriod is when the root certificate expires
	CertificationAuthorityRootValidityPeriod = 87600 * time.Hour // a decade

	// XDSCertificateValidityPeriod is the TTL of the certificates used for Sidecar to xDS communication.
	XDSCertificateValidityPeriod = 87600 * time.Hour // a decade

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

	// SidecarUniqueIDLabelName is the label applied to pods with the unique ID of the sidecar.
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

	// SidecarServiceNodeSeparator is the character separating the strings used to create an Sidecars service node parameter.
	// Example use: envoy --service-node 52883c80-6e0d-4c64-b901-cbcb75134949/bookstore/10.144.2.91/bookstore-v1/bookstore-v1
	SidecarServiceNodeSeparator = "/"

	// OSMMeshConfig is the name of the OSM MeshConfig
	OSMMeshConfig = "osm-mesh-config"

	// CRDConversionWebhookPort is the port of the CRD conversion webhook service
	CRDConversionWebhookPort = 9443
)

// Annotations used by the control plane
const (
	// SidecarInjectionAnnotation is the annotation used for sidecar injection
	SidecarInjectionAnnotation = "openservicemesh.io/sidecar-injection"

	// MetricsAnnotation is the annotation used for enabling/disabling metrics
	MetricsAnnotation = "openservicemesh.io/metrics"
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
	// field key string for context
	LogFieldContext = "context"

	// multicluster field value for context field.
	LogContextMulticluster = "multicluster"
)

const (
	SidecarClasssEnvoy = "envoy"
	SidecarClasssPipy  = "pipy"
)
