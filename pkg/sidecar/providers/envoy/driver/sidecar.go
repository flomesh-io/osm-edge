package driver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/health"
	"github.com/openservicemesh/osm/pkg/injector"
	"github.com/openservicemesh/osm/pkg/sidecar"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/ads"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/registry"
)

const (
	livenessProbePort  = int32(15901)
	readinessProbePort = int32(15902)
	startupProbePort   = int32(15903)
)

// EnvoySidecarDriver is the envoy sidecar driver
type EnvoySidecarDriver struct {
	ctx *driver.ControllerContext
}

// Start is the implement for ControllerDriver.Start
func (sd EnvoySidecarDriver) Start(ctx context.Context) (health.Probes, error) {
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
	// Create and start the ADS gRPC service
	xdsServer := ads.NewADSServer(ctrlCtx.MeshCatalog, proxyRegistry, cfg.IsDebugServerEnabled(), ctrlCtx.OsmNamespace, cfg, certManager, k8sClient, ctrlCtx.MsgBroker)

	ctrlCtx.DebugHandlers["/debug/proxy"] = sd.getProxies(proxyRegistry)
	ctrlCtx.DebugHandlers["/debug/xds"] = sd.getXDSHandler(xdsServer)

	return xdsServer, xdsServer.Start(ctx, cancel, proxyServerPort, proxyServiceCert)
}

// Patch is the implement for InjectorDriver.Patch
func (sd EnvoySidecarDriver) Patch(ctx context.Context, pod *corev1.Pod) ([]*corev1.Secret, error) {
	parentCtx := ctx.Value(&driver.InjectorCtxKey)
	if parentCtx == nil {
		return nil, errors.New("missing Injector Context")
	}
	injCtx := parentCtx.(*driver.InjectorContext)

	var secrets []*corev1.Secret
	// Create the bootstrap configuration for the Sidecar proxy for the given pod
	sidecarBootstrapConfigName := fmt.Sprintf("sidecar-bootstrap-config-%s", injCtx.ProxyUUID)

	// The webhook has a side effect (making out-of-band changes) of creating k8s secret
	// corresponding to the Sidecar bootstrap config. Such a side effect needs to be skipped
	// when the request is a DryRun.
	// Ref: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#side-effects
	if injCtx.DryRun {
		log.Debug().Msgf("Skipping Sidecar bootstrap config creation for dry-run request: service-account=%s, namespace=%s", pod.Spec.ServiceAccountName, injCtx.PodNamespace)
	} else if secret, err := createSidecarBootstrapConfig(*injCtx, sidecarBootstrapConfigName); err != nil {
		log.Error().Err(err).Msgf("Failed to create Sidecar bootstrap config for pod: service-account=%s, namespace=%s, certificate CN=%s", pod.Spec.ServiceAccountName, injCtx.PodNamespace, injCtx.ProxyCommonName)
		return nil, err
	} else {
		secrets = append(secrets, secret)
	}

	// Create volume for sidecar TLS secret
	pod.Spec.Volumes = append(pod.Spec.Volumes, injector.GetVolumeSpec(sidecarBootstrapConfigName)...)

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
	clusterID := fmt.Sprintf("%s.%s", pod.Spec.ServiceAccountName, pod.Namespace)
	securityContext, containerImage := sidecar.GetPlatformSpecificSpecComponents(injCtx.Configurator, injCtx.PodOS)
	sidecarContainer := corev1.Container{
		Name:            constants.SidecarContainerName,
		Image:           containerImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: securityContext,
		Ports:           getSidecarContainerPorts(injCtx.OriginalHealthProbes),
		VolumeMounts: []corev1.VolumeMount{{
			Name:      sidecarBootstrapConfigVolume,
			ReadOnly:  true,
			MountPath: envoyProxyConfigPath,
		}},
		Command:   []string{"envoy"},
		Resources: injCtx.Configurator.GetProxyResources(),
		Args: []string{
			"--log-level", injCtx.Configurator.GetSidecarLogLevel(),
			"--config-path", strings.Join([]string{envoyProxyConfigPath, envoyBootstrapConfigFile}, "/"),
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

	pod.Spec.Containers = append(pod.Spec.Containers, sidecarContainer)

	return secrets, nil
}
