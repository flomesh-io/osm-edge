package driver

import (
	"sync"

	"github.com/openservicemesh/osm/pkg/sidecar"
)

const (
	driverName = `envoy`
)

var (
	driverMutex sync.RWMutex
)

func init() {
	driverMutex.Lock()
	defer driverMutex.Unlock()

	if !sidecar.Exists(driverName) {
		sidecar.Register(`envoy`, new(EnvoySidecarDriver))
	}
}
