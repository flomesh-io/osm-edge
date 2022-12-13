package trafficpolicy

import "fmt"

// PluginPolicy defines plugins for a given backend
type PluginPolicy struct {
	// Name defines the Name of the plugin.
	Name string

	// Script defines pipy script used by the PlugIn.
	Script string
}

// GetPluginURI return the URI of the plugin.
func (plugin *PluginPolicy) GetPluginURI() string {
	return fmt.Sprintf("plugins/%s.js", plugin.Name)
}
