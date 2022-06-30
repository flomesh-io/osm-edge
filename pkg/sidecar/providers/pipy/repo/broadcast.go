package repo

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
)

// Routine which fulfills listening to proxy broadcasts
func (s *Server) broadcastListener() {
	//// Register for proxy config updates broadcasted by the message broker
	//proxyUpdatePubSub := s.msgBroker.GetProxyUpdatePubSub()
	//proxyUpdateChan := proxyUpdatePubSub.Sub(announcements.ProxyUpdate.String())
	//defer s.msgBroker.Unsub(proxyUpdatePubSub, proxyUpdateChan)

	kubePubSub := s.msgBroker.GetKubeEventPubSub()
	podAddChan := kubePubSub.Sub(announcements.PodAdded.String())
	defer s.msgBroker.Unsub(kubePubSub, podAddChan)

	for {
		s.allPodUpdater()
		//<-proxyUpdateChan
		<-podAddChan
	}
}

func (s *Server) allPodUpdater() {
	var allProxies []*pipy.Proxy

	allPods := s.kubecontroller.ListPods()
	for _, pod := range allPods {
		proxy, err := GetProxyFromPod(pod)
		if err != nil {
			log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrGettingProxyFromPod)).
				Msgf("Could not get proxy from pod %s/%s", pod.Namespace, pod.Name)
			continue
		}
		registry.AddCachedMeshPod(proxy.PodIP, proxy.GetCertificateCommonName().String())
		allProxies = append(allProxies, proxy)
	}

	if len(allProxies) > 0 {
		for _, proxy := range allProxies {
			s.firedProxy(proxy.GetCertificateCommonName(), proxy.PodIP)
		}
	}
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
	tempProxy, err := pipy.NewProxy(certificate.CommonName(commonName), "NoSerial", pod.Status.PodIP)

	return tempProxy, err
}
