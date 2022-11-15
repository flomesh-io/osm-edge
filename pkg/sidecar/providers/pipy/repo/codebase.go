package repo

import (
	_ "embed"
)

//go:embed codebase_outbound-tcp-load-balance.js
var codebaseOutboundTCPLoadBalanceJs []byte

//go:embed codebase_logging-init.js
var codebaseLoggingInitJs []byte

//go:embed codebase_utils.js
var codebaseUtilsJs []byte

//go:embed codebase_tracing-init.js
var codebaseTracingInitJs []byte

//go:embed codebase_metrics-http.js
var codebaseMetricsHTTPJs []byte

//go:embed codebase_config.js
var codebaseConfigJs []byte

//go:embed codebase_tracing.js
var codebaseTracingJs []byte

//go:embed codebase_metrics-init.js
var codebaseMetricsInitJs []byte

//go:embed codebase_logging.js
var codebaseLoggingJs []byte

//go:embed codebase_metrics-tcp.js
var codebaseMetricsTCPJs []byte

//go:embed codebase_inbound-throttle.js
var codebaseInboundThrottleJs []byte

//go:embed codebase_main.js
var codebaseMainJs []byte

//go:embed codebase_breaker.js
var codebaseBreakerJs []byte

//go:embed codebase_inbound-mux-http.js
var codebaseInboundMuxHTTPJs []byte

//go:embed codebase_outbound-mux-http.js
var codebaseOutboundMuxHTTPJs []byte

//go:embed codebase_outbound-http-routing.js
var codebaseOutboundHTTPRoutingJs []byte

//go:embed codebase_inbound-demux-http.js
var codebaseInboundDemuxHTTPJs []byte

//go:embed codebase_inbound-tls-termination.js
var codebaseInboundTLSTerminationJs []byte

//go:embed codebase_outbound-breaker.js
var codebaseOutboundBreakerJs []byte

//go:embed codebase_inbound-proxy-tcp.js
var codebaseInboundProxyTCPJs []byte

//go:embed codebase_stats.js
var codebaseStatsJs []byte

//go:embed codebase_outbound-classifier.js
var codebaseOutboundClassifierJs []byte

//go:embed codebase_inbound-http-routing.js
var codebaseInboundHTTPRoutingJs []byte

//go:embed codebase_outbound-proxy-tcp.js
var codebaseOutboundProxyTCPJs []byte

//go:embed codebase_codes.js
var codebaseCodesJs []byte

//go:embed codebase_inbound-classifier.js
var codebaseInboundClassifierJs []byte

//go:embed codebase_inbound-tcp-load-balance.js
var codebaseInboundTCPLoadBalanceJs []byte

//go:embed codebase_outbound-demux-http.js
var codebaseOutboundDemuxHTTPJs []byte

//go:embed codebase_config.json
var codebaseConfigJSON []byte
