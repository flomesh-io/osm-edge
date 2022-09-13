package injector

import (
	"fmt"
	"strconv"
	"strings"

	configv1alpha2 "github.com/openservicemesh/osm/pkg/apis/config/v1alpha2"

	"github.com/openservicemesh/osm/pkg/constants"
)

// iptablesOutboundStaticRules is the list of iptables rules related to outbound traffic interception and redirection
var iptablesOutboundStaticRules = []string{
	// Redirects outbound TCP traffic hitting OSM_PROXY_OUT_REDIRECT chain to Sidecar's tcp outbound listener port
	fmt.Sprintf("-A OSM_PROXY_OUT_REDIRECT -p tcp -j REDIRECT --to-port %d", constants.SidecarTCPOutboundListenerPort),

	// Redirects outbound UDP traffic hitting OSM_PROXY_OUT_REDIRECT chain to Sidecar's udp outbound listener port
	fmt.Sprintf("-A OSM_PROXY_OUT_REDIRECT -p udp -j REDIRECT --to-port %d", constants.SidecarUDPOutboundListenerPort),

	// Traffic to the Proxy Admin port flows to the Proxy -- not redirected
	fmt.Sprintf("-A OSM_PROXY_OUT_REDIRECT -p tcp --dport %d -j ACCEPT", constants.SidecarAdminPort),

	// For outbound TCP traffic jump from OUTPUT chain to OSM_PROXY_OUTBOUND chain
	"-A OUTPUT -p tcp -j OSM_PROXY_OUTBOUND",

	// For outbound UDP traffic jump from OUTPUT chain to OSM_PROXY_OUTBOUND chain
	"-A OUTPUT -p udp -j OSM_PROXY_OUTBOUND",

	// Outbound traffic from Sidecar to the local app over the loopback interface should jump to the inbound proxy redirect chain.
	// So when an app directs traffic to itself via the k8s service, traffic flows as follows:
	// app -> local sidecar's outbound listener -> iptables -> local sidecar's inbound listener -> app
	fmt.Sprintf("-A OSM_PROXY_OUTBOUND -o lo ! -d 127.0.0.1/32 -m owner --uid-owner %d -j OSM_PROXY_IN_REDIRECT", constants.SidecarUID),

	// Outbound traffic from the app to itself over the loopback interface is not be redirected via the proxy.
	// E.g. when app sends traffic to itself via the pod IP.
	fmt.Sprintf("-A OSM_PROXY_OUTBOUND -o lo -m owner ! --uid-owner %d -j RETURN", constants.SidecarUID),

	// Don't redirect Sidecar traffic back to itself, return it to the next chain for processing
	fmt.Sprintf("-A OSM_PROXY_OUTBOUND -m owner --uid-owner %d -j RETURN", constants.SidecarUID),

	// Skip localhost traffic, doesn't need to be routed via the proxy
	"-A OSM_PROXY_OUTBOUND -d 127.0.0.1/32 -j RETURN",
}

// iptablesInboundStaticRules is the list of iptables rules related to inbound traffic interception and redirection
var iptablesInboundStaticRules = []string{
	// Redirects inbound TCP traffic hitting the OSM_PROXY_IN_REDIRECT chain to Sidecar's tcp inbound listener port
	fmt.Sprintf("-A OSM_PROXY_IN_REDIRECT -p tcp -j REDIRECT --to-port %d", constants.SidecarTCPInboundListenerPort),

	// Redirects inbound UDP traffic hitting the OSM_PROXY_IN_REDIRECT chain to Sidecar's udp inbound listener port
	fmt.Sprintf("-A OSM_PROXY_IN_REDIRECT -p udp -j REDIRECT --to-port %d", constants.SidecarUDPInboundListenerPort),

	// For tcp inbound traffic jump from PREROUTING chain to OSM_PROXY_INBOUND chain
	"-A PREROUTING -p tcp -j OSM_PROXY_INBOUND",

	// For udp inbound traffic jump from PREROUTING chain to OSM_PROXY_INBOUND chain
	"-A PREROUTING -p udp -j OSM_PROXY_INBOUND",

	// Skip metrics query traffic being directed to Sidecar's inbound prometheus listener port
	fmt.Sprintf("-A OSM_PROXY_INBOUND -p tcp --dport %d -j RETURN", constants.SidecarPrometheusInboundListenerPort),

	// Skip inbound health probes; These ports will be explicitly handled by listeners configured on the
	// Sidecar proxy IF any health probes have been configured in the Pod Spec.
	// TODO(draychev): Do not add these if no health probes have been defined (https://github.com/openservicemesh/osm/issues/2243)
	fmt.Sprintf("-A OSM_PROXY_INBOUND -p tcp --dport %d -j RETURN", constants.LivenessProbePort),
	fmt.Sprintf("-A OSM_PROXY_INBOUND -p tcp --dport %d -j RETURN", constants.ReadinessProbePort),
	fmt.Sprintf("-A OSM_PROXY_INBOUND -p tcp --dport %d -j RETURN", constants.StartupProbePort),
	// Skip inbound health probes (originally TCPSocket health probes); requests handled by osm-healthcheck
	fmt.Sprintf("-A OSM_PROXY_INBOUND -p tcp --dport %d -j RETURN", constants.HealthcheckPort),

	// Redirect remaining inbound traffic to Sidecar
	"-A OSM_PROXY_INBOUND -p tcp -j OSM_PROXY_IN_REDIRECT",
}

// GenerateIptablesCommands generates a list of iptables commands to set up sidecar interception and redirection
func GenerateIptablesCommands(proxyMode configv1alpha2.LocalProxyMode, outboundIPRangeExclusionList []string, outboundIPRangeInclusionList []string, outboundPortExclusionList []int, inboundPortExclusionList []int, outboundUDPPortExclusionList []int, inboundUDPPortExclusionList []int, networkInterfaceExclusionList []string) string {
	var rules strings.Builder

	fmt.Fprintln(&rules, `# OSM sidecar interception rules
*nat
:OSM_PROXY_INBOUND - [0:0]
:OSM_PROXY_IN_REDIRECT - [0:0]
:OSM_PROXY_OUTBOUND - [0:0]
:OSM_PROXY_OUT_REDIRECT - [0:0]`)
	var cmds []string

	// 1. Create inbound rules
	cmds = append(cmds, iptablesInboundStaticRules...)

	// Ignore inbound traffic on specified interfaces
	for _, iface := range networkInterfaceExclusionList {
		// *Note: it is important to use the insert option '-I' instead of the append option '-A' to ensure the
		// exclusion of traffic to the network interface happens before the rule that redirects traffic to the proxy
		cmds = append(cmds, fmt.Sprintf("-I OSM_PROXY_INBOUND -i %s -j RETURN", iface))
	}

	// 2. Create dynamic inbound ports exclusion rules
	if len(inboundPortExclusionList) > 0 {
		var portExclusionListStr []string
		for _, port := range inboundPortExclusionList {
			portExclusionListStr = append(portExclusionListStr, strconv.Itoa(port))
		}
		inboundPortsToExclude := strings.Join(portExclusionListStr, ",")
		rule := fmt.Sprintf("-I OSM_PROXY_INBOUND -p tcp --match multiport --dports %s -j RETURN", inboundPortsToExclude)
		cmds = append(cmds, rule)
	}

	if len(inboundUDPPortExclusionList) > 0 {
		var portExclusionListStr []string
		for _, port := range inboundUDPPortExclusionList {
			portExclusionListStr = append(portExclusionListStr, strconv.Itoa(port))
		}
		inboundPortsToExclude := strings.Join(portExclusionListStr, ",")
		rule := fmt.Sprintf("-I OSM_PROXY_INBOUND -p udp --match multiport --dports %s -j RETURN", inboundPortsToExclude)
		cmds = append(cmds, rule)
	}

	// 3. Create outbound rules
	cmds = append(cmds, iptablesOutboundStaticRules...)

	if proxyMode == configv1alpha2.LocalProxyModePodIP {
		// For sidecar -> local service container proxying, send traffic to pod IP instead of localhost
		// *Note: it is important to use the insert option '-I' instead of the append option '-A' to ensure the
		// DNAT to the pod ip for sidecar -> localhost traffic happens before the rule that redirects traffic to the proxy
		cmds = append(cmds, fmt.Sprintf("-I OUTPUT -p tcp -o lo -d 127.0.0.1/32 -m owner --uid-owner %d -j DNAT --to-destination $POD_IP", constants.SidecarUID))
	}

	// Ignore outbound traffic in specified interfaces
	for _, iface := range networkInterfaceExclusionList {
		cmds = append(cmds, fmt.Sprintf("-A OSM_PROXY_OUTBOUND -o %s -j RETURN", iface))
	}

	//
	// Create outbound exclusion and inclusion rules.
	// *Note: exclusion rules must be applied before inclusions as order matters
	//

	// 4. Create dynamic outbound IP range exclusion rules
	for _, cidr := range outboundIPRangeExclusionList {
		// *Note: it is important to use the insert option '-I' instead of the append option '-A' to ensure the exclusion
		// rules take precedence over the static redirection rules. Iptables rules are evaluated in order.
		rule := fmt.Sprintf("-A OSM_PROXY_OUTBOUND -d %s -j RETURN", cidr)
		cmds = append(cmds, rule)
	}

	// 5. Create dynamic outbound ports exclusion rules
	if len(outboundPortExclusionList) > 0 {
		var portExclusionListStr []string
		for _, port := range outboundPortExclusionList {
			portExclusionListStr = append(portExclusionListStr, strconv.Itoa(port))
		}
		outboundPortsToExclude := strings.Join(portExclusionListStr, ",")
		rule := fmt.Sprintf("-A OSM_PROXY_OUTBOUND -p tcp --match multiport --dports %s -j RETURN", outboundPortsToExclude)
		cmds = append(cmds, rule)
	}

	if len(outboundUDPPortExclusionList) > 0 {
		var portExclusionListStr []string
		for _, port := range outboundUDPPortExclusionList {
			portExclusionListStr = append(portExclusionListStr, strconv.Itoa(port))
		}
		outboundPortsToExclude := strings.Join(portExclusionListStr, ",")
		rule := fmt.Sprintf("-A OSM_PROXY_OUTBOUND -p udp --match multiport --dports %s -j RETURN", outboundPortsToExclude)
		cmds = append(cmds, rule)
	}

	// 6. Create dynamic outbound IP range inclusion rules
	if len(outboundIPRangeInclusionList) > 0 {
		// Redirect specified IP ranges to the proxy
		for _, cidr := range outboundIPRangeInclusionList {
			rule := fmt.Sprintf("-A OSM_PROXY_OUTBOUND -d %s -j OSM_PROXY_OUT_REDIRECT", cidr)
			cmds = append(cmds, rule)
		}
		// Remaining traffic not belonging to specified inclusion IP ranges are not redirected
		cmds = append(cmds, "-A OSM_PROXY_OUTBOUND -j RETURN")
	} else {
		// Redirect remaining outbound traffic to the proxy
		cmds = append(cmds, "-A OSM_PROXY_OUTBOUND -j OSM_PROXY_OUT_REDIRECT")
	}

	for _, rule := range cmds {
		fmt.Fprintln(&rules, rule)
	}

	fmt.Fprint(&rules, "COMMIT")

	cmd := fmt.Sprintf(`iptables-restore --noflush <<EOF
%s
EOF
`, rules.String())

	return cmd
}
