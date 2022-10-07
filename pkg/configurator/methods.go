package configurator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	configv1alpha2 "github.com/openservicemesh/osm/pkg/apis/config/v1alpha2"

	"github.com/openservicemesh/osm/pkg/auth"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
)

const (
	// defaultServiceCertValidityDuration is the default validity duration for service certificates
	defaultServiceCertValidityDuration = 24 * time.Hour

	// defaultCertKeyBitSize is the default certificate key bit size
	defaultCertKeyBitSize = 2048

	// minCertKeyBitSize is the minimum certificate key bit size
	minCertKeyBitSize = 2048

	// maxCertKeyBitSize is the maximum certificate key bit size
	maxCertKeyBitSize = 4096
)

// The functions in this file implement the configurator.Configurator interface

// GetMeshConfig returns the MeshConfig resource corresponding to the control plane
func (c *client) GetMeshConfig() configv1alpha2.MeshConfig {
	return c.getMeshConfig()
}

// GetOSMNamespace returns the namespace in which the OSM controller pod resides.
func (c *client) GetOSMNamespace() string {
	return c.osmNamespace
}

func marshalConfigToJSON(config configv1alpha2.MeshConfigSpec) (string, error) {
	bytes, err := json.MarshalIndent(&config, "", "    ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetMeshConfigJSON returns the MeshConfig in pretty JSON.
func (c *client) GetMeshConfigJSON() (string, error) {
	cm, err := marshalConfigToJSON(c.getMeshConfig().Spec)
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrMeshConfigMarshaling)).Msgf("Error marshaling MeshConfig %s: %+v", c.getMeshConfigCacheKey(), c.getMeshConfig())
		return "", err
	}
	return cm, nil
}

// IsPermissiveTrafficPolicyMode tells us whether the OSM Control Plane is in permissive mode,
// where all existing traffic is allowed to flow as it is,
// or it is in SMI Spec mode, in which only traffic between source/destinations
// referenced in SMI policies is allowed.
func (c *client) IsPermissiveTrafficPolicyMode() bool {
	return c.getMeshConfig().Spec.Traffic.EnablePermissiveTrafficPolicyMode
}

// IsEgressEnabled determines whether egress is globally enabled in the mesh or not.
func (c *client) IsEgressEnabled() bool {
	return c.getMeshConfig().Spec.Traffic.EnableEgress
}

// IsDebugServerEnabled determines whether osm debug HTTP server is enabled
func (c *client) IsDebugServerEnabled() bool {
	return c.getMeshConfig().Spec.Observability.EnableDebugServer
}

// IsTracingEnabled returns whether tracing is enabled
func (c *client) IsTracingEnabled() bool {
	return c.getMeshConfig().Spec.Observability.Tracing.Enable
}

// GetTracingHost is the host to which we send tracing spans
func (c *client) GetTracingHost() string {
	tracingAddress := c.getMeshConfig().Spec.Observability.Tracing.Address
	if tracingAddress != "" {
		return tracingAddress
	}
	return fmt.Sprintf("%s.%s.svc.cluster.local", constants.DefaultTracingHost, c.GetOSMNamespace())
}

// GetTracingPort returns the tracing listener port
func (c *client) GetTracingPort() uint32 {
	tracingPort := c.getMeshConfig().Spec.Observability.Tracing.Port
	if tracingPort != 0 {
		return uint32(tracingPort)
	}
	return constants.DefaultTracingPort
}

// GetTracingEndpoint returns the listener's collector endpoint
func (c *client) GetTracingEndpoint() string {
	tracingEndpoint := c.getMeshConfig().Spec.Observability.Tracing.Endpoint
	if tracingEndpoint != "" {
		return tracingEndpoint
	}
	return constants.DefaultTracingEndpoint
}

// IsRemoteLoggingEnabled returns whether remote logging is enabled
func (c *client) IsRemoteLoggingEnabled() bool {
	return c.getMeshConfig().Spec.Observability.RemoteLogging.Enable
}

// GetRemoteLoggingHost is the host to which we send logging spans
func (c *client) GetRemoteLoggingHost() string {
	remoteLoggingAddress := c.getMeshConfig().Spec.Observability.RemoteLogging.Address
	if remoteLoggingAddress != "" {
		return remoteLoggingAddress
	}
	return ""
}

// GetRemoteLoggingPort returns the remote logging listener port
func (c *client) GetRemoteLoggingPort() uint32 {
	remoteLoggingPort := c.getMeshConfig().Spec.Observability.RemoteLogging.Port
	if remoteLoggingPort != 0 {
		return uint32(remoteLoggingPort)
	}
	return 0
}

// GetRemoteLoggingEndpoint returns the collector endpoint
func (c *client) GetRemoteLoggingEndpoint() string {
	remoteLoggingEndpoint := c.getMeshConfig().Spec.Observability.RemoteLogging.Endpoint
	if remoteLoggingEndpoint != "" {
		return remoteLoggingEndpoint
	}
	return ""
}

// GetRemoteLoggingAuthorization returns the access entity that allows to authorize someone in remote logging service.
func (c *client) GetRemoteLoggingAuthorization() string {
	remoteLoggingAuthorization := c.getMeshConfig().Spec.Observability.RemoteLogging.Authorization
	if remoteLoggingAuthorization != "" {
		return remoteLoggingAuthorization
	}
	return ""
}

// GetMaxDataPlaneConnections returns the max data plane connections allowed, 0 if disabled
func (c *client) GetMaxDataPlaneConnections() int {
	return c.getMeshConfig().Spec.Sidecar.MaxDataPlaneConnections
}

// GetSidecarLogLevel returns the sidecar log level
func (c *client) GetSidecarLogLevel() string {
	logLevel := c.getMeshConfig().Spec.Sidecar.LogLevel
	if logLevel != "" {
		return logLevel
	}
	return constants.DefaultSidecarLogLevel
}

// GetSidecarClass returns the sidecar class
func (c *client) GetSidecarClass() string {
	class := c.getMeshConfig().Spec.Sidecar.SidecarClass
	if class == "" {
		class = os.Getenv("OSM_DEFAULT_SIDECAR_CLASS")
	}
	if class == "" {
		class = constants.SidecarClassPipy
	}
	return class
}

// GetSidecarImage returns the sidecar image
func (c *client) GetSidecarImage() string {
	image := c.getMeshConfig().Spec.Sidecar.SidecarImage
	if len(image) == 0 {
		sidecarClass := c.getMeshConfig().Spec.Sidecar.SidecarClass
		sidecarDrivers := c.getMeshConfig().Spec.Sidecar.SidecarDrivers
		for _, sidecarDriver := range sidecarDrivers {
			if strings.EqualFold(strings.ToLower(sidecarClass), strings.ToLower(sidecarDriver.SidecarName)) {
				image = sidecarDriver.SidecarImage
				break
			}
		}
	}
	if len(image) == 0 {
		image = os.Getenv("OSM_DEFAULT_SIDECAR_IMAGE")
	}
	return image
}

// GetSidecarWindowsImage returns the sidecar windows image
func (c *client) GetSidecarWindowsImage() string {
	image := c.getMeshConfig().Spec.Sidecar.SidecarWindowsImage
	if len(image) == 0 {
		sidecarClass := c.getMeshConfig().Spec.Sidecar.SidecarClass
		sidecarDrivers := c.getMeshConfig().Spec.Sidecar.SidecarDrivers
		for _, sidecarDriver := range sidecarDrivers {
			if strings.EqualFold(strings.ToLower(sidecarClass), strings.ToLower(sidecarDriver.SidecarName)) {
				image = sidecarDriver.SidecarWindowsImage
				break
			}
		}
	}
	if len(image) == 0 {
		image = os.Getenv("OSM_DEFAULT_SIDECAR_WINDOWS_IMAGE")
	}
	return image
}

// GetInitContainerImage returns the init container image
func (c *client) GetInitContainerImage() string {
	image := c.getMeshConfig().Spec.Sidecar.InitContainerImage
	if len(image) == 0 {
		sidecarClass := c.getMeshConfig().Spec.Sidecar.SidecarClass
		sidecarDrivers := c.getMeshConfig().Spec.Sidecar.SidecarDrivers
		for _, sidecarDriver := range sidecarDrivers {
			if strings.EqualFold(strings.ToLower(sidecarClass), strings.ToLower(sidecarDriver.SidecarName)) {
				image = sidecarDriver.InitContainerImage
				break
			}
		}
	}
	if len(image) == 0 {
		image = os.Getenv("OSM_DEFAULT_INIT_CONTAINER_IMAGE")
	}
	return image
}

// GetProxyServerPort returns the port on which the Discovery Service listens for new connections from Sidecars
func (c *client) GetProxyServerPort() uint32 {
	sidecarClass := c.getMeshConfig().Spec.Sidecar.SidecarClass
	sidecarDrivers := c.getMeshConfig().Spec.Sidecar.SidecarDrivers
	for _, sidecarDriver := range sidecarDrivers {
		if strings.EqualFold(strings.ToLower(sidecarClass), strings.ToLower(sidecarDriver.SidecarName)) {
			return sidecarDriver.ProxyServerPort
		}
	}
	return constants.ProxyServerPort
}

// GetSidecarDisabledMTLS returns the status of mTLS
func (c *client) GetSidecarDisabledMTLS() bool {
	disabledMTLS := false
	sidecarClass := c.getMeshConfig().Spec.Sidecar.SidecarClass
	sidecarDrivers := c.getMeshConfig().Spec.Sidecar.SidecarDrivers
	for _, sidecarDriver := range sidecarDrivers {
		if strings.EqualFold(strings.ToLower(sidecarClass), strings.ToLower(sidecarDriver.SidecarName)) {
			disabledMTLS = sidecarDriver.SidecarDisabledMTLS
			break
		}
	}
	return disabledMTLS
}

// GetServiceCertValidityPeriod returns the validity duration for service certificates, and a default in case of invalid duration
func (c *client) GetServiceCertValidityPeriod() time.Duration {
	durationStr := c.getMeshConfig().Spec.Certificate.ServiceCertValidityDuration
	validityDuration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Error().Err(err).Msgf("Error parsing service certificate validity duration %s", durationStr)
		return defaultServiceCertValidityDuration
	}

	return validityDuration
}

// GetCertKeyBitSize returns the certificate key bit size to be used
func (c *client) GetCertKeyBitSize() int {
	bitSize := c.getMeshConfig().Spec.Certificate.CertKeyBitSize
	if bitSize < minCertKeyBitSize || bitSize > maxCertKeyBitSize {
		log.Error().Msgf("Invalid key bit size: %d", bitSize)
		return defaultCertKeyBitSize
	}

	return bitSize
}

// IsPrivilegedInitContainer returns whether init containers should be privileged
func (c *client) IsPrivilegedInitContainer() bool {
	return c.getMeshConfig().Spec.Sidecar.EnablePrivilegedInitContainer
}

// GetConfigResyncInterval returns the duration for resync interval.
// If error or non-parsable value, returns 0 duration
func (c *client) GetConfigResyncInterval() time.Duration {
	resyncDuration := c.getMeshConfig().Spec.Sidecar.ConfigResyncInterval
	duration, err := time.ParseDuration(resyncDuration)
	if err != nil {
		log.Debug().Err(err).Msgf("Error parsing config resync interval: %s", duration)
		return time.Duration(0)
	}
	return duration
}

// GetProxyResources returns the `Resources` configured for proxies, if any
func (c *client) GetProxyResources() corev1.ResourceRequirements {
	return c.getMeshConfig().Spec.Sidecar.Resources
}

// GetInboundExternalAuthConfig returns the External Authentication configuration for incoming traffic, if any
func (c *client) GetInboundExternalAuthConfig() auth.ExtAuthConfig {
	extAuthConfig := auth.ExtAuthConfig{}
	inboundExtAuthzMeshConfig := c.getMeshConfig().Spec.Traffic.InboundExternalAuthorization

	extAuthConfig.Enable = inboundExtAuthzMeshConfig.Enable
	extAuthConfig.Address = inboundExtAuthzMeshConfig.Address
	extAuthConfig.Port = uint16(inboundExtAuthzMeshConfig.Port)
	extAuthConfig.StatPrefix = inboundExtAuthzMeshConfig.StatPrefix
	extAuthConfig.FailureModeAllow = inboundExtAuthzMeshConfig.FailureModeAllow

	duration, err := time.ParseDuration(inboundExtAuthzMeshConfig.Timeout)
	if err != nil {
		log.Debug().Err(err).Msgf("ExternAuthzTimeout: Not a valid duration %s. defaulting to 1s.", duration)
		duration = 1 * time.Second
	}
	extAuthConfig.AuthzTimeout = duration

	return extAuthConfig
}

// GetFeatureFlags returns OSM's feature flags
func (c *client) GetFeatureFlags() configv1alpha2.FeatureFlags {
	return c.getMeshConfig().Spec.FeatureFlags
}

// GetOSMLogLevel returns the configured OSM log level
func (c *client) GetOSMLogLevel() string {
	return c.getMeshConfig().Spec.Observability.OSMLogLevel
}
