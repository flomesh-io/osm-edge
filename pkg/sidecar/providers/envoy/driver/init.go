package driver

import (
	"github.com/openservicemesh/osm/pkg/sidecar"
)

const (
	driverName = `envoy`
)

func init() {
	sidecar.Register(driverName, new(EnvoySidecarDriver))
}
