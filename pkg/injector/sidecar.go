package injector

import "github.com/openservicemesh/osm/pkg/constants"

var (
	sidecarType = constants.PipySidecar
)

func GetSidecarType() string {
	return sidecarType
}

func SetSidecarType(sidecar string) {
	sidecarType = sidecar
}
