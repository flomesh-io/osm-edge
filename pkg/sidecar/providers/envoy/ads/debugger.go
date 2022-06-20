package ads

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/jinzhu/copier"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy"
)

// GetXDSLog implements XDSDebugger interface and a log of the XDS responses sent to Envoy proxies.
func (s *Server) GetXDSLog() *map[certificate.CommonName]map[envoy.TypeURI][]time.Time {
	var logsCopy map[certificate.CommonName]map[envoy.TypeURI][]time.Time
	var err error

	s.withXdsLogMutex(func() {
		// Making a copy to avoid debugger potential reads while writes are happening from XDS routines
		err = copier.Copy(&logsCopy, &s.xdsLog)
	})

	if err != nil {
		log.Error().Err(err).Msgf("Failed to copy xdsLogMap")
	}

	return &logsCopy
}

// GetDebugHandlers implements ProxyDebugger interface
func (s *Server) GetDebugHandlers() map[string]http.Handler {
	handlers := map[string]http.Handler{
		"/xds": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xdsLog := s.GetXDSLog()

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
		}),
	}
	return handlers
}
