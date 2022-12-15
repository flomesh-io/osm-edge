package repo

import (
	"fmt"
	"sort"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/client"
	"github.com/openservicemesh/osm/pkg/trafficpolicy"
)

func (s *Server) pluginListener() {
	pluginUpdatePubSub := s.msgBroker.GetPluginUpdatePubSub()
	pluginUpdateChan := pluginUpdatePubSub.Sub(announcements.PluginUpdate.String())
	defer s.msgBroker.Unsub(pluginUpdatePubSub, pluginUpdateChan)

	slidingTimer := time.NewTimer(time.Second * 10)
	defer slidingTimer.Stop()

	slidingTimerReset := func() {
		slidingTimer.Reset(time.Second * 5)
	}
	s.retryPluginsJob = slidingTimerReset

	reconfirm := true

	for {
		select {
		case <-pluginUpdateChan:
			// Wait for an informer synchronization period
			slidingTimer.Reset(time.Second * 5)
			// Avoid data omission
			reconfirm = true

		case <-slidingTimer.C:
			if s.updatePlugins() {
				if s.retryProxiesJob != nil {
					s.retryProxiesJob()
				}
			}
			if reconfirm {
				reconfirm = false
				slidingTimer.Reset(time.Second * 10)
			}
		}
	}
}

func (s *Server) updatePlugins() bool {
	var pluginItems []client.BatchItem
	var pluginVers []string
	newPluginSet := mapset.NewSet()

	plugins := s.catalog.GetPlugins()
	for _, pluginItem := range plugins {
		uri := getPluginURI(pluginItem.Name)
		bytes := []byte(pluginItem.Script)
		newPluginSet.Add(pluginItem.Name)
		pluginItems = append(pluginItems, client.BatchItem{
			Filename: uri,
			Content:  bytes,
		})
		pluginVers = append(pluginVers, fmt.Sprintf("%s:%d", uri, hash(bytes)))
	}

	diffSet := s.pluginSet.Difference(newPluginSet)
	diffPlugins := diffSet.ToSlice()
	for _, pluginName := range diffPlugins {
		pluginItems = append(pluginItems, client.BatchItem{
			Filename: getPluginURI(pluginName.(string)),
			Obsolete: true,
		})
	}
	if len(pluginItems) > 0 {
		sort.Strings(pluginVers)
		pluginSetHash := hash([]byte(strings.Join(pluginVers, "")))
		pluginSetVersion := fmt.Sprintf("%d", pluginSetHash)

		s.pluginMutex.Lock()
		defer s.pluginMutex.Unlock()

		if s.pluginSetVersion == pluginSetVersion {
			return false
		}
		_, err := s.repoClient.Batch(pluginSetVersion, []client.Batch{
			{
				Basepath: osmCodebase,
				Items:    pluginItems,
			},
		})
		if err != nil {
			log.Error().Err(err)
		} else {
			s.pluginSet = newPluginSet
			s.pluginSetVersion = pluginSetVersion
		}
		return true
	}

	return false
}

// getPluginURI return the URI of the plugin.
func getPluginURI(name string) string {
	return fmt.Sprintf("plugins/%s.js", name)
}

func matchPluginChain(pluginChain *trafficpolicy.PluginChain, ns *corev1.Namespace, pod *corev1.Pod) bool {
	matchedNamespace := false
	matchedPod := false

	if pluginChain.Selectors.NamespaceSelector != nil {
		labelSelector, errSelector := metav1.LabelSelectorAsSelector(pluginChain.Selectors.NamespaceSelector)
		if errSelector == nil {
			matchedNamespace = labelSelector.Matches(labels.Set(ns.GetLabels()))
		} else {
			log.Err(errSelector).Str("namespace", pluginChain.Namespace).Str("PluginChan", pluginChain.Name)
			return false
		}
	} else {
		matchedNamespace = true
	}

	if pluginChain.Selectors.PodSelector != nil {
		labelSelector, errSelector := metav1.LabelSelectorAsSelector(pluginChain.Selectors.PodSelector)
		if errSelector == nil {
			matchedPod = labelSelector.Matches(labels.Set(pod.GetLabels()))
		} else {
			log.Err(errSelector).Str("namespace", pluginChain.Namespace).Str("PluginChan", pluginChain.Name)
			return false
		}
	} else {
		matchedPod = true
	}

	return matchedNamespace && matchedPod
}
