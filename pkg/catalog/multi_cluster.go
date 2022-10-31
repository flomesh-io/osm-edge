package catalog

import (
	multiclusterv1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"
)

// ListImportedServices lists imported services
func (mc *MeshCatalog) ListImportedServices() []*multiclusterv1alpha1.ServiceImport {
	return mc.multiclusterController.ListImportedServices()
}
