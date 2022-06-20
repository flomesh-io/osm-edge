package registry

import (
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/k8s/events"
)

var (
	// CachedMeshPods cache mesh pods
	CachedMeshPods = make(map[string]string)
	// CachedMeshPodsV the latest version of CachedMeshPodsV
	CachedMeshPodsV = uint64(0)
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
				if pr.releaseCertificateCallback != nil {
					pr.releaseCertificateCallback(podUID, endpointCN)
				}
				certManager.ReleaseCertificate(endpointCN)
			} else {
				log.Warn().Msgf("Pod with UID %s not found in proxy registry", podUID)
			}
		}
	}
}

// SetReleaseCertificateCallback is used to set the callback after certificate is released
func (pr *ProxyRegistry) SetReleaseCertificateCallback(cb func(podUID types.UID, endpointCN certificate.CommonName)) {
	pr.releaseCertificateCallback = cb
}

// CacheMeshPodsHandler handles mesh pods cache
func (pr *ProxyRegistry) CacheMeshPodsHandler(stop <-chan struct{}) {
	kubePubSub := pr.msgBroker.GetKubeEventPubSub()
	podChan := kubePubSub.Sub(announcements.PodUpdated.String())
	defer pr.msgBroker.Unsub(kubePubSub, podChan)
	for {
		select {
		case <-stop:
			return

		case podMsg := <-podChan:
			subMessage, castOk := podMsg.(events.PubSubMessage)
			if !castOk {
				log.Error().Msgf("Error casting to events.PubSubMessage, got type %T", subMessage)
				continue
			}

			podObj, castOk := subMessage.OldObj.(*corev1.Pod)
			if !castOk {
				log.Error().Msgf("Error casting to *corev1.Pod, got type %T", podObj)
				continue
			}
			podUID := podObj.GetObjectMeta().GetUID()
			if podIface, ok := pr.podUIDToCN.Load(podUID); ok {
				endpointCN := podIface.(certificate.CommonName)
				AddCachedMeshPod(podObj.Status.PodIP, endpointCN.String())
			}
		}
	}
}

// AddCachedMeshPod is used to cache mesh pod by addr
func AddCachedMeshPod(addr, cn string) {
	CachedMeshPodsLock.Lock()
	defer CachedMeshPodsLock.Unlock()
	CachedMeshPods[addr] = cn
	CachedMeshPodsV++
}

// RemoveCachedMeshPod is used to remove cached mesh pod by addr
func RemoveCachedMeshPod(addr string) {
	CachedMeshPodsLock.Lock()
	defer CachedMeshPodsLock.Unlock()
	delete(CachedMeshPods, addr)
	CachedMeshPodsV++
}
