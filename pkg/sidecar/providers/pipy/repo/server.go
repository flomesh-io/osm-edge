package repo

import (
	"fmt"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
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

	osmCodebaseConfig = "config.json"
)

var (
	osmCodebase        = "osm-edge-base"
	osmSidecarCodebase = "osm-edge-sidecar"
	osmCodebaseRepo    = fmt.Sprintf("/%s", osmCodebase)
)

// NewRepoServer creates a new Aggregated Discovery Service server
func NewRepoServer(meshCatalog catalog.MeshCataloger, proxyRegistry *registry.ProxyRegistry, _ bool, osmNamespace string, cfg configurator.Configurator, certManager *certificate.Manager, kubecontroller k8s.Controller, msgBroker *messaging.Broker) *Server {
	if len(cfg.GetRepoServerCodebase()) > 0 {
		osmCodebase = fmt.Sprintf("%s/%s", cfg.GetRepoServerCodebase(), osmCodebase)
		osmSidecarCodebase = fmt.Sprintf("%s/%s", cfg.GetRepoServerCodebase(), osmSidecarCodebase)
		osmCodebaseRepo = fmt.Sprintf("/%s", osmCodebase)
	}

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
		pluginSet:      mapset.NewSet(),
		msgBroker:      msgBroker,
		repoClient:     client.NewRepoClient(cfg.GetRepoServerIPAddr(), uint16(cfg.GetProxyServerPort())),
	}

	return &server
}

// Start starts the codebase push server
func (s *Server) Start(_ uint32, _ *certificate.Certificate) error {
	// wait until pipy repo is up
	err := wait.PollImmediate(5*time.Second, 90*time.Second, func() (bool, error) {
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
		return err
	}

	s.repoClient.Restore = func() error {
		_, restoreErr := s.repoClient.Batch(fmt.Sprintf("%d", 0), []client.Batch{
			{
				Basepath: osmCodebase,
				Items:    osmCodebaseItems,
			},
		})
		if restoreErr != nil {
			log.Error().Err(restoreErr)
			return restoreErr
		}
		// wait until base codebase is ready
		restoreErr = wait.PollImmediate(5*time.Second, 90*time.Second, func() (bool, error) {
			success, _, codebaseErr := s.repoClient.GetCodebase(osmCodebase)
			if success {
				log.Info().Msg("Base codebase is READY!")
				return success, nil
			}
			log.Error().Msg("Base codebase is NOT READY, sleeping ...")
			return success, codebaseErr
		})
		if restoreErr != nil {
			log.Error().Err(restoreErr)
			return restoreErr
		}
		return nil
	}

	err = s.repoClient.Restore()
	if err != nil {
		log.Error().Err(err)
		return err
	}

	// Start broadcast listener thread
	go s.broadcastListener()

	s.ready = true

	return nil
}
