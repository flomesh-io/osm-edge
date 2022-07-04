package repo

import (
	"fmt"
	"sync"
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
func (s *Server) broadcastListener() {
	// Register for proxy config updates broadcast by the message broker
	proxyUpdatePubSub := s.msgBroker.GetProxyUpdatePubSub()
	proxyUpdateChan := proxyUpdatePubSub.Sub(announcements.ProxyUpdate.String())
	defer s.msgBroker.Unsub(proxyUpdatePubSub, proxyUpdateChan)

	// Wait for two informer synchronization periods
	slidingTimer := time.NewTimer(time.Second * 20)
	defer slidingTimer.Stop()

	s.retryJob = func() {
		slidingTimer.Reset(time.Second * 10)
	}

	reconfirm := true

	for {
		select {
		case <-proxyUpdateChan:
			// Wait for an informer synchronization period
			slidingTimer.Reset(time.Second * 5)
			// Avoid data omission
			reconfirm = true

		case <-slidingTimer.C:
			fmt.Println("RefreshAllProxiesJson:", time.Now().Format("2006-01-02 15:04:05"))
			proxies := s.fireExistProxies()
			for _, proxy := range proxies {
				if proxy.PodMetadata == nil {
					if err := s.recordPodMetadata(proxy); err != nil {
						slidingTimer.Reset(time.Second * 10)
						continue
					}
				}
				if proxy.PodMetadata == nil || len(proxy.PodIP) == 0 {
					slidingTimer.Reset(time.Second * 10)
					continue
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
			if reconfirm {
				reconfirm = false
				slidingTimer.Reset(time.Second * 10)
			}
		}
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
		s.fireUpdatedPod(s.proxyRegistry, proxy)
		allProxies = append(allProxies, proxy)
	}
	return allProxies
}

func (s *Server) fireUpdatedPod(proxyRegistry *registry.ProxyRegistry, proxy *pipy.Proxy) {
	if _, ok := proxyRegistry.PodCNtoProxy.Load(proxy.GetCertificateCommonName()); !ok {
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
