// Package multicluster implements the Kubernetes client for the resources in the flomesh.io API group
package multicluster

import (
	"k8s.io/client-go/kubernetes"

	multiclusterv1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/k8s/informers"
)

// Client is the type used to represent the Kubernetes Client for the flomesh.io API group
type Client struct {
	informers      *informers.InformerCollection
	kubeClient     kubernetes.Interface
	kubeController k8s.Controller
}

// Controller is the interface for the functionality provided by the resources part of the flomesh.io API group
type Controller interface {
	// ListImportedServices lists imported services
	ListImportedServices() []*multiclusterv1alpha1.ServiceImport
}
