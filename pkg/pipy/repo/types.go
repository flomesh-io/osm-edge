package repo

import (
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/pipy"
	"sync"
	"time"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/logger"
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/pipy/registry"
	"github.com/openservicemesh/osm/pkg/workerpool"
)

var (
	log = logger.New("flomesh-pipy")
)

// Server implements the Sidecar xDS Aggregate Discovery Services
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

type Repo struct {
	server           *Server
	connectedProxies sync.Map
}

type ConnectedProxy struct {
	proxy        *pipy.Proxy
	connectedAt  time.Time
	lastReportAt time.Time
	initError    error
	quit         chan struct{}
}

type PipyReport struct {
	Timestamp uint64 `json:"timestamp"`
	Uuid      string `json:"uuid"`
	Version   string `json:"version"`
}

type Protocol string
type Address string
type Port uint16
type Weight uint32
type ClusterName string
type WeightedEndpoint map[HttpHostPort]Weight
type ClustersConfigs map[ClusterName]*WeightedEndpoint

type Header string
type HeaderRegexp string
type HeadersMatch map[Header]HeaderRegexp

type Method string
type MethodsMatch []Method
type WeightedClusters map[ClusterName]Weight
type URIPathRegexp string
type ServiceName string
type Services []ServiceName

type HttpRouteRule struct {
	Headers         HeadersMatch     `json:"Headers,omitempty"`
	Methods         MethodsMatch     `json:"Methods,omitempty"`
	TargetClusters  WeightedClusters `json:"TargetClusters,omitempty"`
	AllowedServices Services         `json:"AllowedServices,omitempty"`
}
type HttpRouteRules map[URIPathRegexp]*HttpRouteRule
type HttpRouteRuleName string
type HttpServiceRouteRules map[HttpRouteRuleName]*HttpRouteRules

type HttpHostPort string
type HttpHostPort2Service map[HttpHostPort]HttpRouteRuleName

type DestinationIPRange string
type DestinationIPRanges []DestinationIPRange

type SourceIPRange string
type SourceIPRanges []SourceIPRange

type AllowedEndpoints map[Address]ServiceName

type TrafficMatch struct {
	Port                  Port                  `json:"Port"`
	Protocol              Protocol              `json:"Protocol"`
	HttpHostPort2Service  HttpHostPort2Service  `json:"HttpHostPort2Service"`
	HttpServiceRouteRules HttpServiceRouteRules `json:"HttpServiceRouteRules"`
	TargetClusters        WeightedClusters      `json:"TargetClusters"`
}

type InboundTrafficMatch struct {
	SourceIPRanges SourceIPRanges
	TrafficMatch
	AllowedEndpoints AllowedEndpoints
}
type InboundTrafficMatches map[Port]*InboundTrafficMatch

type OutboundTrafficMatch struct {
	DestinationIPRanges DestinationIPRanges
	TrafficMatch
	AllowedEgressTraffic bool
	ServiceIdentity      identity.ServiceIdentity
}
type OutboundTrafficMatches []*OutboundTrafficMatch

type TrafficPolicy struct {
	ClustersConfigs ClustersConfigs `json:"ClustersConfigs"`
}

type InboundTrafficPolicy struct {
	TrafficMatches InboundTrafficMatches `json:"TrafficMatches"`
	TrafficPolicy
}

type OutboundTrafficPolicy struct {
	TrafficMatches OutboundTrafficMatches `json:"TrafficMatches"`
	TrafficPolicy
}

type FeatureFlags struct {
	EnableSidecarActiveHealthChecks bool
}

type TrafficSpec struct {
	EnableEgress                      bool
	enablePermissiveTrafficPolicyMode bool
}

type MeshConfigSpec struct {
	Traffic      TrafficSpec
	FeatureFlags FeatureFlags
	Probes       struct {
		ReadinessProbes []interface{}
		LivenessProbes  []interface{}
		StartupProbes   []interface{}
	}
}

type PipyConf struct {
	Spec              MeshConfigSpec
	Inbound           *InboundTrafficPolicy  `json:"Inbound"`
	Outbound          *OutboundTrafficPolicy `json:"Outbound"`
	AllowedEndpoints  map[string]string      `json:"AllowedEndpoints"`
	allowedEndpointsV uint64
	bytes             []byte
}
