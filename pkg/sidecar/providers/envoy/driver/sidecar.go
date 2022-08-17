package driver

import (
	"context"
	"errors"
	"os"

	corev1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/health"
	"github.com/openservicemesh/osm/pkg/injector"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/ads"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/envoy/registry"
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

	return xdsServer, xdsServer.Start(ctx, cancel, int(proxyServerPort), proxyServiceCert)
}

// Patch is the implement for InjectorDriver.Patch
func (sd EnvoySidecarDriver) Patch(ctx context.Context) error {
	parentCtx := ctx.Value(&driver.InjectorCtxKey)
	if parentCtx == nil {
		return errors.New("missing Injector Context")
	}
	injCtx := parentCtx.(*driver.InjectorContext)
	configurator := injCtx.Configurator
	osmNamespace := injCtx.OsmNamespace
	osmContainerPullPolicy := injCtx.OsmContainerPullPolicy
	namespace := injCtx.PodNamespace
	pod := injCtx.Pod
	podOS := injCtx.PodOS
	proxyUUID := injCtx.ProxyUUID
	bootstrapCertificate := injCtx.BootstrapCertificate
	cnPrefix := injCtx.BootstrapCertificateCNPrefix
	dryRun := injCtx.DryRun

	originalHealthProbes := injector.RewriteHealthProbes(pod)

	// Create the bootstrap configuration for the Envoy proxy for the given pod
	envoyBootstrapConfigName := injector.BootstrapSecretPrefix + proxyUUID.String()

	// This needs to occur before replacing the label below.
	originalUUID, alreadyInjected := injector.GetProxyUUID(pod)
	switch {
	case dryRun:
		// The webhook has a side effect (making out-of-band changes) of creating k8s secret
		// corresponding to the Envoy bootstrap config. Such a side effect needs to be skipped
		// when the request is a DryRun.
		// Ref: https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#side-effects
		log.Debug().Msgf("Skipping envoy bootstrap config creation for dry-run request: service-account=%s, namespace=%s", pod.Spec.ServiceAccountName, namespace)
	case alreadyInjected:
		// Pod definitions can be copied via the `kubectl debug` command, which can lead to a pod being created that
		// has already had injection occur. We could simply do nothing and return early, but that would leave 2 pods
		// with the same UUID, so instead we change the UUID, and create a new bootstrap config, copied from the original,
		// with the proxy UUID changed.
		oldConfigName := injector.BootstrapSecretPrefix + originalUUID
		if _, err := createEnvoyBootstrapFromExisting(injCtx, envoyBootstrapConfigName, oldConfigName, namespace, bootstrapCertificate); err != nil {
			log.Error().Err(err).Msgf("Failed to create Envoy bootstrap config for already-injected pod: service-account=%s, namespace=%s, certificate CN prefix=%s", pod.Spec.ServiceAccountName, namespace, cnPrefix)
			return err
		}
	default:
		if _, err := createEnvoyBootstrapConfig(injCtx, envoyBootstrapConfigName, namespace, osmNamespace, bootstrapCertificate, originalHealthProbes); err != nil {
			log.Error().Err(err).Msgf("Failed to create Envoy bootstrap config for pod: service-account=%s, namespace=%s, certificate CN prefix=%s", pod.Spec.ServiceAccountName, namespace, cnPrefix)
			return err
		}
	}

	if alreadyInjected {
		// replace the volume and we're done.
		for i, volume := range pod.Spec.Volumes {
			// It should be the last, but we check all for posterity.
			if volume.Name == injector.SidecarBootstrapConfigVolume {
				pod.Spec.Volumes[i] = injector.GetVolumeSpec(envoyBootstrapConfigName)
				break
			}
		}
		return nil
	}

	// Create volume for the envoy bootstrap config Secret
	pod.Spec.Volumes = append(pod.Spec.Volumes, injector.GetVolumeSpec(envoyBootstrapConfigName))

	err := injector.ConfigurePodInit(configurator, podOS, pod, osmContainerPullPolicy)
	if err != nil {
		return err
	}

	if originalHealthProbes.UsesTCP() {
		healthcheckContainer := corev1.Container{
			Name:            "osm-healthcheck",
			Image:           os.Getenv("OSM_DEFAULT_HEALTHCHECK_CONTAINER_IMAGE"),
			ImagePullPolicy: osmContainerPullPolicy,
			Args: []string{
				"--verbosity", log.GetLevel().String(),
			},
			Command: []string{
				"/osm-healthcheck",
			},
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: constants.HealthcheckPort,
				},
			},
		}
		pod.Spec.Containers = append(pod.Spec.Containers, healthcheckContainer)
	}

	// Add the Envoy sidecar
	sidecar := getEnvoySidecarContainerSpec(pod, configurator, originalHealthProbes, podOS)
	pod.Spec.Containers = append(pod.Spec.Containers, sidecar)

	return nil
}
