package sidecar

import (
	"time"
)

// TypeURI is a string describing the sidecar xDS payload.
type TypeURI string

func (t TypeURI) String() string {
	return string(t)
}

// ProxyKind is the type used to define the proxy's kind
type ProxyKind string

const (
	// KindSidecar implies the proxy is a sidecar
	KindSidecar ProxyKind = "sidecar"

	// KindGateway implies the proxy is a gateway
	KindGateway ProxyKind = "gateway"
)

// Proxy is an interface providing adaptiving proxies of multiple sidecars
type Proxy interface {
	GetConnectedAt() time.Time
}
