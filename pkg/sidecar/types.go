package sidecar

import (
	"github.com/openservicemesh/osm/pkg/certificate"
	"time"
)

// Proxy is an interface providing adaptiving proxies of multiple sidecars
type Proxy interface {
	GetConnectedAt() time.Time
}

// ProxyRegistry is an interface providing adaptiving Registries of multiple sidecars
type ProxyRegistry interface {
	ListConnectedProxies() map[certificate.CommonName]Proxy
}
