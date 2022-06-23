package injector

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/metricsstore"
	"github.com/openservicemesh/osm/pkg/sidecar"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
)

func (wh *mutatingWebhook) createPatch(pod *corev1.Pod, req *admissionv1.AdmissionRequest, proxyUUID uuid.UUID) ([]byte, error) {
	namespace := req.Namespace

	// Issue a certificate for the proxy sidecar - used for Sidecar to connect to XDS (not Sidecar-to-Sidecar connections)
	cn := sidecar.NewCertCommonName(proxyUUID, sidecar.KindSidecar, pod.Spec.ServiceAccountName, namespace)
	log.Debug().Msgf("Patching POD spec: service-account=%s, namespace=%s with certificate CN=%s", pod.Spec.ServiceAccountName, namespace, cn)
	startTime := time.Now()
	bootstrapCertificate, err := wh.certManager.IssueCertificate(cn, constants.XDSCertificateValidityPeriod)
	if err != nil {
		log.Error().Err(err).Msgf("Error issuing bootstrap certificate for Sidecar with CN=%s", cn)
		return nil, err
	}
	elapsed := time.Since(startTime)

	metricsstore.DefaultMetricsStore.CertIssuedCount.Inc()
	metricsstore.DefaultMetricsStore.CertIssuedTime.
		WithLabelValues().Observe(elapsed.Seconds())
	originalHealthProbes := rewriteHealthProbes(pod)

	// On Windows we cannot use init containers to program HNS because it requires elevated privileges
	// As a result we assume that the HNS redirection policies are already programmed via a CNI plugin.
	// Skip adding the init container and only patch the pod spec with sidecar container.
	podOS := pod.Spec.NodeSelector["kubernetes.io/os"]
	if err = wh.verifyPrerequisites(podOS); err != nil {
		return nil, err
	}

	inboundPortExclusionList, outboundPortExclusionList, outboundIPRangeInclusionList, outboundIPRangeExclusionList,
		err := wh.configurePodPortAndIPRangeInit(podOS, pod, namespace)
	if err != nil {
		return nil, err
	}

	if (originalHealthProbes.GetLiveness() != nil && originalHealthProbes.GetLiveness().IsTCPSocket()) ||
		(originalHealthProbes.GetReadiness() != nil && originalHealthProbes.GetReadiness().IsTCPSocket()) ||
		(originalHealthProbes.GetStartup() != nil && originalHealthProbes.GetStartup().IsTCPSocket()) {
		healthcheckContainer := corev1.Container{
			Name:            "osm-healthcheck",
			Image:           os.Getenv("OSM_DEFAULT_HEALTHCHECK_CONTAINER_IMAGE"),
			ImagePullPolicy: wh.osmContainerPullPolicy,
			Args: []string{
				"--verbosity", log.GetLevel().String(),
			},
			Command: []string{
				"/osm-healthcheck",
			},
			Ports: []corev1.ContainerPort{
				{
					ContainerPort: healthcheckPort,
				},
			},
		}
		pod.Spec.Containers = append(pod.Spec.Containers, healthcheckContainer)
	}

	enableMetrics, err := wh.isMetricsEnabled(namespace)
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

	background := driver.InjectorContext{
		MeshName:                     wh.meshName,
		OsmNamespace:                 wh.osmNamespace,
		PodNamespace:                 namespace,
		PodOS:                        podOS,
		ProxyCommonName:              cn,
		ProxyUUID:                    proxyUUID,
		Configurator:                 wh.configurator,
		BootstrapCertificate:         bootstrapCertificate,
		ContainerPullPolicy:          wh.osmContainerPullPolicy,
		InboundPortExclusionList:     inboundPortExclusionList,
		OutboundPortExclusionList:    outboundPortExclusionList,
		OutboundIPRangeInclusionList: outboundIPRangeInclusionList,
		OutboundIPRangeExclusionList: outboundIPRangeExclusionList,
		OriginalHealthProbes:         originalHealthProbes,
		DryRun:                       req.DryRun != nil && *req.DryRun,
	}
	ctx, cancel := context.WithCancel(&background)
	defer cancel()

	secrets, err := sidecar.Patch(ctx, pod)
	if err != nil {
		return nil, err
	}

	if len(secrets) > 0 {
		for _, secret := range secrets {
			if existing, err := wh.kubeClient.CoreV1().Secrets(namespace).Get(context.Background(), secret.ObjectMeta.Name, metav1.GetOptions{}); err == nil {
				log.Debug().Msgf("Updating bootstrap config Envoy: name=%s, namespace=%s", secret.ObjectMeta.Name, namespace)
				existing.Data = secret.Data
				_, err = wh.kubeClient.CoreV1().Secrets(namespace).Update(context.Background(), existing, metav1.UpdateOptions{})
				if err != nil {
					return nil, err
				}
			}

			log.Debug().Msgf("Creating bootstrap config for Envoy: name=%s, namespace=%s", secret.ObjectMeta.Name, namespace)
			_, err = wh.kubeClient.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
			if err != nil {
				return nil, err
			}
		}
	}

	return json.Marshal(makePatches(req, pod))
}

// verifyPrerequisites verifies if the prerequisites to patch the request are met by returning an error if unmet
func (wh *mutatingWebhook) verifyPrerequisites(podOS string) error {
	isWindows := strings.EqualFold(podOS, constants.OSWindows)

	// Verify that the required images are configured
	if image := wh.configurator.GetSidecarImage(); !isWindows && image == "" {
		// Linux pods require Sidecar Linux image
		return errors.New("MeshConfig sidecar.sidecarImage not set")
	}
	if image := wh.configurator.GetSidecarWindowsImage(); isWindows && image == "" {
		// Windows pods require Sidecar Windows image
		return errors.New("MeshConfig sidecar.sidecarWindowsImage not set")
	}
	if image := wh.configurator.GetInitContainerImage(); !isWindows && image == "" {
		// Linux pods require init container image
		return errors.New("MeshConfig sidecar.initContainerImage not set")
	}

	return nil
}

func (wh *mutatingWebhook) configurePodPortAndIPRangeInit(podOS string, pod *corev1.Pod, namespace string) ([]int, []int, []string, []string, error) {
	if strings.EqualFold(podOS, constants.OSWindows) {
		// No init container for Windows
		return nil, nil, nil, nil, nil
	}

	// Build outbound port exclusion list
	podOutboundPortExclusionList, err := getPortExclusionListForPod(pod, namespace, outboundPortExclusionListAnnotation)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	globalOutboundPortExclusionList := wh.configurator.GetMeshConfig().Spec.Traffic.OutboundPortExclusionList
	outboundPortExclusionList := mergePortExclusionLists(podOutboundPortExclusionList, globalOutboundPortExclusionList)

	// Build inbound port exclusion list
	podInboundPortExclusionList, err := getPortExclusionListForPod(pod, namespace, inboundPortExclusionListAnnotation)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	globalInboundPortExclusionList := wh.configurator.GetMeshConfig().Spec.Traffic.InboundPortExclusionList
	inboundPortExclusionList := mergePortExclusionLists(podInboundPortExclusionList, globalInboundPortExclusionList)

	// Build the outbound IP range exclusion list
	podOutboundIPRangeExclusionList, err := getOutboundIPRangeListForPod(pod, namespace, outboundIPRangeExclusionListAnnotation)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	globalOutboundIPRangeExclusionList := wh.configurator.GetMeshConfig().Spec.Traffic.OutboundIPRangeExclusionList
	outboundIPRangeExclusionList := mergeIPRangeLists(podOutboundIPRangeExclusionList, globalOutboundIPRangeExclusionList)

	// Build the outbound IP range inclusion list
	podOutboundIPRangeInclusionList, err := getOutboundIPRangeListForPod(pod, namespace, outboundIPRangeInclusionListAnnotation)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	globalOutboundIPRangeInclusionList := wh.configurator.GetMeshConfig().Spec.Traffic.OutboundIPRangeInclusionList
	outboundIPRangeInclusionList := mergeIPRangeLists(podOutboundIPRangeInclusionList, globalOutboundIPRangeInclusionList)

	return inboundPortExclusionList, outboundPortExclusionList, outboundIPRangeInclusionList, outboundIPRangeExclusionList, nil
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
