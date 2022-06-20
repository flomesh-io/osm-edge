package registry

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/service"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
)

// ProxyServiceMapper knows how to map Sidecar instances to services.
type ProxyServiceMapper interface {
	ListProxyServices(*pipy.Proxy) ([]service.MeshService, error)
	ListProxyPods() []v1.Pod
}

// ExplicitProxyServiceMapper is a custom ProxyServiceMapper implementation.
type ExplicitProxyServiceMapper func(*pipy.Proxy) ([]service.MeshService, error)

// ListProxyServices executes the given mapping.
func (e ExplicitProxyServiceMapper) ListProxyServices(p *pipy.Proxy) ([]service.MeshService, error) {
	return e(p)
}

// KubeProxyServiceMapper maps an Sidecar instance to services in a Kubernetes cluster.
type KubeProxyServiceMapper struct {
	KubeController k8s.Controller
}

// ListProxyPods maps an Sidecar instance to a number of Kubernetes services.
func (k *KubeProxyServiceMapper) ListProxyPods() []v1.Pod {
	allPods := k.KubeController.ListPods()
	var matchedPods []v1.Pod
	for _, pod := range allPods {
		if _, exists := pod.Labels[constants.SidecarUniqueIDLabelName]; exists {
			matchedPods = append(matchedPods, *pod)
		}
	}
	return matchedPods
}

// ListProxyServices retrives mesh services by proxy
func (k *KubeProxyServiceMapper) ListProxyServices(p *pipy.Proxy) ([]service.MeshService, error) {
	cn := p.GetCertificateCommonName()

	pod, err := pipy.GetPodFromCertificate(cn, k.KubeController)
	if err != nil {
		return nil, err
	}

	services := listServicesForPod(pod, k.KubeController)

	if len(services) == 0 {
		return nil, nil
	}

	meshServices := kubernetesServicesToMeshServices(k.KubeController, services)

	servicesForPod := strings.Join(listServiceNames(meshServices), ",")
	log.Trace().Msgf("Services associated with Pod with UID=%s Name=%s/%s: %+v",
		pod.ObjectMeta.UID, pod.Namespace, pod.Name, servicesForPod)

	return meshServices, nil
}

func kubernetesServicesToMeshServices(kubeController k8s.Controller, kubernetesServices []v1.Service) (meshServices []service.MeshService) {
	for _, svc := range kubernetesServices {
		meshServices = append(meshServices, k8s.ServiceToMeshServices(kubeController, svc)...)
	}
	return meshServices
}

func listServiceNames(meshServices []service.MeshService) (serviceNames []string) {
	for _, meshService := range meshServices {
		serviceNames = append(serviceNames, fmt.Sprintf("%s/%s", meshService.Namespace, meshService.Name))
	}
	return serviceNames
}

// listServicesForPod lists Kubernetes services whose selectors match pod labels
func listServicesForPod(pod *v1.Pod, kubeController k8s.Controller) []v1.Service {
	var serviceList []v1.Service
	svcList := kubeController.ListServices()

	for _, svc := range svcList {
		if svc.Namespace != pod.Namespace {
			continue
		}
		svcRawSelector := svc.Spec.Selector
		// service has no selectors, we do not need to match against the pod label
		if len(svcRawSelector) == 0 {
			continue
		}
		selector := labels.Set(svcRawSelector).AsSelector()
		if selector.Matches(labels.Set(pod.Labels)) {
			serviceList = append(serviceList, *svc)
		}
	}

	return serviceList
}
