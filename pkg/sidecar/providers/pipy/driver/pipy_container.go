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
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/bootstrap"
)

func getPlatformSpecificSpecComponents(cfg configurator.Configurator, podOS string) (podSecurityContext *corev1.SecurityContext, pipyContainer string) {
	if strings.EqualFold(podOS, constants.OSWindows) {
		podSecurityContext = &corev1.SecurityContext{
			WindowsOptions: &corev1.WindowsSecurityContextOptions{
				RunAsUserName: func() *string {
					userName := constants.SidecarWindowsUser
					return &userName
				}(),
			},
		}
		pipyContainer = cfg.GetSidecarWindowsImage()
	} else {
		podSecurityContext = &corev1.SecurityContext{
			AllowPrivilegeEscalation: pointer.BoolPtr(false),
			RunAsUser: func() *int64 {
				uid := constants.SidecarUID
				return &uid
			}(),
		}
		pipyContainer = cfg.GetSidecarImage()
	}
	return
}

func getPipySidecarContainerSpec(injCtx *driver.InjectorContext, pod *corev1.Pod, cfg configurator.Configurator, cnPrefix string, originalHealthProbes models.HealthProbes, podOS string) corev1.Container {
	securityContext, containerImage := getPlatformSpecificSpecComponents(cfg, podOS)

	podControllerKind := ""
	podControllerName := ""
	for _, ref := range pod.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			podControllerKind = ref.Kind
			podControllerName = ref.Name
			break
		}
	}
	// Assume ReplicaSets are controlled by a Deployment unless their names
	// do not contain a hyphen. This aligns with the behavior of the
	// Prometheus config in the OSM Helm chart.
	if podControllerKind == "ReplicaSet" {
		if hyp := strings.LastIndex(podControllerName, "-"); hyp >= 0 {
			podControllerKind = "Deployment"
			podControllerName = podControllerName[:hyp]
		}
	}

	pipyRepo := fmt.Sprintf("%s://%s.%s:%v/repo/osm-edge-sidecar/%s/", constants.ProtocolHTTP,
		constants.OSMControllerName, injCtx.OsmNamespace, cfg.GetProxyServerPort(), cnPrefix)
	sidecarContainer := corev1.Container{
		Name:            constants.SidecarContainerName,
		Image:           containerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: securityContext,
		Ports:           getPipyContainerPorts(originalHealthProbes),
		VolumeMounts: []corev1.VolumeMount{{
			Name:      injector.SidecarBootstrapConfigVolume,
			ReadOnly:  true,
			MountPath: bootstrap.PipyProxyConfigPath,
		}},
		Resources: cfg.GetProxyResources(),
		Args: []string{
			"pipy",
			fmt.Sprintf("--log-level=%s", injCtx.Configurator.GetSidecarLogLevel()),
			fmt.Sprintf("--admin-port=%d", cfg.GetProxyServerPort()),
			pipyRepo,
		},
		Env: []corev1.EnvVar{
			{
				Name:  "MESH_NAME",
				Value: injCtx.MeshName,
			},
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
			{
				Name:  "POD_CONTROLLER_KIND",
				Value: podControllerKind,
			},
			{
				Name:  "POD_CONTROLLER_NAME",
				Value: podControllerName,
			},
		},
	}

	if injCtx.Configurator.IsTracingEnabled() {
		if len(injCtx.Configurator.GetTracingHost()) > 0 && injCtx.Configurator.GetTracingPort() > 0 {
			sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
				Name:  "TRACING_ADDRESS",
				Value: fmt.Sprintf("%s:%d", injCtx.Configurator.GetTracingHost(), injCtx.Configurator.GetTracingPort()),
			})
		}
		if len(injCtx.Configurator.GetTracingEndpoint()) > 0 {
			sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
				Name:  "TRACING_ENDPOINT",
				Value: injCtx.Configurator.GetTracingEndpoint(),
			})
		}
	}

	if injCtx.Configurator.IsRemoteLoggingEnabled() {
		if len(injCtx.Configurator.GetRemoteLoggingHost()) > 0 && injCtx.Configurator.GetRemoteLoggingPort() > 0 {
			sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
				Name:  "REMOTE_LOGGING_ADDRESS",
				Value: fmt.Sprintf("%s:%d", injCtx.Configurator.GetRemoteLoggingHost(), injCtx.Configurator.GetRemoteLoggingPort()),
			})
		}
		if len(injCtx.Configurator.GetRemoteLoggingEndpoint()) > 0 {
			sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
				Name:  "REMOTE_LOGGING_ENDPOINT",
				Value: injCtx.Configurator.GetRemoteLoggingEndpoint(),
			})
		}
		if len(injCtx.Configurator.GetRemoteLoggingAuthorization()) > 0 {
			sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
				Name:  "REMOTE_LOGGING_AUTHORIZATION",
				Value: injCtx.Configurator.GetRemoteLoggingAuthorization(),
			})
		}
	}

	return sidecarContainer
}

func getPipyContainerPorts(originalHealthProbes models.HealthProbes) []corev1.ContainerPort {
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
