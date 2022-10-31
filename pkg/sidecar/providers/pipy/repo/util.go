package repo

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/catalog"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/endpoint"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/service"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy"
	"github.com/openservicemesh/osm/pkg/trafficpolicy"
	"github.com/openservicemesh/osm/pkg/utils"
)

func generatePipyInboundTrafficPolicy(meshCatalog catalog.MeshCataloger, _ identity.ServiceIdentity, pipyConf *PipyConf, inboundPolicy *trafficpolicy.InboundMeshTrafficPolicy, trustDomain string) {
	itp := pipyConf.newInboundTrafficPolicy()

	for _, trafficMatch := range inboundPolicy.TrafficMatches {
		destinationProtocol := strings.ToLower(trafficMatch.DestinationProtocol)
		upstreamSvc := trafficMatchToMeshSvc(trafficMatch)
		cluster := getMeshClusterConfigs(inboundPolicy.ClustersConfigs,
			service.ClusterName(upstreamSvc.SidecarLocalClusterName()))
		tm := itp.newTrafficMatch(Port(cluster.Service.Port))
		tm.setProtocol(Protocol(destinationProtocol))
		tm.setPort(Port(trafficMatch.DestinationPort))
		tm.setTCPServiceRateLimit(trafficMatch.RateLimit)

		if destinationProtocol == constants.ProtocolHTTP ||
			trafficMatch.DestinationProtocol == constants.ProtocolGRPC {
			upstreamSvcFQDN := upstreamSvc.FQDN()

			httpRouteConfig := getInboundHTTPRouteConfigs(inboundPolicy.HTTPRouteConfigsPerPort,
				int(upstreamSvc.TargetPort), upstreamSvcFQDN)
			if httpRouteConfig == nil {
				continue
			}

			ruleName := HTTPRouteRuleName(httpRouteConfig.Name)
			hsrrs := tm.newHTTPServiceRouteRules(ruleName)
			hsrrs.setHTTPServiceRateLimit(trafficMatch.RateLimit)
			hsrrs.setHTTPHeadersRateLimit(trafficMatch.HeaderRateLimit)
			for _, hostname := range httpRouteConfig.Hostnames {
				tm.addHTTPHostPort2Service(HTTPHostPort(hostname), ruleName)
			}
			if len(httpRouteConfig.Rules) > 0 {
				for _, rule := range httpRouteConfig.Rules {
					pathRegexp := URIPathRegexp(rule.Route.HTTPRouteMatch.Path)
					if len(pathRegexp) == 0 {
						continue
					}

					hsrr := hsrrs.newHTTPServiceRouteRule(pathRegexp)
					for k, v := range rule.Route.HTTPRouteMatch.Headers {
						hsrr.addHeaderMatch(Header(k), HeaderRegexp(v))
					}

					if len(rule.Route.HTTPRouteMatch.Methods) == 0 {
						hsrr.addMethodMatch("*")
					} else {
						for _, method := range rule.Route.HTTPRouteMatch.Methods {
							hsrr.addMethodMatch(Method(method))
						}
					}

					for routeCluster := range rule.Route.WeightedClusters.Iter() {
						weightedCluster := routeCluster.(service.WeightedCluster)
						hsrr.addWeightedCluster(ClusterName(weightedCluster.ClusterName),
							Weight(weightedCluster.Weight))
					}

					hsrr.setRateLimit(rule.Route.RateLimit)

					for allowedPrincipal := range rule.AllowedPrincipals.Iter() {
						servicePrincipal := allowedPrincipal.(string)
						serviceIdentity := identity.FromPrincipal(servicePrincipal, trustDomain)
						hsrr.addAllowedService(ServiceName(serviceIdentity))
						if identity.WildcardPrincipal == servicePrincipal || pipyConf.isPermissiveTrafficPolicyMode() {
							continue
						}
						allowedServiceEndpoints := getEndpointsForProxyIdentity(meshCatalog, serviceIdentity)
						if len(allowedServiceEndpoints) > 0 {
							for _, allowedEndpoint := range allowedServiceEndpoints {
								tm.addAllowedEndpoint(Address(allowedEndpoint.IP.String()), ServiceName(serviceIdentity))
							}
						}
					}
				}
			} else {
				pathRegexp := URIPathRegexp(".*")
				hsrr := hsrrs.newHTTPServiceRouteRule(pathRegexp)
				hsrr.addMethodMatch("*")
				hsrr.addWeightedCluster(ClusterName(cluster.Name), Weight(constants.ClusterWeightAcceptAll))
				hsrr.addAllowedService("*")
			}
		} else if destinationProtocol == constants.ProtocolTCP ||
			destinationProtocol == constants.ProtocolTCPServerFirst {
			tm.addWeightedCluster(ClusterName(cluster.Name), Weight(constants.ClusterWeightAcceptAll))
		}
	}

	for _, cluster := range inboundPolicy.ClustersConfigs {
		clusterConfigs := itp.newClusterConfigs(ClusterName(cluster.Name))
		address := Address(cluster.Address)
		port := Port(cluster.Port)
		weight := Weight(constants.ClusterWeightAcceptAll)
		clusterConfigs.addWeightedEndpoint(address, port, weight)
	}
}

func generatePipyOutboundTrafficRoutePolicy(_ catalog.MeshCataloger, proxyIdentity identity.ServiceIdentity, pipyConf *PipyConf, outboundPolicy *trafficpolicy.OutboundMeshTrafficPolicy) map[service.ClusterName]*WeightedCluster {
	if len(outboundPolicy.TrafficMatches) == 0 {
		return nil
	}

	otp := pipyConf.newOutboundTrafficPolicy()
	dependClusters := make(map[service.ClusterName]*WeightedCluster)

	for _, trafficMatch := range outboundPolicy.TrafficMatches {
		destinationProtocol := strings.ToLower(trafficMatch.DestinationProtocol)
		tm := otp.newTrafficMatch(Port(trafficMatch.DestinationPort))
		tm.setProtocol(Protocol(destinationProtocol))
		tm.setPort(Port(trafficMatch.DestinationPort))
		tm.setServiceIdentity(proxyIdentity)

		for _, ipRange := range trafficMatch.DestinationIPRanges {
			tm.addDestinationIPRange(DestinationIPRange(ipRange), nil)
		}

		if destinationProtocol == constants.ProtocolHTTP ||
			destinationProtocol == constants.ProtocolGRPC {
			upstreamSvc := trafficMatchToMeshSvc(trafficMatch)
			upstreamSvcFQDN := upstreamSvc.FQDN()

			httpRouteConfig := getOutboundHTTPRouteConfigs(outboundPolicy.HTTPRouteConfigsPerPort,
				int(upstreamSvc.TargetPort),
				upstreamSvcFQDN)
			if httpRouteConfig == nil {
				continue
			}

			ruleName := HTTPRouteRuleName(httpRouteConfig.Name)
			hsrrs := tm.newHTTPServiceRouteRules(ruleName)
			for _, hostname := range httpRouteConfig.Hostnames {
				tm.addHTTPHostPort2Service(HTTPHostPort(hostname), ruleName)
			}
			for _, route := range httpRouteConfig.Routes {
				pathRegexp := URIPathRegexp(route.HTTPRouteMatch.Path)
				if len(pathRegexp) == 0 {
					pathRegexp = ".*"
				}

				hsrr := hsrrs.newHTTPServiceRouteRule(pathRegexp)
				for k, v := range route.HTTPRouteMatch.Headers {
					hsrr.addHeaderMatch(Header(k), HeaderRegexp(v))
				}

				if len(route.HTTPRouteMatch.Methods) == 0 {
					hsrr.addMethodMatch("*")
				} else {
					for _, method := range route.HTTPRouteMatch.Methods {
						hsrr.addMethodMatch(Method(method))
					}
				}

				for cluster := range route.WeightedClusters.Iter() {
					serviceCluster := cluster.(service.WeightedCluster)
					weightedCluster := new(WeightedCluster)
					weightedCluster.WeightedCluster = serviceCluster
					weightedCluster.RetryPolicy = route.RetryPolicy
					if _, exist := dependClusters[weightedCluster.ClusterName]; !exist {
						dependClusters[weightedCluster.ClusterName] = weightedCluster
					}
					hsrr.addWeightedCluster(ClusterName(weightedCluster.ClusterName), Weight(weightedCluster.Weight))
				}
			}
		} else if destinationProtocol == constants.ProtocolTCP ||
			destinationProtocol == constants.ProtocolTCPServerFirst {
			for _, serviceCluster := range trafficMatch.WeightedClusters {
				weightedCluster := new(WeightedCluster)
				weightedCluster.WeightedCluster = serviceCluster
				if _, exist := dependClusters[weightedCluster.ClusterName]; !exist {
					dependClusters[weightedCluster.ClusterName] = weightedCluster
				}
				tm.addWeightedCluster(ClusterName(weightedCluster.ClusterName), Weight(weightedCluster.Weight))
			}
		}
	}

	return dependClusters
}

func generatePipyEgressTrafficRoutePolicy(meshCatalog catalog.MeshCataloger, _ identity.ServiceIdentity, pipyConf *PipyConf, egressPolicy *trafficpolicy.EgressTrafficPolicy) map[service.ClusterName]*WeightedCluster {
	if len(egressPolicy.TrafficMatches) == 0 {
		return nil
	}

	otp := pipyConf.newOutboundTrafficPolicy()
	dependClusters := make(map[service.ClusterName]*WeightedCluster)
	for _, trafficMatch := range egressPolicy.TrafficMatches {
		destinationProtocol := strings.ToLower(trafficMatch.DestinationProtocol)
		tm := otp.newTrafficMatch(Port(trafficMatch.DestinationPort))
		tm.setProtocol(Protocol(destinationProtocol))
		tm.setPort(Port(trafficMatch.DestinationPort))
		tm.setEgressForwardGateway(trafficMatch.EgressGateWay)

		var destinationSpec *DestinationSecuritySpec
		if clusterConfig := getEgressClusterConfigs(egressPolicy.ClustersConfigs, service.ClusterName(trafficMatch.Cluster)); clusterConfig != nil {
			if clusterConfig.SourceMTLS != nil {
				destinationSpec = new(DestinationSecuritySpec)
				destinationSpec.SourceCert = new(Certificate)
				osmIssued := strings.EqualFold(`osm`, clusterConfig.SourceMTLS.Issuer)
				destinationSpec.SourceCert.OsmIssued = &osmIssued
				if !osmIssued && clusterConfig.SourceMTLS.Cert != nil {
					secretReference := corev1.SecretReference{
						Name:      clusterConfig.SourceMTLS.Cert.Secret.Name,
						Namespace: clusterConfig.SourceMTLS.Cert.Secret.Namespace,
					}
					if secret, err := meshCatalog.GetEgressSourceSecret(secretReference); err == nil {
						destinationSpec.SourceCert.SubjectAltNames = clusterConfig.SourceMTLS.Cert.SubjectAltNames
						destinationSpec.SourceCert.Expiration = clusterConfig.SourceMTLS.Cert.Expiration
						if caCrt, ok := secret.Data["ca.crt"]; ok {
							destinationSpec.SourceCert.IssuingCA = string(caCrt)
						}
						if tlsCrt, ok := secret.Data["tls.crt"]; ok {
							destinationSpec.SourceCert.CertChain = string(tlsCrt)
						}
						if tlsKey, ok := secret.Data["tls.key"]; ok {
							destinationSpec.SourceCert.PrivateKey = string(tlsKey)
						}
					} else {
						log.Error().Err(err)
					}
				}
			}
		}

		for _, ipRange := range trafficMatch.DestinationIPRanges {
			tm.addDestinationIPRange(DestinationIPRange(ipRange), destinationSpec)
		}

		if destinationProtocol == constants.ProtocolHTTP || destinationProtocol == constants.ProtocolGRPC {
			httpRouteConfigs := getEgressHTTPRouteConfigs(egressPolicy.HTTPRouteConfigsPerPort, trafficMatch.DestinationPort)
			if len(httpRouteConfigs) == 0 {
				continue
			}

			for _, httpRouteConfig := range httpRouteConfigs {
				ruleName := HTTPRouteRuleName(httpRouteConfig.Name)
				hsrrs := tm.newHTTPServiceRouteRules(ruleName)
				for _, hostname := range httpRouteConfig.Hostnames {
					tm.addHTTPHostPort2Service(HTTPHostPort(hostname), ruleName)
				}
				for _, rule := range httpRouteConfig.RoutingRules {
					route := rule.Route
					pathRegexp := URIPathRegexp(route.HTTPRouteMatch.Path)
					if len(pathRegexp) == 0 {
						pathRegexp = ".*"
					}

					hsrr := hsrrs.newHTTPServiceRouteRule(pathRegexp)
					for k, v := range route.HTTPRouteMatch.Headers {
						hsrr.addHeaderMatch(Header(k), HeaderRegexp(v))
					}

					if len(route.HTTPRouteMatch.Methods) == 0 {
						hsrr.addMethodMatch("*")
					} else {
						for _, method := range route.HTTPRouteMatch.Methods {
							hsrr.addMethodMatch(Method(method))
						}
					}

					for cluster := range route.WeightedClusters.Iter() {
						serviceCluster := cluster.(service.WeightedCluster)
						weightedCluster := new(WeightedCluster)
						weightedCluster.WeightedCluster = serviceCluster
						weightedCluster.RetryPolicy = route.RetryPolicy
						if _, exist := dependClusters[weightedCluster.ClusterName]; !exist {
							dependClusters[weightedCluster.ClusterName] = weightedCluster
						}
						hsrr.addWeightedCluster(ClusterName(weightedCluster.ClusterName), Weight(weightedCluster.Weight))
					}

					for _, allowedIPRange := range rule.AllowedDestinationIPRanges {
						tm.addDestinationIPRange(DestinationIPRange(allowedIPRange), destinationSpec)
					}
				}
			}
		} else if destinationProtocol == constants.ProtocolHTTPS {
			weightedCluster := new(WeightedCluster)
			weightedCluster.ClusterName = service.ClusterName(trafficMatch.Cluster)
			weightedCluster.Weight = constants.ClusterWeightAcceptAll
			tm.addWeightedCluster(ClusterName(weightedCluster.ClusterName), Weight(weightedCluster.Weight))
			clusterConfigs := otp.newClusterConfigs(ClusterName(weightedCluster.ClusterName.String()))
			for _, serverName := range trafficMatch.ServerNames {
				address := Address(serverName)
				port := Port(trafficMatch.DestinationPort)
				weight := Weight(constants.ClusterWeightAcceptAll)
				clusterConfigs.addWeightedEndpoint(address, port, weight)
			}
		} else if destinationProtocol == constants.ProtocolTCP ||
			destinationProtocol == constants.ProtocolTCPServerFirst {
			tm.setAllowedEgressTraffic(true)
		}
	}

	return dependClusters
}

func generatePipyOutboundTrafficBalancePolicy(meshCatalog catalog.MeshCataloger, _ *pipy.Proxy,
	proxyIdentity identity.ServiceIdentity,
	pipyConf *PipyConf, outboundPolicy *trafficpolicy.OutboundMeshTrafficPolicy,
	dependClusters map[service.ClusterName]*WeightedCluster) bool {
	ready := true
	otp := pipyConf.newOutboundTrafficPolicy()
	for _, cluster := range dependClusters {
		clusterConfig := getMeshClusterConfigs(outboundPolicy.ClustersConfigs, cluster.ClusterName)
		if clusterConfig == nil {
			ready = false
			continue
		}
		clusterConfigs := otp.newClusterConfigs(ClusterName(cluster.ClusterName.String()))
		upstreamEndpoints := getUpstreamEndpoints(meshCatalog, proxyIdentity, cluster.ClusterName)
		if len(upstreamEndpoints) == 0 {
			ready = false
			continue
		}
		for _, upstreamEndpoint := range upstreamEndpoints {
			address := Address(upstreamEndpoint.IP.String())
			port := Port(clusterConfig.Service.Port)
			if targetPort := Port(clusterConfig.Service.TargetPort); targetPort > 0 {
				port = targetPort
			}
			weight := Weight(upstreamEndpoint.Weight)
			clusterConfigs.addWeightedEndpoint(address, port, weight)
			if clusterConfig.UpstreamTrafficSetting != nil {
				if clusterConfig.UpstreamTrafficSetting.Spec.ConnectionSettings != nil {
					clusterConfigs.setConnectionSettings(clusterConfig.UpstreamTrafficSetting.Spec.ConnectionSettings)
				}
			}
			if cluster.RetryPolicy != nil {
				clusterConfigs.setRetryPolicy(cluster.RetryPolicy)
			}
		}
	}
	return ready
}

func generatePipyIngressTrafficRoutePolicy(_ catalog.MeshCataloger, _ identity.ServiceIdentity, pipyConf *PipyConf, ingressPolicy *trafficpolicy.IngressTrafficPolicy) {
	if len(ingressPolicy.TrafficMatches) == 0 {
		return
	}

	if pipyConf.Inbound == nil {
		return
	}

	if len(pipyConf.Inbound.ClustersConfigs) == 0 {
		return
	}

	itp := pipyConf.newInboundTrafficPolicy()

	for _, trafficMatch := range ingressPolicy.TrafficMatches {
		tm := itp.getTrafficMatch(Port(trafficMatch.Port))
		if tm == nil {
			continue
		}

		var securitySpec *SourceSecuritySpec
		if trafficMatch.TLS != nil {
			securitySpec = &SourceSecuritySpec{
				MTLS:                     true,
				SkipClientCertValidation: trafficMatch.TLS.SkipClientCertValidation,
			}
		}

		for _, ipRange := range trafficMatch.SourceIPRanges {
			tm.addSourceIPRange(SourceIPRange(ipRange), securitySpec)
		}

		var authenticatedPrincipals []string
		protocol := strings.ToLower(trafficMatch.Protocol)
		if protocol != constants.ProtocolHTTP && protocol != constants.ProtocolHTTPS && protocol != constants.ProtocolGRPC {
			continue
		}
		for _, httpRouteConfig := range ingressPolicy.HTTPRoutePolicies {
			if len(httpRouteConfig.Rules) == 0 {
				continue
			}
			for _, hostname := range httpRouteConfig.Hostnames {
				ruleName := HTTPRouteRuleName(hostname)
				tm.addHTTPHostPort2Service(HTTPHostPort(hostname), ruleName)

				hsrrs := tm.newHTTPServiceRouteRules(ruleName)
				for _, rule := range httpRouteConfig.Rules {
					pathRegexp := URIPathRegexp(rule.Route.HTTPRouteMatch.Path)
					if len(pathRegexp) == 0 {
						continue
					}

					hsrr := hsrrs.newHTTPServiceRouteRule(pathRegexp)
					for k, v := range rule.Route.HTTPRouteMatch.Headers {
						hsrr.addHeaderMatch(Header(k), HeaderRegexp(v))
					}

					if len(rule.Route.HTTPRouteMatch.Methods) == 0 {
						hsrr.addMethodMatch("*")
					} else {
						for _, method := range rule.Route.HTTPRouteMatch.Methods {
							hsrr.addMethodMatch(Method(method))
						}
					}

					for routeCluster := range rule.Route.WeightedClusters.Iter() {
						weightedCluster := routeCluster.(service.WeightedCluster)
						hsrr.addWeightedCluster(ClusterName(weightedCluster.ClusterName),
							Weight(weightedCluster.Weight))
					}

					for allowedPrincipal := range rule.AllowedPrincipals.Iter() {
						servicePrincipal := allowedPrincipal.(string)
						authenticatedPrincipals = append(authenticatedPrincipals, servicePrincipal)
					}
				}
			}
		}

		if securitySpec != nil {
			securitySpec.AuthenticatedPrincipals = authenticatedPrincipals
		}
	}
}

func generatePipyEgressTrafficForwardPolicy(_ catalog.MeshCataloger, _ identity.ServiceIdentity, pipyConf *PipyConf, egressGatewayPolicy *trafficpolicy.EgressGatewayPolicy) bool {
	if egressGatewayPolicy == nil || (egressGatewayPolicy.Global == nil && (egressGatewayPolicy.Rules == nil || len(egressGatewayPolicy.Rules) == 0)) {
		return true
	}

	success := true
	ftp := pipyConf.newForwardTrafficPolicy()
	if egressGatewayPolicy.Global != nil {
		forwardMatch := ftp.newForwardMatch("*")
		for _, gateway := range egressGatewayPolicy.Global {
			clusterName := fmt.Sprintf("%s.%s", gateway.Service, gateway.Namespace)
			if gateway.Weight != nil {
				forwardMatch[ClusterName(clusterName)] = Weight(*gateway.Weight)
			} else {
				forwardMatch[ClusterName(clusterName)] = Weight(0)
			}
			if len(gateway.Endpoints) > 0 {
				clusterConfigs := ftp.newEgressGateway(ClusterName(clusterName))
				for _, endPeer := range gateway.Endpoints {
					address := Address(endPeer.IP.String())
					port := Port(endPeer.Port)
					weight := Weight(0)
					clusterConfigs.addWeightedEndpoint(address, port, weight)
				}
			}
		}
	}
	if egressGatewayPolicy.Rules != nil {
		for index, rule := range egressGatewayPolicy.Rules {
			ruleName := fmt.Sprintf("%s.%s.%d", rule.Namespace, rule.Name, index)
			forwardMatch := ftp.newForwardMatch(ruleName)
			for _, gateway := range rule.EgressGateways {
				clusterName := fmt.Sprintf("%s.%s", gateway.Service, gateway.Namespace)
				if gateway.Weight != nil {
					forwardMatch[ClusterName(clusterName)] = Weight(*gateway.Weight)
				} else {
					forwardMatch[ClusterName(clusterName)] = constants.ClusterWeightAcceptAll
				}
				if len(gateway.Endpoints) > 0 {
					clusterConfigs := ftp.newEgressGateway(ClusterName(clusterName))
					for _, endPeer := range gateway.Endpoints {
						address := Address(endPeer.IP.String())
						port := Port(endPeer.Port)
						weight := Weight(0)
						clusterConfigs.addWeightedEndpoint(address, port, weight)
					}
				} else {
					success = false
				}
			}
		}
	}

	return success
}

func generatePipyAccessControlTrafficRoutePolicy(_ catalog.MeshCataloger, _ identity.ServiceIdentity, pipyConf *PipyConf, aclPolicy *trafficpolicy.AccessControlTrafficPolicy) {
	if len(aclPolicy.TrafficMatches) == 0 {
		return
	}

	if pipyConf.Inbound == nil {
		return
	}

	if len(pipyConf.Inbound.ClustersConfigs) == 0 {
		return
	}

	itp := pipyConf.newInboundTrafficPolicy()

	for _, trafficMatch := range aclPolicy.TrafficMatches {
		tm := itp.getTrafficMatch(Port(trafficMatch.Port))
		if tm == nil {
			continue
		}

		var securitySpec *SourceSecuritySpec
		if trafficMatch.TLS != nil {
			securitySpec = &SourceSecuritySpec{
				MTLS:                     true,
				SkipClientCertValidation: trafficMatch.TLS.SkipClientCertValidation,
			}
		}

		for _, ipRange := range trafficMatch.SourceIPRanges {
			tm.addSourceIPRange(SourceIPRange(ipRange), securitySpec)
		}

		var authenticatedPrincipals []string
		protocol := strings.ToLower(trafficMatch.Protocol)
		if protocol != constants.ProtocolHTTP && protocol != constants.ProtocolHTTPS && protocol != constants.ProtocolGRPC {
			continue
		}
		for _, httpRouteConfig := range aclPolicy.HTTPRoutePolicies {
			if len(httpRouteConfig.Rules) == 0 {
				continue
			}
			for _, hostname := range httpRouteConfig.Hostnames {
				ruleName := HTTPRouteRuleName(hostname)
				tm.addHTTPHostPort2Service(HTTPHostPort(hostname), ruleName)

				hsrrs := tm.newHTTPServiceRouteRules(ruleName)
				for _, rule := range httpRouteConfig.Rules {
					pathRegexp := URIPathRegexp(rule.Route.HTTPRouteMatch.Path)
					if len(pathRegexp) == 0 {
						continue
					}

					hsrr := hsrrs.newHTTPServiceRouteRule(pathRegexp)
					for k, v := range rule.Route.HTTPRouteMatch.Headers {
						hsrr.addHeaderMatch(Header(k), HeaderRegexp(v))
					}

					if len(rule.Route.HTTPRouteMatch.Methods) == 0 {
						hsrr.addMethodMatch("*")
					} else {
						for _, method := range rule.Route.HTTPRouteMatch.Methods {
							hsrr.addMethodMatch(Method(method))
						}
					}

					for routeCluster := range rule.Route.WeightedClusters.Iter() {
						weightedCluster := routeCluster.(service.WeightedCluster)
						hsrr.addWeightedCluster(ClusterName(weightedCluster.ClusterName),
							Weight(weightedCluster.Weight))
					}

					for allowedPrincipal := range rule.AllowedPrincipals.Iter() {
						servicePrincipal := allowedPrincipal.(string)
						authenticatedPrincipals = append(authenticatedPrincipals, servicePrincipal)
					}
				}
			}
		}

		if securitySpec != nil {
			securitySpec.AuthenticatedPrincipals = authenticatedPrincipals
		}
	}
}

func generatePipyEgressTrafficBalancePolicy(meshCatalog catalog.MeshCataloger, _ *pipy.Proxy, serviceIdentity identity.ServiceIdentity, pipyConf *PipyConf, egressPolicy *trafficpolicy.EgressTrafficPolicy, dependClusters map[service.ClusterName]*WeightedCluster) bool {
	ready := true
	otp := pipyConf.newOutboundTrafficPolicy()
	for _, cluster := range dependClusters {
		clusterConfig := getEgressClusterConfigs(egressPolicy.ClustersConfigs, cluster.ClusterName)
		if clusterConfig == nil {
			ready = false
			continue
		}
		clusterConfigs := otp.newClusterConfigs(ClusterName(cluster.ClusterName.String()))
		address := Address(clusterConfig.Name)
		port := Port(clusterConfig.Port)
		weight := Weight(constants.ClusterWeightAcceptAll)
		clusterConfigs.addWeightedEndpoint(address, port, weight)
		if clusterConfig.UpstreamTrafficSetting != nil {
			clusterConfigs.setConnectionSettings(clusterConfig.UpstreamTrafficSetting.Spec.ConnectionSettings)
		}
		if clusterConfig.SourceMTLS != nil {
			clusterConfigs.SourceCert = new(Certificate)
			osmIssued := strings.EqualFold(`osm`, clusterConfig.SourceMTLS.Issuer)
			clusterConfigs.SourceCert.OsmIssued = &osmIssued
			if !osmIssued && clusterConfig.SourceMTLS.Cert != nil {
				secretReference := corev1.SecretReference{
					Name:      clusterConfig.SourceMTLS.Cert.Secret.Name,
					Namespace: clusterConfig.SourceMTLS.Cert.Secret.Namespace,
				}
				if secret, err := meshCatalog.GetEgressSourceSecret(secretReference); err == nil {
					clusterConfigs.SourceCert.SubjectAltNames = clusterConfig.SourceMTLS.Cert.SubjectAltNames
					clusterConfigs.SourceCert.Expiration = clusterConfig.SourceMTLS.Cert.Expiration
					if caCrt, ok := secret.Data["ca.crt"]; ok {
						clusterConfigs.SourceCert.IssuingCA = string(caCrt)
					}
					if tlsCrt, ok := secret.Data["tls.crt"]; ok {
						clusterConfigs.SourceCert.CertChain = string(tlsCrt)
					}
					if tlsKey, ok := secret.Data["tls.key"]; ok {
						clusterConfigs.SourceCert.PrivateKey = string(tlsKey)
					}
				} else {
					log.Error().Err(err)
				}
			}
		}
		if cluster.RetryPolicy != nil {
			clusterConfigs.setRetryPolicy(cluster.RetryPolicy)
		} else if upstreamSvc, err := hostToMeshSvc(cluster.ClusterName.String()); err == nil {
			if retryPolicy := meshCatalog.GetRetryPolicy(serviceIdentity, upstreamSvc); retryPolicy != nil {
				clusterConfigs.setRetryPolicy(retryPolicy)
			}
		}
	}
	return ready
}

func getInboundHTTPRouteConfigs(httpRouteConfigsPerPort map[int][]*trafficpolicy.InboundTrafficPolicy,
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

func getOutboundHTTPRouteConfigs(httpRouteConfigsPerPort map[int][]*trafficpolicy.OutboundTrafficPolicy,
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

func getEgressHTTPRouteConfigs(httpRouteConfigsPerPort map[int][]*trafficpolicy.EgressHTTPRouteConfig,
	targetPort int) []*trafficpolicy.EgressHTTPRouteConfig {
	if httpRouteConfigs, ok := httpRouteConfigsPerPort[targetPort]; ok {
		return httpRouteConfigs
	}
	return nil
}

func trafficMatchToMeshSvc(trafficMatch *trafficpolicy.TrafficMatch) *service.MeshService {
	splitFunc := func(r rune) bool {
		return r == '_'
	}

	chunks := strings.FieldsFunc(trafficMatch.Name, splitFunc)
	if len(chunks) != 4 {
		log.Error().Msgf("Invalid traffic match name. Expected: xxx_<namespace>/<name>_<port>_<protocol>, got: %s",
			trafficMatch.Name)
		return nil
	}

	namespacedName, err := k8s.NamespacedNameFrom(chunks[1])
	if err != nil {
		log.Error().Err(err).Msgf("Error retrieving NamespacedName from TrafficMatch")
		return nil
	}
	return &service.MeshService{
		Namespace:  namespacedName.Namespace,
		Name:       namespacedName.Name,
		Protocol:   strings.ToLower(trafficMatch.DestinationProtocol),
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

// hostToMeshSvc returns the MeshService associated with the given host name
func hostToMeshSvc(cluster string) (service.MeshService, error) {
	splitFunc := func(r rune) bool {
		return r == '.' || r == ':'
	}

	chunks := strings.FieldsFunc(cluster, splitFunc)
	if len(chunks) > 4 && strings.EqualFold("svc", chunks[3]) {
		return service.MeshService{},
			errors.Errorf("Invalid host. Expected: <name>.<namespace>.svc.trustdomain:<port>, got: %s", cluster)
	}

	port, err := strconv.ParseUint(chunks[len(chunks)-1], 10, 16)
	if err != nil {
		return service.MeshService{}, errors.Errorf("Invalid cluster port %s, expected int value: %s", chunks[len(chunks)-1], err)
	}

	return service.MeshService{
		Namespace: chunks[1],
		Name:      chunks[0],
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

func hash(bytes []byte) uint64 {
	if hashCode, err := utils.HashFromString(string(bytes)); err == nil {
		return hashCode
	}
	return uint64(time.Now().Nanosecond())
}
