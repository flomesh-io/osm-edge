package informers

import (
	"errors"
	"time"

	"k8s.io/client-go/tools/cache"
)

// InformerKey stores the different Informers we keep for K8s resources
type InformerKey string

const (
	// InformerKeyNamespace is the InformerKey for a Namespace informer
	InformerKeyNamespace InformerKey = "Namespace"
	// InformerKeyService is the InformerKey for a Service informer
	InformerKeyService InformerKey = "Service"
	// InformerKeyPod is the InformerKey for a Pod informer
	InformerKeyPod InformerKey = "Pod"
	// InformerKeyEndpoints is the InformerKey for a Endpoints informer
	InformerKeyEndpoints InformerKey = "Endpoints"
	// InformerKeyServiceAccount is the InformerKey for a ServiceAccount informer
	InformerKeyServiceAccount InformerKey = "ServiceAccount"

	// InformerKeyTrafficSplit is the InformerKey for a TrafficSplit informer
	InformerKeyTrafficSplit InformerKey = "TrafficSplit"
	// InformerKeyTrafficTarget is the InformerKey for a TrafficTarget informer
	InformerKeyTrafficTarget InformerKey = "TrafficTarget"
	// InformerKeyHTTPRouteGroup is the InformerKey for a HTTPRouteGroup informer
	InformerKeyHTTPRouteGroup InformerKey = "HTTPRouteGroup"
	// InformerKeyTCPRoute is the InformerKey for a TCPRoute informer
	InformerKeyTCPRoute InformerKey = "TCPRoute"

	// InformerKeyMeshConfig is the InformerKey for a MeshConfig informer
	InformerKeyMeshConfig InformerKey = "MeshConfig"
	// InformerKeyMeshRootCertificate is the InformerKey for a MeshRootCertificate informer
	InformerKeyMeshRootCertificate InformerKey = "MeshRootCertificate"

	// InformerKeyEgress is the InformerKey for an Egress informer
	InformerKeyEgress InformerKey = "Egress"
	// InformerKeyEgressGateway is the InformerKey for an EgressGateway informer
	InformerKeyEgressGateway InformerKey = "EgressGateway"
	// InformerKeyIngressBackend is the InformerKey for a IngressBackend informer
	InformerKeyIngressBackend InformerKey = "IngressBackend"
	// InformerKeyUpstreamTrafficSetting is the InformerKey for a UpstreamTrafficSetting informer
	InformerKeyUpstreamTrafficSetting InformerKey = "UpstreamTrafficSetting"
	// InformerKeyRetry is the InformerKey for a Retry informer
	InformerKeyRetry InformerKey = "Retry"
	// InformerKeyAccessControl is the InformerKey for a AccessControl informer
	InformerKeyAccessControl InformerKey = "AccessControl"
	// InformerKeyAccessCert is the InformerKey for a AccessCert informer
	InformerKeyAccessCert InformerKey = "AccessCert"
	// InformerKeyServiceImport is the InformerKey for a ServiceImport informer
	InformerKeyServiceImport InformerKey = "ServiceImport"
	// InformerKeyServiceExport is the InformerKey for a ServiceExport informer
	InformerKeyServiceExport InformerKey = "ServiceExport"
	// InformerKeyGlobalTrafficPolicy is the InformerKey for a GlobalTrafficPolicy informer
	InformerKeyGlobalTrafficPolicy InformerKey = "GlobalTrafficPolicy"
	// InformerKeyIngressClass is the InformerKey for a IngressClass informer
	InformerKeyIngressClass InformerKey = "IngressClass"

	// InformerKeyPlugin is the InformerKey for a Plugin informer
	InformerKeyPlugin InformerKey = "Plugin"
	// InformerKeyPluginChain is the InformerKey for a PluginChain informer
	InformerKeyPluginChain InformerKey = "PluginChain"
	// InformerKeyPluginConfig is the InformerKey for a PluginConfig informer
	InformerKeyPluginConfig InformerKey = "PluginConfig"
)

const (
	// DefaultKubeEventResyncInterval is the default resync interval for k8s events
	// This is set to 0 because we do not need resyncs from k8s client, and have our
	// own Ticker to turn on periodic resyncs.
	DefaultKubeEventResyncInterval = 0 * time.Second
)

var (
	errInitInformers = errors.New("informer not initialized")
	errSyncingCaches = errors.New("failed initial cache sync for informers")
)

// InformerCollection is an abstraction around a set of informers
// initialized with the clients stored in its fields. This data
// type should only be passed around as a pointer
type InformerCollection struct {
	informers map[InformerKey]cache.SharedIndexInformer
	meshName  string
}
