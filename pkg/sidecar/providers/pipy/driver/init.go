package driver

import (
	"github.com/openservicemesh/osm/pkg/sidecar"
)

const (
	driverName = `pipy`
)

func init() {
	sidecar.Register(driverName, new(PipySidecarDriver))
}
