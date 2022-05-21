package repo

import (
	"fmt"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/pipy/registry"
	"regexp"
	"strings"
)

var (
	addrWithPort, _ = regexp.Compile(":\\d+$")
)

func (p *PipyConf) SetEnableSidecarActiveHealthChecks(enableSidecarActiveHealthChecks bool) {
	p.Spec.FeatureFlags.EnableSidecarActiveHealthChecks = enableSidecarActiveHealthChecks
}

func (p *PipyConf) SetEnableEgress(enableEgress bool) {
	p.Spec.Traffic.EnableEgress = enableEgress
}

func (p *PipyConf) setEnablePermissiveTrafficPolicyMode(enablePermissiveTrafficPolicyMode bool) {
	p.Spec.Traffic.enablePermissiveTrafficPolicyMode = enablePermissiveTrafficPolicyMode
}

func (p *PipyConf) isPermissiveTrafficPolicyMode() bool {
	return p.Spec.Traffic.enablePermissiveTrafficPolicyMode
}

func (p *PipyConf) NewInboundTrafficPolicy() *InboundTrafficPolicy {
	if p.Inbound == nil {
		p.Inbound = new(InboundTrafficPolicy)
	}
	return p.Inbound
}

func (p *PipyConf) NewOutboundTrafficPolicy() *OutboundTrafficPolicy {
	if p.Outbound == nil {
		p.Outbound = new(OutboundTrafficPolicy)
	}
	return p.Outbound
}

func (p *PipyConf) RebalanceOutboundClusters() {
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

func (p *PipyConf) CopyAllowedEndpoints() {
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
			ingressIp := strings.TrimSuffix(string(ipRange), "/32")
			p.AllowedEndpoints[ingressIp] = "Ingress Controller"
		}
	}
}

func (itm *InboundTrafficMatch) AddSourceIPRange(ipRange SourceIPRange) {
	itm.SourceIPRanges = append(itm.SourceIPRanges, ipRange)
}

func (itm *InboundTrafficMatch) AddAllowedEndpoint(address Address, serviceName ServiceName) {
	if itm.AllowedEndpoints == nil {
		itm.AllowedEndpoints = make(AllowedEndpoints)
	}
	if _, exists := itm.AllowedEndpoints[address]; !exists {
		itm.AllowedEndpoints[address] = serviceName
	}
}

func (otm *OutboundTrafficMatch) AddDestinationIPRange(ipRange DestinationIPRange) {
	otm.DestinationIPRanges = append(otm.DestinationIPRanges, ipRange)
}

func (otm *OutboundTrafficMatch) SetServiceIdentity(serviceIdentity identity.ServiceIdentity) {
	otm.ServiceIdentity = serviceIdentity
}

func (otm *OutboundTrafficMatch) SetAllowedEgressTraffic(allowedEgressTraffic bool) {
	otm.AllowedEgressTraffic = allowedEgressTraffic
}

func (tm *TrafficMatch) SetPort(port Port) {
	tm.Port = port
}

func (tm *TrafficMatch) SetProtocol(protocol Protocol) {
	if constants.ProtocolTCPServerFirst == protocol {
		tm.Protocol = constants.ProtocolTCP
	} else {
		tm.Protocol = protocol
	}
}

func (tm *TrafficMatch) AddWeightedCluster(clusterName ClusterName, weight Weight) {
	if tm.TargetClusters == nil {
		tm.TargetClusters = make(WeightedClusters)
	}
	tm.TargetClusters[clusterName] = weight
}

func (tm *TrafficMatch) AddHttpHostPort2Service(hostPort HttpHostPort, ruleName HttpRouteRuleName) {
	if tm.HttpHostPort2Service == nil {
		tm.HttpHostPort2Service = make(HttpHostPort2Service)
	}
	tm.HttpHostPort2Service[hostPort] = ruleName
}

func (tm *TrafficMatch) NewHttpServiceRouteRules(httpRouteRuleName HttpRouteRuleName) *HttpRouteRules {
	if tm.HttpServiceRouteRules == nil {
		tm.HttpServiceRouteRules = make(HttpServiceRouteRules)
	}
	if len(httpRouteRuleName) == 0 {
		return nil
	}
	if rules, exist := tm.HttpServiceRouteRules[httpRouteRuleName]; !exist || rules == nil {
		newCluster := make(HttpRouteRules, 0)
		tm.HttpServiceRouteRules[httpRouteRuleName] = &newCluster
		return &newCluster
	} else {
		return rules
	}
}

func (itp *InboundTrafficPolicy) NewTrafficMatch(port Port) *InboundTrafficMatch {
	if itp.TrafficMatches == nil {
		itp.TrafficMatches = make(InboundTrafficMatches)
	}
	if trafficMatch, exist := itp.TrafficMatches[port]; !exist || trafficMatch == nil {
		trafficMatch = new(InboundTrafficMatch)
		itp.TrafficMatches[port] = trafficMatch
		return trafficMatch
	} else {
		return trafficMatch
	}
}

func (itp *InboundTrafficPolicy) GetTrafficMatch(port Port) *InboundTrafficMatch {
	if itp.TrafficMatches == nil {
		return nil
	}
	if trafficMatch, exist := itp.TrafficMatches[port]; exist {
		return trafficMatch
	} else {
		return nil
	}
}

func (otp *OutboundTrafficPolicy) NewTrafficMatch(port Port) *OutboundTrafficMatch {
	trafficMatch := new(OutboundTrafficMatch)
	if otp.TrafficMatches == nil {
		otp.TrafficMatches = make(OutboundTrafficMatches)
	}
	trafficMatches, _ := otp.TrafficMatches[port]
	trafficMatches = append(trafficMatches, trafficMatch)
	otp.TrafficMatches[port] = trafficMatches
	return trafficMatch
}

func (hrrs *HttpRouteRules) NewHttpServiceRouteRule(pathReg URIPathRegexp) *HttpRouteRule {
	if routeRule, exist := (*hrrs)[pathReg]; !exist || routeRule == nil {
		routeRule = new(HttpRouteRule)
		(*hrrs)[pathReg] = routeRule
		return routeRule
	} else {
		return routeRule
	}
}

func (hrr *HttpRouteRule) AddHeaderMatch(header Header, headerRegexp HeaderRegexp) {
	if hrr.Headers == nil {
		hrr.Headers = make(HeadersMatch)
	}
	hrr.Headers[header] = headerRegexp
}

func (hrr *HttpRouteRule) AddMethodMatch(method Method) {
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

func (hrr *HttpRouteRule) AddWeightedCluster(clusterName ClusterName, weight Weight) {
	if hrr.TargetClusters == nil {
		hrr.TargetClusters = make(WeightedClusters)
	}
	hrr.TargetClusters[clusterName] = weight
}

func (hrr *HttpRouteRule) AddAllowedService(serviceName ServiceName) {
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

func (tp *TrafficPolicy) NewClusterConfigs(clusterName ClusterName) *WeightedEndpoint {
	if tp.ClustersConfigs == nil {
		tp.ClustersConfigs = make(ClustersConfigs)
	}
	if cluster, exist := tp.ClustersConfigs[clusterName]; !exist || cluster == nil {
		newCluster := make(WeightedEndpoint, 0)
		tp.ClustersConfigs[clusterName] = &newCluster
		return &newCluster
	} else {
		return cluster
	}
}

func (we *WeightedEndpoint) AddWeightedEndpoint(
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
