package repo

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/models"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
)

// Routine which fulfills listening to proxy broadcasts
func (s *Server) broadcastListener() {
	// Register for proxy config updates broadcast by the message broker
	proxyUpdatePubSub := s.msgBroker.GetProxyUpdatePubSub()
	proxyUpdateChan := proxyUpdatePubSub.Sub(announcements.ProxyUpdate.String())
	defer s.msgBroker.Unsub(proxyUpdatePubSub, proxyUpdateChan)

	// Wait for two informer synchronization periods
	slidingTimer := time.NewTimer(time.Second * 20)
	defer slidingTimer.Stop()

	slidingTimerReset := func() {
		slidingTimer.Reset(time.Second * 5)
	}

	s.retryJob = slidingTimerReset
	s.proxyRegistry.UpdateProxies = slidingTimerReset

	reconfirm := true

	for {
		select {
		case <-proxyUpdateChan:
			// Wait for an informer synchronization period
			slidingTimer.Reset(time.Second * 5)
			// Avoid data omission
			reconfirm = true

		case <-slidingTimer.C:
			connectedProxies := make(map[string]*pipy.Proxy)
			disconnectedProxies := make(map[string]*pipy.Proxy)
			proxies := s.fireExistProxies()
			for _, proxy := range proxies {
				if proxy.PodMetadata == nil {
					if err := s.recordPodMetadata(proxy); err != nil {
						slidingTimer.Reset(time.Second * 5)
						continue
					}
				}
				if proxy.PodMetadata == nil || proxy.Addr == nil || len(proxy.GetAddr()) == 0 {
					slidingTimer.Reset(time.Second * 5)
					continue
				}
				connectedProxies[proxy.UUID.String()] = proxy
			}

			s.proxyRegistry.RangeConnectedProxy(func(key, value interface{}) bool {
				proxyUUID := key.(string)
				if _, exists := connectedProxies[proxyUUID]; !exists {
					disconnectedProxies[proxyUUID] = value.(*pipy.Proxy)
				}
				return true
			})

			if len(connectedProxies) > 0 {
				for _, proxy := range connectedProxies {
					newJob := func() *PipyConfGeneratorJob {
						return &PipyConfGeneratorJob{
							proxy:      proxy,
							repoServer: s,
							done:       make(chan struct{}),
						}
					}
					<-s.workQueues.AddJob(newJob())
				}
			}

			if reconfirm {
				reconfirm = false
				slidingTimer.Reset(time.Second * 10)
			}

			go func() {
				if len(disconnectedProxies) > 0 {
					for certCommonName, proxy := range disconnectedProxies {
						s.proxyRegistry.UnregisterProxy(proxy)
						s.repoClient.Delete(fmt.Sprintf("%s/%s", osmSidecarCodebase, certCommonName))
					}
				}
			}()
		}
	}
}

func (s *Server) fireExistProxies() []*pipy.Proxy {
	var allProxies []*pipy.Proxy
	allPods := s.kubeController.ListPods()
	for _, pod := range allPods {
		proxy, err := GetProxyFromPod(pod)
		if err != nil {
			continue
		}
		s.fireUpdatedPod(s.proxyRegistry, proxy)
		allProxies = append(allProxies, proxy)
	}
	return allProxies
}

func (s *Server) fireUpdatedPod(proxyRegistry *registry.ProxyRegistry, proxy *pipy.Proxy) {
	if v := proxyRegistry.GetConnectedProxy(proxy.UUID.String()); v == nil {
		s.informProxy(proxy)
	}
}

func (s *Server) informProxy(proxy *pipy.Proxy) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if aggregatedErr := s.informTrafficPolicies(proxy, &wg); aggregatedErr != nil {
			log.Error().Err(aggregatedErr).Msgf("Pipy Aggregated Traffic Policies Error.")
		}
	}()
	wg.Wait()
}

// GetProxyFromPod infers and creates a Proxy data structure from a Pod.
// This is a temporary workaround as proxy is required and expected in any vertical call,
// however snapshotcache has no need to provide visibility on proxies whatsoever.
// All verticals use the proxy structure to infer the pod later, so the actual only mandatory
// data for the verticals to be functional is the common name, which links proxy <-> pod
func GetProxyFromPod(pod *v1.Pod) (*pipy.Proxy, error) {
	uuidString, uuidFound := pod.Labels[constants.SidecarUniqueIDLabelName]
	if !uuidFound {
		return nil, fmt.Errorf("UUID not found for pod %s/%s, not a mesh pod", pod.Namespace, pod.Name)
	}
	proxyUUID, err := uuid.Parse(uuidString)
	if err != nil {
		return nil, fmt.Errorf("Could not parse UUID label into UUID type (%s): %w", uuidString, err)
	}

	sa := pod.Spec.ServiceAccountName
	namespace := pod.Namespace

	return pipy.NewProxy(models.KindSidecar, proxyUUID, identity.New(sa, namespace), nil), nil
}

// GetProxyUUIDFromPod infers and creates a Proxy UUID from a Pod.
func GetProxyUUIDFromPod(pod *v1.Pod) (string, error) {
	uuidString, uuidFound := pod.Labels[constants.SidecarUniqueIDLabelName]
	if !uuidFound {
		return "", fmt.Errorf("UUID not found for pod %s/%s, not a mesh pod", pod.Namespace, pod.Name)
	}
	proxyUUID, err := uuid.Parse(uuidString)
	if err != nil {
		return "", fmt.Errorf("Could not parse UUID label into UUID type (%s): %w", uuidString, err)
	}
	return proxyUUID.String(), nil
}
