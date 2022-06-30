package registry

import (
	"sync"

	corev1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/k8s/events"
)

var (
	// CachedMeshPods cache mesh pods
	CachedMeshPods = make(map[string]string)
	// CachedMeshPodsLock the lock of CachedMeshPods
	CachedMeshPodsLock = sync.RWMutex{}
)

// ReleaseCertificateHandler releases certificates based on podDelete events
func (pr *ProxyRegistry) ReleaseCertificateHandler(certManager certificate.Manager, stop <-chan struct{}) {
	kubePubSub := pr.msgBroker.GetKubeEventPubSub()
	podDeleteChan := kubePubSub.Sub(announcements.PodDeleted.String())
	defer pr.msgBroker.Unsub(kubePubSub, podDeleteChan)

	for {
		select {
		case <-stop:
			return

		case podDeletedMsg := <-podDeleteChan:
			psubMessage, castOk := podDeletedMsg.(events.PubSubMessage)
			if !castOk {
				log.Error().Msgf("Error casting to events.PubSubMessage, got type %T", psubMessage)
				continue
			}

			// guaranteed can only be a PodDeleted event
			deletedPodObj, castOk := psubMessage.OldObj.(*corev1.Pod)
			if !castOk {
				log.Error().Msgf("Error casting to *corev1.Pod, got type %T", deletedPodObj)
				continue
			}
			RemoveCachedMeshPod(deletedPodObj.Status.PodIP)
			podUID := deletedPodObj.GetObjectMeta().GetUID()
			if podIface, ok := pr.podUIDToCN.Load(podUID); ok {
				endpointCN := podIface.(certificate.CommonName)
				log.Warn().Msgf("Pod with UID %s found in proxy registry; releasing certificate %s", podUID, endpointCN)
				certManager.ReleaseCertificate(endpointCN)
			} else {
				log.Warn().Msgf("Pod with UID %s not found in proxy registry", podUID)
			}
		}
	}
}

// AddCachedMeshPod is used to cache mesh pod by addr
func AddCachedMeshPod(addr, cn string) {
	CachedMeshPodsLock.Lock()
	defer CachedMeshPodsLock.Unlock()
	CachedMeshPods[addr] = cn
}

// RemoveCachedMeshPod is used to remove cached mesh pod by addr
func RemoveCachedMeshPod(addr string) {
	CachedMeshPodsLock.Lock()
	defer CachedMeshPodsLock.Unlock()
	delete(CachedMeshPods, addr)
}
