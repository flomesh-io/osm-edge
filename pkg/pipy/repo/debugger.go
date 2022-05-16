package repo

import (
	"time"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/envoy"
)

// GetXDSLog implements XDSDebugger interface.
func (s *Server) GetXDSLog() *map[certificate.CommonName]map[envoy.TypeURI][]time.Time {
	return nil
}
