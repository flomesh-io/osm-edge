package repo

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/k8s/events"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
)

// Routine which fulfills listening to proxy broadcasts
func (s *Server) broadcastListener(proxyRegistry *registry.ProxyRegistry, stop <-chan struct{}) {
	// Register for proxy config updates broadcasted by the message broker
	proxyUpdatePubSub := s.msgBroker.GetProxyUpdatePubSub()
	proxyUpdateChan := proxyUpdatePubSub.Sub(announcements.ProxyUpdate.String())
	defer s.msgBroker.Unsub(proxyUpdatePubSub, proxyUpdateChan)

	kubePubSub := s.msgBroker.GetKubeEventPubSub()
	podUpdatedChan := kubePubSub.Sub(announcements.PodUpdated.String())
	defer s.msgBroker.Unsub(kubePubSub, podUpdatedChan)

	podDeleteChan := kubePubSub.Sub(announcements.PodDeleted.String())
	defer s.msgBroker.Unsub(kubePubSub, podDeleteChan)

	s.fireExistPods()

	for {
		select {
		case <-stop:
			return

		case podDeletedMsg := <-podDeleteChan:
			subMessage, castOk := podDeletedMsg.(events.PubSubMessage)
			if !castOk {
				log.Error().Msgf("Error casting to events.PubSubMessage, got type %T", subMessage)
				continue
			}

			// guaranteed can only be a PodDeleted event
			deletedPodObj, castOk := subMessage.OldObj.(*corev1.Pod)
			if !castOk {
				log.Error().Msgf("Error casting to *corev1.Pod, got type %T", deletedPodObj)
				continue
			}
			podUID := deletedPodObj.GetObjectMeta().GetUID()
			proxyRegistry.PodUIDToCertificateSerialNumber.Delete(podUID)
			if podCN, ok := proxyRegistry.PodUIDToCN.LoadAndDelete(podUID); ok {
				endpointCN := podCN.(certificate.CommonName)
				if podProxy, ok := proxyRegistry.PodCNtoProxy.LoadAndDelete(endpointCN); ok {
					proxy := podProxy.(*pipy.Proxy)
					proxy.Quit <- true
				}
				log.Warn().Msgf("Pod with UID %s found in proxy registry; releasing certificate %s", podUID, endpointCN)
			} else {
				log.Warn().Msgf("Pod with UID %s not found in proxy registry", podUID)
			}

		case podUpdatedMsg := <-podUpdatedChan:
			subMessage, castOk := podUpdatedMsg.(events.PubSubMessage)
			if !castOk {
				log.Error().Msgf("Error casting to events.PubSubMessage, got type %T", subMessage)
				continue
			}

			// guaranteed can only be a PodUpdated event
			updatedPodObj, castOk := subMessage.NewObj.(*corev1.Pod)
			if !castOk {
				log.Error().Msgf("Error casting to *corev1.Pod, got type %T", updatedPodObj)
				continue
			}
			proxy, err := GetProxyFromPod(updatedPodObj)
			if err != nil {
				log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrGettingProxyFromPod)).
					Msgf("Could not get proxy from pod %s/%s", updatedPodObj.Namespace, updatedPodObj.Name)
				continue
			}
			s.fireUpdatedPod(proxyRegistry, proxy, updatedPodObj)

		case <-proxyUpdateChan:
			s.fireExistPods()
		}
	}
}

func (s *Server) fireExistPods() {
	allPods := s.kubeController.ListPods()
	for _, pod := range allPods {
		proxy, err := GetProxyFromPod(pod)
		if err != nil {
			log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrGettingProxyFromPod)).
				Msgf("Could not get proxy from pod %s/%s", pod.Namespace, pod.Name)
			continue
		}
		s.fireUpdatedPod(s.proxyRegistry, proxy, pod)
	}
}

func (s *Server) fireUpdatedPod(proxyRegistry *registry.ProxyRegistry, proxy *pipy.Proxy, updatedPodObj *v1.Pod) {
	if podProxy, ok := proxyRegistry.PodCNtoProxy.Load(proxy.GetCertificateCommonName()); !ok {
		s.firedProxy(proxy)
	} else {
		proxy = podProxy.(*pipy.Proxy)
	}

	if proxy.PodIP != updatedPodObj.Status.PodIP {
		proxy.PodIP = updatedPodObj.Status.PodIP
	}

	newJob := func() *PipyConfGeneratorJob {
		return &PipyConfGeneratorJob{
			proxy:      proxy,
			repoServer: s,
			done:       make(chan struct{}),
		}
	}
	<-s.workQueues.AddJob(newJob())
}

func (s *Server) firedProxy(proxy *pipy.Proxy) {
	go func() {
		if aggregatedErr := s.informTrafficPolicies(proxy); aggregatedErr != nil {
			log.Error().Err(aggregatedErr).Msgf("Pipy Aggregated Traffic Policies Error.")
		}
	}()
}

// GetProxyFromPod infers and creates a Proxy data structure from a Pod.
// This is a temporary workaround as proxy is required and expected in any vertical call to XDS,
// however snapshotcache has no need to provide visibility on proxies whatsoever.
// All verticals use the proxy structure to infer the pod later, so the actual only mandatory
// data for the verticals to be functional is the common name, which links proxy <-> pod
func GetProxyFromPod(pod *v1.Pod) (*pipy.Proxy, error) {
	var serviceAccount string
	var namespace string

	uuidString, uuidFound := pod.Labels[constants.SidecarUniqueIDLabelName]
	if !uuidFound {
		return nil, errors.Errorf("UUID not found for pod %s/%s, not a mesh pod", pod.Namespace, pod.Name)
	}
	proxyUUID, err := uuid.Parse(uuidString)
	if err != nil {
		return nil, errors.Errorf("Could not parse UUID label into UUID type (%s): %v", uuidString, err)
	}

	serviceAccount = pod.Spec.ServiceAccountName
	namespace = pod.Namespace

	// construct CN for this pod/proxy
	// TODO: Infer proxy type from Pod
	commonName := pipy.NewCertCommonName(proxyUUID, pipy.KindSidecar, serviceAccount, namespace)
	tempProxy, err := pipy.NewProxy(commonName, "NoSerial", pod.Status.PodIP)

	return tempProxy, err
}
