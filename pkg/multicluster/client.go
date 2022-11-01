package multicluster

import (
	"encoding/json"
	"fmt"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/service"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"

	multiclusterv1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"
	"github.com/openservicemesh/osm/pkg/k8s/informers"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/messaging"
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
	fmt.Println("GetService:", svc)
	// client-go cache uses <namespace>/<name> as key
	importedServiceIf, exists, err := c.informers.GetByKey(informers.InformerKeyServiceImport, svc.NamespacedKey())
	if exists && err == nil {
		importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)

		svc := new(corev1.Service)
		svc.Namespace = importedService.Namespace
		svc.Name = importedService.Name
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = map[string]string{"app": importedService.Name}
		svc.Spec.ClusterIP = "192.168.127.91"
		svc.Spec.ClusterIPs = append(svc.Spec.ClusterIPs, svc.Spec.ClusterIP)
		for _, port := range importedService.Spec.Ports {
			svcPort := corev1.ServicePort{
				Name:        port.Name,
				Protocol:    port.Protocol,
				AppProtocol: port.AppProtocol,
				Port:        port.Port,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8091,
				},
			}
			svc.Spec.Ports = append(svc.Spec.Ports, svcPort)
		}
		defer func() {
			bytes, _ := json.Marshal(svc)
			fmt.Println("ListServices:", string(bytes))
		}()
		return svc
	}
	return nil
}

// ListServices returns a list of services that are imported from other clusters.
func (c *Client) ListServices() []*corev1.Service {
	var services []*corev1.Service

	defer func() {
		bytes, _ := json.Marshal(services)
		fmt.Println("ListServices:", string(bytes))
	}()

	for _, importedServiceIf := range c.informers.List(informers.InformerKeyServiceImport) {
		importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)

		svc := new(corev1.Service)
		svc.Namespace = importedService.Namespace
		svc.Name = importedService.Name
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = map[string]string{"app": importedService.Name}
		svc.Spec.ClusterIP = "192.168.127.91"
		svc.Spec.ClusterIPs = append(svc.Spec.ClusterIPs, svc.Spec.ClusterIP)
		for _, port := range importedService.Spec.Ports {
			svcPort := corev1.ServicePort{
				Name:        port.Name,
				Protocol:    port.Protocol,
				AppProtocol: port.AppProtocol,
				Port:        port.Port,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 8080,
				},
			}
			svc.Spec.Ports = append(svc.Spec.Ports, svcPort)
		}
		services = append(services, svc)
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
	fmt.Println("GetNamespace:", ns)
	for _, importedServiceIf := range c.informers.List(informers.InformerKeyServiceImport) {
		importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)

		if importedService.Namespace == ns {
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
	var pods []*corev1.Pod

	//for _, podInterface := range c.informers.List(osminformers.InformerKeyPod) {
	//	pod := podInterface.(*corev1.Pod)
	//	if !c.IsMonitoredNamespace(pod.Namespace) {
	//		continue
	//	}
	//	pods = append(pods, pod)
	//}
	return pods
}

// GetEndpoints returns the endpoint for a given service, otherwise returns nil if not found
// or error if the API errored out.
func (c *Client) GetEndpoints(svc service.MeshService) (*corev1.Endpoints, error) {
	importedServiceIf, exists, err := c.informers.GetByKey(informers.InformerKeyServiceImport, svc.NamespacedKey())
	if exists && err == nil {
		importedService := importedServiceIf.(*multiclusterv1alpha1.ServiceImport)

		endpoints := new(corev1.Endpoints)
		endpoints.Annotations = make(map[string]string)
		endpoints.Annotations[ServiceImportClusterKeyAnnotation] = "default/default/default/cluster1"
		endpoints.Annotations[ServiceImportContextPathAnnotation] = "/pipy"
		endpoints.Namespace = importedService.Namespace
		endpoints.Name = importedService.Name
		endpoints.Subsets = append(endpoints.Subsets, corev1.EndpointSubset{
			Addresses: []corev1.EndpointAddress{
				{
					IP:       "192.168.127.91",
					Hostname: "demo.flomesh.internal",
				},
			},
			Ports: []corev1.EndpointPort{
				{
					Name:     "pipy",
					Protocol: "tcp",
					//AppProtocol: "http",
					Port: 8080,
				},
			},
		})
		defer func() {
			bytes, _ := json.Marshal(endpoints)
			fmt.Println("GetEndpoints:", string(bytes))
		}()
		return endpoints, nil
	}
	return nil, nil
}

// ListServiceIdentitiesForService lists ServiceAccounts associated with the given service
func (c *Client) ListServiceIdentitiesForService(svc service.MeshService) ([]identity.K8sServiceAccount, error) {
	fmt.Println("ListServiceIdentitiesForService:", svc)
	var svcAccounts []identity.K8sServiceAccount

	//k8sSvc := c.GetService(svc)
	//if k8sSvc == nil {
	//	return nil, fmt.Errorf("Error fetching service %q: %s", svc, errServiceNotFound)
	//}
	//
	//svcAccountsSet := mapset.NewSet()
	//pods := c.ListPods()
	//for _, pod := range pods {
	//	svcRawSelector := k8sSvc.Spec.Selector
	//	selector := labels.Set(svcRawSelector).AsSelector()
	//	// service has no selectors, we do not need to match against the pod label
	//	if len(svcRawSelector) == 0 {
	//		continue
	//	}
	//	if selector.Matches(labels.Set(pod.Labels)) {
	//		podSvcAccount := identity.K8sServiceAccount{
	//			Name:      pod.Spec.ServiceAccountName,
	//			Namespace: pod.Namespace, // ServiceAccount must belong to the same namespace as the pod
	//		}
	//		svcAccountsSet.Add(podSvcAccount)
	//	}
	//}
	//
	//for svcAcc := range svcAccountsSet.Iter() {
	//	svcAccounts = append(svcAccounts, svcAcc.(identity.K8sServiceAccount))
	//}
	return svcAccounts, nil
}
