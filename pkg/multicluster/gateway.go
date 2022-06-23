package multicluster

import (
	"github.com/google/uuid"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/sidecar"
)

// GetMulticlusterGatewaySubjectCommonName creates a unique certificate.CommonName
// specifically for a Multicluster Gateway. Each gateway will have its own unique
// cert. The kind of Envoy (gateway) is encoded in the cert CN by convention.
func GetMulticlusterGatewaySubjectCommonName(serviceAccount, namespace string) certificate.CommonName {
	gatewayUID := uuid.New()
	sidecarType := sidecar.KindGateway
	return sidecar.NewCertCommonName(gatewayUID, sidecarType, serviceAccount, namespace)
}
