package repo

import (
	_ "embed"
)

//go:embed codebase_main.js
var codebaseMainJS []byte

//go:embed codebase_codes.js
var codebaseCodesJS []byte

//go:embed codebase_breaker.js
var codebaseBreakerJS []byte

//go:embed codebase_config.js
var codebaseConfigJS []byte

//go:embed codebase_metrics.js
var codebaseMetricsJS []byte

//go:embed codebase_gather.js
var codebaseGatherJS []byte

//go:embed codebase_stats.js
var codebaseStatsJS []byte

//go:embed codebase_pipy.json
var codebasePipyJSON []byte

//go:embed codebase_inbound-proxy-tcp.js
var codebaseInboundProxyTCPJS []byte

//go:embed codebase_inbound-recv-http.js
var codebaseInboundRecvHTTPJS []byte

//go:embed codebase_inbound-recv-tcp.js
var codebaseInboundRecvTCPJS []byte

//go:embed codebase_inbound-throttle.js
var codebaseInboundThrottleJS []byte

//go:embed codebase_outbound-breaker.js
var codebaseOutboundBreakerJS []byte

//go:embed codebase_outbound-mux-http.js
var codebaseOutboundMuxHTTPJS []byte

//go:embed codebase_outbound-proxy-tcp.js
var codebaseOutboundProxyTCPJS []byte

//go:embed codebase_outbound-recv-http.js
var codebaseOutboundRecvHTTPJS []byte
