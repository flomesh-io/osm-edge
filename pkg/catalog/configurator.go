package catalog

import (
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/endpoint"
	"github.com/openservicemesh/osm/pkg/identity"
)

// GetConfigurator converts private variable to public
func (mc *MeshCatalog) GetConfigurator() *configurator.Configurator {
	return &mc.configurator
}

// ListEndpointsForServiceIdentity converts private method to public
func (mc *MeshCatalog) ListEndpointsForServiceIdentity(serviceIdentity identity.ServiceIdentity) []endpoint.Endpoint {
	return mc.listEndpointsForServiceIdentity(serviceIdentity)
}
