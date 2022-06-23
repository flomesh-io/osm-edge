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
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/ads"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/registry"
)

const (
	specificProxyQueryKey = "proxy"
	proxyConfigQueryKey   = "cfg"
)

func (sd EnvoySidecarDriver) configDebug(proxyRegistry *registry.ProxyRegistry, xdsServer *ads.Server) {
	for uri, handler := range map[string]http.Handler{
		"/debug/proxy": sd.getProxies(proxyRegistry),
		"/debug/xds":   sd.getXDSHandler(xdsServer),
	} {
		sd.ctx.DebugHandlers[uri] = handler
	}
}

func (sd EnvoySidecarDriver) getProxies(proxyRegistry *registry.ProxyRegistry) http.Handler {
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

func (sd EnvoySidecarDriver) getConfigDump(cn certificate.CommonName, w http.ResponseWriter) {
	pod, err := envoy.GetPodFromCertificate(cn, sd.ctx.MeshCatalog.GetKubeController())
	if err != nil {
		log.Error().Err(err).Msgf("Error getting Pod from certificate with CN=%s", cn)
	}
	w.Header().Set("Content-Type", "application/json")
	envoyConfig := sd.getEnvoyConfig(pod, "config_dump")
	_, _ = fmt.Fprintf(w, "%s", envoyConfig)
}

func (sd EnvoySidecarDriver) getProxy(cn certificate.CommonName, w http.ResponseWriter) {
	pod, err := envoy.GetPodFromCertificate(cn, sd.ctx.MeshCatalog.GetKubeController())
	if err != nil {
		log.Error().Err(err).Msgf("Error getting Pod from certificate with CN=%s", cn)
	}
	w.Header().Set("Content-Type", "application/json")
	envoyConfig := sd.getEnvoyConfig(pod, "certs")
	_, _ = fmt.Fprintf(w, "%s", envoyConfig)
}

func (sd EnvoySidecarDriver) getEnvoyConfig(pod *v1.Pod, url string) string {
	log.Debug().Msgf("Getting Envoy config on Pod with UID=%s", pod.ObjectMeta.UID)

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

func (sd EnvoySidecarDriver) getXDSHandler(xdsServer *ads.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xdsLog := xdsServer.GetXDSLog()

		var proxies []string
		for proxyCN := range *xdsLog {
			proxies = append(proxies, proxyCN.String())
		}

		sort.Strings(proxies)

		for _, proxyCN := range proxies {
			xdsTypeWithTimestamps := (*xdsLog)[certificate.CommonName(proxyCN)]
			_, _ = fmt.Fprintf(w, "---[ %s\n", proxyCN)

			var xdsTypes []string
			for xdsType := range xdsTypeWithTimestamps {
				xdsTypes = append(xdsTypes, xdsType.String())
			}

			sort.Strings(xdsTypes)

			for _, xdsType := range xdsTypes {
				timeStamps := xdsTypeWithTimestamps[envoy.TypeURI(xdsType)]

				_, _ = fmt.Fprintf(w, "\t %s (%d):\n", xdsType, len(timeStamps))

				sort.Slice(timeStamps, func(i, j int) bool {
					return timeStamps[i].After(timeStamps[j])
				})
				for _, timeStamp := range timeStamps {
					_, _ = fmt.Fprintf(w, "\t\t%+v (%+v ago)\n", timeStamp, time.Since(timeStamp))
				}
				_, _ = fmt.Fprint(w, "\n")
			}
			_, _ = fmt.Fprint(w, "\n")
		}
	})
}
