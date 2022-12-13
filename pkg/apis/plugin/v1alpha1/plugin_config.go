package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PluginConfig is the type used to represent a plugin config policy.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PluginConfig struct {
	// Object's type metadata
	metav1.TypeMeta `json:",inline"`

	// Object's metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the PlugIn specification
	// +optional
	Spec PluginConfigSpec `json:"spec,omitempty"`

	// Status is the status of the plugin config configuration.
	// +optional
	Status PluginConfigStatus `json:"status,omitempty"`
}

// PluginConfigSpec is the type used to represent the plugin config specification.
type PluginConfigSpec struct {
	// Onload is the type used to represent the plugin service onload chain.
	// +optional
	Onload *OnloadChainSpec `json:"onload,omitempty"`

	// Inbound is the type used to represent the plugin service inbound chain.
	// +optional
	Inbound *InboundChainSpec `json:"inbound,omitempty"`

	// Outbound is the type used to represent the plugin service outbound chain.
	// +optional
	Outbound *OutboundChainSpec `json:"outbound,omitempty"`

	// Unload is the type used to represent the plugin service unload chain.
	// +optional
	Unload *UnloadChainSpec `json:"unload,omitempty"`
}

// OnloadChainSpec is the type used to represent the plugin service onload chain.
type OnloadChainSpec struct {
	// Plugins is a list of mounted plugins applied
	Plugins []MountedPlugin `json:"plugins,omitempty"`
}

// UnloadChainSpec is the type used to represent the plugin service unload chain.
type UnloadChainSpec struct {
	// Plugins is a list of mounted plugins applied
	Plugins []MountedPlugin `json:"plugins,omitempty"`
}

// InboundChainSpec is the type used to represent the plugin service inbound chain.
type InboundChainSpec struct {
	// TargetRoutes are the traffic target routes to allow
	// +optional
	TargetRoutes []TrafficTargetRoute `json:"targetRoutes,omitempty"`

	// Plugins is a list of mounted plugins applied
	Plugins []MountedPlugin `json:"plugins,omitempty"`
}

// OutboundChainSpec is the type used to represent the plugin service outbound chain.
type OutboundChainSpec struct {
	// TargetServices are the traffic services to allow
	// +optional
	TargetServices []TrafficTargetService `json:"targetServices,omitempty"`

	// Plugins is a list of mounted plugins applied
	Plugins []MountedPlugin `json:"plugins,omitempty"`
}

// TrafficTargetService is the Traffic Service to allow for a TrafficTarget
type TrafficTargetService struct {
	// Kind is the kind of TrafficSpec to allow
	Kind string `json:"kind,omitempty"`

	// Name of the TrafficSpec to use
	Name string `json:"name,omitempty"`

	// Namespace defines the space within which each name must be unique. An empty namespace is
	// equivalent to the "default" namespace, but "default" is the canonical representation.
	// Not all objects are required to be scoped to a namespace - the value of this field for
	// those objects will be empty.
	//
	// Must be a DNS_LABEL.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/namespaces
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// TargetRoutes are the traffic target routes to allow
	// +optional
	TargetRoutes []TrafficTargetRoute `json:"targetRoutes,omitempty"`

	// Plugins is a list of mounted plugins applied
	Plugins []MountedPlugin `json:"plugins,omitempty"`
}

// TrafficTargetRoute is the Traffic Route to allow for a TrafficTarget
type TrafficTargetRoute struct {
	// Kind is the kind of TrafficSpec to allow
	Kind string `json:"kind,omitempty"`

	// Name of the TrafficSpec to use
	Name string `json:"name,omitempty"`

	// Namespace defines the space within which each name must be unique. An empty namespace is
	// equivalent to the "default" namespace, but "default" is the canonical representation.
	// Not all objects are required to be scoped to a namespace - the value of this field for
	// those objects will be empty.
	//
	// Must be a DNS_LABEL.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/namespaces
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Matches is a list of TrafficSpec routes to allow traffic for
	// +optional
	Matches []string `json:"matches,omitempty"`

	// Plugins is a list of mounted plugins applied
	Plugins []MountedPlugin `json:"plugins,omitempty"`
}

// MountedPlugin is the type used to represent the mounted plugin.
type MountedPlugin struct {
	PluginIdentity

	// MountPoint defines the mount point of the plugin.
	MountPoint string `json:"mountpoint,omitempty"`
}

// PluginConfigList defines the list of PluginConfig objects.
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PluginConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PluginConfig `json:"items"`
}

// PluginConfigStatus is the type used to represent the status of a plugin service resource.
type PluginConfigStatus struct {
	// CurrentStatus defines the current status of an AccessCert resource.
	// +optional
	CurrentStatus string `json:"currentStatus,omitempty"`

	// Reason defines the reason for the current status of an AccessCert resource.
	// +optional
	Reason string `json:"reason,omitempty"`
}

// PluginIdentity is the type used to represent the plugin identity.
type PluginIdentity struct {
	// Namespace defines the namespace of the plugin.
	Namespace string `json:"namespace,omitempty"`

	// Name defines the Name of the plugin.
	Name string `json:"name,omitempty"`
}

// GetPluginURI return the URI of the plugin.
func (plugin *PluginIdentity) GetPluginURI() string {
	return fmt.Sprintf("plugins/%s-%s.js", plugin.Namespace, plugin.Name)
}
