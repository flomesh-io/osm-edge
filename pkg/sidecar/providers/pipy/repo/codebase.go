package repo

import (
	_ "embed"
)

//go:embed codebase/outbound-tcp-load-balance.js
var codebaseOutboundTCPLoadBalanceJs []byte

//go:embed codebase/logging-init.js
var codebaseLoggingInitJs []byte

//go:embed codebase/utils.js
var codebaseUtilsJs []byte

//go:embed codebase/tracing-init.js
var codebaseTracingInitJs []byte

//go:embed codebase/metrics-http.js
var codebaseMetricsHTTPJs []byte

//go:embed codebase/config.js
var codebaseConfigJs []byte

//go:embed codebase/tracing.js
var codebaseTracingJs []byte

//go:embed codebase/metrics-init.js
var codebaseMetricsInitJs []byte

//go:embed codebase/logging.js
var codebaseLoggingJs []byte

//go:embed codebase/metrics-tcp.js
var codebaseMetricsTCPJs []byte

//go:embed codebase/inbound-throttle.js
var codebaseInboundThrottleJs []byte

//go:embed codebase/main.js
var codebaseMainJs []byte

//go:embed codebase/breaker.js
var codebaseBreakerJs []byte

//go:embed codebase/inbound-mux-http.js
var codebaseInboundMuxHTTPJs []byte

//go:embed codebase/outbound-mux-http.js
var codebaseOutboundMuxHTTPJs []byte

//go:embed codebase/outbound-http-routing.js
var codebaseOutboundHTTPRoutingJs []byte

//go:embed codebase/inbound-demux-http.js
var codebaseInboundDemuxHTTPJs []byte

//go:embed codebase/inbound-tls-termination.js
var codebaseInboundTLSTerminationJs []byte

//go:embed codebase/outbound-breaker.js
var codebaseOutboundBreakerJs []byte

//go:embed codebase/inbound-proxy-tcp.js
var codebaseInboundProxyTCPJs []byte

//go:embed codebase/stats.js
var codebaseStatsJs []byte

//go:embed codebase/outbound-classifier.js
var codebaseOutboundClassifierJs []byte

//go:embed codebase/inbound-http-routing.js
var codebaseInboundHTTPRoutingJs []byte

//go:embed codebase/outbound-proxy-tcp.js
var codebaseOutboundProxyTCPJs []byte

//go:embed codebase/codes.js
var codebaseCodesJs []byte

//go:embed codebase/inbound-classifier.js
var codebaseInboundClassifierJs []byte

//go:embed codebase/inbound-tcp-load-balance.js
var codebaseInboundTCPLoadBalanceJs []byte

//go:embed codebase/outbound-demux-http.js
var codebaseOutboundDemuxHTTPJs []byte

//go:embed codebase/config.json
var codebaseConfigJSON []byte
