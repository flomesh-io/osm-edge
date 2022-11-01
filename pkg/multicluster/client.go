package multicluster

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	multiclusterv1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/k8s/informers"
	"github.com/openservicemesh/osm/pkg/messaging"
	"github.com/openservicemesh/osm/pkg/service"
)

// NewMultiClusterController returns a multicluster.Controller interface related to functionality provided by the resources in the flomesh.io API group
func NewMultiClusterController(informerCollection *informers.InformerCollection, kubeClient kubernetes.Interface, kubeController k8s.Controller, msgBroker *messaging.Broker) *Client {
	client := &Client{
		informers:      informerCollection,
		kubeClient:     kubeClient,
		kubeController: kubeController,
	}

	shouldObserve := func(obj interface{}) bool {
		object, ok := obj.(metav1.Object)
		if !ok {
			return false
		}
		return kubeController.IsMonitoredNamespace(object.GetNamespace())
	}

	svcImportEventTypes := k8s.EventTypes{
		Add:    announcements.ServiceImportAdded,
		Update: announcements.ServiceImportUpdated,
		Delete: announcements.ServiceImportDeleted,
	}
	client.informers.AddEventHandler(informers.InformerKeyServiceImport, k8s.GetEventHandlerFuncs(shouldObserve, svcImportEventTypes, msgBroker))

	return client
}

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

// ListServiceAccounts returns a list of service accounts that are part of monitored namespaces
func (c *Client) ListServiceAccounts() []*corev1.ServiceAccount {
	fmt.Println("ListServiceAccounts:")
	return nil
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
			namespace.Spec.Finalizers = append(namespace.Spec.Finalizers, "Flomesh")
			return namespace
		}
	}

	return nil
}

// ListPods returns a list of pods part of the mesh
// Kubecontroller does not currently segment pod notifications, hence it receives notifications
// for all k8s Pods.
func (c *Client) ListPods() []*corev1.Pod {
	fmt.Println("ListPods:")
	return nil
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
	fmt.Println("ListServiceIdentitiesForService:")
	return nil, nil
}
