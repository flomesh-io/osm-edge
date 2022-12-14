package repo

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/client"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
	"github.com/openservicemesh/osm/pkg/workerpool"
)

const (
	// ServerType is the type identifier for the ADS server
	ServerType = "pipy-Repo"

	// workerPoolSize is the default number of workerpool workers (0 is GOMAXPROCS)
	workerPoolSize = 0

	osmCodebase        = "/osm-edge"
	osmSidecarCodebase = "/osm-edge-sidecar"
	osmCodebaseConfig  = "config.json"
)

// NewRepoServer creates a new Aggregated Discovery Service server
func NewRepoServer(meshCatalog catalog.MeshCataloger, proxyRegistry *registry.ProxyRegistry, _ bool, osmNamespace string,
	cfg configurator.Configurator, certManager *certificate.Manager, kubecontroller k8s.Controller, msgBroker *messaging.Broker) *Server {
	server := Server{
		catalog:        meshCatalog,
		proxyRegistry:  proxyRegistry,
		osmNamespace:   osmNamespace,
		cfg:            cfg,
		certManager:    certManager,
		workQueues:     workerpool.NewWorkerPool(workerPoolSize),
		kubeController: kubecontroller,
		configVerMutex: sync.Mutex{},
		configVersion:  make(map[string]uint64),
		msgBroker:      msgBroker,
		repoClient:     client.NewRepoClient("127.0.0.1", uint16(cfg.GetProxyServerPort())),
	}

	return &server
}

// Start starts the codebase push server
func (s *Server) Start(_ uint32, _ *certificate.Certificate) error {
	// wait until pipy repo is up
	err := wait.PollImmediate(5*time.Second, 60*time.Second, func() (bool, error) {
		success, err := s.repoClient.IsRepoUp()
		if success {
			log.Info().Msg("Repo is READY!")
			return success, nil
		}
		log.Error().Msg("Repo is not up, sleeping ...")
		return success, err
	})
	if err != nil {
		log.Error().Err(err)
	}

	_, err = s.repoClient.Batch(fmt.Sprintf("%d", 0), []client.Batch{
		{
			Basepath: osmCodebase,
			Items: []client.BatchItem{

				{Filename: "config.js", Content: codebaseConfigJs},
				{Filename: "main.js", Content: codebaseMainJs},
				{Filename: "modules/inbound-http-default.js", Content: codebaseModulesInboundHTTPDefaultJs},
				{Filename: "modules/inbound-http-load-balancing.js", Content: codebaseModulesInboundHTTPLoadBalancingJs},
				{Filename: "modules/inbound-http-routing.js", Content: codebaseModulesInboundHTTPRoutingJs},
				{Filename: "modules/inbound-logging-http.js", Content: codebaseModulesInboundLoggingHTTPJs},
				{Filename: "modules/inbound-main.js", Content: codebaseModulesInboundMainJs},
				{Filename: "modules/inbound-make-connection.js", Content: codebaseModulesInboundMakeConnectionJs},
				{Filename: "modules/inbound-metrics-http.js", Content: codebaseModulesInboundMetricsHTTPJs},
				{Filename: "modules/inbound-metrics-tcp.js", Content: codebaseModulesInboundMetricsTCPJs},
				{Filename: "modules/inbound-tcp-default.js", Content: codebaseModulesInboundTCPDefaultJs},
				{Filename: "modules/inbound-tcp-load-balancing.js", Content: codebaseModulesInboundTCPLoadBalancingJs},
				{Filename: "modules/inbound-throttle-route.js", Content: codebaseModulesInboundThrottleRouteJs},
				{Filename: "modules/inbound-throttle-service.js", Content: codebaseModulesInboundThrottleServiceJs},
				{Filename: "modules/inbound-tls-termination.js", Content: codebaseModulesInboundTLSTerminationJs},
				{Filename: "modules/inbound-tracing-http.js", Content: codebaseModulesInboundTracingHTTPJs},
				{Filename: "modules/logging.js", Content: codebaseModulesLoggingJs},
				{Filename: "modules/metrics.js", Content: codebaseModulesMetricsJs},
				{Filename: "modules/outbound-circuit-breaker.js", Content: codebaseModulesOutboundCircuitBreakerJs},
				{Filename: "modules/outbound-http-default.js", Content: codebaseModulesOutboundHTTPDefaultJs},
				{Filename: "modules/outbound-http-load-balancing.js", Content: codebaseModulesOutboundHTTPLoadBalancingJs},
				{Filename: "modules/outbound-http-routing.js", Content: codebaseModulesOutboundHTTPRoutingJs},
				{Filename: "modules/outbound-logging-http.js", Content: codebaseModulesOutboundLoggingHTTPJs},
				{Filename: "modules/outbound-main.js", Content: codebaseModulesOutboundMainJs},
				{Filename: "modules/outbound-metrics-http.js", Content: codebaseModulesOutboundMetricsHTTPJs},
				{Filename: "modules/outbound-metrics-tcp.js", Content: codebaseModulesOutboundMetricsTCPJs},
				{Filename: "modules/outbound-tcp-default.js", Content: codebaseModulesOutboundTCPDefaultJs},
				{Filename: "modules/outbound-tcp-load-balancing.js", Content: codebaseModulesOutboundTCPLoadBalancingJs},
				{Filename: "modules/outbound-tls-initiation.js", Content: codebaseModulesOutboundTLSInitiationJs},
				{Filename: "modules/outbound-tracing-http.js", Content: codebaseModulesOutboundTracingHTTPJs},
				{Filename: "modules/tracing.js", Content: codebaseModulesTracingJs},
				{Filename: "probes.js", Content: codebaseProbesJs},
				{Filename: "stats.js", Content: codebaseStatsJs},

				{
					Filename: osmCodebaseConfig,
					Content:  codebaseConfigJSON,
				},
			},
		},
	})
	if err != nil {
		log.Error().Err(err)
	}

	// Start broadcast listener thread
	go s.broadcastListener()

	s.ready = true

	return nil
}
