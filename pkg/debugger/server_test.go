package debugger

import (
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	tassert "github.com/stretchr/testify/assert"
	testclient "k8s.io/client-go/kubernetes/fake"

	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/registry"
)

type dummyHandler struct{}

func (h *dummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

// Tests GetHandlers returns the expected debug endpoints and non-nil handlers
func TestGetHandlers(t *testing.T) {
	assert := tassert.New(t)
	mockCtrl := gomock.NewController(t)

	mockCertDebugger := NewMockCertificateManagerDebugger(mockCtrl)
	mockProxyDebugger := driver.NewMockProxyDebugger(mockCtrl)
	mockCatalogDebugger := NewMockMeshCatalogDebugger(mockCtrl)
	mockConfig := configurator.NewMockConfigurator(mockCtrl)
	client := testclient.NewSimpleClientset()
	mockKubeController := k8s.NewMockController(mockCtrl)
	proxyRegistry := registry.NewProxyRegistry(nil, nil)

	mockProxyDebugger.EXPECT().GetDebugHandlers().Return(map[string]http.Handler{
		"/certs":      new(dummyHandler),
		"/xds":        new(dummyHandler),
		"/proxy":      new(dummyHandler),
		"/policies":   new(dummyHandler),
		"/config":     new(dummyHandler),
		"/namespaces": new(dummyHandler),
		// Pprof handlers
		"/pprof/":        new(dummyHandler),
		"/pprof/cmdline": new(dummyHandler),
		"/pprof/profile": new(dummyHandler),
		"/pprof/symbol":  new(dummyHandler),
		"/pprof/trace":   new(dummyHandler),
	})

	ds := NewDebugConfig(mockCertDebugger,
		mockProxyDebugger,
		mockCatalogDebugger,
		proxyRegistry,
		nil,
		client,
		mockConfig,
		mockKubeController,
		nil)

	handlers := ds.GetHandlers()

	debugEndpoints := []string{
		"/debug/certs",
		"/debug/xds",
		"/debug/proxy",
		"/debug/policies",
		"/debug/config",
		"/debug/namespaces",
		// Pprof handlers
		"/debug/pprof/",
		"/debug/pprof/cmdline",
		"/debug/pprof/profile",
		"/debug/pprof/symbol",
		"/debug/pprof/trace",
	}

	for _, endpoint := range debugEndpoints {
		handler, found := handlers[endpoint]
		assert.True(found)
		assert.NotNil(handler)
	}
}
