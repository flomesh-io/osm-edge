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
	"k8s.io/apimachinery/pkg/runtime"

	policyv1alpha1 "github.com/openservicemesh/osm/pkg/apis/policy/v1alpha1"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/service"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
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
	newPluginPri := make(map[string]uint16)

	plugins := s.catalog.GetPlugins()
	for _, pluginItem := range plugins {
		uri := getPluginURI(pluginItem.Name)
		bytes := []byte(pluginItem.Script)
		newPluginSet.Add(pluginItem.Name)
		newPluginPri[pluginItem.Name] = pluginItem.Priority
		pluginItems = append(pluginItems, client.BatchItem{
			Filename: uri,
			Content:  bytes,
		})
		pluginVers = append(pluginVers, fmt.Sprintf("%s:%d:%d", uri, pluginItem.Priority, hash(bytes)))
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
			s.pluginPri = newPluginPri
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

func walkPluginChain(pluginChains []*trafficpolicy.PluginChain, ns *corev1.Namespace, pod *corev1.Pod, pluginSet mapset.Set, s *Server, proxy *pipy.Proxy) (plugin2MountPoints map[string]*map[string]*runtime.RawExtension, mountPoint2Plugins map[string]mapset.Set) {
	plugin2MountPoints = make(map[string]*map[string]*runtime.RawExtension)
	mountPoint2Plugins = make(map[string]mapset.Set)

	for _, pluginChain := range pluginChains {
		matched := matchPluginChain(pluginChain, ns, pod)
		if !matched {
			continue
		}
		for _, chain := range pluginChain.Chains {
			for _, pluginName := range chain.Plugins {
				if !pluginSet.Contains(pluginName) {
					if len(s.pluginSetVersion) > 0 {
						log.Warn().Str("proxy", proxy.String()).
							Str("plugin", pluginName).
							Msg("Could not find plugin for connecting proxy.")
					}
					if s.retryPluginsJob != nil {
						s.retryPluginsJob()
					}
					continue
				}

				mountPointSet, existPointSet := plugin2MountPoints[pluginName]
				if !existPointSet {
					mountPointMap := make(map[string]*runtime.RawExtension)
					mountPointSet = &mountPointMap
					plugin2MountPoints[pluginName] = mountPointSet
				}
				if _, exist := (*mountPointSet)[chain.Name]; !exist {
					(*mountPointSet)[chain.Name] = nil
				}

				mountedPluginSet, existPluginSet := mountPoint2Plugins[chain.Name]
				if !existPluginSet {
					mountedPluginSet = mapset.NewSet()
					mountPoint2Plugins[chain.Name] = mountedPluginSet
				}
				if !mountedPluginSet.Contains(pluginName) {
					mountedPluginSet.Add(pluginName)
				}
			}
		}
	}
	return
}

func walkPluginConfig(cataloger catalog.MeshCataloger, plugin2MountPoint2Config map[string]*map[string]*runtime.RawExtension) map[string]map[string]*map[string]*runtime.RawExtension {
	meshSvc2Plugin2MountPoint2Config := make(map[string]map[string]*map[string]*runtime.RawExtension)
	pluginConfigs := cataloger.GetPluginConfigs()
	if len(pluginConfigs) > 0 {
		for _, pluginConfig := range pluginConfigs {
			mountPoint2ConfigItem, existMountPoint2Config := plugin2MountPoint2Config[pluginConfig.Plugin]
			if !existMountPoint2Config {
				continue
			}
			for mountPoint := range *mountPoint2ConfigItem {
				(*mountPoint2ConfigItem)[mountPoint] = &pluginConfig.Config
			}
			for _, destinationRef := range pluginConfig.DestinationRefs {
				if destinationRef.Kind == policyv1alpha1.KindService {
					meshSvc := service.MeshService{
						Namespace: destinationRef.Namespace,
						Name:      destinationRef.Name,
					}
					plugin2MountPoint2ConfigItem, exist := meshSvc2Plugin2MountPoint2Config[meshSvc.String()]
					if !exist {
						plugin2MountPoint2ConfigItem = make(map[string]*map[string]*runtime.RawExtension)
						meshSvc2Plugin2MountPoint2Config[meshSvc.String()] = plugin2MountPoint2ConfigItem
					}
					plugin2MountPoint2ConfigItem[pluginConfig.Plugin] = mountPoint2ConfigItem
				}
			}
		}
	}
	return meshSvc2Plugin2MountPoint2Config
}

func setSidecarChain(pipyConf *PipyConf, pluginPri map[string]uint16, mountPoint2Plugins map[string]mapset.Set) {
	pipyConf.Chains = nil
	if len(mountPoint2Plugins) > 0 {
		pipyConf.Chains = make(map[string][]string)
		for mountPoint, plugins := range mountPoint2Plugins {
			var pluginItems PluginSlice
			pluginSlice := plugins.ToSlice()
			for _, item := range pluginSlice {
				if pri, exist := pluginPri[item.(string)]; exist {
					pluginItems = append(pluginItems, trafficpolicy.Plugin{
						Name:     item.(string),
						Priority: pri,
					})
				}
			}
			if len(pluginItems) > 0 {
				var pluginURIs []string
				sort.Sort(&pluginItems)
				for _, pluginItem := range pluginItems {
					pluginURIs = append(pluginURIs, getPluginURI(pluginItem.Name))
				}
				pipyConf.Chains[mountPoint] = pluginURIs
			}
		}
	}
}

func (p *PipyConf) getTrafficMatchPluginConfigs(trafficMatch string) map[string]*runtime.RawExtension {
	segs := strings.Split(trafficMatch, "_")
	meshSvc := segs[1]

	direct := segs[0]
	switch segs[0] {
	case "ingress":
	case "acl":
	case "exp":
		direct = "inbound"
	}
	mountPoint := fmt.Sprintf("%s-%s", direct, segs[3])

	plugin2MountPoint2Config, exist := p.pluginPolicies[meshSvc]
	if !exist {
		return nil
	}
	var pluginConfigs map[string]*runtime.RawExtension
	for pluginName, mountPoint2Config := range plugin2MountPoint2Config {
		if configLoop, existConfig := (*mountPoint2Config)[mountPoint]; existConfig {
			config := configLoop
			if pluginConfigs == nil {
				pluginConfigs = make(map[string]*runtime.RawExtension)
			}
			pluginConfigs[pluginName] = config
		}
	}
	return pluginConfigs
}
