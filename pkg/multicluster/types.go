// Package multicluster implements the Kubernetes client for the resources in the flomesh.io API group
package multicluster

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	multiclusterv1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"

	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/k8s/informers"
	"github.com/openservicemesh/osm/pkg/service"
)

const (
	// ServiceImportClusterKeyAnnotation is the annotation used to configure context path for imported service
	ServiceImportClusterKeyAnnotation = "flomesh.io/ServiceImport/ClusterKey"

	// ServiceImportContextPathAnnotation is the annotation used to configure context path for imported service
	ServiceImportContextPathAnnotation = "flomesh.io/ServiceImport/ContextPath"

	// AnyServiceAccount defines wildcard service account
	AnyServiceAccount = "*"
)

// Client is the type used to represent the Kubernetes Client for the flomesh.io API group
type Client struct {
	informers      *informers.InformerCollection
	kubeClient     kubernetes.Interface
	kubeController k8s.Controller
}

// Controller is the interface for the functionality provided by the resources part of the flomesh.io API group
type Controller interface {
	// ListServices returns a list of all (monitored-namespace filtered) services in the mesh
	ListServices() []*corev1.Service

	// GetService returns a corev1 Service representation if the MeshService exists in cache, otherwise nil
	GetService(service.MeshService) *corev1.Service

	// ListPods returns a list of pods part of the mesh
	ListPods() []*corev1.Pod

	// ListServiceIdentitiesForService lists ServiceAccounts associated with the given service
	ListServiceIdentitiesForService(service.MeshService) ([]identity.K8sServiceAccount, error)

	// GetEndpoints returns the endpoints for a given service, if found
	GetEndpoints(service.MeshService) (*corev1.Endpoints, error)

	// GetIngressControllerServices returns ingress controller services.
	GetIngressControllerServices() []service.MeshService

	// GetExportedRule retrieves the export rule for the given MeshService
	GetExportedRule(svc service.MeshService) (*multiclusterv1alpha1.ServiceExportRule, error)

	//GetTargetPortForServicePort retrieves target for service
	GetTargetPortForServicePort(types.NamespacedName, uint16) (uint16, error)

	// GetGlobalTrafficPolicies retrieves global traffic policies
	GetGlobalTrafficPolicies(svc service.MeshService) (*multiclusterv1alpha1.GlobalTrafficPolicy, error)
}
