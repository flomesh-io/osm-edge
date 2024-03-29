// Package debugger implements functionality to provide information to debug the control plane via the debug HTTP server.
package debugger

import (
	access "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha3"
	spec "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha4"
	split "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha4"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/logger"
	"github.com/openservicemesh/osm/pkg/messaging"
)

var log = logger.New("debugger")

// DebugConfig implements the DebugServer interface.
type DebugConfig struct {
	certDebugger        CertificateManagerDebugger
	meshCatalogDebugger MeshCatalogDebugger
	kubeConfig          *rest.Config
	kubeClient          kubernetes.Interface
	kubeController      k8s.Controller
	configurator        configurator.Configurator
	msgBroker           *messaging.Broker
}

// CertificateManagerDebugger is an interface with methods for debugging certificate issuance.
type CertificateManagerDebugger interface {
	// ListIssuedCertificates returns the current list of certificates in OSM's cache.
	ListIssuedCertificates() []*certificate.Certificate
}

// MeshCatalogDebugger is an interface with methods for debugging Mesh Catalog.
type MeshCatalogDebugger interface {
	// ListSMIPolicies lists the SMI policies detected by OSM.
	ListSMIPolicies() ([]*split.TrafficSplit, []identity.K8sServiceAccount, []*spec.HTTPRouteGroup, []*access.TrafficTarget)
}
