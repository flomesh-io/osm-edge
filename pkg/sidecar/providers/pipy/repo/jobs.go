package repo

import (
	"encoding/json"
	"fmt"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/client"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
)

// PipyConfGeneratorJob is the job to generate pipy policy json
type PipyConfGeneratorJob struct {
	proxy      *pipy.Proxy
	repoServer *Server

	// Optional waiter
	done chan struct{}
}

// GetDoneCh returns the channel, which when closed, indicates the job has been finished.
func (job *PipyConfGeneratorJob) GetDoneCh() <-chan struct{} {
	return job.done
}

// Run is the logic unit of job
func (job *PipyConfGeneratorJob) Run() {
	defer close(job.done)
	if job.proxy == nil {
		return
	}

	s := job.repoServer
	proxy := job.proxy

	proxy.Mutex.Lock()
	defer proxy.Mutex.Unlock()

	proxyIdentity, err := pipy.GetServiceIdentityFromProxyCertificate(proxy.GetCertificateCommonName())
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrGettingServiceIdentity)).
			Msgf("Error looking up Service Account for Sidecar with serial number=%q", proxy.GetCertificateSerialNumber())
		return
	}
	proxy.ProxyIdentity = proxyIdentity

	proxyServices, err := s.proxyRegistry.ListProxyServices(proxy)
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrFetchingServiceList)).
			Msgf("Error looking up services for Sidecar with serial number=%q", proxy.GetCertificateSerialNumber())
		return
	}

	cataloger := s.catalog
	pipyConf := new(PipyConf)

	if proxy.PodMetadata != nil {
		if len(proxy.PodMetadata.StartupProbes) > 0 {
			for idx := range proxy.PodMetadata.StartupProbes {
				pipyConf.Spec.Probes.StartupProbes = append(pipyConf.Spec.Probes.StartupProbes, *proxy.PodMetadata.StartupProbes[idx])
			}
		}
		if len(proxy.PodMetadata.LivenessProbes) > 0 {
			for idx := range proxy.PodMetadata.LivenessProbes {
				pipyConf.Spec.Probes.LivenessProbes = append(pipyConf.Spec.Probes.LivenessProbes, *proxy.PodMetadata.LivenessProbes[idx])
			}
		}
		if len(proxy.PodMetadata.ReadinessProbes) > 0 {
			for idx := range proxy.PodMetadata.ReadinessProbes {
				pipyConf.Spec.Probes.ReadinessProbes = append(pipyConf.Spec.Probes.ReadinessProbes, *proxy.PodMetadata.ReadinessProbes[idx])
			}
		}
	}

	if mc, ok := s.catalog.(*catalog.MeshCatalog); ok {
		meshConf := mc.GetConfigurator()
		proxy.MeshConf = meshConf
		pipyConf.setSidecarLogLevel((*meshConf).GetSidecarLogLevel())
		pipyConf.setEnableSidecarActiveHealthChecks((*meshConf).GetFeatureFlags().EnableSidecarActiveHealthChecks)
		pipyConf.setEnableEgress((*meshConf).IsEgressEnabled())
		pipyConf.setEnablePermissiveTrafficPolicyMode((*meshConf).IsPermissiveTrafficPolicyMode())
		if !(*meshConf).GetSidecarDisabledMTLS() {
			if proxy.SidecarCert == nil || proxy.SidecarCert.ShouldRotate() {
				pipyConf.Certificate = nil
				sidecarCert, certErr := s.certManager.IssueCertificate(certificate.CommonName(proxyIdentity), s.cfg.GetServiceCertValidityPeriod())
				if certErr != nil {
					log.Error().Err(certErr).Str("proxy", proxy.String()).Msgf("Error issuing a certificate for proxy")
					proxy.SidecarCert = nil
				} else {
					proxy.SidecarCert = sidecarCert
				}
			}
		} else {
			proxy.SidecarCert = nil
		}
	}
	if proxy.SidecarCert != nil {
		pipyConf.Certificate = &Certificate{
			CommonName:   proxy.SidecarCert.CommonName,
			SerialNumber: proxy.SidecarCert.SerialNumber,
			Expiration:   proxy.SidecarCert.Expiration.Format("2006-01-02 15:04:05"),
			CertChain:    string(proxy.SidecarCert.CertChain),
			PrivateKey:   string(proxy.SidecarCert.PrivateKey),
			IssuingCA:    string(proxy.SidecarCert.IssuingCA),
		}
	} else {
		pipyConf.Certificate = nil
	}

	// Build inbound mesh route configurations. These route configurations allow
	// the services associated with this proxy to accept traffic from downstream
	// clients on allowed routes.
	inboundTrafficPolicy := cataloger.GetInboundMeshTrafficPolicy(proxyIdentity, proxyServices)
	generatePipyInboundTrafficPolicy(cataloger, proxyIdentity, pipyConf, inboundTrafficPolicy)
	if len(proxyServices) > 0 {
		for _, svc := range proxyServices {
			if ingressTrafficPolicy, ingressErr := cataloger.GetIngressTrafficPolicy(svc); ingressErr == nil {
				if ingressTrafficPolicy != nil {
					generatePipyIngressTrafficRoutePolicy(cataloger, proxyIdentity, pipyConf,
						ingressTrafficPolicy)
				}
			}
		}
	}

	outboundTrafficPolicy := cataloger.GetOutboundMeshTrafficPolicy(proxyIdentity)
	outboundDependClusters := generatePipyOutboundTrafficRoutePolicy(cataloger, proxyIdentity, pipyConf,
		outboundTrafficPolicy)
	if len(outboundDependClusters) > 0 {
		if ready := generatePipyOutboundTrafficBalancePolicy(cataloger, proxy, proxyIdentity, pipyConf,
			outboundTrafficPolicy, outboundDependClusters); !ready {
			if s.retryJob != nil {
				s.retryJob()
			}
			return
		}
	}

	egressTrafficPolicy, egressErr := cataloger.GetEgressTrafficPolicy(proxyIdentity)
	if egressErr != nil {
		if s.retryJob != nil {
			s.retryJob()
		}
		return
	}

	if egressTrafficPolicy != nil {
		egressDependClusters := generatePipyEgressTrafficRoutePolicy(cataloger, proxyIdentity, pipyConf,
			egressTrafficPolicy)
		if len(egressDependClusters) > 0 {
			if ready := generatePipyEgressTrafficBalancePolicy(cataloger, proxy, proxyIdentity, pipyConf,
				egressTrafficPolicy, egressDependClusters); !ready {
				if s.retryJob != nil {
					s.retryJob()
				}
				return
			}
		}
	}

	pipyConf.rebalanceOutboundClusters()

	ready := pipyConf.copyAllowedEndpoints(s.kubeController)
	if !ready {
		if s.retryJob != nil {
			s.retryJob()
		}
	}

	job.publishSidecarConf(s.repoClient, s.proxyRegistry, proxy, pipyConf)
}

func (job *PipyConfGeneratorJob) publishSidecarConf(repoClient *client.PipyRepoClient, proxyRegistry *registry.ProxyRegistry, proxy *pipy.Proxy, pipyConf *PipyConf) {
	if bytes, jsonErr := json.MarshalIndent(pipyConf, "", " "); jsonErr == nil {
		codebasePreV := int64(0)
		certCommonName := proxy.GetCertificateCommonName()
		if etag, ok := proxyRegistry.PodCNtoETag.Load(certCommonName); ok {
			codebasePreV = etag.(int64)
		}
		codebaseCurV := hash(bytes)
		if codebaseCurV != codebasePreV {
			err := repoClient.DeriveCodebase(fmt.Sprintf("%s/%s", osmSidecarCodebase, proxy.GetCertificateCommonName()), osmCodebase)
			if err == nil {
				err = repoClient.Batch([]client.Batch{
					{
						Basepath: fmt.Sprintf("%s/%s", osmSidecarCodebase, proxy.GetCertificateCommonName()),
						Items: []client.BatchItem{
							{
								Filename: "pipy.json",
								Content:  bytes,
							},
						},
					},
				})
			}
			if err != nil {
				log.Error().Err(err)
			} else {
				proxyRegistry.PodCNtoETag.Store(certCommonName, codebaseCurV)
			}
		}
	}
}

// JobName implementation for this job, for logging purposes
func (job *PipyConfGeneratorJob) JobName() string {
	return fmt.Sprintf("pipyJob-%s", job.proxy.GetCertificateSerialNumber())
}

// Hash implementation for this job to hash into the worker queues
func (job *PipyConfGeneratorJob) Hash() uint64 {
	// Uses proxy hash to always serialize work for the same proxy to the same worker,
	// this avoid out-of-order mishandling of sidecar updates by multiple workers
	return job.proxy.GetHash()
}
