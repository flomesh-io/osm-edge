package repo

import (
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

	_, err = s.repoClient.Batch(0, []client.Batch{
		{
			Basepath: osmCodebase,
			Items: []client.BatchItem{
				{
					Filename: "main.js",
					Content:  codebaseMainJS,
				},
				{
					Filename: "config.js",
					Content:  codebaseConfigJS,
				},
				{
					Filename: "metrics.js",
					Content:  codebaseMetricsJS,
				},
				{
					Filename: osmCodebaseConfig,
					Content:  codebaseConfigJSON,
				},
				{
					Filename: "codes.js",
					Content:  codebaseCodesJS,
				},
				{
					Filename: "breaker.js",
					Content:  codebaseBreakerJS,
				},
				{
					Filename: "gather.js",
					Content:  codebaseGatherJS,
				},
				{
					Filename: "stats.js",
					Content:  codebaseStatsJS,
				},
				{
					Filename: "inbound-proxy-tcp.js",
					Content:  codebaseInboundProxyTCPJS,
				},
				{
					Filename: "inbound-recv-http.js",
					Content:  codebaseInboundRecvHTTPJS,
				},
				{
					Filename: "inbound-recv-tcp.js",
					Content:  codebaseInboundRecvTCPJS,
				},
				{
					Filename: "inbound-throttle.js",
					Content:  codebaseInboundThrottleJS,
				},
				{
					Filename: "outbound-breaker.js",
					Content:  codebaseOutboundBreakerJS,
				},
				{
					Filename: "outbound-mux-http.js",
					Content:  codebaseOutboundMuxHTTPJS,
				},
				{
					Filename: "outbound-proxy-tcp.js",
					Content:  codebaseOutboundProxyTCPJS,
				},
				{
					Filename: "outbound-recv-http.js",
					Content:  codebaseOutboundRecvHTTPJS,
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
