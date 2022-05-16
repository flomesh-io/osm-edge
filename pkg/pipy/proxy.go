package pipy

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/openservicemesh/osm/pkg/certificate"
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

	// The Serial Number of the certificate used for Sidecar to XDS communication.
	xDSCertificateSerialNumber certificate.SerialNumber

	net.Addr

	// The time this Proxy connected to the OSM control plane
	connectedAt time.Time

	codebaseReady       bool
	latestRepoResources map[RepoResource]*RepoResourceV
	latestRepoCodebaseV string
	reportRepoCodebaseV string

	codebaseLock sync.RWMutex
	codebaseConf interface{}

	// hash is based on CommonName
	hash uint64

	// kind is the proxy's kind (ex. sidecar, gateway)
	kind ProxyKind

	// Records metadata around the Kubernetes Pod on which this Sidecar Proxy is installed.
	// This could be nil if the Sidecar is not operating in a Kubernetes cluster (VM for example)
	// NOTE: This field may be not be set at the time Proxy struct is initialized. This would
	// eventually be set when the metadata arrives via the xDS protocol.
	PodMetadata *PodMetadata
}

func (p *Proxy) GetCodebase() (interface{}, string, bool) {
	p.codebaseLock.RLock()
	defer p.codebaseLock.RUnlock()
	return p.codebaseConf, p.latestRepoCodebaseV, p.codebaseReady
}

func (p *Proxy) SetCodebase(codebaseConf interface{}, latestRepoCodebaseV string, codebaseReady bool) {
	p.codebaseLock.Lock()
	defer p.codebaseLock.Unlock()
	p.codebaseConf = codebaseConf
	p.latestRepoCodebaseV = latestRepoCodebaseV
	p.codebaseReady = codebaseReady
}

func (p *Proxy) String() string {
	return fmt.Sprintf("[Serial=%s], [Pod metadata=%s]", p.xDSCertificateSerialNumber, p.PodMetadataString())
}

// PodMetadata is a struct holding information on the Pod on which a given Sidecar proxy is installed
// This struct is initialized *eventually*, when the metadata arrives via xDS.
type PodMetadata struct {
	UID             string
	Name            string
	Namespace       string
	IP              string
	ServiceAccount  identity.K8sServiceAccount
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
	return p.xDSCertificateSerialNumber
}

// GetHash returns the proxy hash based on its xDSCertificateCommonName
func (p *Proxy) GetHash() uint64 {
	return p.hash
}

// GetConnectedAt returns the timestamp of when the given proxy connected to the control plane.
func (p *Proxy) GetConnectedAt() time.Time {
	return p.connectedAt
}

// GetIP returns the IP address of the Sidecar proxy connected to xDS.
func (p *Proxy) GetIP() net.Addr {
	return p.Addr
}

func (p *Proxy) GetLatestRepoResources(repoResource RepoResource) (*RepoResourceV, bool) {
	repoResourceV, ok := p.latestRepoResources[repoResource]
	return repoResourceV, ok
}

func (p *Proxy) SetLatestRepoResources(repoResource RepoResource, resource *RepoResourceV) {
	if hashcode, err := utils.HashFromString(resource.Content); err == nil {
		resource.Version = fmt.Sprintf("%d", hashcode)
	}
	p.latestRepoResources[repoResource] = resource
}

func (p *Proxy) SetReportRepoCodebaseV(reportRepoCodebaseV string) {
	p.reportRepoCodebaseV = reportRepoCodebaseV
}

// Kind return the proxy's kind
func (p *Proxy) Kind() ProxyKind {
	return p.kind
}

// NewProxy creates a new instance of an Sidecar proxy connected to the xDS servers.
func NewProxy(certCommonName certificate.CommonName, certSerialNumber certificate.SerialNumber, ip net.Addr) (*Proxy, error) {
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
		certificateCommonName:      certCommonName,
		xDSCertificateSerialNumber: certSerialNumber,
		UUID:                       cnMeta.ProxyUUID,

		Addr: ip,

		connectedAt: time.Now(),
		hash:        hash,

		latestRepoResources: make(map[RepoResource]*RepoResourceV),

		kind: cnMeta.ProxyKind,
	}, nil
}

// NewXDSCertCommonName returns a newly generated CommonName for a certificate of the form: <ProxyUUID>.<kind>.<serviceAccount>.<namespace>
func NewXDSCertCommonName(proxyUUID uuid.UUID, kind ProxyKind, serviceAccount, namespace string) certificate.CommonName {
	return certificate.CommonName(fmt.Sprintf("%s.%s.%s.%s.%s", proxyUUID.String(), kind, serviceAccount, namespace, identity.ClusterLocalTrustDomain))
}
