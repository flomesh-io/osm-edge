package repo

import (
	_ "embed"
)

//go:embed codebase/config.js
var codebaseConfigJs []byte

//go:embed codebase/main.js
var codebaseMainJs []byte

//go:embed codebase/modules/inbound-http-default.js
var codebaseModulesInboundHTTPDefaultJs []byte

//go:embed codebase/modules/inbound-http-load-balancing.js
var codebaseModulesInboundHTTPLoadBalancingJs []byte

//go:embed codebase/modules/inbound-http-routing.js
var codebaseModulesInboundHTTPRoutingJs []byte

//go:embed codebase/modules/inbound-logging-http.js
var codebaseModulesInboundLoggingHTTPJs []byte

//go:embed codebase/modules/inbound-main.js
var codebaseModulesInboundMainJs []byte

//go:embed codebase/modules/inbound-make-connection.js
var codebaseModulesInboundMakeConnectionJs []byte

//go:embed codebase/modules/inbound-metrics-http.js
var codebaseModulesInboundMetricsHTTPJs []byte

//go:embed codebase/modules/inbound-metrics-tcp.js
var codebaseModulesInboundMetricsTCPJs []byte

//go:embed codebase/modules/inbound-tcp-default.js
var codebaseModulesInboundTCPDefaultJs []byte

//go:embed codebase/modules/inbound-tcp-load-balancing.js
var codebaseModulesInboundTCPLoadBalancingJs []byte

//go:embed codebase/modules/inbound-throttle-route.js
var codebaseModulesInboundThrottleRouteJs []byte

//go:embed codebase/modules/inbound-throttle-service.js
var codebaseModulesInboundThrottleServiceJs []byte

//go:embed codebase/modules/inbound-tls-termination.js
var codebaseModulesInboundTLSTerminationJs []byte

//go:embed codebase/modules/inbound-tracing-http.js
var codebaseModulesInboundTracingHTTPJs []byte

//go:embed codebase/modules/logging.js
var codebaseModulesLoggingJs []byte

//go:embed codebase/modules/metrics.js
var codebaseModulesMetricsJs []byte

//go:embed codebase/modules/outbound-circuit-breaker.js
var codebaseModulesOutboundCircuitBreakerJs []byte

//go:embed codebase/modules/outbound-http-default.js
var codebaseModulesOutboundHTTPDefaultJs []byte

//go:embed codebase/modules/outbound-http-load-balancing.js
var codebaseModulesOutboundHTTPLoadBalancingJs []byte

//go:embed codebase/modules/outbound-http-routing.js
var codebaseModulesOutboundHTTPRoutingJs []byte

//go:embed codebase/modules/outbound-logging-http.js
var codebaseModulesOutboundLoggingHTTPJs []byte

//go:embed codebase/modules/outbound-main.js
var codebaseModulesOutboundMainJs []byte

//go:embed codebase/modules/outbound-metrics-http.js
var codebaseModulesOutboundMetricsHTTPJs []byte

//go:embed codebase/modules/outbound-metrics-tcp.js
var codebaseModulesOutboundMetricsTCPJs []byte

//go:embed codebase/modules/outbound-tcp-default.js
var codebaseModulesOutboundTCPDefaultJs []byte

//go:embed codebase/modules/outbound-tcp-load-balancing.js
var codebaseModulesOutboundTCPLoadBalancingJs []byte

//go:embed codebase/modules/outbound-tls-initiation.js
var codebaseModulesOutboundTLSInitiationJs []byte

//go:embed codebase/modules/outbound-tracing-http.js
var codebaseModulesOutboundTracingHTTPJs []byte

//go:embed codebase/modules/tracing.js
var codebaseModulesTracingJs []byte

//go:embed codebase/probes.js
var codebaseProbesJs []byte

//go:embed codebase/stats.js
var codebaseStatsJs []byte

//go:embed codebase/config.json
var codebaseConfigJSON []byte
