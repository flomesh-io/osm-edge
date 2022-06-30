package repo

import (
	_ "embed"
	"sync"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
)

const (
	certificateNoSerial = "NoSerial"
)

// codebasePluginSource is the main pjs.
// Its value is embedded at build time.
//go:embed repo.main.js
var codebasePluginSource []byte

func (r *Server) registerProxy(proxy *pipy.Proxy) (connectedProxy *ConnectedProxy, exists bool) {
	var actual interface{}
	connectedProxy = &ConnectedProxy{
		proxy: proxy,
	}
	actual, exists = r.connectedProxies.LoadOrStore(proxy.GetCertificateCommonName(), connectedProxy)
	if exists {
		connectedProxy = actual.(*ConnectedProxy)
	} else {
		log.Debug().Str("proxy", proxy.String()).Msg("Registered new proxy")
	}
	return
}

func (r *Server) unregisterProxy(p *pipy.Proxy) {
	r.connectedProxies.Delete(p.GetCertificateCommonName())
	log.Debug().Msgf("Unregistered proxy %s", p.String())
}

func (r *Server) firedProxy(certCommonName certificate.CommonName, podIP string) {
	proxy, err := pipy.NewProxy(certCommonName, certificateNoSerial, podIP)
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrInitializingProxy)).
			Msgf("Error initializing proxy with certificate Common Name=%s", certCommonName)
		return
	}

	connectedProxy, exists := r.registerProxy(proxy)
	if !exists {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			if aggregatedErr := r.informTrafficPolicies(wg, connectedProxy); aggregatedErr != nil {
				wg.Done()
				log.Error().Err(aggregatedErr).Msgf("Pipy Aggregated Traffic Policies Error.")
			}
		}()
		wg.Wait()

		if connectedProxy.initError != nil {
			log.Error().Err(connectedProxy.initError).Msgf("Pipy Aggregated Traffic Policies Error.")
		}
	}
}
