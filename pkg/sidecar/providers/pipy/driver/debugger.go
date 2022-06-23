package driver

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/repo"
)

const (
	specificProxyQueryKey = "proxy"
	proxyConfigQueryKey   = "cfg"
)

func (sd PipySidecarDriver) configDebug(proxyRegistry *registry.ProxyRegistry, _ *repo.Server) {
	for uri, handler := range map[string]http.Handler{
		"/debug/proxy": sd.getProxies(proxyRegistry),
	} {
		sd.ctx.DebugHandlers[uri] = handler
	}
}

func (sd PipySidecarDriver) getProxies(proxyRegistry *registry.ProxyRegistry) http.Handler {
	// This function is needed to convert the list of connected proxies to
	// the type (map) required by the printProxies function.
	listConnected := func() map[certificate.CommonName]time.Time {
		proxies := make(map[certificate.CommonName]time.Time)
		for cn, proxy := range proxyRegistry.ListConnectedProxies() {
			proxies[cn] = proxy.GetConnectedAt()
		}
		return proxies
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if proxyConfigDump, ok := r.URL.Query()[proxyConfigQueryKey]; ok {
			sd.getConfigDump(certificate.CommonName(proxyConfigDump[0]), w)
		} else if specificProxy, ok := r.URL.Query()[specificProxyQueryKey]; ok {
			sd.getProxy(certificate.CommonName(specificProxy[0]), w)
		} else {
			printProxies(w, listConnected(), "Connected")
		}
	})
}

func (sd PipySidecarDriver) getConfigDump(cn certificate.CommonName, w http.ResponseWriter) {
	pod, err := pipy.GetPodFromCertificate(cn, sd.ctx.MeshCatalog.GetKubeController())
	if err != nil {
		log.Error().Err(err).Msgf("Error getting Pod from certificate with CN=%s", cn)
	}
	w.Header().Set("Content-Type", "application/json")
	sidecarConfig := sd.getSidecarConfig(pod, "config_dump")
	_, _ = fmt.Fprintf(w, "%s", sidecarConfig)
}

func (sd PipySidecarDriver) getProxy(cn certificate.CommonName, w http.ResponseWriter) {
	pod, err := pipy.GetPodFromCertificate(cn, sd.ctx.MeshCatalog.GetKubeController())
	if err != nil {
		log.Error().Err(err).Msgf("Error getting Pod from certificate with CN=%s", cn)
	}
	w.Header().Set("Content-Type", "application/json")
	sidecarConfig := sd.getSidecarConfig(pod, "certs")
	_, _ = fmt.Fprintf(w, "%s", sidecarConfig)
}

func (sd PipySidecarDriver) getSidecarConfig(pod *v1.Pod, url string) string {
	log.Debug().Msgf("Getting Pipy config on Pod with UID=%s", pod.ObjectMeta.UID)

	minPort := 16000
	maxPort := 18000

	// #nosec G404
	portFwdRequest := portForward{
		Pod:       pod,
		LocalPort: rand.Intn(maxPort-minPort) + minPort,
		PodPort:   15000,
		Stop:      make(chan struct{}),
		Ready:     make(chan struct{}),
	}
	go sd.forwardPort(portFwdRequest)

	<-portFwdRequest.Ready

	client := &http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://%s:%d/%s", "localhost", portFwdRequest.LocalPort, url))
	if err != nil {
		log.Error().Err(err).Msgf("Error getting Pod with UID=%s", pod.ObjectMeta.UID)
		return fmt.Sprintf("Error: %s", err)
	}

	defer func() {
		portFwdRequest.Stop <- struct{}{}
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("Error getting Envoy config on Pod with UID=%s; HTTP Error %d", pod.ObjectMeta.UID, resp.StatusCode)
		portFwdRequest.Stop <- struct{}{}
		return fmt.Sprintf("Error: %s", err)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting Pod with UID=%s", pod.ObjectMeta.UID)
		return fmt.Sprintf("Error: %s", err)
	}

	return string(bodyBytes)
}

func printProxies(w http.ResponseWriter, proxies map[certificate.CommonName]time.Time, category string) {
	var commonNames []string
	for cn := range proxies {
		commonNames = append(commonNames, cn.String())
	}

	sort.Strings(commonNames)

	_, _ = fmt.Fprintf(w, "<h1>%s Proxies (%d):</h1>", category, len(proxies))
	_, _ = fmt.Fprint(w, `<table>`)
	_, _ = fmt.Fprint(w, "<tr><td>#</td><td>Envoy's certificate CN</td><td>Connected At</td><td>How long ago</td><td>tools</td></tr>")
	for idx, cn := range commonNames {
		ts := proxies[certificate.CommonName(cn)]
		_, _ = fmt.Fprintf(w, `<tr><td>%d:</td><td>%s</td><td>%+v</td><td>(%+v ago)</td><td><a href="/debug/proxy?%s=%s">certs</a></td><td><a href="/debug/proxy?%s=%s">cfg</a></td></tr>`,
			idx, cn, ts, time.Since(ts), specificProxyQueryKey, cn, proxyConfigQueryKey, cn)
	}
	_, _ = fmt.Fprint(w, `</table>`)
}
