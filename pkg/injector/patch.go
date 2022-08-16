package injector

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/metricsstore"
	"github.com/openservicemesh/osm/pkg/models"
	"github.com/openservicemesh/osm/pkg/sidecar"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
)

func (wh *mutatingWebhook) createPatch(pod *corev1.Pod, req *admissionv1.AdmissionRequest, proxyUUID uuid.UUID) ([]byte, error) {
	namespace := req.Namespace

	// On Windows we cannot use init containers to program HNS because it requires elevated privileges
	// As a result we assume that the HNS redirection policies are already programmed via a CNI plugin.
	// Skip adding the init container and only patch the pod spec with sidecar container.
	podOS := pod.Spec.NodeSelector["kubernetes.io/os"]
	if err := wh.verifyPrerequisites(podOS); err != nil {
		return nil, err
	}

	enableMetrics, err := IsMetricsEnabled(wh.kubeController, namespace)
	if err != nil {
		log.Error().Err(err).Msgf("Error checking if namespace %s is enabled for metrics", namespace)
		return nil, err
	}
	if enableMetrics {
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		pod.Annotations[constants.PrometheusScrapeAnnotation] = strconv.FormatBool(true)
		pod.Annotations[constants.PrometheusPortAnnotation] = strconv.Itoa(constants.SidecarPrometheusInboundListenerPort)
		pod.Annotations[constants.PrometheusPathAnnotation] = constants.PrometheusScrapePath
	}

	// This will append a label to the pod, which points to the unique Sidecar ID used in the
	// xDS certificate for that Sidecar. This label will help xDS match the actual pod to the Sidecar that
	// connects to xDS (with the certificate's CN matching this label).
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[constants.SidecarUniqueIDLabelName] = proxyUUID.String()

	// Issue a certificate for the proxy sidecar - used for Sidecar to connect to XDS (not Sidecar-to-Sidecar connections)
	cnPrefix := sidecar.NewCertCNPrefix(proxyUUID, models.KindSidecar, identity.New(pod.Spec.ServiceAccountName, namespace))
	log.Debug().Msgf("Patching POD spec: service-account=%s, namespace=%s with certificate CN prefix=%s", pod.Spec.ServiceAccountName, namespace, cnPrefix)
	startTime := time.Now()
	bootstrapCertificate, err := wh.certManager.IssueCertificate(cnPrefix, certificate.Internal)
	if err != nil {
		log.Error().Err(err).Msgf("Error issuing bootstrap certificate for Sidecar with CN prefix=%s", cnPrefix)
		return nil, err
	}
	elapsed := time.Since(startTime)

	metricsstore.DefaultMetricsStore.CertIssuedCount.Inc()
	metricsstore.DefaultMetricsStore.CertIssuedTime.WithLabelValues().Observe(elapsed.Seconds())

	background := driver.InjectorContext{
		KubeClient:                   wh.kubeClient,
		MeshName:                     wh.meshName,
		OsmNamespace:                 wh.osmNamespace,
		OsmContainerPullPolicy:       wh.osmContainerPullPolicy,
		Configurator:                 wh.configurator,
		Pod:                          pod,
		PodOS:                        podOS,
		PodNamespace:                 namespace,
		ProxyUUID:                    proxyUUID,
		BootstrapCertificateCNPrefix: cnPrefix,
		BootstrapCertificate:         bootstrapCertificate,
		DryRun:                       req.DryRun != nil && *req.DryRun,
	}
	ctx, cancel := context.WithCancel(&background)
	defer cancel()

	if err = sidecar.Patch(ctx); err != nil {
		return nil, err
	}

	return json.Marshal(makePatches(req, pod))
}

// verifyPrerequisites verifies if the prerequisites to patch the request are met by returning an error if unmet
func (wh *mutatingWebhook) verifyPrerequisites(podOS string) error {
	isWindows := strings.EqualFold(podOS, constants.OSWindows)

	// Verify that the required images are configured
	if image := wh.configurator.GetSidecarImage(); !isWindows && image == "" {
		// Linux pods require Sidecar Linux image
		return fmt.Errorf("MeshConfig sidecar.sidecarImage not set")
	}
	if image := wh.configurator.GetSidecarWindowsImage(); isWindows && image == "" {
		// Windows pods require Sidecar Windows image
		return fmt.Errorf("MeshConfig sidecar.sidecarWindowsImage not set")
	}
	if image := wh.configurator.GetInitContainerImage(); !isWindows && image == "" {
		// Linux pods require init container image
		return fmt.Errorf("MeshConfig sidecar.initContainerImage not set")
	}

	return nil
}

// ConfigurePodInit patch the init container to pod.
func ConfigurePodInit(cfg configurator.Configurator, podOS string, pod *corev1.Pod, osmContainerPullPolicy corev1.PullPolicy) error {
	if strings.EqualFold(podOS, constants.OSWindows) {
		// No init container for Windows
		return nil
	}

	// Build outbound port exclusion list
	podOutboundPortExclusionList, err := GetPortExclusionListForPod(pod, OutboundPortExclusionListAnnotation)
	if err != nil {
		return err
	}
	globalOutboundPortExclusionList := cfg.GetMeshConfig().Spec.Traffic.OutboundPortExclusionList
	outboundPortExclusionList := MergePortExclusionLists(podOutboundPortExclusionList, globalOutboundPortExclusionList)

	// Build inbound port exclusion list
	podInboundPortExclusionList, err := GetPortExclusionListForPod(pod, InboundPortExclusionListAnnotation)
	if err != nil {
		return err
	}
	globalInboundPortExclusionList := cfg.GetMeshConfig().Spec.Traffic.InboundPortExclusionList
	inboundPortExclusionList := MergePortExclusionLists(podInboundPortExclusionList, globalInboundPortExclusionList)

	// Build the outbound IP range exclusion list
	podOutboundIPRangeExclusionList, err := GetOutboundIPRangeListForPod(pod, OutboundIPRangeExclusionListAnnotation)
	if err != nil {
		return err
	}
	globalOutboundIPRangeExclusionList := cfg.GetMeshConfig().Spec.Traffic.OutboundIPRangeExclusionList
	outboundIPRangeExclusionList := MergeIPRangeLists(podOutboundIPRangeExclusionList, globalOutboundIPRangeExclusionList)

	// Build the outbound IP range inclusion list
	podOutboundIPRangeInclusionList, err := GetOutboundIPRangeListForPod(pod, OutboundIPRangeInclusionListAnnotation)
	if err != nil {
		return err
	}
	globalOutboundIPRangeInclusionList := cfg.GetMeshConfig().Spec.Traffic.OutboundIPRangeInclusionList
	outboundIPRangeInclusionList := MergeIPRangeLists(podOutboundIPRangeInclusionList, globalOutboundIPRangeInclusionList)

	networkInterfaceExclusionList := cfg.GetMeshConfig().Spec.Traffic.NetworkInterfaceExclusionList

	// Add the init container to the pod spec
	initContainer := GetInitContainerSpec(constants.InitContainerName, cfg, outboundIPRangeExclusionList, outboundIPRangeInclusionList, outboundPortExclusionList, inboundPortExclusionList, cfg.IsPrivilegedInitContainer(), osmContainerPullPolicy, networkInterfaceExclusionList)
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, initContainer)

	return nil
}

func makePatches(req *admissionv1.AdmissionRequest, pod *corev1.Pod) []jsonpatch.JsonPatchOperation {
	original := req.Object.Raw
	current, err := json.Marshal(pod)
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrMarshallingKubernetesResource)).
			Msgf("Error marshaling Pod with UID=%s", pod.ObjectMeta.UID)
	}
	admissionResponse := admission.PatchResponseFromRaw(original, current)
	return admissionResponse.Patches
}

// GetProxyUUID return proxy uuid retrieved from sidecar bootstrap config volume.
func GetProxyUUID(pod *corev1.Pod) (string, bool) {
	// kubectl debug does not recreate the object with the same metadata
	for _, volume := range pod.Spec.Volumes {
		if volume.Name == SidecarBootstrapConfigVolume {
			return strings.TrimPrefix(volume.Secret.SecretName, BootstrapSecretPrefix), true
		}
	}
	return "", false
}
