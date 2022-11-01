// Package fsm implements MulticlusterClient's methods.
package fsm

import (
	"fmt"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/multicluster"
	"github.com/rs/zerolog/log"
	"k8s.io/utils/pointer"
	"net"
	"strings"

	mapset "github.com/deckarep/golang-set"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/endpoint"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/service"
)

// Ensure interface compliance
var _ endpoint.Provider = (*client)(nil)
var _ service.Provider = (*client)(nil)

// NewClient returns a client that has all components necessary to connect to and maintain state of a multi cluster.
func NewClient(multiclusterController multicluster.Controller, cfg configurator.Configurator) *client { //nolint: revive // unexported-return
	return &client{
		multiclusterController: multiclusterController,
		meshConfigurator:       cfg,
	}
}

// GetID returns a string descriptor / identifier of the compute provider.
// Required by interfaces: EndpointsProvider, ServiceProvider
func (c *client) GetID() string {
	return providerName
}

// ListEndpointsForService retrieves the list of IP addresses for the given service
func (c *client) ListEndpointsForService(svc service.MeshService) []endpoint.Endpoint {
	log.Trace().Msgf("Getting Endpoints for MeshService %s on MultiClusters", svc)

	kubernetesEndpoints, err := c.multiclusterController.GetEndpoints(svc)
	if err != nil || kubernetesEndpoints == nil {
		log.Info().Msgf("No k8s endpoints found for MeshService %s", svc)
		return nil
	}

	var endpoints []endpoint.Endpoint
	for _, kubernetesEndpoint := range kubernetesEndpoints.Subsets {
		for _, port := range kubernetesEndpoint.Ports {
			// If a TargetPort is specified for the service, filter the endpoint by this port.
			// This is required to ensure we do not attempt to filter the endpoints when the endpoints
			// are being listed for a MeshService whose TargetPort is not known.
			if svc.TargetPort != 0 && port.Port != int32(svc.TargetPort) {
				// k8s service's port does not match MeshService port, ignore this port
				continue
			}
			for _, address := range kubernetesEndpoint.Addresses {
				if svc.Subdomain() != "" && svc.Subdomain() != address.Hostname {
					// if there's a subdomain on this meshservice, make sure it matches the endpoint's hostname
					continue
				}
				ip := net.ParseIP(address.IP)
				if ip == nil {
					log.Error().Msgf("Error parsing endpoint IP address %s for MeshService %s", address.IP, svc)
					continue
				}
				ept := endpoint.Endpoint{
					IP:      ip,
					Port:    endpoint.Port(port.Port),
					Cluster: kubernetesEndpoints.Annotations[multicluster.ServiceImportClusterKeyAnnotation],
					Path:    kubernetesEndpoints.Annotations[multicluster.ServiceImportContextPathAnnotation],
				}
				endpoints = append(endpoints, ept)
			}
		}
	}

	log.Trace().Msgf("Endpoints for MeshService %s: %v", svc, endpoints)

	return endpoints
}

// ListEndpointsForIdentity retrieves the list of IP addresses for the given service account
// Note: ServiceIdentity must be in the format "name.namespace" [https://github.com/openservicemesh/osm/issues/3188]
func (c *client) ListEndpointsForIdentity(serviceIdentity identity.ServiceIdentity) []endpoint.Endpoint {
	sa := serviceIdentity.ToK8sServiceAccount()
	log.Trace().Msgf("[%s] (ListEndpointsForIdentity) Getting Endpoints for service account %s on Kubernetes", c.GetID(), sa)

	var endpoints []endpoint.Endpoint
	for _, pod := range c.multiclusterController.ListPods() {
		if pod.Namespace != sa.Namespace {
			continue
		}
		if pod.Spec.ServiceAccountName != sa.Name {
			continue
		}

		for _, podIP := range pod.Status.PodIPs {
			ip := net.ParseIP(podIP.IP)
			if ip == nil {
				log.Error().Msgf("[%s] Error parsing IP address %s", c.GetID(), podIP.IP)
				break
			}
			ept := endpoint.Endpoint{IP: ip}
			endpoints = append(endpoints, ept)
		}
	}

	log.Trace().Msgf("[%s][ListEndpointsForIdentity] Endpoints for service identity (serviceAccount=%s) %s: %+v", c.GetID(), serviceIdentity, sa, endpoints)

	return endpoints
}

// GetServicesForServiceIdentity retrieves a list of services for the given service identity.
func (c *client) GetServicesForServiceIdentity(svcIdentity identity.ServiceIdentity) []service.MeshService {
	var meshServices []service.MeshService
	svcSet := mapset.NewSet() // mapset is used to avoid duplicate elements in the output list

	svcAccount := svcIdentity.ToK8sServiceAccount()

	for _, pod := range c.multiclusterController.ListPods() {
		if pod.Namespace != svcAccount.Namespace {
			continue
		}

		if pod.Spec.ServiceAccountName != svcAccount.Name {
			continue
		}

		podLabels := pod.ObjectMeta.Labels
		meshServicesForPod := c.getServicesByLabels(podLabels, pod.Namespace)
		for _, svc := range meshServicesForPod {
			if added := svcSet.Add(svc); added {
				meshServices = append(meshServices, svc)
			}
		}
	}

	log.Trace().Msgf("[%s] Services for service account %s: %v", c.GetID(), svcAccount, meshServices)
	return meshServices
}

// getServicesByLabels gets Kubernetes services whose selectors match the given labels
func (c *client) getServicesByLabels(podLabels map[string]string, targetNamespace string) []service.MeshService {
	var finalList []service.MeshService
	serviceList := c.multiclusterController.ListServices()

	for _, svc := range serviceList {
		// TODO: #1684 Introduce APIs to dynamically allow applying selectors, instead of callers implementing
		// filtering themselves
		if svc.Namespace != targetNamespace {
			continue
		}

		svcRawSelector := svc.Spec.Selector
		// service has no selectors, we do not need to match against the pod label
		if len(svcRawSelector) == 0 {
			continue
		}
		selector := labels.Set(svcRawSelector).AsSelector()
		if selector.Matches(labels.Set(podLabels)) {
			finalList = append(finalList, ServiceToMeshServices(c.multiclusterController, *svc)...)
		}
	}

	return finalList
}

// GetResolvableEndpointsForService returns the expected endpoints that are to be reached when the service
// FQDN is resolved
func (c *client) GetResolvableEndpointsForService(svc service.MeshService) []endpoint.Endpoint {
	var endpoints []endpoint.Endpoint

	// Check if the service has been given Cluster IP
	kubeService := c.multiclusterController.GetService(svc)
	if kubeService == nil {
		log.Info().Msgf("No k8s services found for MeshService %s", svc)
		return nil
	}

	if len(kubeService.Spec.ClusterIP) == 0 || kubeService.Spec.ClusterIP == corev1.ClusterIPNone {
		// If service has no cluster IP or cluster IP is <none>, use final endpoint as resolvable destinations
		return c.ListEndpointsForService(svc)
	}

	// Cluster IP is present
	ip := net.ParseIP(kubeService.Spec.ClusterIP)
	if ip == nil {
		log.Error().Msgf("[%s] Could not parse Cluster IP %s", c.GetID(), kubeService.Spec.ClusterIP)
		return nil
	}

	for _, svcPort := range kubeService.Spec.Ports {
		endpoints = append(endpoints, endpoint.Endpoint{
			IP:   ip,
			Port: endpoint.Port(svcPort.Port),
		})
	}

	return endpoints
}

// ListServices returns a list of services that are part of monitored namespaces
func (c *client) ListServices() []service.MeshService {
	var services []service.MeshService
	for _, svc := range c.multiclusterController.ListServices() {
		services = append(services, ServiceToMeshServices(c.multiclusterController, *svc)...)
	}
	return services
}

// ListServiceIdentitiesForService lists the service identities associated with the given mesh service.
func (c *client) ListServiceIdentitiesForService(svc service.MeshService) []identity.ServiceIdentity {
	serviceAccounts, err := c.multiclusterController.ListServiceIdentitiesForService(svc)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting ServiceAccounts for Service %s", svc)
		return nil
	}

	var serviceIdentities []identity.ServiceIdentity
	for _, svcAccount := range serviceAccounts {
		serviceIdentity := svcAccount.ToServiceIdentity()
		serviceIdentities = append(serviceIdentities, serviceIdentity)
	}

	return serviceIdentities
}

// ServiceToMeshServices translates a k8s service with one or more ports to one or more
// MeshService objects per port.
func ServiceToMeshServices(c k8s.CoreController, svc corev1.Service) []service.MeshService {
	var meshServices []service.MeshService

	for _, portSpec := range svc.Spec.Ports {
		meshSvc := service.MeshService{
			Namespace: svc.Namespace,
			Name:      svc.Name,
			Port:      uint16(portSpec.Port),
		}

		// attempt to parse protocol from port name
		// Order of Preference is:
		// 1. port.appProtocol field
		// 2. protocol prefixed to port name (e.g. tcp-my-port)
		// 3. default to http
		protocol := constants.ProtocolHTTP
		for _, p := range constants.SupportedProtocolsInMesh {
			if strings.HasPrefix(portSpec.Name, p+"-") {
				protocol = p
				break
			}
		}

		// use port.appProtocol if specified, else use port protocol
		meshSvc.Protocol = pointer.StringDeref(portSpec.AppProtocol, protocol)

		// The endpoints for the kubernetes service carry information that allows
		// us to retrieve the TargetPort for the MeshService.
		endpoints, _ := c.GetEndpoints(meshSvc)
		if endpoints != nil {
			meshSvc.TargetPort = GetTargetPortFromEndpoints(portSpec.Name, *endpoints)
		} else {
			log.Warn().Msgf("k8s service %s/%s does not have endpoints but is being represented as a MeshService", svc.Namespace, svc.Name)
		}

		if !k8s.IsHeadlessService(svc) || endpoints == nil {
			meshServices = append(meshServices, meshSvc)
			continue
		}

		for _, subset := range endpoints.Subsets {
			for _, address := range subset.Addresses {
				if address.Hostname == "" {
					continue
				}
				meshServices = append(meshServices, service.MeshService{
					Namespace:  svc.Namespace,
					Name:       fmt.Sprintf("%s.%s", address.Hostname, svc.Name),
					Port:       meshSvc.Port,
					TargetPort: meshSvc.TargetPort,
					Protocol:   meshSvc.Protocol,
				})
			}
		}
	}

	return meshServices
}

// GetTargetPortFromEndpoints returns the endpoint port corresponding to the given endpoint name and endpoints
func GetTargetPortFromEndpoints(endpointName string, endpoints corev1.Endpoints) (endpointPort uint16) {
	// Per https://pkg.go.dev/k8s.io/api/core/v1#ServicePort and
	// https://pkg.go.dev/k8s.io/api/core/v1#EndpointPort, if a service has multiple
	// ports, then ServicePort.Name must match EndpointPort.Name when considering
	// matching endpoints for the service's port. ServicePort.Name and EndpointPort.Name
	// can be unset when the service has a single port exposed, in which case we are
	// guaranteed to have the same port specified in the list of EndpointPort.Subsets.
	//
	// The logic below works as follows:
	// If the service has multiple ports, retrieve the matching endpoint port using
	// the given ServicePort.Name specified by `endpointName`.
	// Otherwise, simply return the only port referenced in EndpointPort.Subsets.
	for _, subset := range endpoints.Subsets {
		for _, port := range subset.Ports {
			if endpointName == "" || len(subset.Ports) == 1 {
				// ServicePort.Name is not passed or a single port exists on the service.
				// Both imply that this service has a single ServicePort and EndpointPort.
				endpointPort = uint16(port.Port)
				return
			}

			// If more than 1 port is specified
			if port.Name == endpointName {
				endpointPort = uint16(port.Port)
				return
			}
		}
	}
	return
}
