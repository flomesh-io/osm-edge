package repo

import (
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
)

// Routine which fulfills listening to proxy broadcasts
func (s *Server) broadcastListener(proxyRegistry *registry.ProxyRegistry, stop <-chan struct{}) {
	// Register for proxy config updates broadcast by the message broker
	proxyUpdatePubSub := s.msgBroker.GetProxyUpdatePubSub()
	proxyUpdateChan := proxyUpdatePubSub.Sub(announcements.ProxyUpdate.String())
	defer s.msgBroker.Unsub(proxyUpdatePubSub, proxyUpdateChan)

	for {
		proxies := s.fireExistProxies()
		for _, proxy := range proxies {
			newJob := func() *PipyConfGeneratorJob {
				return &PipyConfGeneratorJob{
					proxy:      proxy,
					repoServer: s,
					done:       make(chan struct{}),
				}
			}
			<-s.workQueues.AddJob(newJob())
		}
		<-proxyUpdateChan
		// Wait for an informer synchronization period
		time.Sleep(time.Second * 5)
	}
}

func (s *Server) fireExistProxies() []*pipy.Proxy {
	var allProxies []*pipy.Proxy
	allPods := s.kubeController.ListPods()
	for _, pod := range allPods {
		proxy, err := GetProxyFromPod(pod)
		if err != nil {
			log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrGettingProxyFromPod)).
				Msgf("Could not get proxy from pod %s/%s", pod.Namespace, pod.Name)
			continue
		}
		s.fireUpdatedPod(s.proxyRegistry, proxy, pod)
		allProxies = append(allProxies, proxy)
	}
	return allProxies
}

func (s *Server) fireUpdatedPod(proxyRegistry *registry.ProxyRegistry, proxy *pipy.Proxy, updatedPodObj *v1.Pod) {
	if podProxy, ok := proxyRegistry.PodCNtoProxy.Load(proxy.GetCertificateCommonName()); !ok {
		s.informProxy(proxy)
	} else {
		proxy = podProxy.(*pipy.Proxy)
		if err := s.recordPodMetadata(proxy); err != nil {
			log.Err(err)
		}
	}

	if proxy.PodIP != updatedPodObj.Status.PodIP {
		proxy.PodIP = updatedPodObj.Status.PodIP
	}
}

func (s *Server) informProxy(proxy *pipy.Proxy) {
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
