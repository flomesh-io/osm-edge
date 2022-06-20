package sidecar

import (
	"time"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/identity"
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

// ProxyRegistry is an interface providing adaptiving Registries of multiple sidecars
type ProxyRegistry interface {
	ListConnectedProxies() map[certificate.CommonName]Proxy
}

// PodMetadata is a struct holding information on the Pod on which a given Envoy proxy is installed
// This struct is initialized *eventually*, when the metadata arrives via xDS.
type PodMetadata struct {
	UID            string
	Name           string
	Namespace      string
	IP             string
	ServiceAccount identity.K8sServiceAccount
	Cluster        string
	EnvoyNodeID    string
	WorkloadKind   string
	WorkloadName   string
}
