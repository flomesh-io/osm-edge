package repo

import (
	"context"
	"sync"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/pipy/registry"
	"github.com/openservicemesh/osm/pkg/workerpool"
)

const (
	// ServerType is the type identifier for the ADS server
	ServerType = "ADS"

	// workerPoolSize is the default number of workerpool workers (0 is GOMAXPROCS)
	workerPoolSize = 0
)

// NewADSServer creates a new Aggregated Discovery Service server
func NewADSServer(meshCatalog catalog.MeshCataloger, proxyRegistry *registry.ProxyRegistry, _ bool, osmNamespace string,
	cfg configurator.Configurator, certManager certificate.Manager, kubecontroller k8s.Controller, msgBroker *messaging.Broker) *Server {
	server := Server{
		catalog:        meshCatalog,
		proxyRegistry:  proxyRegistry,
		osmNamespace:   osmNamespace,
		cfg:            cfg,
		certManager:    certManager,
		workqueues:     workerpool.NewWorkerPool(workerPoolSize),
		kubecontroller: kubecontroller,
		configVerMutex: sync.Mutex{},
		configVersion:  make(map[string]uint64),
		msgBroker:      msgBroker,
	}

	return &server
}

// Start starts the ADS server
func (s *Server) Start(_ context.Context, cancel context.CancelFunc, port int, _ *certificate.Certificate) error {
	// Start broadcast listener thread
	go s.broadcastListener()

	err := s.pipyRepoHTTPServer(uint16(port))
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrStartingADSServer)).
			Msg("Error starting ADS server")
		cancel()
		return err
	}

	s.ready = true

	return nil
}
