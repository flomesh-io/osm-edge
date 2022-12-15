package trafficpolicy

import (
	"fmt"

	pluginv1alpha1 "github.com/openservicemesh/osm/pkg/apis/plugin/v1alpha1"
)

// Plugin defines plugin
type Plugin struct {
	// Name defines the Name of the plugin.
	Name string

	// Script defines pipy script used by the PlugIn.
	Script string
}

// GetPluginURI return the URI of the plugin.
func (plugin *Plugin) GetPluginURI() string {
	return fmt.Sprintf("plugins/%s.js", plugin.Name)
}

// PluginChain defines plugin chain
type PluginChain struct {
	pluginv1alpha1.PluginChainSpec
}

// PluginConfig defines plugin config
type PluginConfig struct {
	pluginv1alpha1.PluginConfigSpec
}
