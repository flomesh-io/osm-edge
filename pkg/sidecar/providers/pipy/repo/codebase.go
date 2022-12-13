package repo

import (
	_ "embed"

	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/client"
)

//go:embed codebase/modules/outbound-http-routing.js
var codebaseModulesOutboundHTTPRoutingJs []byte

//go:embed codebase/modules/outbound-main.js
var codebaseModulesOutboundMainJs []byte

//go:embed codebase/modules/outbound-tcp-load-balancing.js
var codebaseModulesOutboundTCPLoadBalancingJs []byte

//go:embed codebase/modules/outbound-http-load-balancing.js
var codebaseModulesOutboundHTTPLoadBalancingJs []byte

//go:embed codebase/modules/inbound-http-routing.js
var codebaseModulesInboundHTTPRoutingJs []byte

//go:embed codebase/modules/inbound-http-load-balancing.js
var codebaseModulesInboundHTTPLoadBalancingJs []byte

//go:embed codebase/modules/inbound-main.js
var codebaseModulesInboundMainJs []byte

//go:embed codebase/modules/outbound-http-default.js
var codebaseModulesOutboundHTTPDefaultJs []byte

//go:embed codebase/modules/inbound-tcp-load-balancing.js
var codebaseModulesInboundTCPLoadBalancingJs []byte

//go:embed codebase/modules/inbound-http-default.js
var codebaseModulesInboundHTTPDefaultJs []byte

//go:embed codebase/modules/inbound-tls-termination.js
var codebaseModulesInboundTLSTerminationJs []byte

//go:embed codebase/modules/outbound-tcp-default.js
var codebaseModulesOutboundTCPDefaultJs []byte

//go:embed codebase/modules/inbound-tcp-default.js
var codebaseModulesInboundTCPDefaultJs []byte

//go:embed codebase/modules/inbound-throttle-service.js
var codebaseModulesInboundThrottleServiceJs []byte

//go:embed codebase/modules/inbound-throttle-route.js
var codebaseModulesInboundThrottleRouteJs []byte

//go:embed codebase/modules/outbound-circuit-breaker.js
var codebaseModulesOutboundCircuitBreakerJS []byte

//go:embed codebase/modules/outbound-tls-initiation.js
var codebaseModulesOutboundTLSInitiationJS []byte

//go:embed codebase/modules/inbound-metrics-http.js
var codebaseModulesInboundMetricsHTTPJS []byte

//go:embed codebase/modules/outbound-metrics-http.js
var codebaseModulesOutboundMetricsHTTPJS []byte

//go:embed codebase/modules/metrics.js
var codebaseModulesMetricsJS []byte

//go:embed codebase/main.js
var codebaseMainJs []byte

//go:embed codebase/probes.js
var codebaseProbesJs []byte

//go:embed codebase/stats.js
var codebaseStatsJs []byte

//go:embed codebase/config.json
var codebaseConfigJSON []byte

//go:embed codebase/config.js
var codebaseConfigJs []byte

var osmCodebaseItems = []client.BatchItem{
	{Filename: "modules/outbound-http-routing.js", Content: codebaseModulesOutboundHTTPRoutingJs},
	{Filename: "modules/outbound-main.js", Content: codebaseModulesOutboundMainJs},
	{Filename: "modules/outbound-tcp-load-balancing.js", Content: codebaseModulesOutboundTCPLoadBalancingJs},
	{Filename: "modules/outbound-http-load-balancing.js", Content: codebaseModulesOutboundHTTPLoadBalancingJs},
	{Filename: "modules/inbound-http-routing.js", Content: codebaseModulesInboundHTTPRoutingJs},
	{Filename: "modules/inbound-http-load-balancing.js", Content: codebaseModulesInboundHTTPLoadBalancingJs},
	{Filename: "modules/inbound-main.js", Content: codebaseModulesInboundMainJs},
	{Filename: "modules/outbound-http-default.js", Content: codebaseModulesOutboundHTTPDefaultJs},
	{Filename: "modules/inbound-tcp-load-balancing.js", Content: codebaseModulesInboundTCPLoadBalancingJs},
	{Filename: "modules/inbound-http-default.js", Content: codebaseModulesInboundHTTPDefaultJs},
	{Filename: "modules/inbound-tls-termination.js", Content: codebaseModulesInboundTLSTerminationJs},
	{Filename: "modules/outbound-tcp-default.js", Content: codebaseModulesOutboundTCPDefaultJs},
	{Filename: "modules/inbound-tcp-default.js", Content: codebaseModulesInboundTCPDefaultJs},
	{Filename: "modules/inbound-throttle-service.js", Content: codebaseModulesInboundThrottleServiceJs},
	{Filename: "modules/inbound-throttle-route.js", Content: codebaseModulesInboundThrottleRouteJs},
	{Filename: "modules/outbound-circuit-breaker.js", Content: codebaseModulesOutboundCircuitBreakerJS},
	{Filename: "modules/outbound-tls-initiation.js", Content: codebaseModulesOutboundTLSInitiationJS},
	{Filename: "modules/inbound-metrics-http.js", Content: codebaseModulesInboundMetricsHTTPJS},
	{Filename: "modules/outbound-metrics-http.js", Content: codebaseModulesOutboundMetricsHTTPJS},
	{Filename: "modules/metrics.js", Content: codebaseModulesMetricsJS},
	{Filename: "main.js", Content: codebaseMainJs},
	{Filename: "probes.js", Content: codebaseProbesJs},
	{Filename: "stats.js", Content: codebaseStatsJs},
	{Filename: "config.js", Content: codebaseConfigJs},
	{Filename: osmCodebaseConfig, Content: codebaseConfigJSON},
}
