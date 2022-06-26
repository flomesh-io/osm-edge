package driver

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/health"
	"github.com/openservicemesh/osm/pkg/injector"
	"github.com/openservicemesh/osm/pkg/sidecar"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/repo"
)

const (
	livenessProbePort  = int32(15901)
	readinessProbePort = int32(15902)
	startupProbePort   = int32(15903)

	pipyAdminPort = 6060
)

// PipySidecarDriver is the pipy sidecar driver
type PipySidecarDriver struct {
	ctx *driver.ControllerContext
}

// Start is the implement for ControllerDriver.Start
func (sd PipySidecarDriver) Start(ctx context.Context) (health.Probes, error) {
	parentCtx := ctx.Value(&driver.ControllerCtxKey)
	if parentCtx == nil {
		return nil, errors.New("missing Controller Context")
	}
	ctrlCtx := parentCtx.(*driver.ControllerContext)
	cancel := ctrlCtx.CancelFunc
	cfg := ctrlCtx.Configurator
	certManager := ctrlCtx.CertManager
	k8sClient := ctrlCtx.MeshCatalog.GetKubeController()
	proxyServerPort := ctrlCtx.ProxyServerPort
	proxyServiceCert := ctrlCtx.ProxyServiceCert
	sd.ctx = ctrlCtx

	proxyMapper := &registry.KubeProxyServiceMapper{KubeController: k8sClient}
	proxyRegistry := registry.NewProxyRegistry(proxyMapper, ctrlCtx.MsgBroker)
	go proxyRegistry.ReleaseCertificateHandler(certManager, ctrlCtx.Stop)
	go proxyRegistry.CacheMeshPodsHandler(ctrlCtx.Stop)
	// Create and start the pipy repo http service
	repoServer := repo.NewRepoServer(ctrlCtx.MeshCatalog, proxyRegistry, cfg.IsDebugServerEnabled(), ctrlCtx.OsmNamespace, cfg, certManager, k8sClient, ctrlCtx.MsgBroker)

	ctrlCtx.DebugHandlers["/debug/proxy"] = sd.getProxies(proxyRegistry)

	return repoServer, repoServer.Start(ctx, cancel, proxyServerPort, proxyServiceCert)
}

// Patch is the implement for InjectorDriver.Patch
func (sd PipySidecarDriver) Patch(ctx context.Context, pod *corev1.Pod) ([]*corev1.Secret, error) {
	parentCtx := ctx.Value(&driver.InjectorCtxKey)
	if parentCtx == nil {
		return nil, errors.New("missing Injector Context")
	}
	injCtx := parentCtx.(*driver.InjectorContext)

	iptablesInitCommand := injector.GenerateIptablesCommands(injCtx.OutboundIPRangeExclusionList, injCtx.OutboundIPRangeInclusionList, injCtx.OutboundPortExclusionList, injCtx.InboundPortExclusionList)
	enablePrivilegedInitContainer := injCtx.Configurator.IsPrivilegedInitContainer()
	initContainer := corev1.Container{
		Name:            constants.InitContainerName,
		Image:           injCtx.Configurator.GetInitContainerImage(),
		ImagePullPolicy: injCtx.ContainerPullPolicy,
		SecurityContext: &corev1.SecurityContext{
			Privileged: &enablePrivilegedInitContainer,
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{
					"NET_ADMIN",
				},
			},
			RunAsNonRoot: pointer.BoolPtr(false),
			// User ID 0 corresponds to root
			RunAsUser: pointer.Int64Ptr(0),
		},
		Command: []string{"/bin/sh"},
		Args: []string{
			"-c",
			iptablesInitCommand,
		},
	}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainer)

	// cluster ID will be used as an identifier to the tracing sink
	securityContext, containerImage := sidecar.GetPlatformSpecificSpecComponents(injCtx.Configurator, injCtx.PodOS)
	pipyRepo := fmt.Sprintf("%s://%s.%s.svc.cluster.local:%v/repo/%s/", constants.ProtocolHTTP,
		constants.OSMControllerName, injCtx.OsmNamespace, constants.ProxyServerPort, injCtx.BootstrapCertificate.GetCommonName())

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

	sidecarContainer := corev1.Container{
		Name:            constants.SidecarContainerName,
		Image:           containerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: securityContext,
		Ports:           getSidecarContainerPorts(injCtx.OriginalHealthProbes),
		Resources:       injCtx.Configurator.GetProxyResources(),
		Args: []string{
			"pipy",
			fmt.Sprintf("--log-level=%s", injCtx.Configurator.GetSidecarLogLevel()),
			fmt.Sprintf("--admin-port=%d", pipyAdminPort),
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
				Name:  "POD_CONTROLLER_KIND",
				Value: podControllerKind,
			},
			{
				Name:  "POD_CONTROLLER_NAME",
				Value: podControllerName,
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

	if len(injCtx.Configurator.GetTracingHost()) > 0 {
		sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
			Name:  "TRACING_ADDRESS",
			Value: fmt.Sprintf("%s:%d", injCtx.Configurator.GetTracingHost(), injCtx.Configurator.GetTracingPort()),
		})
		sidecarContainer.Env = append(sidecarContainer.Env, corev1.EnvVar{
			Name:  "TRACING_ENDPOINT",
			Value: injCtx.Configurator.GetTracingEndpoint(),
		})
	}

	pod.Spec.Containers = append(pod.Spec.Containers, sidecarContainer)

	return nil, nil
}

func getSidecarContainerPorts(originalHealthProbes driver.HealthProbes) []corev1.ContainerPort {
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

	if originalHealthProbes.GetLiveness() != nil {
		livenessPort := corev1.ContainerPort{
			// Name must be no more than 15 characters
			Name:          "liveness-port",
			ContainerPort: livenessProbePort,
		}
		containerPorts = append(containerPorts, livenessPort)
	}

	if originalHealthProbes.GetReadiness() != nil {
		readinessPort := corev1.ContainerPort{
			// Name must be no more than 15 characters
			Name:          "readiness-port",
			ContainerPort: readinessProbePort,
		}
		containerPorts = append(containerPorts, readinessPort)
	}

	if originalHealthProbes.GetStartup() != nil {
		startupPort := corev1.ContainerPort{
			// Name must be no more than 15 characters
			Name:          "startup-port",
			ContainerPort: startupProbePort,
		}
		containerPorts = append(containerPorts, startupPort)
	}

	return containerPorts
}
