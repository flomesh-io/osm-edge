package catalog

import (
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/endpoint"
	"github.com/openservicemesh/osm/pkg/identity"
)

func (mc *MeshCatalog) GetConfigurator() *configurator.Configurator {
	return &mc.configurator
}

func (mc *MeshCatalog) ListEndpointsForServiceIdentity(serviceIdentity identity.ServiceIdentity) []endpoint.Endpoint {
	return mc.listEndpointsForServiceIdentity(serviceIdentity)
}
