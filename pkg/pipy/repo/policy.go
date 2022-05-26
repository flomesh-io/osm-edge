package repo

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/pipy/registry"
)

var (
	addrWithPort, _ = regexp.Compile(`:\d+$`)
)

func (p *PipyConf) setEnableSidecarActiveHealthChecks(enableSidecarActiveHealthChecks bool) {
	p.Spec.FeatureFlags.EnableSidecarActiveHealthChecks = enableSidecarActiveHealthChecks
}

func (p *PipyConf) setEnableEgress(enableEgress bool) {
	p.Spec.Traffic.EnableEgress = enableEgress
}

func (p *PipyConf) setEnablePermissiveTrafficPolicyMode(enablePermissiveTrafficPolicyMode bool) {
	p.Spec.Traffic.enablePermissiveTrafficPolicyMode = enablePermissiveTrafficPolicyMode
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

func (p *PipyConf) rebalanceOutboundClusters() {
	if p.Outbound == nil {
		return
	}
	if p.Outbound.ClustersConfigs == nil || len(p.Outbound.ClustersConfigs) == 0 {
		return
	}
	for _, weightedEndpoints := range p.Outbound.ClustersConfigs {
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

func (p *PipyConf) copyAllowedEndpoints() {
	p.AllowedEndpoints = make(map[string]string)
	registry.CachedMeshPodsLock.RLock()
	p.allowedEndpointsV = registry.CachedMeshPodsV
	for k, v := range registry.CachedMeshPods {
		p.AllowedEndpoints[k] = v
	}
	registry.CachedMeshPodsLock.RUnlock()
	if p.Inbound == nil {
		return
	}
	if len(p.Inbound.TrafficMatches) == 0 {
		return
	}
	for _, trafficMatch := range p.Inbound.TrafficMatches {
		if len(trafficMatch.SourceIPRanges) == 0 {
			continue
		}
		for _, ipRange := range trafficMatch.SourceIPRanges {
			ingressIP := strings.TrimSuffix(string(ipRange), "/32")
			p.AllowedEndpoints[ingressIP] = "Ingress Controller"
		}
	}
}

func (itm *InboundTrafficMatch) addSourceIPRange(ipRange SourceIPRange) {
	itm.SourceIPRanges = append(itm.SourceIPRanges, ipRange)
}

func (itm *InboundTrafficMatch) addAllowedEndpoint(address Address, serviceName ServiceName) {
	if itm.AllowedEndpoints == nil {
		itm.AllowedEndpoints = make(AllowedEndpoints)
	}
	if _, exists := itm.AllowedEndpoints[address]; !exists {
		itm.AllowedEndpoints[address] = serviceName
	}
}

func (otm *OutboundTrafficMatch) addDestinationIPRange(ipRange DestinationIPRange) {
	otm.DestinationIPRanges = append(otm.DestinationIPRanges, ipRange)
}

func (otm *OutboundTrafficMatch) setServiceIdentity(serviceIdentity identity.ServiceIdentity) {
	otm.ServiceIdentity = serviceIdentity
}

func (otm *OutboundTrafficMatch) setAllowedEgressTraffic(allowedEgressTraffic bool) {
	otm.AllowedEgressTraffic = allowedEgressTraffic
}

func (tm *TrafficMatch) setPort(port Port) {
	tm.Port = port
}

func (tm *TrafficMatch) setProtocol(protocol Protocol) {
	if constants.ProtocolTCPServerFirst == protocol {
		tm.Protocol = constants.ProtocolTCP
	} else {
		tm.Protocol = protocol
	}
}

func (tm *TrafficMatch) addWeightedCluster(clusterName ClusterName, weight Weight) {
	if tm.TargetClusters == nil {
		tm.TargetClusters = make(WeightedClusters)
	}
	tm.TargetClusters[clusterName] = weight
}

func (tm *TrafficMatch) addHTTPHostPort2Service(hostPort HttpHostPort, ruleName HttpRouteRuleName) {
	if tm.HttpHostPort2Service == nil {
		tm.HttpHostPort2Service = make(HttpHostPort2Service)
	}
	tm.HttpHostPort2Service[hostPort] = ruleName
}

func (tm *TrafficMatch) newHTTPServiceRouteRules(httpRouteRuleName HttpRouteRuleName) *HttpRouteRules {
	if tm.HttpServiceRouteRules == nil {
		tm.HttpServiceRouteRules = make(HttpServiceRouteRules)
	}
	if len(httpRouteRuleName) == 0 {
		return nil
	}
	rules, exist := tm.HttpServiceRouteRules[httpRouteRuleName]
	if !exist || rules == nil {
		newCluster := make(HttpRouteRules, 0)
		tm.HttpServiceRouteRules[httpRouteRuleName] = &newCluster
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

func (hrrs *HttpRouteRules) newHTTPServiceRouteRule(pathReg URIPathRegexp) *HttpRouteRule {
	routeRule, exist := (*hrrs)[pathReg]
	if !exist || routeRule == nil {
		routeRule = new(HttpRouteRule)
		(*hrrs)[pathReg] = routeRule
		return routeRule
	}
	return routeRule
}

func (hrr *HttpRouteRule) addHeaderMatch(header Header, headerRegexp HeaderRegexp) {
	if hrr.Headers == nil {
		hrr.Headers = make(HeadersMatch)
	}
	hrr.Headers[header] = headerRegexp
}

func (hrr *HttpRouteRule) addMethodMatch(method Method) {
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

func (hrr *HttpRouteRule) addWeightedCluster(clusterName ClusterName, weight Weight) {
	if hrr.TargetClusters == nil {
		hrr.TargetClusters = make(WeightedClusters)
	}
	hrr.TargetClusters[clusterName] = weight
}

func (hrr *HttpRouteRule) addAllowedService(serviceName ServiceName) {
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

func (tp *TrafficPolicy) newClusterConfigs(clusterName ClusterName) *WeightedEndpoint {
	if tp.ClustersConfigs == nil {
		tp.ClustersConfigs = make(ClustersConfigs)
	}
	cluster, exist := tp.ClustersConfigs[clusterName]
	if !exist || cluster == nil {
		newCluster := make(WeightedEndpoint, 0)
		tp.ClustersConfigs[clusterName] = &newCluster
		return &newCluster
	}
	return cluster
}

func (we *WeightedEndpoint) addWeightedEndpoint(
	address Address,
	port Port,
	weight Weight) {
	if addrWithPort.MatchString(string(address)) {
		httpHostPort := HttpHostPort(address)
		(*we)[httpHostPort] = weight
	} else {
		httpHostPort := HttpHostPort(fmt.Sprintf("%s:%d", address, port))
		(*we)[httpHostPort] = weight
	}
}
