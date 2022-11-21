package fsm

import (
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/multicluster"
)

const (
	// providerName is the name of the Flomesh client that implements service.Provider and endpoint.Provider interfaces
	providerName = "flomesh"
)

// client is the type used to represent the k8s client for endpoints and service provider
type client struct {
	multiclusterController multicluster.Controller
	meshConfigurator       configurator.Configurator
}
