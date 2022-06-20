package driver

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xds_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	xds_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/bootstrap"
	"github.com/openservicemesh/osm/pkg/utils"
	"github.com/openservicemesh/osm/pkg/version"
)

func getEnvoyConfigYAML(config sidecarBootstrapConfigMeta, cfg configurator.Configurator) ([]byte, error) {
	bootstrapConfig, err := bootstrap.BuildFromConfig(bootstrap.Config{
		NodeID:                config.NodeID,
		AdminPort:             constants.SidecarAdminPort,
		XDSClusterName:        constants.OSMControllerName,
		TrustedCA:             config.RootCert,
		CertificateChain:      config.Cert,
		PrivateKey:            config.Key,
		XDSHost:               config.XDSHost,
		XDSPort:               config.XDSPort,
		TLSMinProtocolVersion: config.TLSMinProtocolVersion,
		TLSMaxProtocolVersion: config.TLSMaxProtocolVersion,
		CipherSuites:          config.CipherSuites,
		ECDHCurves:            config.ECDHCurves,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Error building Envoy boostrap config")
		return nil, err
	}

	probeListeners, probeClusters, err := getProbeResources(config)
	if err != nil {
		return nil, err
	}
	bootstrapConfig.StaticResources.Listeners = append(bootstrapConfig.StaticResources.Listeners, probeListeners...)
	bootstrapConfig.StaticResources.Clusters = append(bootstrapConfig.StaticResources.Clusters, probeClusters...)

	configYAML, err := utils.ProtoToYAML(bootstrapConfig)
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrMarshallingProtoToYAML)).
			Msgf("Failed to marshal envoy bootstrap config to yaml")
		return nil, err
	}
	return configYAML, nil
}

// getProbeResources returns the listener and cluster objects that are statically configured to serve
// startup, readiness and liveness probes.
// These will not change during the lifetime of the Pod.
// If the original probe defined a TCPSocket action, listener and cluster objects are not configured
// to serve that probe.
func getProbeResources(config sidecarBootstrapConfigMeta) ([]*xds_listener.Listener, []*xds_cluster.Cluster, error) {
	// This slice is the list of listeners for liveness, readiness, startup IF these have been configured in the Pod Spec
	var listeners []*xds_listener.Listener
	var clusters []*xds_cluster.Cluster

	// Is there a liveness probe in the Pod Spec?
	if config.OriginalHealthProbes.GetLiveness() != nil && !config.OriginalHealthProbes.GetLiveness().IsTCPSocket() {
		listener, err := getLivenessListener(config.OriginalHealthProbes.GetLiveness())
		if err != nil {
			log.Error().Err(err).Msgf("Error getting liveness listener")
			return nil, nil, err
		}
		listeners = append(listeners, listener)
		clusters = append(clusters, getLivenessCluster(config.OriginalHealthProbes.GetLiveness()))
	}

	// Is there a readiness probe in the Pod Spec?
	if config.OriginalHealthProbes.GetReadiness() != nil && !config.OriginalHealthProbes.GetReadiness().IsTCPSocket() {
		listener, err := getReadinessListener(config.OriginalHealthProbes.GetReadiness())
		if err != nil {
			log.Error().Err(err).Msgf("Error getting readiness listener")
			return nil, nil, err
		}
		listeners = append(listeners, listener)
		clusters = append(clusters, getReadinessCluster(config.OriginalHealthProbes.GetReadiness()))
	}

	// Is there a startup probe in the Pod Spec?
	if config.OriginalHealthProbes.GetStartup() != nil && !config.OriginalHealthProbes.GetStartup().IsTCPSocket() {
		listener, err := getStartupListener(config.OriginalHealthProbes.GetStartup())
		if err != nil {
			log.Error().Err(err).Msgf("Error getting startup listener")
			return nil, nil, err
		}
		listeners = append(listeners, listener)
		clusters = append(clusters, getStartupCluster(config.OriginalHealthProbes.GetStartup()))
	}

	return listeners, clusters, nil
}

func createSidecarBootstrapConfig(ctx driver.InjectorContext, sidecarBootstrapConfigName string) (*corev1.Secret, error) {
	configMeta := sidecarBootstrapConfigMeta{
		SidecarAdminPort: constants.SidecarAdminPort,
		XDSClusterName:   constants.OSMControllerName,
		NodeID:           ctx.BootstrapCertificate.GetCommonName().String(),

		RootCert: ctx.BootstrapCertificate.GetIssuingCA(),
		Cert:     ctx.BootstrapCertificate.GetCertificateChain(),
		Key:      ctx.BootstrapCertificate.GetPrivateKey(),

		XDSHost: fmt.Sprintf("%s.%s.svc.cluster.local", constants.OSMControllerName, ctx.OsmNamespace),
		XDSPort: constants.ProxyServerPort,

		// OriginalHealthProbes stores the path and port for liveness, readiness, and startup health probes as initially
		// defined on the Pod Spec.
		OriginalHealthProbes: ctx.OriginalHealthProbes,

		TLSMinProtocolVersion: ctx.Configurator.GetMeshConfig().Spec.Sidecar.TLSMinProtocolVersion,
		TLSMaxProtocolVersion: ctx.Configurator.GetMeshConfig().Spec.Sidecar.TLSMaxProtocolVersion,
		CipherSuites:          ctx.Configurator.GetMeshConfig().Spec.Sidecar.CipherSuites,
		ECDHCurves:            ctx.Configurator.GetMeshConfig().Spec.Sidecar.ECDHCurves,
	}
	yamlContent, err := getEnvoyConfigYAML(configMeta, ctx.Configurator)
	if err != nil {
		log.Error().Err(err).Msg("Error creating Sidecar bootstrap YAML")
		return nil, err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: sidecarBootstrapConfigName,
			Labels: map[string]string{
				constants.OSMAppNameLabelKey:     constants.OSMAppNameLabelValue,
				constants.OSMAppInstanceLabelKey: ctx.MeshName,
				constants.OSMAppVersionLabelKey:  version.Version,
			},
		},
		Data: map[string][]byte{
			envoyBootstrapConfigFile: yamlContent,
		},
	}
	return secret, nil
}
