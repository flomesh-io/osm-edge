package repo

import (
	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s/events"
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/metricsstore"
	"github.com/openservicemesh/osm/pkg/pipy"
	"strings"
	"sync"
)

func (s *Server) InformTrafficPolicies(repo *Repo, wg *sync.WaitGroup, connectedProxy *ConnectedProxy) error {
	// If maxDataPlaneConnections is enabled i.e. not 0, then check that the number of Sidecar connections is less than maxDataPlaneConnections
	if s.cfg.GetMaxDataPlaneConnections() != 0 && s.proxyRegistry.GetConnectedProxyCount() >= s.cfg.GetMaxDataPlaneConnections() {
		connectedProxy.initError = errTooManyConnections
		return errTooManyConnections
	}

	metricsstore.DefaultMetricsStore.ProxyConnectCount.Inc()

	proxy := connectedProxy.proxy

	if connectedProxy.initError = s.recordPodMetadata(proxy); connectedProxy.initError == errServiceAccountMismatch {
		// Service Account mismatch
		log.Error().Err(connectedProxy.initError).Str("proxy", proxy.String()).Msg("Mismatched service account for proxy")
		return connectedProxy.initError
	}

	s.proxyRegistry.RegisterProxy(proxy)

	defer s.proxyRegistry.UnregisterProxy(proxy)
	defer repo.UnregisterProxy(proxy)

	connectedProxy.quit = make(chan struct{})
	// Subscribe to both broadcast and proxy UUID specific events
	proxyUpdatePubSub := s.msgBroker.GetProxyUpdatePubSub()
	proxyUpdateChan := proxyUpdatePubSub.Sub(announcements.ProxyUpdate.String(), messaging.GetPubSubTopicForProxyUUID(proxy.UUID.String()))
	defer s.msgBroker.Unsub(proxyUpdatePubSub, proxyUpdateChan)

	// Register for certificate rotation updates
	certPubSub := s.msgBroker.GetCertPubSub()
	certRotateChan := certPubSub.Sub(announcements.CertificateRotated.String())
	defer s.msgBroker.Unsub(certPubSub, certRotateChan)

	newJob := func() *PipyConfGeneratorJob {
		return &PipyConfGeneratorJob{
			proxy:     proxy,
			xdsServer: s,
			done:      make(chan struct{}),
		}
	}
	wg.Done()

	for {
		select {
		case <-connectedProxy.quit:
			log.Info().Str("proxy", proxy.String()).Msgf("Pipy Restful session closed")
			metricsstore.DefaultMetricsStore.ProxyConnectCount.Dec()
			return nil

		case <-proxyUpdateChan:
			log.Info().Str("proxy", proxy.String()).Msg("Broadcast update received")

			// Queue a full configuration update
			// Do not send SDS, let sidecar figure out what certs does it want.
			<-s.workqueues.AddJob(newJob())

		case certRotateMsg := <-certRotateChan:
			cert := certRotateMsg.(events.PubSubMessage).NewObj.(*certificate.Certificate)
			if isCNforProxy(proxy, cert.GetCommonName()) {
				// The CN whose corresponding certificate was updated (rotated) by the certificate provider is associated
				// with this proxy, so update the secrets corresponding to this certificate via SDS.
				log.Debug().Str("proxy", proxy.String()).Msg("Certificate has been updated for proxy")

				// Empty DiscoveryRequest should create the SDS specific request
				// Prepare to queue the SDS proxy response job on the worker pool
				<-s.workqueues.AddJob(newJob())
			}
		}
	}
}

// isCNforProxy returns true if the given CN for the workload certificate matches the given proxy's identity.
// Proxy identity corresponds to the k8s service account, while the workload certificate is of the form
// <svc-account>.<namespace>.<trust-domain>.
func isCNforProxy(proxy *pipy.Proxy, cn certificate.CommonName) bool {
	proxyIdentity, err := pipy.GetServiceIdentityFromProxyCertificate(proxy.GetCertificateCommonName())
	if err != nil {
		log.Error().Str("proxy", proxy.String()).Err(err).Msg("Error looking up proxy identity")
		return false
	}

	// Workload certificate CN is of the form <svc-account>.<namespace>.<trust-domain>
	chunks := strings.Split(cn.String(), constants.DomainDelimiter)
	if len(chunks) < 3 {
		return false
	}

	identityForCN := identity.K8sServiceAccount{Name: chunks[0], Namespace: chunks[1]}
	return identityForCN == proxyIdentity.ToK8sServiceAccount()
}

// recordPodMetadata records pod metadata and verifies the certificate issued for this pod
// is for the same service account as seen on the pod's service account
func (s *Server) recordPodMetadata(p *pipy.Proxy) error {
	if p.Kind() == pipy.KindGateway {
		log.Debug().Str(constants.LogFieldContext, constants.LogContextMulticluster).Str("proxy", p.String()).
			Msgf("Proxy is a Multicluster gateway, skipping recording pod metadata")
		return nil
	}

	pod, err := pipy.GetPodFromCertificate(p.GetCertificateCommonName(), s.kubecontroller)
	if err != nil {
		log.Warn().Str("proxy", p.String()).Msg("Could not find pod for connecting proxy. No metadata was recorded.")
		return nil
	}

	workloadKind := ""
	workloadName := ""
	for _, ref := range pod.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			workloadKind = ref.Kind
			workloadName = ref.Name
			break
		}
	}

	p.PodMetadata = &pipy.PodMetadata{
		UID:       string(pod.UID),
		Name:      pod.Name,
		Namespace: pod.Namespace,
		ServiceAccount: identity.K8sServiceAccount{
			Namespace: pod.Namespace,
			Name:      pod.Spec.ServiceAccountName,
		},
		WorkloadKind: workloadKind,
		WorkloadName: workloadName,
	}

	for idx := range pod.Spec.Containers {
		if pod.Spec.Containers[idx].ReadinessProbe != nil {
			p.PodMetadata.ReadinessProbes = append(p.PodMetadata.ReadinessProbes, pod.Spec.Containers[idx].ReadinessProbe)
		}
		if pod.Spec.Containers[idx].LivenessProbe != nil {
			p.PodMetadata.LivenessProbes = append(p.PodMetadata.LivenessProbes, pod.Spec.Containers[idx].LivenessProbe)
		}
		if pod.Spec.Containers[idx].StartupProbe != nil {
			p.PodMetadata.StartupProbes = append(p.PodMetadata.StartupProbes, pod.Spec.Containers[idx].StartupProbe)
		}
	}

	// Verify Service account matches (cert to pod Service Account)
	cn := p.GetCertificateCommonName()
	certSA, err := pipy.GetServiceIdentityFromProxyCertificate(cn)
	if err != nil {
		log.Error().Err(err).Str("proxy", p.String()).Msgf("Error getting service account from XDS certificate with CommonName=%s", cn)
		return err
	}

	if certSA.ToK8sServiceAccount() != p.PodMetadata.ServiceAccount {
		log.Error().Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrMismatchedServiceAccount)).Str("proxy", p.String()).
			Msgf("Service Account referenced in NodeID (%s) does not match Service Account in Certificate (%s). This proxy is not allowed to join the mesh.", p.PodMetadata.ServiceAccount, certSA)
		return errServiceAccountMismatch
	}

	return nil
}
