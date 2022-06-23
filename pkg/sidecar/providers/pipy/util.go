package pipy

import (
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
)

var (
	// ErrInvalidCertificateCN is an error for when a certificate has a CommonName, which does not match expected string format.
	ErrInvalidCertificateCN = errors.New("invalid cn")

	// ErrMoreThanOnePodForCertificate is an error for when OSM finds more than one pod for a given xDS certificate. There should always be exactly one Pod for a given xDS certificate.
	ErrMoreThanOnePodForCertificate = errors.New("found more than one pod for xDS certificate")

	// ErrDidNotFindPodForCertificate is an error for when OSM cannot not find a pod for the given xDS certificate.
	ErrDidNotFindPodForCertificate = errors.New("did not find pod for certificate")

	// ErrServiceAccountDoesNotMatchCertificate is an error for when the service account of a Pod does not match the xDS certificate.
	ErrServiceAccountDoesNotMatchCertificate = errors.New("service account does not match certificate")

	// ErrNamespaceDoesNotMatchCertificate is an error for when the namespace of the Pod does not match the xDS certificate.
	ErrNamespaceDoesNotMatchCertificate = errors.New("namespace does not match certificate")
)

// certificateCommonNameMeta is the type that stores the metadata present in the CommonName field in a proxy's certificate
type certificateCommonNameMeta struct {
	ProxyUUID uuid.UUID
	ProxyKind ProxyKind
	// TODO(draychev): Change this to ServiceIdentity type (instead of string)
	ServiceIdentity identity.ServiceIdentity
}

func getCertificateCommonNameMeta(cn certificate.CommonName) (*certificateCommonNameMeta, error) {
	// XDS cert CN is of the form <proxy-UUID>.<kind>.<proxy-identity>
	chunks := strings.SplitN(cn.String(), constants.DomainDelimiter, 3)
	if len(chunks) < 3 {
		return nil, ErrInvalidCertificateCN
	}
	proxyUUID, err := uuid.Parse(chunks[0])
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrParsingXDSCertCN)).
			Msgf("Error parsing %s into uuid.UUID", chunks[0])
		return nil, err
	}

	return &certificateCommonNameMeta{
		ProxyUUID:       proxyUUID,
		ProxyKind:       ProxyKind(chunks[1]),
		ServiceIdentity: identity.ServiceIdentity(chunks[2]),
	}, nil
}

// GetPodFromCertificate returns the Kubernetes Pod object for a given certificate.
func GetPodFromCertificate(cn certificate.CommonName, kubecontroller k8s.Controller) (*v1.Pod, error) {
	cnMeta, err := getCertificateCommonNameMeta(cn)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("Looking for pod with label %q=%q", constants.SidecarUniqueIDLabelName, cnMeta.ProxyUUID)
	podList := kubecontroller.ListPods()
	var pods []v1.Pod
	for _, pod := range podList {
		if pod.Namespace != cnMeta.ServiceIdentity.ToK8sServiceAccount().Namespace {
			continue
		}
		if proxyUUID, labelFound := pod.Labels[constants.SidecarUniqueIDLabelName]; labelFound && proxyUUID == cnMeta.ProxyUUID.String() {
			pods = append(pods, *pod)
		}
	}

	if len(pods) == 0 {
		log.Error().Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrFetchingPodFromCert)).
			Msgf("Did not find Pod with label %s = %s in namespace %s",
				constants.SidecarUniqueIDLabelName, cnMeta.ProxyUUID, cnMeta.ServiceIdentity.ToK8sServiceAccount().Namespace)
		return nil, ErrDidNotFindPodForCertificate
	}

	// --- CONVENTION ---
	// By Open Service Mesh convention the number of services a pod can belong to is 1
	// This is a limitation we set in place in order to make the mesh easy to understand and reason about.
	// When a pod belongs to more than one service XDS will not program the Sidecar proxy, leaving it out of the mesh.
	if len(pods) > 1 {
		log.Error().Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrPodBelongsToMultipleServices)).
			Msgf("Found more than one pod with label %s = %s in namespace %s. There can be only one!",
				constants.SidecarUniqueIDLabelName, cnMeta.ProxyUUID, cnMeta.ServiceIdentity.ToK8sServiceAccount().Namespace)
		return nil, ErrMoreThanOnePodForCertificate
	}

	pod := pods[0]
	log.Trace().Msgf("Found Pod with UID=%s for proxyID %s", pod.ObjectMeta.UID, cnMeta.ProxyUUID)

	// Ensure the Namespace encoded in the certificate matches that of the Pod
	if pod.Namespace != cnMeta.ServiceIdentity.ToK8sServiceAccount().Namespace {
		log.Warn().Msgf("Pod with UID=%s belongs to Namespace %s. The pod's xDS certificate was issued for Namespace %s",
			pod.ObjectMeta.UID, pod.Namespace, cnMeta.ServiceIdentity.ToK8sServiceAccount().Namespace)
		return nil, ErrNamespaceDoesNotMatchCertificate
	}

	// Ensure the Name encoded in the certificate matches that of the Pod
	// TODO(draychev): check that the Kind matches too! [https://github.com/openservicemesh/osm/issues/3173]
	if pod.Spec.ServiceAccountName != cnMeta.ServiceIdentity.ToK8sServiceAccount().Name {
		// Since we search for the pod in the namespace we obtain from the certificate -- these namespaces will always match.
		log.Warn().Msgf("Pod with UID=%s belongs to ServiceAccount=%s. The pod's xDS certificate was issued for ServiceAccount=%s",
			pod.ObjectMeta.UID, pod.Spec.ServiceAccountName, cnMeta.ServiceIdentity.ToK8sServiceAccount())
		return nil, ErrServiceAccountDoesNotMatchCertificate
	}

	return &pod, nil
}

// GetServiceIdentityFromProxyCertificate returns the ServiceIdentity information encoded in the certificate CN
func GetServiceIdentityFromProxyCertificate(cn certificate.CommonName) (identity.ServiceIdentity, error) {
	cnMeta, err := getCertificateCommonNameMeta(cn)
	if err != nil {
		return "", err
	}

	return cnMeta.ServiceIdentity, nil
}
