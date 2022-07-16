package pipy

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/utils"
)

// Proxy is a representation of an Sidecar proxy connected to the xDS server.
// This should at some point have a 1:1 match to an Endpoint (which is a member of a meshed service).
type Proxy struct {
	// The Subject Common Name of the certificate used for Sidecar to communication.
	certificateCommonName certificate.CommonName

	// UUID of the proxy
	uuid.UUID

	// IP of the pod
	PodIP string

	// The time this Proxy connected to the OSM control plane
	connectedAt time.Time

	// The Serial Number of the certificate used for Sidecar.
	CertificateSerialNumber certificate.SerialNumber

	// hash is based on CommonName
	hash uint64

	// kind is the proxy's kind (ex. sidecar, gateway)
	kind ProxyKind

	// Records metadata around the Kubernetes Pod on which this Sidecar Proxy is installed.
	// This could be nil if the Sidecar is not operating in a Kubernetes cluster (VM for example)
	// NOTE: This field may be not be set at the time Proxy struct is initialized. This would
	// eventually be set when the metadata arrives via the xDS protocol.
	PodMetadata *PodMetadata

	MeshConf    *configurator.Configurator
	SidecarCert *certificate.Certificate

	Mutex sync.RWMutex
	Quit  chan bool
}

func (p *Proxy) String() string {
	return fmt.Sprintf("[Serial=%s], [Pod metadata=%s]", p.CertificateSerialNumber, p.PodMetadataString())
}

// PodMetadata is a struct holding information on the Pod on which a given Sidecar proxy is installed
// This struct is initialized *eventually*, when the metadata arrives via xDS.
type PodMetadata struct {
	UID             string
	Name            string
	Namespace       string
	IP              string
	ServiceAccount  identity.K8sServiceAccount
	CreationTime    time.Time
	Cluster         string
	SidecarNodeID   string
	WorkloadKind    string
	WorkloadName    string
	ReadinessProbes []*v1.Probe
	LivenessProbes  []*v1.Probe
	StartupProbes   []*v1.Probe
}

// HasPodMetadata answers the question - has the Pod metadata been recorded for the given Sidecar proxy
func (p *Proxy) HasPodMetadata() bool {
	return p.PodMetadata != nil
}

// StatsHeaders returns the headers required for SMI metrics
func (p *Proxy) StatsHeaders() map[string]string {
	unknown := "unknown"
	podName := unknown
	podNamespace := unknown
	podControllerKind := unknown
	podControllerName := unknown

	if p.PodMetadata != nil {
		if len(p.PodMetadata.Name) > 0 {
			podName = p.PodMetadata.Name
		}
		if len(p.PodMetadata.Namespace) > 0 {
			podNamespace = p.PodMetadata.Namespace
		}
		if len(p.PodMetadata.WorkloadKind) > 0 {
			podControllerKind = p.PodMetadata.WorkloadKind
		}
		if len(p.PodMetadata.WorkloadName) > 0 {
			podControllerName = p.PodMetadata.WorkloadName
		}
	}

	// Assume ReplicaSets are controlled by a Deployment unless their names
	// do not contain a hyphen. This aligns with the behavior of the
	// Prometheus config in the OSM Helm chart.
	if podControllerKind == "ReplicaSet" {
		if hyp := strings.LastIndex(podControllerName, "-"); hyp >= 0 {
			podControllerKind = "Deployment"
			podControllerName = podControllerName[:hyp]
		}
	}

	return map[string]string{
		"osm-stats-pod":       podName,
		"osm-stats-namespace": podNamespace,
		"osm-stats-kind":      podControllerKind,
		"osm-stats-name":      podControllerName,
	}
}

// PodMetadataString returns relevant pod metadata as a string
func (p *Proxy) PodMetadataString() string {
	if p.PodMetadata == nil {
		return ""
	}
	return fmt.Sprintf("UID=%s, Namespace=%s, Name=%s, ServiceAccount=%s", p.PodMetadata.UID, p.PodMetadata.Namespace, p.PodMetadata.Name, p.PodMetadata.ServiceAccount.Name)
}

// GetCertificateCommonName returns the Subject Common Name from the mTLS certificate of the Sidecar proxy connected to xDS.
func (p *Proxy) GetCertificateCommonName() certificate.CommonName {
	return p.certificateCommonName
}

// GetCertificateSerialNumber returns the Serial Number of the certificate for the connected Sidecar proxy.
func (p *Proxy) GetCertificateSerialNumber() certificate.SerialNumber {
	return p.CertificateSerialNumber
}

// GetHash returns the proxy hash based on its xDSCertificateCommonName
func (p *Proxy) GetHash() uint64 {
	return p.hash
}

// Kind return the proxy's kind
func (p *Proxy) Kind() ProxyKind {
	return p.kind
}

// GetConnectedAt returns the timestamp of when the given proxy connected to the control plane.
func (p *Proxy) GetConnectedAt() time.Time {
	return p.connectedAt
}

// NewProxy creates a new instance of an Sidecar proxy connected to the servers.
func NewProxy(certCommonName certificate.CommonName, certSerialNumber certificate.SerialNumber, podIP string) (*Proxy, error) {
	// Get CommonName hash for this proxy
	hash, err := utils.HashFromString(certCommonName.String())
	if err != nil {
		log.Warn().Err(err).Msgf("Failed to get hash for proxy serial %s, 0 hash will be used", certSerialNumber)
	}

	cnMeta, err := getCertificateCommonNameMeta(certCommonName)
	if err != nil {
		return nil, ErrInvalidCertificateCN
	}

	return &Proxy{
		certificateCommonName:   certCommonName,
		CertificateSerialNumber: certSerialNumber,
		UUID:                    cnMeta.ProxyUUID,
		PodIP:                   podIP,
		connectedAt:             time.Now(),
		hash:                    hash,
		kind:                    cnMeta.ProxyKind,
	}, nil
}

// NewCertCommonName returns a newly generated CommonName for a certificate of the form: <ProxyUUID>.<kind>.<serviceAccount>.<namespace>
func NewCertCommonName(proxyUUID uuid.UUID, kind ProxyKind, serviceAccount, namespace string) certificate.CommonName {
	return certificate.CommonName(fmt.Sprintf("%s.%s.%s.%s.%s", proxyUUID.String(), kind, serviceAccount, namespace, identity.ClusterLocalTrustDomain))
}
