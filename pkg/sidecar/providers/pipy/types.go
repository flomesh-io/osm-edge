// Package pipy implements utility routines related to Pipy proxy, and models an instance of a proxy
// to be able to generate XDS configurations for it.
package pipy

import (
	"github.com/openservicemesh/osm/pkg/logger"
)

var (
	log = logger.New("flomesh-pipy")
)

// ProxyKind is the type used to define the proxy's kind
type ProxyKind string

const (
	// KindSidecar implies the proxy is a sidecar
	KindSidecar ProxyKind = "sidecar"

	// KindGateway implies the proxy is a gateway
	KindGateway ProxyKind = "gateway"
)
