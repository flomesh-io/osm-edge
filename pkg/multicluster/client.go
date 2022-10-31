package multicluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// ListImportedServices lists imported services
func (c *Client) ListImportedServices() []*multiclusterv1alpha1.ServiceImport {
	var importedServices []*multiclusterv1alpha1.ServiceImport
	for _, importedServiceIface := range c.informers.List(informers.InformerKeyServiceImport) {
		egressGateway := importedServiceIface.(*multiclusterv1alpha1.ServiceImport)
		importedServices = append(importedServices, egressGateway)
	}
	return importedServices
}
