package injector

import (
	"fmt"
	"github.com/openservicemesh/osm/pkg/certificate"
	corev1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/constants"
)

const (
	PipyAdminPort       = 6060
	pipyProxyConfigPath = "/etc/pipy"
)

func getPipySidecarContainerSpec(_ *corev1.Pod, cfg configurator.Configurator, originalHealthProbes healthProbes,
	podOS string, bootstrapCertificate *certificate.Certificate, osmNamespace string) corev1.Container {
	// cluster ID will be used as an identifier to the tracing sink
	securityContext, containerImage := getPlatformSpecificSpecComponents(cfg, podOS)
	pipyRepo := fmt.Sprintf("%s://%s.%s.svc.cluster.local:%v/repo/%s/", constants.ProtocolHTTP,
		constants.OSMControllerName, osmNamespace, constants.ADSServerPort, bootstrapCertificate.GetCommonName())

	return corev1.Container{
		Name:            constants.SidecarContainerName,
		Image:           containerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: securityContext,
		Ports:           getSidecarContainerPorts(originalHealthProbes),
		VolumeMounts: []corev1.VolumeMount{{
			Name:      sidecarBootstrapConfigVolume,
			ReadOnly:  true,
			MountPath: pipyProxyConfigPath,
		}},
		Command:   []string{"/usr/local/bin/pipy"},
		Resources: cfg.GetProxyResources(),
		Args: []string{
			fmt.Sprintf("--log-level=%s", cfg.GetSidecarLogLevel()),
			fmt.Sprintf("--admin-port=%d", PipyAdminPort),
			pipyRepo,
		},
		Env: []corev1.EnvVar{
			{
				Name: "POD_UID",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.uid",
					},
				},
			},
			{
				Name: "POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name: "POD_NAMESPACE",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
			{
				Name: "POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
			{
				Name: "SERVICE_ACCOUNT",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "spec.serviceAccountName",
					},
				},
			},
		},
	}
}
