package driver

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/injector"
	"github.com/openservicemesh/osm/pkg/models"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/bootstrap"
)

func getPlatformSpecificSpecComponents(cfg configurator.Configurator, podOS string) (podSecurityContext *corev1.SecurityContext, envoyContainer string) {
	if strings.EqualFold(podOS, constants.OSWindows) {
		podSecurityContext = &corev1.SecurityContext{
			WindowsOptions: &corev1.WindowsSecurityContextOptions{
				RunAsUserName: func() *string {
					userName := constants.SidecarWindowsUser
					return &userName
				}(),
			},
		}
		envoyContainer = cfg.GetSidecarWindowsImage()
	} else {
		podSecurityContext = &corev1.SecurityContext{
			AllowPrivilegeEscalation: pointer.BoolPtr(false),
			RunAsUser: func() *int64 {
				uid := constants.SidecarUID
				return &uid
			}(),
		}
		envoyContainer = cfg.GetSidecarImage()
	}
	return
}

func getEnvoySidecarContainerSpec(pod *corev1.Pod, cfg configurator.Configurator, originalHealthProbes models.HealthProbes, podOS string) corev1.Container {
	// cluster ID will be used as an identifier to the tracing sink
	clusterID := fmt.Sprintf("%s.%s", pod.Spec.ServiceAccountName, pod.Namespace)
	securityContext, containerImage := getPlatformSpecificSpecComponents(cfg, podOS)

	return corev1.Container{
		Name:            constants.SidecarContainerName,
		Image:           containerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: securityContext,
		Ports:           getEnvoyContainerPorts(originalHealthProbes),
		VolumeMounts: []corev1.VolumeMount{{
			Name:      injector.SidecarBootstrapConfigVolume,
			ReadOnly:  true,
			MountPath: bootstrap.EnvoyProxyConfigPath,
		}},
		Command:   []string{"envoy"},
		Resources: cfg.GetProxyResources(),
		Args: []string{
			"--log-level", cfg.GetSidecarLogLevel(),
			"--config-path", strings.Join([]string{bootstrap.EnvoyProxyConfigPath, bootstrap.EnvoyBootstrapConfigFile}, "/"),
			"--service-cluster", clusterID,
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

func getEnvoyContainerPorts(originalHealthProbes models.HealthProbes) []corev1.ContainerPort {
	containerPorts := []corev1.ContainerPort{
		{
			Name:          constants.SidecarAdminPortName,
			ContainerPort: constants.SidecarAdminPort,
		},
		{
			Name:          constants.SidecarInboundListenerPortName,
			ContainerPort: constants.SidecarInboundListenerPort,
		},
		{
			Name:          constants.SidecarInboundPrometheusListenerPortName,
			ContainerPort: constants.SidecarPrometheusInboundListenerPort,
		},
	}

	if originalHealthProbes.Liveness != nil {
		livenessPort := corev1.ContainerPort{
			// Name must be no more than 15 characters
			Name:          "liveness-port",
			ContainerPort: constants.LivenessProbePort,
		}
		containerPorts = append(containerPorts, livenessPort)
	}

	if originalHealthProbes.Readiness != nil {
		readinessPort := corev1.ContainerPort{
			// Name must be no more than 15 characters
			Name:          "readiness-port",
			ContainerPort: constants.ReadinessProbePort,
		}
		containerPorts = append(containerPorts, readinessPort)
	}

	if originalHealthProbes.Startup != nil {
		startupPort := corev1.ContainerPort{
			// Name must be no more than 15 characters
			Name:          "startup-port",
			ContainerPort: constants.StartupProbePort,
		}
		containerPorts = append(containerPorts, startupPort)
	}

	return containerPorts
}
