package driver

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/bootstrap"
	"github.com/openservicemesh/osm/pkg/utils"
	"github.com/openservicemesh/osm/pkg/version"
)

const (
	envoyBootstrapConfigFile = "bootstrap.yaml"
)

func getPipyConfigYAML(config sidecarBootstrapConfigMeta, _ configurator.Configurator) ([]byte, error) {
	bootstrapConfig, err := bootstrap.BuildFromConfig(bootstrap.Config{
		NodeID:                config.NodeID,
		AdminPort:             constants.SidecarAdminPort,
		XDSClusterName:        constants.OSMControllerName,
		TrustedCA:             config.RootCert,
		CertificateChain:      config.Cert,
		PrivateKey:            config.Key,
		XDSHost:               config.ProxyServerHost,
		XDSPort:               config.ProxyServerPort,
		TLSMinProtocolVersion: config.TLSMinProtocolVersion,
		TLSMaxProtocolVersion: config.TLSMaxProtocolVersion,
		CipherSuites:          config.CipherSuites,
		ECDHCurves:            config.ECDHCurves,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Error building Envoy boostrap config")
		return nil, err
	}

	configYAML, err := utils.ProtoToYAML(bootstrapConfig)
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrMarshallingProtoToYAML)).
			Msgf("Failed to marshal envoy bootstrap config to yaml")
		return nil, err
	}
	return configYAML, nil
}

func createSidecarBootstrapConfig(ctx driver.InjectorContext, sidecarBootstrapConfigName string) (*corev1.Secret, error) {
	configMeta := sidecarBootstrapConfigMeta{
		SidecarAdminPort: constants.SidecarAdminPort,
		OSMClusterName:   constants.OSMControllerName,
		NodeID:           ctx.BootstrapCertificate.GetCommonName().String(),

		RootCert: ctx.BootstrapCertificate.GetIssuingCA(),
		Cert:     ctx.BootstrapCertificate.GetCertificateChain(),
		Key:      ctx.BootstrapCertificate.GetPrivateKey(),

		ProxyServerHost: fmt.Sprintf("%s.%s.svc.cluster.local", constants.OSMControllerName, ctx.OsmNamespace),
		ProxyServerPort: ctx.Configurator.GetProxyServerPort(),

		// OriginalHealthProbes stores the path and port for liveness, readiness, and startup health probes as initially
		// defined on the Pod Spec.
		OriginalHealthProbes: ctx.OriginalHealthProbes,

		TLSMinProtocolVersion: ctx.Configurator.GetMeshConfig().Spec.Sidecar.TLSMinProtocolVersion,
		TLSMaxProtocolVersion: ctx.Configurator.GetMeshConfig().Spec.Sidecar.TLSMaxProtocolVersion,
		CipherSuites:          ctx.Configurator.GetMeshConfig().Spec.Sidecar.CipherSuites,
		ECDHCurves:            ctx.Configurator.GetMeshConfig().Spec.Sidecar.ECDHCurves,
	}
	yamlContent, err := getPipyConfigYAML(configMeta, ctx.Configurator)
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
