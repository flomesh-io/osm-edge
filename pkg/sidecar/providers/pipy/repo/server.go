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
		repoClient:     client.NewRepoClient(fmt.Sprintf("127.0.0.1:%d", cfg.GetProxyServerPort())),
	}

	return &server
}

// Start starts the codebase push server
func (s *Server) Start(_ uint32, _ *certificate.Certificate) error {
	// wait until pipy repo is up
	err := wait.PollImmediate(5*time.Second, 60*time.Second, func() (bool, error) {
		if s.repoClient.IsRepoUp() {
			log.Info().Msg("Repo is READY!")
			return true, nil
		}
		log.Info().Msg("Repo is not up, sleeping ...")
		return false, nil
	})
	if err != nil {
		log.Error().Err(err)
	}

	err = s.repoClient.Batch(0, []client.Batch{
		{
			Basepath: osmCodebase,
			Items: []client.BatchItem{
				{
					Filename: "main.js",
					Content:  codebaseMainJS,
				},
				{
					Filename: "f_config.js",
					Content:  codebaseConfigJS,
				},
				{
					Filename: "f_metrics.js",
					Content:  codebaseMetricsJS,
				},
				{
					Filename: "pipy.json",
					Content:  codebasePipyJSON,
				},
				{
					Filename: "f_codes.js",
					Content:  codebaseCodesJS,
				},
				{
					Filename: "f_breaker.js",
					Content:  codebaseBreakerJS,
				},
				{
					Filename: "p_gather.js",
					Content:  codebaseGatherJS,
				},
				{
					Filename: "p_stats.js",
					Content:  codebaseStatsJS,
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
