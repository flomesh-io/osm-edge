package repo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/service"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/client"
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

	proxyServices, err := s.proxyRegistry.ListProxyServices(proxy)
	if err != nil {
		log.Warn().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrFetchingServiceList)).
			Msgf("Error looking up services for Sidecar with name=%s", proxy.GetName())
		return
	}

	cataloger := s.catalog
	pipyConf := new(PipyConf)

	probes(proxy, pipyConf)
	features(s, proxy, pipyConf)
	certs(s, proxy, pipyConf)
	inbound(cataloger, proxy.Identity, proxyServices, pipyConf, s.certManager.GetTrustDomain())
	outbound(cataloger, proxy.Identity, pipyConf, proxy, s)
	egress(cataloger, proxy.Identity, s, pipyConf, proxy)
	forward(cataloger, proxy.Identity, s, pipyConf, proxy)
	balance(pipyConf)
	endpoints(pipyConf, s)

	job.publishSidecarConf(s.repoClient, proxy, pipyConf)
}

func endpoints(pipyConf *PipyConf, s *Server) {
	ready := pipyConf.copyAllowedEndpoints(s.kubeController, s.proxyRegistry)
	if !ready {
		if s.retryJob != nil {
			s.retryJob()
		}
	}
}

func balance(pipyConf *PipyConf) {
	pipyConf.rebalancedOutboundClusters()
	pipyConf.rebalancedForwardClusters()
}

func egress(cataloger catalog.MeshCataloger, serviceIdentity identity.ServiceIdentity, s *Server, pipyConf *PipyConf, proxy *pipy.Proxy) bool {
	egressTrafficPolicy, egressErr := cataloger.GetEgressTrafficPolicy(serviceIdentity)
	if egressErr != nil {
		if s.retryJob != nil {
			s.retryJob()
		}
		return false
	}

	if egressTrafficPolicy != nil {
		egressDependClusters := generatePipyEgressTrafficRoutePolicy(cataloger, serviceIdentity, pipyConf,
			egressTrafficPolicy)
		if len(egressDependClusters) > 0 {
			if ready := generatePipyEgressTrafficBalancePolicy(cataloger, proxy, serviceIdentity, pipyConf,
				egressTrafficPolicy, egressDependClusters); !ready {
				if s.retryJob != nil {
					s.retryJob()
				}
				return false
			}
		}
	}
	return true
}

func forward(cataloger catalog.MeshCataloger, serviceIdentity identity.ServiceIdentity, s *Server, pipyConf *PipyConf, _ *pipy.Proxy) bool {
	egressGatewayPolicy, egressErr := cataloger.GetEgressGatewayPolicy()
	if egressErr != nil {
		if s.retryJob != nil {
			s.retryJob()
		}
		return false
	}
	if egressGatewayPolicy != nil {
		if ready := generatePipyEgressTrafficForwardPolicy(cataloger, serviceIdentity, pipyConf,
			egressGatewayPolicy); !ready {
			if s.retryJob != nil {
				s.retryJob()
			}
			return false
		}
	}
	return true
}

func outbound(cataloger catalog.MeshCataloger, serviceIdentity identity.ServiceIdentity, pipyConf *PipyConf, proxy *pipy.Proxy, s *Server) bool {
	outboundTrafficPolicy := cataloger.GetOutboundMeshTrafficPolicy(serviceIdentity)
	outboundDependClusters := generatePipyOutboundTrafficRoutePolicy(cataloger, serviceIdentity, pipyConf,
		outboundTrafficPolicy)
	if len(outboundDependClusters) > 0 {
		if ready := generatePipyOutboundTrafficBalancePolicy(cataloger, proxy, serviceIdentity, pipyConf,
			outboundTrafficPolicy, outboundDependClusters); !ready {
			if s.retryJob != nil {
				s.retryJob()
			}
			return false
		}
	}
	return true
}

func inbound(cataloger catalog.MeshCataloger, serviceIdentity identity.ServiceIdentity, proxyServices []service.MeshService, pipyConf *PipyConf, trustDomain string) {
	// Build inbound mesh route configurations. These route configurations allow
	// the services associated with this proxy to accept traffic from downstream
	// clients on allowed routes.
	inboundTrafficPolicy := cataloger.GetInboundMeshTrafficPolicy(serviceIdentity, proxyServices)
	generatePipyInboundTrafficPolicy(cataloger, serviceIdentity, pipyConf, inboundTrafficPolicy, trustDomain)
	if len(proxyServices) > 0 {
		for _, svc := range proxyServices {
			if ingressTrafficPolicy, ingressErr := cataloger.GetIngressTrafficPolicy(svc); ingressErr == nil {
				if ingressTrafficPolicy != nil {
					generatePipyIngressTrafficRoutePolicy(cataloger, serviceIdentity, pipyConf, ingressTrafficPolicy)
				}
			}
			if aclTrafficPolicy, aclErr := cataloger.GetAccessControlTrafficPolicy(svc); aclErr == nil {
				if aclTrafficPolicy != nil {
					generatePipyAccessControlTrafficRoutePolicy(cataloger, serviceIdentity, pipyConf, aclTrafficPolicy, trustDomain)
				}
			}
		}
	}
}

func certs(s *Server, proxy *pipy.Proxy, pipyConf *PipyConf) {
	if mc, ok := s.catalog.(*catalog.MeshCatalog); ok {
		meshConf := mc.GetConfigurator()
		if !(*meshConf).GetSidecarDisabledMTLS() {
			cnPrefix := proxy.Identity.String()
			if proxy.SidecarCert == nil {
				pipyConf.Certificate = nil
				sidecarCert := s.certManager.GetCertificate(cnPrefix)
				if sidecarCert == nil {
					proxy.SidecarCert = nil
				} else {
					proxy.SidecarCert = sidecarCert
				}
			}
			if proxy.SidecarCert == nil || s.certManager.ShouldRotate(proxy.SidecarCert) {
				pipyConf.Certificate = nil
				ct := proxy.PodMetadata.CreationTime
				now := time.Now()
				certValidityPeriod := s.cfg.GetServiceCertValidityPeriod()
				aliveDuration := now.Sub(ct)
				expirationDuration := (aliveDuration + certValidityPeriod/2).Round(certValidityPeriod)
				certExpiration := ct.Add(expirationDuration)
				certValidityPeriod = certExpiration.Sub(now)
				sidecarCert, certErr := s.certManager.IssueCertificate(cnPrefix, certificate.Service, certificate.ValidityDurationProvided(&certValidityPeriod))
				if certErr != nil {
					proxy.SidecarCert = nil
				} else {
					sidecarCert.Expiration = certExpiration
					proxy.SidecarCert = sidecarCert
				}
			}
		} else {
			proxy.SidecarCert = nil
		}
	}
}

func features(s *Server, proxy *pipy.Proxy, pipyConf *PipyConf) {
	if mc, ok := s.catalog.(*catalog.MeshCatalog); ok {
		meshConf := mc.GetConfigurator()
		proxy.MeshConf = meshConf
		pipyConf.setSidecarLogLevel((*meshConf).GetMeshConfig().Spec.Sidecar.LogLevel)
		pipyConf.setEnableSidecarActiveHealthChecks((*meshConf).GetFeatureFlags().EnableSidecarActiveHealthChecks)
		pipyConf.setEnableEgress((*meshConf).IsEgressEnabled())
		pipyConf.setEnablePermissiveTrafficPolicyMode((*meshConf).IsPermissiveTrafficPolicyMode())
	}
}

func probes(proxy *pipy.Proxy, pipyConf *PipyConf) {
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
}

func (job *PipyConfGeneratorJob) publishSidecarConf(repoClient *client.PipyRepoClient, proxy *pipy.Proxy, pipyConf *PipyConf) {
	pipyConf.Ts = nil
	pipyConf.Version = nil
	pipyConf.Certificate = nil
	if proxy.SidecarCert != nil {
		pipyConf.Certificate = &Certificate{
			Expiration: proxy.SidecarCert.Expiration.Format("2006-01-02 15:04:05"),
		}
	}
	bytes, jsonErr := json.Marshal(pipyConf)

	if jsonErr == nil {
		codebasePreV := proxy.ETag
		codebaseCurV := hash(bytes)
		if codebaseCurV != codebasePreV {
			codebase := fmt.Sprintf("%s/%s", osmSidecarCodebase, proxy.GetCNPrefix())
			err := repoClient.DeriveCodebase(codebase, osmCodebase)
			if err == nil {
				ts := time.Now()
				pipyConf.Ts = &ts
				version := fmt.Sprintf("%d", codebaseCurV)
				pipyConf.Version = &version
				if proxy.SidecarCert != nil {
					pipyConf.Certificate.CommonName = &proxy.SidecarCert.CommonName
					pipyConf.Certificate.CertChain = string(proxy.SidecarCert.CertChain)
					pipyConf.Certificate.PrivateKey = string(proxy.SidecarCert.PrivateKey)
					pipyConf.Certificate.IssuingCA = string(proxy.SidecarCert.IssuingCA)
				}
				bytes, _ = json.MarshalIndent(pipyConf, "", " ")
				err = repoClient.Batch(codebaseCurV, []client.Batch{
					{
						Basepath: codebase,
						Items: []client.BatchItem{
							{
								Filename: osmCodebaseConfig,
								Content:  bytes,
							},
						},
					},
				})
			}
			if err != nil {
				log.Error().Err(err)
			} else {
				proxy.ETag = codebaseCurV
			}
		}
	}
}

// JobName implementation for this job, for logging purposes
func (job *PipyConfGeneratorJob) JobName() string {
	return fmt.Sprintf("pipyJob-%s", job.proxy.GetName())
}
