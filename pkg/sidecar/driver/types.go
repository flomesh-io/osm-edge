package driver

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/health"
	"github.com/openservicemesh/osm/pkg/messaging"
)

// Driver is an interface that must be implemented by a sidecar driver.
// Patch method is invoked by osm-injector and Start method is invoked by osm-controller
type Driver interface {
	Patch(ctx context.Context) error
	Start(ctx context.Context) (health.Probes, error)
}

// HealthProbes is to serve as an indication how to probe the sidecar driver's health status
type HealthProbes struct {
	liveness, readiness, startup *HealthProbe
}

// HealthProbe is an API endpoint to indicate the current status of the sidecar driver.
type HealthProbe struct {
	path      string
	port      int32
	http      bool
	timeout   time.Duration
	tcpSocket bool
}

// InjectorCtxKey the pointer is the key that a InjectorContext returns itself for.
var InjectorCtxKey int

// InjectorContext carries the arguments for invoking InjectorDriver.Patch
type InjectorContext struct {
	context.Context

	Pod                          *corev1.Pod
	MeshName                     string
	OsmNamespace                 string
	PodNamespace                 string
	PodOS                        string
	ProxyCommonName              certificate.CommonName
	ProxyUUID                    uuid.UUID
	Configurator                 configurator.Configurator
	KubeClient                   kubernetes.Interface
	BootstrapCertificate         *certificate.Certificate
	ContainerPullPolicy          corev1.PullPolicy
	InboundPortExclusionList     []int
	OutboundPortExclusionList    []int
	OutboundIPRangeInclusionList []string
	OutboundIPRangeExclusionList []string
	OriginalHealthProbes         HealthProbes

	DryRun bool
}

// ControllerCtxKey the pointer is the key that a ControllerContext returns itself for.
var ControllerCtxKey int

// ControllerContext carries the arguments for invoking ControllerDriver.Start
type ControllerContext struct {
	context.Context

	ProxyServerPort  uint32
	ProxyServiceCert *certificate.Certificate
	OsmNamespace     string
	KubeConfig       *rest.Config
	Configurator     configurator.Configurator
	MeshCatalog      catalog.MeshCataloger
	CertManager      certificate.Manager
	MsgBroker        *messaging.Broker
	DebugHandlers    map[string]http.Handler
	CancelFunc       func()
	Stop             chan struct {
	}
}
