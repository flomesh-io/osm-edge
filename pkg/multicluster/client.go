package multicluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/k8s/informers"
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

	svcExportEventTypes := k8s.EventTypes{
		Add:    announcements.ServiceExportAdded,
		Update: announcements.ServiceExportUpdated,
		Delete: announcements.ServiceExportDeleted,
	}
	client.informers.AddEventHandler(informers.InformerKeyServiceExport, k8s.GetEventHandlerFuncs(shouldObserve, svcExportEventTypes, msgBroker))

	svcImportEventTypes := k8s.EventTypes{
		Add:    announcements.ServiceImportAdded,
		Update: announcements.ServiceImportUpdated,
		Delete: announcements.ServiceImportDeleted,
	}
	client.informers.AddEventHandler(informers.InformerKeyServiceImport, k8s.GetEventHandlerFuncs(shouldObserve, svcImportEventTypes, msgBroker))

	ingressClassEventTypes := k8s.EventTypes{
		Add:    announcements.IngressClassAdded,
		Update: announcements.IngressClassUpdated,
		Delete: announcements.IngressClassDeleted,
	}
	client.informers.AddEventHandler(informers.InformerKeyIngressClass, k8s.GetEventHandlerFuncs(shouldObserve, ingressClassEventTypes, msgBroker))

	return client
}