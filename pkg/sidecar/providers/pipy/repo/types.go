package repo

import (
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/apis/policy/v1alpha1"
	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/logger"
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/service"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/client"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
	"github.com/openservicemesh/osm/pkg/workerpool"
)

var (
	log = logger.New("flomesh-pipy")
)

// Server implements the Aggregate Discovery Services
type Server struct {
	catalog        catalog.MeshCataloger
	proxyRegistry  *registry.ProxyRegistry
	osmNamespace   string
	cfg            configurator.Configurator
	certManager    *certificate.Manager
	ready          bool
	workQueues     *workerpool.WorkerPool
	kubeController k8s.Controller

	// When snapshot cache is enabled, we (currently) don't keep track of proxy information, however different
	// config versions have to be provided to the cache as we keep adding snapshots. The following map
	// tracks at which version we are at given a proxy UUID
	configVerMutex sync.Mutex
	configVersion  map[string]uint64

	msgBroker *messaging.Broker

	repoClient *client.PipyRepoClient

	retryJob func()
}

// Protocol is a string wrapper type
type Protocol string

// Address is a string wrapper type
type Address string

// Port is a uint16 wrapper type
type Port uint16

// Weight is a uint32 wrapper type
type Weight uint32

// ClusterName is a string wrapper type
type ClusterName string

// WeightedEndpoint is a wrapper type of map[HTTPHostPort]Weight
type WeightedEndpoint map[HTTPHostPort]Weight

// Header is a string wrapper type
type Header string

// HeaderRegexp is a string wrapper type
type HeaderRegexp string

// Headers is a wrapper type of map[Header]HeaderRegexp
type Headers map[Header]HeaderRegexp

// Method is a string wrapper type
type Method string

// Methods is a wrapper type of []Method
type Methods []Method

// WeightedClusters is a wrapper type of map[ClusterName]Weight
type WeightedClusters map[ClusterName]Weight

// URIPathRegexp is a string wrapper type
type URIPathRegexp string

// ServiceName is a string wrapper type
type ServiceName string

// Services is a wrapper type of []ServiceName
type Services []ServiceName

// HTTPRouteRule http route rule
type HTTPRouteRule struct {
	Headers         Headers          `json:"Headers"`
	Methods         Methods          `json:"Methods"`
	TargetClusters  WeightedClusters `json:"TargetClusters"`
	AllowedServices Services         `json:"AllowedServices"`

	allowedAnyService bool
	allowedAnyMethod  bool
}

// HTTPRouteRuleName is a string wrapper type
type HTTPRouteRuleName string

// HTTPHostPort is a string wrapper type
type HTTPHostPort string

// HTTPHostPort2Service is a wrapper type of map[HTTPHostPort]HTTPRouteRuleName
type HTTPHostPort2Service map[HTTPHostPort]HTTPRouteRuleName

// DestinationIPRange is a string wrapper type
type DestinationIPRange string

// DestinationSecuritySpec is the security spec of destination
type DestinationSecuritySpec struct {
	SourceCert *Certificate `json:"SourceCert,omitempty"`
}

// DestinationIPRanges is a wrapper type of map[DestinationIPRange]*DestinationSecuritySpec
type DestinationIPRanges map[DestinationIPRange]*DestinationSecuritySpec

// SourceIPRange is a string wrapper type
type SourceIPRange string

// SourceSecuritySpec is the security spec of source
type SourceSecuritySpec struct {
	MTLS                     bool `json:"mTLS"`
	SkipClientCertValidation bool
	AuthenticatedPrincipals  []string
}

// SourceIPRanges is a wrapper type of map[SourceIPRange]*SourceSecuritySpec
type SourceIPRanges map[SourceIPRange]*SourceSecuritySpec

// AllowedEndpoints is a wrapper type of map[Address]ServiceName
type AllowedEndpoints map[Address]ServiceName

// PipyConf is a policy used by pipy sidecar
type PipyConf struct {
	Ts               *time.Time
	Version          *string
	Spec             MeshConfigSpec
	Certificate      *Certificate
	Inbound          *InboundTrafficPolicy    `json:"Inbound"`
	Outbound         *OutboundTrafficPolicy   `json:"Outbound"`
	Forward          *ForwardTrafficPolicy    `json:"Forward"`
	AllowedEndpoints map[string]string        `json:"AllowedEndpoints"`
	DNSResolveDB     map[string][]interface{} `json:"DNSResolveDB,omitempty"`
}

// FeatureFlags represents the flags of feature
type FeatureFlags struct {
	EnableSidecarActiveHealthChecks bool
}

// TrafficSpec represents the spec of traffic
type TrafficSpec struct {
	EnableEgress                      bool
	enablePermissiveTrafficPolicyMode bool
}

// UpstreamDNSServers defines upstream DNS servers for local DNS Proxy.
type UpstreamDNSServers struct {
	// Primary defines a primary upstream DNS server for local DNS Proxy.
	Primary *string `json:"Primary,omitempty"`
	// Secondary defines a secondary upstream DNS server for local DNS Proxy.
	Secondary *string `json:"Secondary,omitempty"`
}

// LocalDNSProxy is the type to represent OSM's local DNS proxy configuration.
type LocalDNSProxy struct {
	// UpstreamDNSServers defines upstream DNS servers for local DNS Proxy.
	UpstreamDNSServers *UpstreamDNSServers `json:"UpstreamDNSServers,omitempty"`
}

// MeshConfigSpec represents the spec of mesh config
type MeshConfigSpec struct {
	SidecarLogLevel string
	Traffic         TrafficSpec
	FeatureFlags    FeatureFlags
	Probes          struct {
		ReadinessProbes []v1.Probe `json:"ReadinessProbes,omitempty"`
		LivenessProbes  []v1.Probe `json:"LivenessProbes,omitempty"`
		StartupProbes   []v1.Probe `json:"StartupProbes,omitempty"`
	}
	LocalDNSProxy *LocalDNSProxy `json:"LocalDNSProxy,omitempty"`
}

// Certificate represents an x509 certificate.
type Certificate struct {
	// If issued by osm ca
	OsmIssued *bool `json:"OsmIssued,omitempty"`

	// The CommonName of the certificate
	CommonName *certificate.CommonName `json:"CommonName,omitempty"`

	// SubjectAltNames defines the Subject Alternative Names (domain names and IP addresses) secured by the certificate.
	SubjectAltNames []string `json:"SubjectAltNames,omitempty"`

	// When the cert expires
	Expiration string

	// PEM encoded Certificate and Key (byte arrays)
	CertChain  string
	PrivateKey string

	// Certificate authority signing this certificate
	IssuingCA string
}

// RetryPolicy is the type used to represent the retry policy specified in the Retry policy specification.
type RetryPolicy struct {
	// RetryOn defines the policies to retry on, delimited by comma.
	RetryOn string `json:"RetryOn"`

	// PerTryTimeout defines the time allowed for a retry before it's considered a failed attempt.
	// +optional
	PerTryTimeout *float64 `json:"PerTryTimeout"`

	// NumRetries defines the max number of retries to attempt.
	// +optional
	NumRetries *uint32 `json:"NumRetries"`

	// RetryBackoffBaseInterval defines the base interval for exponential retry backoff.
	// +optional
	RetryBackoffBaseInterval *float64 `json:"RetryBackoffBaseInterval"`
}

// WeightedCluster is a struct of a cluster and is weight that is backing a service
type WeightedCluster struct {
	service.WeightedCluster
	RetryPolicy *v1alpha1.RetryPolicySpec
}

// InboundHTTPRouteRule http route rule
type InboundHTTPRouteRule struct {
	HTTPRouteRule
	RateLimit *HTTPPerRouteRateLimit `json:"RateLimit"`
}

// InboundHTTPRouteRules is a wrapper type
type InboundHTTPRouteRules struct {
	RouteRules       map[URIPathRegexp]*InboundHTTPRouteRule `json:"RouteRules"`
	RateLimit        *HTTPRateLimit                          `json:"RateLimit"`
	HeaderRateLimits []*HTTPHeaderRateLimit                  `json:"HeaderRateLimits"`
}

// InboundHTTPServiceRouteRules is a wrapper type of map[HTTPRouteRuleName]*InboundHTTPRouteRules
type InboundHTTPServiceRouteRules map[HTTPRouteRuleName]*InboundHTTPRouteRules

// InboundTrafficMatch represents the match of InboundTraffic
type InboundTrafficMatch struct {
	Port                  Port     `json:"Port"`
	Protocol              Protocol `json:"Protocol"`
	SourceIPRanges        SourceIPRanges
	HTTPHostPort2Service  HTTPHostPort2Service         `json:"HttpHostPort2Service"`
	HTTPServiceRouteRules InboundHTTPServiceRouteRules `json:"HttpServiceRouteRules"`
	TargetClusters        WeightedClusters             `json:"TargetClusters"`
	AllowedEndpoints      AllowedEndpoints
	RateLimit             *TCPRateLimit `json:"RateLimit"`
}

// InboundTrafficMatches is a wrapper type of map[Port]*InboundTrafficMatch
type InboundTrafficMatches map[Port]*InboundTrafficMatch

// OutboundHTTPRouteRules is a wrapper type of map[URIPathRegexp]*HTTPRouteRule
type OutboundHTTPRouteRules map[URIPathRegexp]*HTTPRouteRule

// OutboundHTTPServiceRouteRules is a wrapper type of map[HTTPRouteRuleName]*HTTPRouteRules
type OutboundHTTPServiceRouteRules map[HTTPRouteRuleName]*OutboundHTTPRouteRules

// OutboundTrafficMatch represents the match of OutboundTraffic
type OutboundTrafficMatch struct {
	DestinationIPRanges   DestinationIPRanges
	Port                  Port                          `json:"Port"`
	Protocol              Protocol                      `json:"Protocol"`
	HTTPHostPort2Service  HTTPHostPort2Service          `json:"HttpHostPort2Service"`
	HTTPServiceRouteRules OutboundHTTPServiceRouteRules `json:"HttpServiceRouteRules"`
	TargetClusters        WeightedClusters              `json:"TargetClusters"`
	ServiceIdentity       identity.ServiceIdentity
	AllowedEgressTraffic  bool
	EgressForwardGateway  *string
}

// OutboundTrafficMatches is a wrapper type of map[Port][]*OutboundTrafficMatch
type OutboundTrafficMatches map[Port][]*OutboundTrafficMatch

// namedOutboundTrafficMatches is a wrapper type of map[string]*OutboundTrafficMatch
type namedOutboundTrafficMatches map[string]*OutboundTrafficMatch

// InboundTrafficPolicy represents the policy of InboundTraffic
type InboundTrafficPolicy struct {
	TrafficMatches  InboundTrafficMatches             `json:"TrafficMatches"`
	ClustersConfigs map[ClusterName]*WeightedEndpoint `json:"ClustersConfigs"`
}

// WeightedZoneEndpoint represents the endpoint with zone and weight
type WeightedZoneEndpoint struct {
	Weight      Weight `json:"Weight"`
	Cluster     string `json:"Cluster,omitempty"`
	ContextPath string `json:"Path,omitempty"`
}

// WeightedEndpoints is a wrapper type of map[HTTPHostPort]WeightedZoneEndpoint
type WeightedEndpoints map[HTTPHostPort]*WeightedZoneEndpoint

// ClusterConfigs represents the configs of Cluster
type ClusterConfigs struct {
	Endpoints          *WeightedEndpoints  `json:"Endpoints"`
	ConnectionSettings *ConnectionSettings `json:"ConnectionSettings,omitempty"`
	RetryPolicy        *RetryPolicy        `json:"RetryPolicy,omitempty"`
	SourceCert         *Certificate        `json:"SourceCert,omitempty"`
}

// OutboundTrafficPolicy represents the policy of OutboundTraffic
type OutboundTrafficPolicy struct {
	namedTrafficMatches namedOutboundTrafficMatches
	TrafficMatches      OutboundTrafficMatches          `json:"TrafficMatches"`
	ClustersConfigs     map[ClusterName]*ClusterConfigs `json:"ClustersConfigs"`
}

// ForwardTrafficMatches is a wrapper type of map[Port][]WeightedClusters
type ForwardTrafficMatches map[string]WeightedClusters

// ForwardTrafficPolicy represents the policy of Egress Gateway
type ForwardTrafficPolicy struct {
	ForwardMatches ForwardTrafficMatches           `json:"ForwardMatches"`
	EgressGateways map[ClusterName]*ClusterConfigs `json:"EgressGateways"`
}

// ConnectionSettings defines the connection settings for an
// upstream host.
type ConnectionSettings struct {
	// TCP specifies the TCP level connection settings.
	// Applies to both TCP and HTTP connections.
	// +optional
	TCP *TCPConnectionSettings `json:"tcp,omitempty"`

	// HTTP specifies the HTTP level connection settings.
	// +optional
	HTTP *HTTPConnectionSettings `json:"http,omitempty"`
}

// TCPConnectionSettings defines the TCP connection settings for an
// upstream host.
type TCPConnectionSettings struct {
	// MaxConnections specifies the maximum number of TCP connections
	// allowed to the upstream host.
	// Defaults to 4294967295 (2^32 - 1) if not specified.
	// +optional
	MaxConnections *uint32 `json:"MaxConnections,omitempty"`

	// ConnectTimeout specifies the TCP connection timeout.
	// Defaults to 5s if not specified.
	// +optional
	ConnectTimeout *float64 `json:"ConnectTimeout,omitempty"`
}

// HTTPCircuitBreaking defines the HTTP Circuit Breaking settings for an
// upstream host.
type HTTPCircuitBreaking struct {
	// StatTimeWindow specifies statistical time period of circuit breaking
	StatTimeWindow *float64 `json:"StatTimeWindow"`

	// MinRequestAmount specifies minimum number of requests (in an active statistic time span) that can trigger circuit breaking.
	MinRequestAmount uint32 `json:"MinRequestAmount"`

	// DegradedTimeWindow specifies the duration of circuit breaking
	DegradedTimeWindow *float64 `json:"DegradedTimeWindow"`

	// SlowTimeThreshold specifies the time threshold of slow request
	SlowTimeThreshold *float64 `json:"SlowTimeThreshold,omitempty"`

	// SlowAmountThreshold specifies the amount threshold of slow request
	SlowAmountThreshold *uint32 `json:"SlowAmountThreshold,omitempty"`

	// SlowRatioThreshold specifies the ratio threshold of slow request
	SlowRatioThreshold *float32 `json:"SlowRatioThreshold,omitempty"`

	// ErrorAmountThreshold specifies the amount threshold of error request
	ErrorAmountThreshold *uint32 `json:"ErrorAmountThreshold,omitempty"`

	// ErrorRatioThreshold specifies the ratio threshold of error request
	ErrorRatioThreshold *float32 `json:"ErrorRatioThreshold,omitempty"`

	// DegradedStatusCode specifies the degraded http status code of circuit breaking
	DegradedStatusCode *int32 `json:"DegradedStatusCode,omitempty"`

	// DegradedResponseContent specifies the degraded http response content of circuit breaking
	DegradedResponseContent *string `json:"DegradedResponseContent,omitempty"`
}

// HTTPConnectionSettings defines the HTTP connection settings for an
// upstream host.
type HTTPConnectionSettings struct {
	// MaxRequests specifies the maximum number of parallel requests
	// allowed to the upstream host.
	// Defaults to 4294967295 (2^32 - 1) if not specified.
	// +optional
	MaxRequests *uint32 `json:"MaxRequests,omitempty"`

	// MaxRequestsPerConnection specifies the maximum number of requests
	// per connection allowed to the upstream host.
	// Defaults to unlimited if not specified.
	// +optional
	MaxRequestsPerConnection *uint32 `json:"MaxRequestsPerConnection,omitempty"`

	// MaxPendingRequests specifies the maximum number of pending HTTP
	// requests allowed to the upstream host. For HTTP/2 connections,
	// if `maxRequestsPerConnection` is not configured, all requests will
	// be multiplexed over the same connection so this circuit breaker
	// will only be hit when no connection is already established.
	// Defaults to 4294967295 (2^32 - 1) if not specified.
	// +optional
	MaxPendingRequests *uint32 `json:"MaxPendingRequests,omitempty"`

	// MaxRetries specifies the maximum number of parallel retries
	// allowed to the upstream host.
	// Defaults to 4294967295 (2^32 - 1) if not specified.
	// +optional
	MaxRetries *uint32 `json:"MaxRetries,omitempty"`

	// CircuitBreaking specifies the HTTP connection circuit breaking setting.
	CircuitBreaking *HTTPCircuitBreaking `json:"CircuitBreaking,omitempty"`
}
