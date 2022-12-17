package driver

import (
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"

	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
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

	repoServerIPAddr := cfg.GetRepoServerIPAddr()
	if strings.HasPrefix(repoServerIPAddr, "127.") || strings.EqualFold(strings.ToLower(repoServerIPAddr), "localhost") {
		repoServerIPAddr = fmt.Sprintf("%s.%s", constants.OSMControllerName, injCtx.OsmNamespace)
	}
	repoServer := fmt.Sprintf("%s://%s:%v/repo%s/osm-edge-sidecar/%s/",
		constants.ProtocolHTTP, repoServerIPAddr, cfg.GetProxyServerPort(), cfg.GetRepoServerCodebase(), cnPrefix)
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
			repoServer,
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

	if injCtx.Configurator.IsLocalDNSProxyEnabled() {
		sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
			Name:  "LOCAL_DNS_PROXY",
			Value: "true",
		})
		if len(injCtx.Configurator.GetLocalDNSProxyPrimaryUpstream()) > 0 {
			sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
				Name:  "LOCAL_DNS_PROXY_PRIMARY_UPSTREAM",
				Value: injCtx.Configurator.GetLocalDNSProxyPrimaryUpstream(),
			})
		}
		if len(injCtx.Configurator.GetLocalDNSProxySecondaryUpstream()) > 0 {
			sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
				Name:  "LOCAL_DNS_PROXY_SECONDARY_UPSTREAM",
				Value: injCtx.Configurator.GetLocalDNSProxySecondaryUpstream(),
			})
		}
		if osmControllerSvc, err := getOSMControllerSvc(injCtx.KubeClient, injCtx.OsmNamespace); err == nil {
			pod.Spec.HostAliases = append(pod.Spec.HostAliases, corev1.HostAlias{
				IP:        osmControllerSvc.Spec.ClusterIP,
				Hostnames: []string{fmt.Sprintf("%s.%s", constants.OSMControllerName, injCtx.OsmNamespace)},
			})
		}

		pod.Spec.DNSPolicy = "None"
		trustDomain := injCtx.CertManager.GetTrustDomain()
		ndots := "5"
		pod.Spec.DNSConfig = &corev1.PodDNSConfig{
			Nameservers: []string{"127.0.0.153"},
			Searches:    []string{fmt.Sprintf("svc.%s", trustDomain), trustDomain},
			Options: []corev1.PodDNSConfigOption{
				{Name: "ndots", Value: &ndots},
			},
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

// getOSMControllerSvc returns the osm-controller service.
// The pod name is inferred from the 'CONTROLLER_SVC_NAME' env variable which is set during deployment.
func getOSMControllerSvc(kubeClient kubernetes.Interface, osmNamespace string) (*corev1.Service, error) {
	svcName := os.Getenv("CONTROLLER_SVC_NAME")
	if svcName == "" {
		return nil, fmt.Errorf("CONTROLLER_SVC_NAME env variable cannot be empty")
	}

	svc, err := kubeClient.CoreV1().Services(osmNamespace).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		// TODO(#3962): metric might not be scraped before process restart resulting from this error
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrFetchingControllerSvc)).
			Msgf("Error retrieving osm-controller service %s", svcName)
		return nil, err
	}

	return svc, nil
}
