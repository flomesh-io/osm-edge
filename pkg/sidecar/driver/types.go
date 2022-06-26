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

// Driver is the interface that must be implemented by a sidecar driver.
type Driver interface {
	Patch(ctx context.Context) error
	Start(ctx context.Context) (health.Probes, error)
}

// HealthProbes is to serve as an indication whether the given healthProbe has been rewritten
type HealthProbes struct {
	liveness, readiness, startup *HealthProbe
}

// HealthProbe is an API endpoint to indicate the current status of the server.
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

	ProxyServerPort  int
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
