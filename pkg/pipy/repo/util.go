package repo

import (
	"fmt"
	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/endpoint"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/pipy"
	"github.com/openservicemesh/osm/pkg/service"
	"github.com/openservicemesh/osm/pkg/trafficpolicy"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

func generatePipyInboundTrafficPolicy(meshCatalog catalog.MeshCataloger, _ identity.ServiceIdentity,
	pipyConf *PipyConf, inboundPolicy *trafficpolicy.InboundMeshTrafficPolicy) {

	itp := pipyConf.NewInboundTrafficPolicy()

	for _, trafficMatch := range inboundPolicy.TrafficMatches {
		upstreamSvc := trafficMatchToMeshSvc(trafficMatch)
		cluster := getMeshClusterConfigs(inboundPolicy.ClustersConfigs,
			service.ClusterName(upstreamSvc.SidecarLocalClusterName()))
		tm := itp.NewTrafficMatch(Port(cluster.Service.Port))
		tm.SetProtocol(Protocol(trafficMatch.DestinationProtocol))
		tm.SetPort(Port(trafficMatch.DestinationPort))

		if trafficMatch.DestinationProtocol == constants.ProtocolHTTP {
			upstreamSvcFQDN := upstreamSvc.FQDN()

			httpRouteConfig := getInboundHttpRouteConfigs(inboundPolicy.HTTPRouteConfigsPerPort,
				int(upstreamSvc.TargetPort),
				upstreamSvcFQDN)
			if httpRouteConfig == nil {
				continue
			}

			ruleName := HttpRouteRuleName(httpRouteConfig.Name)
			hsrrs := tm.NewHttpServiceRouteRules(ruleName)
			for _, hostname := range httpRouteConfig.Hostnames {
				tm.AddHttpHostPort2Service(HttpHostPort(hostname), ruleName)
			}
			if len(httpRouteConfig.Rules) > 0 {
				for _, rule := range httpRouteConfig.Rules {
					pathRegexp := URIPathRegexp(rule.Route.HTTPRouteMatch.Path)
					if len(pathRegexp) == 0 {
						continue
					}

					hsrr := hsrrs.NewHttpServiceRouteRule(pathRegexp)
					for k, v := range rule.Route.HTTPRouteMatch.Headers {
						hsrr.AddHeaderMatch(Header(k), HeaderRegexp(v))
					}

					if len(rule.Route.HTTPRouteMatch.Methods) == 0 {
						hsrr.AddMethodMatch("*")
					} else {
						for _, method := range rule.Route.HTTPRouteMatch.Methods {
							hsrr.AddMethodMatch(Method(method))
						}
					}

					for routeCluster := range rule.Route.WeightedClusters.Iter() {
						weightedCluster := routeCluster.(service.WeightedCluster)
						hsrr.AddWeightedCluster(ClusterName(weightedCluster.ClusterName),
							Weight(weightedCluster.Weight))
					}

					for allowedServiceIdentity := range rule.AllowedServiceIdentities.Iter() {
						serviceIdentity := allowedServiceIdentity.(identity.ServiceIdentity)
						hsrr.AddAllowedService(ServiceName(serviceIdentity))
						if pipyConf.isPermissiveTrafficPolicyMode() {
							continue
						}
						allowedServiceEndpoints := getEndpointsForProxyIdentity(meshCatalog, serviceIdentity)
						if len(allowedServiceEndpoints) > 0 {
							for _, allowedEndpoint := range allowedServiceEndpoints {
								tm.AddAllowedEndpoint(Address(allowedEndpoint.IP.String()), ServiceName(serviceIdentity))
							}
						}
					}
				}
			} else {
				pathRegexp := URIPathRegexp(".*")
				hsrr := hsrrs.NewHttpServiceRouteRule(pathRegexp)
				hsrr.AddMethodMatch("*")
				hsrr.AddWeightedCluster(ClusterName(cluster.Name), Weight(constants.ClusterWeightAcceptAll))
				hsrr.AddAllowedService("*")
			}
		} else if trafficMatch.DestinationProtocol == constants.ProtocolTCP ||
			trafficMatch.DestinationProtocol == constants.ProtocolTCPServerFirst {
			tm.AddWeightedCluster(ClusterName(cluster.Name), Weight(constants.ClusterWeightAcceptAll))
		}
	}

	for _, cluster := range inboundPolicy.ClustersConfigs {
		clusterConfigs := itp.NewClusterConfigs(ClusterName(cluster.Name))
		address := Address(cluster.Address)
		port := Port(cluster.Port)
		weight := Weight(constants.ClusterWeightAcceptAll)
		clusterConfigs.AddWeightedEndpoint(address, port, weight)
	}
}

func generatePipyOutboundTrafficRoutePolicy(_ catalog.MeshCataloger, proxyIdentity identity.ServiceIdentity,
	pipyConf *PipyConf,
	outboundPolicy *trafficpolicy.OutboundMeshTrafficPolicy) map[service.ClusterName]service.WeightedCluster {

	if len(outboundPolicy.TrafficMatches) == 0 {
		return nil
	}

	otp := pipyConf.NewOutboundTrafficPolicy()
	dependClusters := make(map[service.ClusterName]service.WeightedCluster)

	for _, trafficMatch := range outboundPolicy.TrafficMatches {
		tm := otp.NewTrafficMatch(Port(trafficMatch.DestinationPort))
		tm.SetProtocol(Protocol(trafficMatch.DestinationProtocol))
		tm.SetPort(Port(trafficMatch.DestinationPort))
		tm.SetServiceIdentity(proxyIdentity)

		for _, ipRange := range trafficMatch.DestinationIPRanges {
			tm.AddDestinationIPRange(DestinationIPRange(ipRange))
		}

		if trafficMatch.DestinationProtocol == constants.ProtocolHTTP {
			upstreamSvc := trafficMatchToMeshSvc(trafficMatch)
			upstreamSvcFQDN := upstreamSvc.FQDN()

			httpRouteConfig := getOutboundHttpRouteConfigs(outboundPolicy.HTTPRouteConfigsPerPort,
				int(upstreamSvc.TargetPort),
				upstreamSvcFQDN)
			if httpRouteConfig == nil {
				continue
			}

			ruleName := HttpRouteRuleName(httpRouteConfig.Name)
			hsrrs := tm.NewHttpServiceRouteRules(ruleName)
			for _, hostname := range httpRouteConfig.Hostnames {
				tm.AddHttpHostPort2Service(HttpHostPort(hostname), ruleName)
			}
			for _, route := range httpRouteConfig.Routes {
				pathRegexp := URIPathRegexp(route.HTTPRouteMatch.Path)
				if len(pathRegexp) == 0 {
					pathRegexp = ".*"
				}

				hsrr := hsrrs.NewHttpServiceRouteRule(pathRegexp)
				for k, v := range route.HTTPRouteMatch.Headers {
					hsrr.AddHeaderMatch(Header(k), HeaderRegexp(v))
				}

				if len(route.HTTPRouteMatch.Methods) == 0 {
					hsrr.AddMethodMatch("*")
				} else {
					for _, method := range route.HTTPRouteMatch.Methods {
						hsrr.AddMethodMatch(Method(method))
					}
				}

				for cluster := range route.WeightedClusters.Iter() {
					weightedCluster := cluster.(service.WeightedCluster)
					if _, exist := dependClusters[weightedCluster.ClusterName]; !exist {
						dependClusters[weightedCluster.ClusterName] = weightedCluster
					}
					hsrr.AddWeightedCluster(ClusterName(weightedCluster.ClusterName), Weight(weightedCluster.Weight))
				}
			}
		} else if trafficMatch.DestinationProtocol == constants.ProtocolTCP ||
			trafficMatch.DestinationProtocol == constants.ProtocolTCPServerFirst {
			for _, cluster := range trafficMatch.WeightedClusters {
				if _, exist := dependClusters[cluster.ClusterName]; !exist {
					dependClusters[cluster.ClusterName] = cluster
				}
				tm.AddWeightedCluster(ClusterName(cluster.ClusterName), Weight(cluster.Weight))
			}
		}
	}

	return dependClusters
}

func generatePipyEgressTrafficRoutePolicy(_ catalog.MeshCataloger, _ identity.ServiceIdentity,
	pipyConf *PipyConf,
	egressPolicy *trafficpolicy.EgressTrafficPolicy) map[service.ClusterName]*service.WeightedCluster {

	if len(egressPolicy.TrafficMatches) == 0 {
		return nil
	}

	otp := pipyConf.NewOutboundTrafficPolicy()
	dependClusters := make(map[service.ClusterName]*service.WeightedCluster)

	for _, trafficMatch := range egressPolicy.TrafficMatches {
		tm := otp.NewTrafficMatch(Port(trafficMatch.DestinationPort))
		tm.SetProtocol(Protocol(trafficMatch.DestinationProtocol))
		tm.SetPort(Port(trafficMatch.DestinationPort))

		for _, ipRange := range trafficMatch.DestinationIPRanges {
			tm.AddDestinationIPRange(DestinationIPRange(ipRange))
		}

		if trafficMatch.DestinationProtocol == constants.ProtocolHTTP {

			httpRouteConfigs := getEgressHttpRouteConfigs(egressPolicy.HTTPRouteConfigsPerPort,
				trafficMatch.DestinationPort, trafficMatch.Name)
			if len(httpRouteConfigs) == 0 {
				continue
			}

			for _, httpRouteConfig := range httpRouteConfigs {
				ruleName := HttpRouteRuleName(httpRouteConfig.Name)
				hsrrs := tm.NewHttpServiceRouteRules(ruleName)
				for _, hostname := range httpRouteConfig.Hostnames {
					tm.AddHttpHostPort2Service(HttpHostPort(hostname), ruleName)
				}
				for _, rule := range httpRouteConfig.RoutingRules {
					route := rule.Route
					pathRegexp := URIPathRegexp(route.HTTPRouteMatch.Path)
					if len(pathRegexp) == 0 {
						pathRegexp = ".*"
					}

					hsrr := hsrrs.NewHttpServiceRouteRule(pathRegexp)
					for k, v := range route.HTTPRouteMatch.Headers {
						hsrr.AddHeaderMatch(Header(k), HeaderRegexp(v))
					}

					if len(route.HTTPRouteMatch.Methods) == 0 {
						hsrr.AddMethodMatch("*")
					} else {
						for _, method := range route.HTTPRouteMatch.Methods {
							hsrr.AddMethodMatch(Method(method))
						}
					}

					for cluster := range route.WeightedClusters.Iter() {
						weightedCluster := cluster.(service.WeightedCluster)
						if _, exist := dependClusters[weightedCluster.ClusterName]; !exist {
							dependClusters[weightedCluster.ClusterName] = &weightedCluster
						}
						hsrr.AddWeightedCluster(ClusterName(weightedCluster.ClusterName), Weight(weightedCluster.Weight))
					}

					for _, allowedIPRange := range rule.AllowedDestinationIPRanges {
						tm.AddDestinationIPRange(DestinationIPRange(allowedIPRange))
					}
				}
			}
		} else if trafficMatch.DestinationProtocol == constants.ProtocolHTTPS {
			cluster := service.WeightedCluster{
				ClusterName: service.ClusterName(trafficMatch.Cluster),
				Weight:      constants.ClusterWeightAcceptAll,
			}
			tm.AddWeightedCluster(ClusterName(cluster.ClusterName), Weight(cluster.Weight))
			clusterConfigs := otp.NewClusterConfigs(ClusterName(cluster.ClusterName.String()))
			for _, serverName := range trafficMatch.ServerNames {
				address := Address(serverName)
				port := Port(trafficMatch.DestinationPort)
				weight := Weight(constants.ClusterWeightAcceptAll)
				clusterConfigs.AddWeightedEndpoint(address, port, weight)
			}
		} else if trafficMatch.DestinationProtocol == constants.ProtocolTCP ||
			trafficMatch.DestinationProtocol == constants.ProtocolTCPServerFirst {
			tm.SetAllowedEgressTraffic(true)
		}
	}

	return dependClusters
}

func generatePipyOutboundTrafficBalancePolicy(meshCatalog catalog.MeshCataloger, _ *pipy.Proxy,
	proxyIdentity identity.ServiceIdentity,
	pipyConf *PipyConf, outboundPolicy *trafficpolicy.OutboundMeshTrafficPolicy,
	dependClusters map[service.ClusterName]service.WeightedCluster) bool {
	otp := pipyConf.NewOutboundTrafficPolicy()
	for _, cluster := range dependClusters {
		clusterConfig := getMeshClusterConfigs(outboundPolicy.ClustersConfigs, cluster.ClusterName)
		if clusterConfig == nil {
			return false
		}
		upstreamEndpoints := getUpstreamEndpoints(meshCatalog, proxyIdentity, cluster.ClusterName)
		if len(upstreamEndpoints) == 0 {
			return false
		}
		clusterConfigs := otp.NewClusterConfigs(ClusterName(cluster.ClusterName.String()))
		for _, upstreamEndpoint := range upstreamEndpoints {
			address := Address(fmt.Sprintf("%s", upstreamEndpoint.IP))
			port := Port(clusterConfig.Service.Port)
			weight := Weight(upstreamEndpoint.Weight)
			clusterConfigs.AddWeightedEndpoint(address, port, weight)
		}
	}

	return true
}

func generatePipyIngressTrafficRoutePolicy(_ catalog.MeshCataloger, _ identity.ServiceIdentity,
	pipyConf *PipyConf,
	ingressPolicy *trafficpolicy.IngressTrafficPolicy) {

	if len(ingressPolicy.TrafficMatches) == 0 {
		return
	}

	if pipyConf.Inbound == nil {
		return
	}

	if len(pipyConf.Inbound.ClustersConfigs) == 0 {
		return
	}

	itp := pipyConf.NewInboundTrafficPolicy()

	for _, trafficMatch := range ingressPolicy.TrafficMatches {
		tm := itp.GetTrafficMatch(Port(trafficMatch.Port))
		if tm == nil {
			continue
		}

		if trafficMatch.Protocol != constants.ProtocolHTTP {
			continue
		}
		for _, ipRange := range trafficMatch.SourceIPRanges {
			tm.AddSourceIPRange(SourceIPRange(ipRange))
		}
		for _, httpRouteConfig := range ingressPolicy.HTTPRoutePolicies {
			if len(httpRouteConfig.Rules) == 0 {
				continue
			}
			for _, hostname := range httpRouteConfig.Hostnames {
				ruleName := HttpRouteRuleName(hostname)
				tm.AddHttpHostPort2Service(HttpHostPort(hostname), ruleName)

				hsrrs := tm.NewHttpServiceRouteRules(ruleName)
				for _, rule := range httpRouteConfig.Rules {
					pathRegexp := URIPathRegexp(rule.Route.HTTPRouteMatch.Path)
					if len(pathRegexp) == 0 {
						continue
					}

					hsrr := hsrrs.NewHttpServiceRouteRule(pathRegexp)
					for k, v := range rule.Route.HTTPRouteMatch.Headers {
						hsrr.AddHeaderMatch(Header(k), HeaderRegexp(v))
					}

					if len(rule.Route.HTTPRouteMatch.Methods) == 0 {
						hsrr.AddMethodMatch("*")
					} else {
						for _, method := range rule.Route.HTTPRouteMatch.Methods {
							hsrr.AddMethodMatch(Method(method))
						}
					}

					for routeCluster := range rule.Route.WeightedClusters.Iter() {
						weightedCluster := routeCluster.(service.WeightedCluster)
						hsrr.AddWeightedCluster(ClusterName(weightedCluster.ClusterName),
							Weight(weightedCluster.Weight))
					}

					for allowedServiceIdentitiy := range rule.AllowedServiceIdentities.Iter() {
						serviceIdentity := allowedServiceIdentitiy.(identity.ServiceIdentity)
						hsrr.AddAllowedService(ServiceName(serviceIdentity))
					}
				}
			}
		}
	}
}

func generatePipyEgressTrafficBalancePolicy(_ catalog.MeshCataloger, _ *pipy.Proxy,
	_ identity.ServiceIdentity,
	pipyConf *PipyConf, egressPolicy *trafficpolicy.EgressTrafficPolicy,
	dependClusters map[service.ClusterName]*service.WeightedCluster) bool {

	otp := pipyConf.NewOutboundTrafficPolicy()
	for _, cluster := range dependClusters {
		clusterConfig := getEgressClusterConfigs(egressPolicy.ClustersConfigs, cluster.ClusterName)
		if clusterConfig == nil {
			return false
		}
		clusterConfigs := otp.NewClusterConfigs(ClusterName(cluster.ClusterName.String()))
		address := Address(clusterConfig.Name)
		port := Port(clusterConfig.Port)
		weight := Weight(constants.ClusterWeightAcceptAll)
		clusterConfigs.AddWeightedEndpoint(address, port, weight)
	}

	return true
}

func getInboundHttpRouteConfigs(httpRouteConfigsPerPort map[int][]*trafficpolicy.InboundTrafficPolicy,
	targetPort int, upstreamSvcFQDN string) *trafficpolicy.InboundTrafficPolicy {
	if httpRouteConfigs, ok := httpRouteConfigsPerPort[targetPort]; ok {
		for _, httpRouteConfig := range httpRouteConfigs {
			if httpRouteConfig.Name == upstreamSvcFQDN {
				return httpRouteConfig
			}
		}
	}
	return nil
}

func getOutboundHttpRouteConfigs(httpRouteConfigsPerPort map[int][]*trafficpolicy.OutboundTrafficPolicy,
	targetPort int, upstreamSvcFQDN string) *trafficpolicy.OutboundTrafficPolicy {
	if httpRouteConfigs, ok := httpRouteConfigsPerPort[targetPort]; ok {
		for _, httpRouteConfig := range httpRouteConfigs {
			if httpRouteConfig.Name == upstreamSvcFQDN {
				return httpRouteConfig
			}
		}
	}
	return nil
}

func getEgressHttpRouteConfigs(httpRouteConfigsPerPort map[int][]*trafficpolicy.EgressHTTPRouteConfig,
	targetPort int, egressSvcName string) []*trafficpolicy.EgressHTTPRouteConfig {
	if httpRouteConfigs, ok := httpRouteConfigsPerPort[targetPort]; ok {
		if len(egressSvcName) == 0 {
			return httpRouteConfigs
		}
		routeConfigs := make([]*trafficpolicy.EgressHTTPRouteConfig, 0)
		for _, httpRouteConfig := range httpRouteConfigs {
			if httpRouteConfig.Name == egressSvcName {
				routeConfigs = append(routeConfigs, httpRouteConfig)
				return routeConfigs
			} else if len(httpRouteConfig.Hostnames) > 0 {
				for _, hostname := range httpRouteConfig.Hostnames {
					if hostname == egressSvcName {
						routeConfigs = append(routeConfigs, httpRouteConfig)
						return routeConfigs
					}
				}
			}
		}
	}
	return nil
}

func trafficMatchToMeshSvc(trafficMatch *trafficpolicy.TrafficMatch) *service.MeshService {
	splitFunc := func(r rune) bool {
		return r == '_'
	}

	chunks := strings.FieldsFunc(trafficMatch.Name, splitFunc)
	if len(chunks) != 3 {
		log.Error().Msgf("Invalid traffic match name. Expected: <namespace>/<name>_<port>_<protocol>, got: %s",
			trafficMatch.Name)
		return nil
	}

	namespacedName, err := k8s.NamespacedNameFrom(chunks[0])
	if err != nil {
		log.Error().Err(err).Msgf("Error retrieving NamespacedName from TrafficMatch")
		return nil
	}
	return &service.MeshService{
		Namespace:  namespacedName.Namespace,
		Name:       namespacedName.Name,
		Protocol:   trafficMatch.DestinationProtocol,
		TargetPort: uint16(trafficMatch.DestinationPort),
	}
}

func getMeshClusterConfigs(clustersConfigs []*trafficpolicy.MeshClusterConfig,
	clusterName service.ClusterName) *trafficpolicy.MeshClusterConfig {
	if len(clustersConfigs) == 0 {
		return nil
	}

	for _, clustersConfig := range clustersConfigs {
		if clusterName.String() == clustersConfig.Name {
			return clustersConfig
		}
	}

	return nil
}

func getEgressClusterConfigs(clustersConfigs []*trafficpolicy.EgressClusterConfig,
	clusterName service.ClusterName) *trafficpolicy.EgressClusterConfig {
	if len(clustersConfigs) == 0 {
		return nil
	}

	for _, clustersConfig := range clustersConfigs {
		if clusterName.String() == clustersConfig.Name {
			return clustersConfig
		}
	}

	return nil
}

func getUpstreamEndpoints(meshCatalog catalog.MeshCataloger, proxyIdentity identity.ServiceIdentity,
	clusterName service.ClusterName) []endpoint.Endpoint {
	if dstSvc, err := clusterToMeshSvc(clusterName.String()); err == nil {
		return meshCatalog.ListAllowedUpstreamEndpointsForService(proxyIdentity, dstSvc)
	}
	return nil
}

// clusterToMeshSvc returns the MeshService associated with the given cluster name
func clusterToMeshSvc(cluster string) (service.MeshService, error) {
	splitFunc := func(r rune) bool {
		return r == '/' || r == '|'
	}

	chunks := strings.FieldsFunc(cluster, splitFunc)
	if len(chunks) != 3 {
		return service.MeshService{},
			errors.Errorf("Invalid cluster name. Expected: <namespace>/<name>|<port>, got: %s", cluster)
	}

	port, err := strconv.ParseUint(chunks[2], 10, 16)
	if err != nil {
		return service.MeshService{}, errors.Errorf("Invalid cluster port %s, expected int value: %s", chunks[2], err)
	}

	return service.MeshService{
		Namespace: chunks[0],
		Name:      chunks[1],
		// The port always maps to MeshServer.TargetPort and not MeshService.Port because
		// endpoints of a service are derived from it's TargetPort and not Port.
		TargetPort: uint16(port),
	}, nil
}

func getEndpointsForProxyIdentity(meshCatalog catalog.MeshCataloger, proxyIdentity identity.ServiceIdentity) []endpoint.Endpoint {
	if mc, ok := meshCatalog.(*catalog.MeshCatalog); ok {
		return mc.ListEndpointsForServiceIdentity(proxyIdentity)
	}
	return nil
}
