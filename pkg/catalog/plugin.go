package catalog

import (
	"github.com/openservicemesh/osm/pkg/trafficpolicy"
)

// GetPlugins returns the plugin policies
func (mc *MeshCatalog) GetPlugins() []*trafficpolicy.Plugin {
	if !mc.configurator.GetFeatureFlags().EnablePluginPolicy {
		return nil
	}

	plugins := mc.pluginController.GetPlugins()
	if plugins == nil {
		log.Trace().Msg("Did not find any plugin")
		return nil
	}

	var pluginPolicies []*trafficpolicy.Plugin
	for _, plugin := range plugins {
		policy := new(trafficpolicy.Plugin)
		policy.Name = plugin.Name
		policy.Script = plugin.Spec.PipyScript
		pluginPolicies = append(pluginPolicies, policy)
	}
	return pluginPolicies
}

// GetPluginConfigs lists plugin configs
func (mc *MeshCatalog) GetPluginConfigs() []*trafficpolicy.PluginConfig {
	if !mc.configurator.GetFeatureFlags().EnablePluginPolicy {
		return nil
	}

	pluginConfigs := mc.pluginController.GetPluginConfigs()
	if pluginConfigs == nil {
		log.Trace().Msg("Did not find any plugin config")
		return nil
	}

	var pluginConfigPolicies []*trafficpolicy.PluginConfig
	for _, pluginConfig := range pluginConfigs {
		policy := new(trafficpolicy.PluginConfig)
		policy.PluginConfigSpec = pluginConfig.Spec
		pluginConfigPolicies = append(pluginConfigPolicies, policy)
	}

	return pluginConfigPolicies
}

// GetPluginChains lists plugin chains
func (mc *MeshCatalog) GetPluginChains() []*trafficpolicy.PluginChain {
	if !mc.configurator.GetFeatureFlags().EnablePluginPolicy {
		return nil
	}

	pluginChains := mc.pluginController.GetPluginChains()
	if pluginChains == nil {
		log.Trace().Msg("Did not find any plugin chain")
		return nil
	}

	var pluginChainPolicies []*trafficpolicy.PluginChain
	for _, pluginConfig := range pluginChains {
		policy := new(trafficpolicy.PluginChain)
		policy.PluginChainSpec = pluginConfig.Spec
		pluginChainPolicies = append(pluginChainPolicies, policy)
	}

	return pluginChainPolicies
}
