package driver

import (
	"sync"

	"github.com/openservicemesh/osm/pkg/sidecar"
)

const (
	driverName = `pipy`
)

var (
	driverMutex sync.RWMutex
)

func init() {
	driverMutex.Lock()
	defer driverMutex.Unlock()

	if !sidecar.Exists(driverName) {
		sidecar.Register(driverName, new(PipySidecarDriver))
	}
}
