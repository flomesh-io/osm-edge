package repo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/openservicemesh/osm/pkg/apis/policy/v1alpha1"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/k8s"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/registry"
)

var (
	addrWithPort, _ = regexp.Compile(`:\d+$`)
)

func (p *PipyConf) setSidecarLogLevel(sidecarLogLevel string) (update bool) {
	if update = !strings.EqualFold(p.Spec.SidecarLogLevel, sidecarLogLevel); update {
		p.Spec.SidecarLogLevel = sidecarLogLevel
	}
	return
}

func (p *PipyConf) setEnableSidecarActiveHealthChecks(enableSidecarActiveHealthChecks bool) (update bool) {
	if update = p.Spec.FeatureFlags.EnableSidecarActiveHealthChecks != enableSidecarActiveHealthChecks; update {
		p.Spec.FeatureFlags.EnableSidecarActiveHealthChecks = enableSidecarActiveHealthChecks
	}
	return
}

func (p *PipyConf) setEnableEgress(enableEgress bool) (update bool) {
	if update = p.Spec.Traffic.EnableEgress != enableEgress; update {
		p.Spec.Traffic.EnableEgress = enableEgress
	}
	return
}

func (p *PipyConf) setEnablePermissiveTrafficPolicyMode(enablePermissiveTrafficPolicyMode bool) (update bool) {
	if update = p.Spec.Traffic.enablePermissiveTrafficPolicyMode != enablePermissiveTrafficPolicyMode; update {
		p.Spec.Traffic.enablePermissiveTrafficPolicyMode = enablePermissiveTrafficPolicyMode
	}
	return
}

func (p *PipyConf) isPermissiveTrafficPolicyMode() bool {
	return p.Spec.Traffic.enablePermissiveTrafficPolicyMode
}

func (p *PipyConf) newInboundTrafficPolicy() *InboundTrafficPolicy {
	if p.Inbound == nil {
		p.Inbound = new(InboundTrafficPolicy)
	}
	return p.Inbound
}

func (p *PipyConf) newOutboundTrafficPolicy() *OutboundTrafficPolicy {
	if p.Outbound == nil {
		p.Outbound = new(OutboundTrafficPolicy)
	}
	return p.Outbound
}

func (p *PipyConf) newForwardTrafficPolicy() *ForwardTrafficPolicy {
	if p.Forward == nil {
		p.Forward = new(ForwardTrafficPolicy)
	}
	return p.Forward
}

func (p *PipyConf) rebalancedOutboundClusters() {
	if p.Outbound == nil {
		return
	}
	if p.Outbound.ClustersConfigs == nil || len(p.Outbound.ClustersConfigs) == 0 {
		return
	}
	for _, clusterConfigs := range p.Outbound.ClustersConfigs {
		weightedEndpoints := clusterConfigs.Endpoints
		if weightedEndpoints == nil || len(*weightedEndpoints) == 0 {
			continue
		}
		missingWeightNb := 0
		availableWeight := uint32(100)
		for _, weight := range *weightedEndpoints {
			if weight == 0 {
				missingWeightNb++
			} else {
				availableWeight = availableWeight - uint32(weight)
			}
		}

		if missingWeightNb == len(*weightedEndpoints) {
			for upstreamEndpoint, weight := range *weightedEndpoints {
				if weight == 0 {
					weight = Weight(availableWeight / uint32(missingWeightNb))
					missingWeightNb--
					availableWeight = availableWeight - uint32(weight)
					(*weightedEndpoints)[upstreamEndpoint] = weight
				}
			}
		}
	}
}

func (p *PipyConf) rebalancedForwardClusters() {
	if p.Forward == nil {
		return
	}
	if p.Forward.ForwardMatches != nil && len(p.Forward.ForwardMatches) > 0 {
		for _, weightedEndpoints := range p.Forward.ForwardMatches {
			if len(weightedEndpoints) == 0 {
				continue
			}
			missingWeightNb := 0
			availableWeight := uint32(100)
			for _, weight := range weightedEndpoints {
				if weight == 0 {
					missingWeightNb++
				} else {
					availableWeight = availableWeight - uint32(weight)
				}
			}

			if missingWeightNb == len(weightedEndpoints) {
				for upstreamEndpoint, weight := range weightedEndpoints {
					if weight == 0 {
						weight = Weight(availableWeight / uint32(missingWeightNb))
						missingWeightNb--
						availableWeight = availableWeight - uint32(weight)
						(weightedEndpoints)[upstreamEndpoint] = weight
					}
				}
			}
		}
	}
	if p.Forward.EgressGateways != nil && len(p.Forward.EgressGateways) > 0 {
		for _, clusterConfigs := range p.Forward.EgressGateways {
			weightedEndpoints := clusterConfigs.Endpoints
			if weightedEndpoints == nil || len(*weightedEndpoints) == 0 {
				continue
			}
			missingWeightNb := 0
			availableWeight := uint32(100)
			for _, weight := range *weightedEndpoints {
				if weight == 0 {
					missingWeightNb++
				} else {
					availableWeight = availableWeight - uint32(weight)
				}
			}

			if missingWeightNb == len(*weightedEndpoints) {
				for upstreamEndpoint, weight := range *weightedEndpoints {
					if weight == 0 {
						weight = Weight(availableWeight / uint32(missingWeightNb))
						missingWeightNb--
						availableWeight = availableWeight - uint32(weight)
						(*weightedEndpoints)[upstreamEndpoint] = weight
					}
				}
			}
		}
	}
}

func (p *PipyConf) copyAllowedEndpoints(kubeController k8s.Controller, proxyRegistry *registry.ProxyRegistry) bool {
	ready := true
	p.AllowedEndpoints = make(map[string]string)
	allPods := kubeController.ListPods()
	for _, pod := range allPods {
		proxyUUID, err := GetProxyUUIDFromPod(pod)
		if err != nil {
			ready = false
			continue
		}
		proxy := proxyRegistry.GetConnectedProxy(proxyUUID)
		if proxy == nil {
			ready = false
			continue
		}
		p.AllowedEndpoints[proxy.GetAddr()] = fmt.Sprintf("%s.%s", pod.Namespace, pod.Name)
		if len(proxy.GetAddr()) == 0 {
			ready = false
		}
	}
	if p.Inbound == nil {
		return ready
	}
	if len(p.Inbound.TrafficMatches) == 0 {
		return ready
	}
	for _, trafficMatch := range p.Inbound.TrafficMatches {
		if len(trafficMatch.SourceIPRanges) == 0 {
			continue
		}
		for ipRange := range trafficMatch.SourceIPRanges {
			ingressIP := strings.TrimSuffix(string(ipRange), "/32")
			p.AllowedEndpoints[ingressIP] = "Ingress Controller"
		}
	}
	return ready
}

func (itm *InboundTrafficMatch) addSourceIPRange(ipRange SourceIPRange, sourceSpec *SourceSecuritySpec) {
	if itm.SourceIPRanges == nil {
		itm.SourceIPRanges = make(map[SourceIPRange]*SourceSecuritySpec)
	}
	if _, exists := itm.SourceIPRanges[ipRange]; !exists {
		itm.SourceIPRanges[ipRange] = sourceSpec
	}
}

func (itm *InboundTrafficMatch) addAllowedEndpoint(address Address, serviceName ServiceName) {
	if itm.AllowedEndpoints == nil {
		itm.AllowedEndpoints = make(AllowedEndpoints)
	}
	if _, exists := itm.AllowedEndpoints[address]; !exists {
		itm.AllowedEndpoints[address] = serviceName
	}
}

func (itm *InboundTrafficMatch) setTCPServiceRateLimit(rateLimit *v1alpha1.RateLimitSpec) {
	if rateLimit == nil || rateLimit.Local == nil {
		itm.RateLimit = nil
	} else {
		itm.RateLimit = newTCPRateLimit(rateLimit.Local)
	}
}

func (otm *OutboundTrafficMatch) addDestinationIPRange(ipRange DestinationIPRange, destinationSpec *DestinationSecuritySpec) {
	if otm.DestinationIPRanges == nil {
		otm.DestinationIPRanges = make(map[DestinationIPRange]*DestinationSecuritySpec)
	}
	if _, exists := otm.DestinationIPRanges[ipRange]; !exists {
		otm.DestinationIPRanges[ipRange] = destinationSpec
	}
}

func (otm *OutboundTrafficMatch) setServiceIdentity(serviceIdentity identity.ServiceIdentity) {
	otm.ServiceIdentity = serviceIdentity
}

func (otm *OutboundTrafficMatch) setAllowedEgressTraffic(allowedEgressTraffic bool) {
	otm.AllowedEgressTraffic = allowedEgressTraffic
}

func (itm *InboundTrafficMatch) setPort(port Port) {
	itm.Port = port
}

func (otm *OutboundTrafficMatch) setPort(port Port) {
	otm.Port = port
}

func (otm *OutboundTrafficMatch) setEgressForwardGateway(egresssGateway *string) {
	otm.EgressForwardGateway = egresssGateway
}

func (itm *InboundTrafficMatch) setProtocol(protocol Protocol) {
	protocol = Protocol(strings.ToLower(string(protocol)))
	if constants.ProtocolTCPServerFirst == protocol {
		itm.Protocol = constants.ProtocolTCP
	} else {
		itm.Protocol = protocol
	}
}

func (otm *OutboundTrafficMatch) setProtocol(protocol Protocol) {
	protocol = Protocol(strings.ToLower(string(protocol)))
	if constants.ProtocolTCPServerFirst == protocol {
		otm.Protocol = constants.ProtocolTCP
	} else {
		otm.Protocol = protocol
	}
}

func (itm *InboundTrafficMatch) addWeightedCluster(clusterName ClusterName, weight Weight) {
	if itm.TargetClusters == nil {
		itm.TargetClusters = make(WeightedClusters)
	}
	itm.TargetClusters[clusterName] = weight
}

func (otm *OutboundTrafficMatch) addWeightedCluster(clusterName ClusterName, weight Weight) {
	if otm.TargetClusters == nil {
		otm.TargetClusters = make(WeightedClusters)
	}
	otm.TargetClusters[clusterName] = weight
}

func (itm *InboundTrafficMatch) addHTTPHostPort2Service(hostPort HTTPHostPort, ruleName HTTPRouteRuleName) {
	if itm.HTTPHostPort2Service == nil {
		itm.HTTPHostPort2Service = make(HTTPHostPort2Service)
	}
	itm.HTTPHostPort2Service[hostPort] = ruleName
}

func (otm *OutboundTrafficMatch) addHTTPHostPort2Service(hostPort HTTPHostPort, ruleName HTTPRouteRuleName) {
	if otm.HTTPHostPort2Service == nil {
		otm.HTTPHostPort2Service = make(HTTPHostPort2Service)
	}
	otm.HTTPHostPort2Service[hostPort] = ruleName
}

func (itm *InboundTrafficMatch) newHTTPServiceRouteRules(httpRouteRuleName HTTPRouteRuleName) *InboundHTTPRouteRules {
	if itm.HTTPServiceRouteRules == nil {
		itm.HTTPServiceRouteRules = make(InboundHTTPServiceRouteRules)
	}
	if len(httpRouteRuleName) == 0 {
		return nil
	}
	rules, exist := itm.HTTPServiceRouteRules[httpRouteRuleName]
	if !exist || rules == nil {
		newCluster := new(InboundHTTPRouteRules)
		itm.HTTPServiceRouteRules[httpRouteRuleName] = newCluster
		return newCluster
	}
	return rules
}

func (otm *OutboundTrafficMatch) newHTTPServiceRouteRules(httpRouteRuleName HTTPRouteRuleName) *OutboundHTTPRouteRules {
	if otm.HTTPServiceRouteRules == nil {
		otm.HTTPServiceRouteRules = make(OutboundHTTPServiceRouteRules)
	}
	if len(httpRouteRuleName) == 0 {
		return nil
	}
	rules, exist := otm.HTTPServiceRouteRules[httpRouteRuleName]
	if !exist || rules == nil {
		newCluster := make(OutboundHTTPRouteRules, 0)
		otm.HTTPServiceRouteRules[httpRouteRuleName] = &newCluster
		return &newCluster
	}
	return rules
}

func (itp *InboundTrafficPolicy) newTrafficMatch(port Port) *InboundTrafficMatch {
	if itp.TrafficMatches == nil {
		itp.TrafficMatches = make(InboundTrafficMatches)
	}
	trafficMatch, exist := itp.TrafficMatches[port]
	if !exist || trafficMatch == nil {
		trafficMatch = new(InboundTrafficMatch)
		itp.TrafficMatches[port] = trafficMatch
		return trafficMatch
	}
	return trafficMatch
}

func (itp *InboundTrafficPolicy) getTrafficMatch(port Port) *InboundTrafficMatch {
	if itp.TrafficMatches == nil {
		return nil
	}
	if trafficMatch, exist := itp.TrafficMatches[port]; exist {
		return trafficMatch
	}
	return nil
}

func (otp *OutboundTrafficPolicy) newTrafficMatch(port Port) *OutboundTrafficMatch {
	trafficMatch := new(OutboundTrafficMatch)
	if otp.TrafficMatches == nil {
		otp.TrafficMatches = make(OutboundTrafficMatches)
	}
	trafficMatches := otp.TrafficMatches[port]
	trafficMatches = append(trafficMatches, trafficMatch)
	otp.TrafficMatches[port] = trafficMatches
	return trafficMatch
}

func (hrrs *InboundHTTPRouteRules) setHTTPServiceRateLimit(rateLimit *v1alpha1.RateLimitSpec) {
	if rateLimit == nil || rateLimit.Local == nil {
		hrrs.RateLimit = nil
	} else {
		hrrs.RateLimit = newHTTPRateLimit(rateLimit.Local)
	}
}

func (hrrs *InboundHTTPRouteRules) setHTTPHeadersRateLimit(rateLimit *[]v1alpha1.HTTPHeaderSpec) {
	if rateLimit == nil {
		hrrs.HeaderRateLimits = nil
	} else {
		hrrs.HeaderRateLimits = newHTTPHeaderRateLimit(rateLimit)
	}
}

func (hrrs *InboundHTTPRouteRules) newHTTPServiceRouteRule(pathReg URIPathRegexp) *InboundHTTPRouteRule {
	if hrrs.RouteRules == nil {
		hrrs.RouteRules = make(map[URIPathRegexp]*InboundHTTPRouteRule)
	}
	routeRule, exist := hrrs.RouteRules[pathReg]
	if !exist || routeRule == nil {
		routeRule = new(InboundHTTPRouteRule)
		hrrs.RouteRules[pathReg] = routeRule
		return routeRule
	}
	return routeRule
}

func (hrrs *OutboundHTTPRouteRules) newHTTPServiceRouteRule(pathReg URIPathRegexp) *HTTPRouteRule {
	routeRule, exist := (*hrrs)[pathReg]
	if !exist || routeRule == nil {
		routeRule = new(HTTPRouteRule)
		(*hrrs)[pathReg] = routeRule
		return routeRule
	}
	return routeRule
}

func (hrr *HTTPRouteRule) addHeaderMatch(header Header, headerRegexp HeaderRegexp) {
	if hrr.Headers == nil {
		hrr.Headers = make(Headers)
	}
	hrr.Headers[header] = headerRegexp
}

func (hrr *HTTPRouteRule) addMethodMatch(method Method) {
	if hrr.allowedAnyMethod {
		return
	}
	if "*" == method {
		hrr.allowedAnyMethod = true
	}
	if hrr.allowedAnyMethod {
		hrr.Methods = nil
	} else {
		hrr.Methods = append(hrr.Methods, method)
	}
}

func (hrr *HTTPRouteRule) addWeightedCluster(clusterName ClusterName, weight Weight) {
	if hrr.TargetClusters == nil {
		hrr.TargetClusters = make(WeightedClusters)
	}
	hrr.TargetClusters[clusterName] = weight
}

func (hrr *HTTPRouteRule) addAllowedService(serviceName ServiceName) {
	if hrr.allowedAnyService {
		return
	}
	if "*" == serviceName {
		hrr.allowedAnyService = true
	}
	if hrr.allowedAnyService {
		hrr.AllowedServices = nil
	} else {
		hrr.AllowedServices = append(hrr.AllowedServices, serviceName)
	}
}

func (ihrr *InboundHTTPRouteRule) setRateLimit(rateLimit *v1alpha1.HTTPPerRouteRateLimitSpec) {
	ihrr.RateLimit = newHTTPPerRouteRateLimit(rateLimit)
}

func (itp *InboundTrafficPolicy) newClusterConfigs(clusterName ClusterName) *WeightedEndpoint {
	if itp.ClustersConfigs == nil {
		itp.ClustersConfigs = make(map[ClusterName]*WeightedEndpoint)
	}
	cluster, exist := itp.ClustersConfigs[clusterName]
	if !exist || cluster == nil {
		newCluster := make(WeightedEndpoint, 0)
		itp.ClustersConfigs[clusterName] = &newCluster
		return &newCluster
	}
	return cluster
}

func (otp *OutboundTrafficPolicy) newClusterConfigs(clusterName ClusterName) *ClusterConfigs {
	if otp.ClustersConfigs == nil {
		otp.ClustersConfigs = make(map[ClusterName]*ClusterConfigs)
	}
	cluster, exist := otp.ClustersConfigs[clusterName]
	if !exist || cluster == nil {
		newCluster := new(ClusterConfigs)
		otp.ClustersConfigs[clusterName] = newCluster
		return newCluster
	}
	return cluster
}

func (otp *ClusterConfigs) addWeightedEndpoint(address Address, port Port, weight Weight) {
	if otp.Endpoints == nil {
		weightedEndpoints := make(WeightedEndpoint)
		otp.Endpoints = &weightedEndpoints
	}
	otp.Endpoints.addWeightedEndpoint(address, port, weight)
}

func (we *WeightedEndpoint) addWeightedEndpoint(address Address, port Port, weight Weight) {
	if addrWithPort.MatchString(string(address)) {
		httpHostPort := HTTPHostPort(address)
		(*we)[httpHostPort] = weight
	} else {
		httpHostPort := HTTPHostPort(fmt.Sprintf("%s:%d", address, port))
		(*we)[httpHostPort] = weight
	}
}

func (otp *ClusterConfigs) setConnectionSettings(connectionSettings *v1alpha1.ConnectionSettingsSpec) {
	if connectionSettings == nil {
		otp.ConnectionSettings = nil
		return
	}
	otp.ConnectionSettings = new(ConnectionSettings)
	if connectionSettings.TCP != nil {
		otp.ConnectionSettings.TCP = new(TCPConnectionSettings)
		otp.ConnectionSettings.TCP.MaxConnections = connectionSettings.TCP.MaxConnections
		if connectionSettings.TCP.ConnectTimeout != nil {
			duration := connectionSettings.TCP.ConnectTimeout.Seconds()
			otp.ConnectionSettings.TCP.ConnectTimeout = &duration
		}
	}
	if connectionSettings.HTTP != nil {
		otp.ConnectionSettings.HTTP = new(HTTPConnectionSettings)
		otp.ConnectionSettings.HTTP.MaxRequests = connectionSettings.HTTP.MaxRequests
		otp.ConnectionSettings.HTTP.MaxRequestsPerConnection = connectionSettings.HTTP.MaxRequestsPerConnection
		otp.ConnectionSettings.HTTP.MaxPendingRequests = connectionSettings.HTTP.MaxPendingRequests
		otp.ConnectionSettings.HTTP.MaxRetries = connectionSettings.HTTP.MaxRetries
		if connectionSettings.HTTP.CircuitBreaking != nil {
			otp.ConnectionSettings.HTTP.CircuitBreaking = new(HTTPCircuitBreaking)
			if connectionSettings.HTTP.CircuitBreaking.StatTimeWindow != nil {
				duration := connectionSettings.HTTP.CircuitBreaking.StatTimeWindow.Seconds()
				otp.ConnectionSettings.HTTP.CircuitBreaking.StatTimeWindow = &duration
			}
			otp.ConnectionSettings.HTTP.CircuitBreaking.MinRequestAmount = connectionSettings.HTTP.CircuitBreaking.MinRequestAmount
			if connectionSettings.HTTP.CircuitBreaking.DegradedTimeWindow != nil {
				duration := connectionSettings.HTTP.CircuitBreaking.DegradedTimeWindow.Seconds()
				otp.ConnectionSettings.HTTP.CircuitBreaking.DegradedTimeWindow = &duration
			}
			if connectionSettings.HTTP.CircuitBreaking.SlowTimeThreshold != nil {
				duration := connectionSettings.HTTP.CircuitBreaking.SlowTimeThreshold.Seconds()
				otp.ConnectionSettings.HTTP.CircuitBreaking.SlowTimeThreshold = &duration
			}
			otp.ConnectionSettings.HTTP.CircuitBreaking.SlowAmountThreshold = connectionSettings.HTTP.CircuitBreaking.SlowAmountThreshold
			otp.ConnectionSettings.HTTP.CircuitBreaking.SlowRatioThreshold = connectionSettings.HTTP.CircuitBreaking.SlowRatioThreshold
			otp.ConnectionSettings.HTTP.CircuitBreaking.ErrorAmountThreshold = connectionSettings.HTTP.CircuitBreaking.ErrorAmountThreshold
			otp.ConnectionSettings.HTTP.CircuitBreaking.ErrorRatioThreshold = connectionSettings.HTTP.CircuitBreaking.ErrorRatioThreshold
			otp.ConnectionSettings.HTTP.CircuitBreaking.DegradedStatusCode = connectionSettings.HTTP.CircuitBreaking.DegradedStatusCode
			otp.ConnectionSettings.HTTP.CircuitBreaking.DegradedResponseContent = connectionSettings.HTTP.CircuitBreaking.DegradedResponseContent
		}
	}
}

func (otp *ClusterConfigs) setRetryPolicy(retryPolicy *v1alpha1.RetryPolicySpec) {
	if retryPolicy == nil {
		otp.RetryPolicy = nil
		return
	}
	otp.RetryPolicy = new(RetryPolicy)
	otp.RetryPolicy.RetryOn = retryPolicy.RetryOn
	otp.RetryPolicy.NumRetries = retryPolicy.NumRetries
	perTryTimeout := retryPolicy.PerTryTimeout.Seconds()
	otp.RetryPolicy.PerTryTimeout = &perTryTimeout
	retryBackoffBaseInterval := retryPolicy.RetryBackoffBaseInterval.Seconds()
	otp.RetryPolicy.RetryBackoffBaseInterval = &retryBackoffBaseInterval
}

func (ftp *ForwardTrafficPolicy) newForwardMatch(rule string) WeightedClusters {
	if ftp.ForwardMatches == nil {
		ftp.ForwardMatches = make(ForwardTrafficMatches)
	}
	forwardMatch, exist := ftp.ForwardMatches[rule]
	if !exist || forwardMatch == nil {
		forwardMatch = make(WeightedClusters)
		ftp.ForwardMatches[rule] = forwardMatch
		return forwardMatch
	}
	return forwardMatch
}

func (ftp *ForwardTrafficPolicy) newEgressGateway(clusterName ClusterName) *ClusterConfigs {
	if ftp.EgressGateways == nil {
		ftp.EgressGateways = make(map[ClusterName]*ClusterConfigs)
	}
	cluster, exist := ftp.EgressGateways[clusterName]
	if !exist || cluster == nil {
		newCluster := new(ClusterConfigs)
		ftp.EgressGateways[clusterName] = newCluster
		return newCluster
	}
	return cluster
}
