package repo

import (
	"encoding/json"
	"fmt"
	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/pipy"
	"github.com/openservicemesh/osm/pkg/pipy/registry"
	"github.com/openservicemesh/osm/pkg/utils"
)

type PipyConfGeneratorJob struct {
	proxy     *pipy.Proxy
	xdsServer *Server

	// Optional waiter
	done chan struct{}
}

// GetDoneCh returns the channel, which when closed, indicates the job has been finished.
func (job *PipyConfGeneratorJob) GetDoneCh() <-chan struct{} {
	return job.done
}

func (job *PipyConfGeneratorJob) Run() {
	defer close(job.done)
	if job.proxy == nil {
		return
	}

	s := job.xdsServer
	proxy := job.proxy

	proxyIdentity, err := pipy.GetServiceIdentityFromProxyCertificate(proxy.GetCertificateCommonName())
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrGettingServiceIdentity)).
			Msgf("Error looking up Service Account for Sidecar with serial number=%q", proxy.GetCertificateSerialNumber())
		return
	}

	proxyServices, err := s.proxyRegistry.ListProxyServices(proxy)
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrFetchingServiceList)).
			Msgf("Error looking up services for Sidecar with serial number=%q", proxy.GetCertificateSerialNumber())
		return
	}

	cataloger := s.catalog

	pipyConf := new(PipyConf)

	if mc, ok := cataloger.(*catalog.MeshCatalog); ok {
		meshConf := *mc.GetConfigurator()
		flags := meshConf.GetFeatureFlags()
		pipyConf.SetEnableSidecarActiveHealthChecks(flags.EnableSidecarActiveHealthChecks)
		pipyConf.SetEnableEgress(meshConf.IsEgressEnabled())
		pipyConf.setEnablePermissiveTrafficPolicyMode(meshConf.IsPermissiveTrafficPolicyMode())
	}

	outboundTrafficPolicy := cataloger.GetOutboundMeshTrafficPolicy(proxyIdentity)
	outboundDependClusters := generatePipyOutboundTrafficRoutePolicy(cataloger, proxyIdentity, pipyConf,
		outboundTrafficPolicy)
	if len(outboundDependClusters) > 0 {
		outboundDependClustersReady := generatePipyOutboundTrafficBalancePolicy(cataloger, proxy, proxyIdentity, pipyConf,
			outboundTrafficPolicy, outboundDependClusters)
		if !outboundDependClustersReady {
			return
		}
	}

	egressTrafficPolicy, egressErr := cataloger.GetEgressTrafficPolicy(proxyIdentity)
	if egressErr != nil {
		return
	}

	if egressTrafficPolicy != nil {
		egressDependClusters := generatePipyEgressTrafficRoutePolicy(cataloger, proxyIdentity, pipyConf,
			egressTrafficPolicy)
		if len(egressDependClusters) > 0 {
			egressDependClustersReady := generatePipyEgressTrafficBalancePolicy(cataloger, proxy, proxyIdentity, pipyConf,
				egressTrafficPolicy, egressDependClusters)
			if !egressDependClustersReady {
				return
			}
		}
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

	if proxy.PodMetadata != nil {
		if len(proxy.PodMetadata.StartupProbes) > 0 {
			for idx := range proxy.PodMetadata.StartupProbes {
				pipyConf.Spec.Probes.StartupProbes = append(pipyConf.Spec.Probes.StartupProbes, proxy.PodMetadata.StartupProbes[idx])
			}
		}
		if len(proxy.PodMetadata.LivenessProbes) > 0 {
			for idx := range proxy.PodMetadata.LivenessProbes {
				pipyConf.Spec.Probes.LivenessProbes = append(pipyConf.Spec.Probes.LivenessProbes, proxy.PodMetadata.LivenessProbes[idx])
			}
		}
		if len(proxy.PodMetadata.ReadinessProbes) > 0 {
			for idx := range proxy.PodMetadata.ReadinessProbes {
				pipyConf.Spec.Probes.ReadinessProbes = append(pipyConf.Spec.Probes.ReadinessProbes, proxy.PodMetadata.ReadinessProbes[idx])
			}
		}
	}

	pipyConf.RebalanceOutboundClusters()
	pipyConf.CopyAllowedEndpoints()

	if bytes, jsonErr := json.Marshal(pipyConf); jsonErr == nil {
		if hashCode, hashErr := utils.HashFromString(string(bytes)); hashErr == nil {
			pipyConf.bytes = bytes
			proxy.SetCodebase(pipyConf, fmt.Sprintf("%d", hashCode), true)
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

func RefreshPipyConf(proxy *pipy.Proxy, pipyConf *PipyConf) string {
	if pipyConf.allowedEndpointsV != registry.CachedMeshPodsV {
		prevAllowedEndpointsV := pipyConf.allowedEndpointsV
		pipyConf.CopyAllowedEndpoints()
		if bytes, jsonErr := json.Marshal(pipyConf); jsonErr == nil {
			if hashCode, hashErr := utils.HashFromString(string(bytes)); hashErr == nil {
				pipyConf.bytes = bytes
				etag := fmt.Sprintf("%d", hashCode)
				proxy.SetCodebase(pipyConf, etag, true)
				return etag
			}
		}
		pipyConf.allowedEndpointsV = prevAllowedEndpointsV
	}
	return ""
}
