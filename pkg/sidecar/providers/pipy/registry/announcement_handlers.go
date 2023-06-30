// Package registry implements handler's methods.
package registry

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/k8s/events"
	"github.com/openservicemesh/osm/pkg/sidecar"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
)

// ReleaseCertificateHandler releases certificates based on podDelete events
func (pr *ProxyRegistry) ReleaseCertificateHandler(certManager certificateReleaser, stop <-chan struct{}) {
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

			proxyUUID := deletedPodObj.Labels[constants.SidecarUniqueIDLabelName]
			if proxyIface, ok := connectedProxies.Load(proxyUUID); ok {
				proxy := proxyIface.(*pipy.Proxy)
				log.Info().Msgf("Pod with label %s: %s found in proxy registry; releasing certificate for proxy %s", constants.SidecarUniqueIDLabelName, proxyUUID, proxy.Identity)
				certManager.ReleaseCertificate(sidecar.NewCertCNPrefix(proxy.UUID, proxy.Kind(), proxy.Identity))
				if pr.UpdateProxies != nil {
					pr.UpdateProxies()
				}
			} else {
				log.Info().Msgf("Pod with label %s: %s not found in proxy registry", constants.SidecarUniqueIDLabelName, proxyUUID)
			}
		}
	}
}
