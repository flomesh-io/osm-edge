package repo

import (
	"sync"
	"time"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/logger"
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/pipy"
	"github.com/openservicemesh/osm/pkg/pipy/registry"
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
	certManager    certificate.Manager
	ready          bool
	workqueues     *workerpool.WorkerPool
	kubecontroller k8s.Controller

	// When snapshot cache is enabled, we (currently) don't keep track of proxy information, however different
	// config versions have to be provided to the cache as we keep adding snapshots. The following map
	// tracks at which version we are at given a proxy UUID
	configVerMutex sync.Mutex
	configVersion  map[string]uint64

	msgBroker *messaging.Broker
}

// Repo pipy repo server wrapper
type Repo struct {
	server           *Server
	connectedProxies sync.Map
}

// ConnectedProxy is the proxy object of connected pipy sidecar
type ConnectedProxy struct {
	proxy        *pipy.Proxy
	connectedAt  time.Time
	lastReportAt time.Time
	initError    error
	quit         chan struct{}
}

// PipyReport is data reported by pipy sidecar
type PipyReport struct {
	Timestamp uint64 `json:"timestamp"`
	UUID      string `json:"uuid"`
	Version   string `json:"version"`
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

// ClustersConfigs is a wrapper type of map[ClusterName]*WeightedEndpoint
type ClustersConfigs map[ClusterName]*WeightedEndpoint

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

// HTTPRouteRules is a wrapper type of map[URIPathRegexp]*HTTPRouteRule
type HTTPRouteRules map[URIPathRegexp]*HTTPRouteRule

// HTTPRouteRuleName is a string wrapper type
type HTTPRouteRuleName string

// HTTPServiceRouteRules is a wrapper type of map[HTTPRouteRuleName]*HTTPRouteRules
type HTTPServiceRouteRules map[HTTPRouteRuleName]*HTTPRouteRules

// HTTPHostPort is a string wrapper type
type HTTPHostPort string

// HTTPHostPort2Service is a wrapper type of map[HTTPHostPort]HTTPRouteRuleName
type HTTPHostPort2Service map[HTTPHostPort]HTTPRouteRuleName

// DestinationIPRange is a string wrapper type
type DestinationIPRange string

// DestinationIPRanges is a wrapper type of []DestinationIPRange
type DestinationIPRanges []DestinationIPRange

// SourceIPRange is a string wrapper type
type SourceIPRange string

// SourceIPRanges is a wrapper type of []SourceIPRange
type SourceIPRanges []SourceIPRange

// AllowedEndpoints is a wrapper type of map[Address]ServiceName
type AllowedEndpoints map[Address]ServiceName

// TrafficMatch represents the base match of traffic
type TrafficMatch struct {
	Port                  Port                  `json:"Port"`
	Protocol              Protocol              `json:"Protocol"`
	HTTPHostPort2Service  HTTPHostPort2Service  `json:"HttpHostPort2Service"`
	HTTPServiceRouteRules HTTPServiceRouteRules `json:"HttpServiceRouteRules"`
	TargetClusters        WeightedClusters      `json:"TargetClusters"`
}

// InboundTrafficMatch represents the match of InboundTraffic
type InboundTrafficMatch struct {
	SourceIPRanges SourceIPRanges
	TrafficMatch
	AllowedEndpoints AllowedEndpoints
}

// InboundTrafficMatches is a wrapper type of map[Port]*InboundTrafficMatch
type InboundTrafficMatches map[Port]*InboundTrafficMatch

// OutboundTrafficMatch represents the match of OutboundTraffic
type OutboundTrafficMatch struct {
	DestinationIPRanges DestinationIPRanges
	TrafficMatch
	AllowedEgressTraffic bool
	ServiceIdentity      identity.ServiceIdentity
}

// OutboundTrafficMatches is a wrapper type of map[Port][]*OutboundTrafficMatch
type OutboundTrafficMatches map[Port][]*OutboundTrafficMatch

// TrafficPolicy represents the base policy of traffic
type TrafficPolicy struct {
	ClustersConfigs ClustersConfigs `json:"ClustersConfigs"`
}

// InboundTrafficPolicy represents the policy of InboundTraffic
type InboundTrafficPolicy struct {
	TrafficMatches InboundTrafficMatches `json:"TrafficMatches"`
	TrafficPolicy
}

// OutboundTrafficPolicy represents the policy of OutboundTraffic
type OutboundTrafficPolicy struct {
	TrafficMatches OutboundTrafficMatches `json:"TrafficMatches"`
	TrafficPolicy
}

// FeatureFlags represents the flags of feature
type FeatureFlags struct {
	SidecarLogLevel                 string
	EnableSidecarActiveHealthChecks bool
}

// TrafficSpec represents the spec of traffic
type TrafficSpec struct {
	EnableEgress                      bool
	enablePermissiveTrafficPolicyMode bool
}

// MeshConfigSpec represents the spec of mesh config
type MeshConfigSpec struct {
	Traffic      TrafficSpec
	FeatureFlags FeatureFlags
	Probes       struct {
		ReadinessProbes []interface{}
		LivenessProbes  []interface{}
		StartupProbes   []interface{}
	}
}

// PipyConf is a policy used by pipy sidecar
type PipyConf struct {
	Spec              MeshConfigSpec
	Inbound           *InboundTrafficPolicy  `json:"Inbound"`
	Outbound          *OutboundTrafficPolicy `json:"Outbound"`
	AllowedEndpoints  map[string]string      `json:"AllowedEndpoints"`
	allowedEndpointsV uint64
	bytes             []byte
}
