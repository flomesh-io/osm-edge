package debugger

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/sidecar"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
)

// GetHandlers implements DebugConfig interface and returns the rest of URLs and the handling functions.
func (ds DebugConfig) GetHandlers() map[string]http.Handler {
	handlers := map[string]http.Handler{
		"/debug/certs":         ds.getCertHandler(),
		"/debug/proxy":         ds.getProxies(),
		"/debug/policies":      ds.getSMIPoliciesHandler(),
		"/debug/config":        ds.getOSMConfigHandler(),
		"/debug/namespaces":    ds.getMonitoredNamespacesHandler(),
		"/debug/feature-flags": ds.getFeatureFlags(),

		// Pprof handlers
		"/debug/pprof/":        http.HandlerFunc(pprof.Index),
		"/debug/pprof/cmdline": http.HandlerFunc(pprof.Cmdline),
		"/debug/pprof/profile": http.HandlerFunc(pprof.Profile),
		"/debug/pprof/symbol":  http.HandlerFunc(pprof.Symbol),
		"/debug/pprof/trace":   http.HandlerFunc(pprof.Trace),
	}

	if ds.proxyDebugger != nil {
		proxyDebugHandlers := ds.proxyDebugger.GetDebugHandlers()
		if len(proxyDebugHandlers) > 0 {
			for path, handler := range proxyDebugHandlers {
				if len(path) > 0 && handler != nil {
					if strings.HasPrefix(path, "/") {
						handlers["/debug"+path] = handler
					} else {
						handlers["/debug/"+path] = handler
					}
				}
			}
		}
	}

	// provides an index of the available /debug endpoints
	handlers["/debug"] = ds.getDebugIndex(handlers)

	return handlers
}

// NewDebugConfig returns an implementation of DebugConfig interface.
func NewDebugConfig(certDebugger CertificateManagerDebugger, proxyDebugger driver.ProxyDebugger, meshCatalogDebugger MeshCatalogDebugger,
	proxyRegistry sidecar.ProxyRegistry, kubeConfig *rest.Config, kubeClient kubernetes.Interface,
	cfg configurator.Configurator, kubeController k8s.Controller, msgBroker *messaging.Broker) DebugConfig {
	return DebugConfig{
		certDebugger:        certDebugger,
		proxyDebugger:       proxyDebugger,
		meshCatalogDebugger: meshCatalogDebugger,
		proxyRegistry:       proxyRegistry,
		kubeClient:          kubeClient,
		kubeController:      kubeController,

		// We need the Kubernetes config to be able to establish port forwarding to the Envoy pod we want to debug.
		kubeConfig: kubeConfig,

		configurator: cfg,

		msgBroker: msgBroker,
	}
}
