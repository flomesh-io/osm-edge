package multicluster

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"strings"

	mapset "github.com/deckarep/golang-set"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"

	multiclusterv1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"

	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s/informers"
	"github.com/openservicemesh/osm/pkg/service"
)

// GetService retrieves the Kubernetes Services resource for the given MeshService
func (c *Client) GetService(svc service.MeshService) *corev1.Service {
	importedServiceIf, exists, err := c.informers.GetByKey(informers.InformerKeyServiceImport, svc.NamespacedKey())
	if !exists || err != nil {
		return nil
	}

	importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)
	if len(importedService.Spec.Ports) == 0 {
		return nil
	}

	for _, port := range importedService.Spec.Ports {
		if strings.EqualFold(importedService.Name, svc.Name) &&
			uint16(port.Port) == svc.Port &&
			len(port.Endpoints) > 0 {
			for _, endpoint := range port.Endpoints {
				if svc.TargetPort == uint16(endpoint.Target.Port) {
					targetSvc := new(corev1.Service)
					targetSvc.Namespace = importedService.Namespace
					targetSvc.Name = importedService.Name
					targetSvc.Spec.Type = corev1.ServiceTypeClusterIP
					targetSvc.Spec.Selector = map[string]string{"app": importedService.Name}
					targetSvc.Spec.ClusterIP = endpoint.Target.IP
					targetSvc.Spec.ClusterIPs = append(targetSvc.Spec.ClusterIPs, targetSvc.Spec.ClusterIP)
					targetSvcPort := corev1.ServicePort{
						Name:        port.Name,
						Protocol:    port.Protocol,
						AppProtocol: port.AppProtocol,
						Port:        port.Port,
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: endpoint.Target.Port,
						},
					}
					targetSvc.Spec.Ports = append(targetSvc.Spec.Ports, targetSvcPort)
					return targetSvc
				}
			}
		}
	}
	return nil
}

// ListServices returns a list of services that are imported from other clusters.
func (c *Client) ListServices() []*corev1.Service {
	importedServiceIfs := c.informers.List(informers.InformerKeyServiceImport)
	if len(importedServiceIfs) == 0 {
		return nil
	}

	var services []*corev1.Service

	for _, importedServiceIf := range importedServiceIfs {
		importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)
		if len(importedService.Spec.Ports) == 0 {
			continue
		}

		for _, port := range importedService.Spec.Ports {
			if len(port.Endpoints) > 0 {
				for _, endpoint := range port.Endpoints {
					targetSvc := new(corev1.Service)
					targetSvc.Namespace = importedService.Namespace
					targetSvc.Name = importedService.Name
					targetSvc.Spec.Type = corev1.ServiceTypeClusterIP
					targetSvc.Spec.Selector = map[string]string{"app": importedService.Name}
					targetSvc.Spec.ClusterIP = endpoint.Target.IP
					targetSvc.Spec.ClusterIPs = append(targetSvc.Spec.ClusterIPs, targetSvc.Spec.ClusterIP)
					targetSvcPort := corev1.ServicePort{
						Name:        port.Name,
						Protocol:    port.Protocol,
						AppProtocol: port.AppProtocol,
						Port:        port.Port,
						TargetPort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: endpoint.Target.Port,
						},
					}
					targetSvc.Spec.Ports = append(targetSvc.Spec.Ports, targetSvcPort)
					services = append(services, targetSvc)
				}
			}
		}
	}
	return services
}

// GetNamespace returns a Namespace resource if found, nil otherwise.
func (c *Client) GetNamespace(ns string) *corev1.Namespace {
	importedServiceIfs := c.informers.List(informers.InformerKeyServiceImport)
	if len(importedServiceIfs) == 0 {
		return nil
	}

	for _, importedServiceIf := range importedServiceIfs {
		importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)
		if strings.EqualFold(importedService.Namespace, ns) {
			namespace := new(corev1.Namespace)
			namespace.Name = importedService.Namespace
			namespace.Spec.Finalizers = append(namespace.Spec.Finalizers, "flomesh.io")
			return namespace
		}
	}

	return nil
}

// ListPods returns a list of pods part of the mesh
// Kubecontroller does not currently segment pod notifications, hence it receives notifications
// for all k8s Pods.
func (c *Client) ListPods() []*corev1.Pod {
	importedServiceIfs := c.informers.List(informers.InformerKeyServiceImport)
	if len(importedServiceIfs) == 0 {
		return nil
	}

	var pods []*corev1.Pod

	for _, importedServiceIf := range importedServiceIfs {
		importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)
		if len(importedService.Spec.Ports) == 0 {
			continue
		}
		for _, port := range importedService.Spec.Ports {
			if len(port.Endpoints) == 0 {
				continue
			}
			for _, endpoint := range port.Endpoints {
				pod := new(corev1.Pod)
				pod.Namespace = importedService.Namespace
				pod.Name = endpoint.Target.Host
				pod.Labels = make(map[string]string)
				pod.Labels["app"] = importedService.Name
				pod.Spec.ServiceAccountName = importedService.Spec.ServiceAccountName
				pod.Status.PodIP = endpoint.Target.IP
				pod.Status.PodIPs = append(pod.Status.PodIPs, corev1.PodIP{IP: pod.Status.PodIP})
				pods = append(pods, pod)
			}
		}
	}
	return pods
}

// GetEndpoints returns the endpoint for a given service, otherwise returns nil if not found
// or error if the API errored out.
func (c *Client) GetEndpoints(svc service.MeshService) (*corev1.Endpoints, error) {
	importedServiceIf, exists, err := c.informers.GetByKey(informers.InformerKeyServiceImport, svc.NamespacedKey())
	if err != nil || !exists {
		return nil, nil
	}

	importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)
	if len(importedService.Spec.Ports) == 0 {
		return nil, nil
	}

	for _, port := range importedService.Spec.Ports {
		if strings.EqualFold(importedService.Name, svc.Name) &&
			(svc.Port == 0 || svc.Port == uint16(port.Port)) &&
			len(port.Endpoints) > 0 {
			for _, endpoint := range port.Endpoints {
				if svc.TargetPort > 0 && svc.TargetPort != uint16(endpoint.Target.Port) {
					continue
				}
				targetEndpoints := new(corev1.Endpoints)
				targetEndpoints.Annotations = make(map[string]string)
				targetEndpoints.Annotations[ServiceImportClusterKeyAnnotation] = endpoint.ClusterKey
				targetEndpoints.Annotations[ServiceImportContextPathAnnotation] = endpoint.Target.Path
				targetEndpoints.Namespace = importedService.Namespace
				targetEndpoints.Name = importedService.Name
				targetEndpoints.Subsets = append(targetEndpoints.Subsets, corev1.EndpointSubset{
					Addresses: []corev1.EndpointAddress{
						{
							IP:       endpoint.Target.IP,
							Hostname: endpoint.Target.Host,
						},
					},
					Ports: []corev1.EndpointPort{
						{
							Name:        port.Name,
							Protocol:    port.Protocol,
							AppProtocol: port.AppProtocol,
							Port:        endpoint.Target.Port,
						},
					},
				})
				return targetEndpoints, nil
			}
		}
	}

	return nil, nil
}

// ListServiceIdentitiesForService lists ServiceAccounts associated with the given service
func (c *Client) ListServiceIdentitiesForService(svc service.MeshService) ([]identity.K8sServiceAccount, error) {
	var svcAccounts []identity.K8sServiceAccount

	k8sSvc := c.GetService(svc)
	if k8sSvc == nil {
		return nil, fmt.Errorf("error fetching service %q: %s", svc, errServiceNotFound)
	}

	svcAccountsSet := mapset.NewSet()
	pods := c.ListPods()
	for _, pod := range pods {
		svcRawSelector := k8sSvc.Spec.Selector
		selector := labels.Set(svcRawSelector).AsSelector()
		// service has no selectors, we do not need to match against the pod label
		if len(svcRawSelector) == 0 {
			continue
		}
		if selector.Matches(labels.Set(pod.Labels)) {
			podSvcAccount := identity.K8sServiceAccount{
				Name:      pod.Spec.ServiceAccountName,
				Namespace: pod.Namespace, // ServiceAccount must belong to the same namespace as the pod
			}
			svcAccountsSet.Add(podSvcAccount)
		}
	}

	for svcAcc := range svcAccountsSet.Iter() {
		svcAccounts = append(svcAccounts, svcAcc.(identity.K8sServiceAccount))
	}
	return svcAccounts, nil
}

// GetTargetPortForServicePort returns the TargetPort corresponding to the Port used by clients
// to communicate with it.
func (c Client) GetTargetPortForServicePort(namespacedSvc types.NamespacedName, port uint16) (uint16, error) {
	meshService := service.MeshService{
		Namespace: namespacedSvc.Namespace, // Backends belong to the same namespace as the apex service
		Name:      namespacedSvc.Name,
	}
	importedServiceIf, exists, err := c.informers.GetByKey(informers.InformerKeyServiceImport, meshService.NamespacedKey())
	if !exists || err != nil {
		return 0, fmt.Errorf("service %s not found in multi cluster cache", namespacedSvc)
	}

	importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)
	if len(importedService.Spec.Ports) == 0 {
		return 0, fmt.Errorf("service port %s not found in multi cluster cache", namespacedSvc)
	}

	for _, svcPort := range importedService.Spec.Ports {
		if strings.EqualFold(importedService.Name, meshService.Name) &&
			uint16(svcPort.Port) == port &&
			len(svcPort.Endpoints) > 0 {
			for _, endpoint := range svcPort.Endpoints {
				return uint16(endpoint.Target.Port), nil
			}
		}
	}

	return 0, fmt.Errorf("endpoint for service %s not found in multi cluster cache", namespacedSvc)
}
