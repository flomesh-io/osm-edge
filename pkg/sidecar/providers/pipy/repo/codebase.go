package repo

import (
	_ "embed"
)

//go:embed codebase/main.js
var codebaseMainJS []byte

//go:embed codebase/codes.js
var codebaseCodesJS []byte

//go:embed codebase/breaker.js
var codebaseBreakerJS []byte

//go:embed codebase/config.js
var codebaseConfigJS []byte

//go:embed codebase/metrics.js
var codebaseMetricsJS []byte

//go:embed codebase/gather.js
var codebaseGatherJS []byte

//go:embed codebase/stats.js
var codebaseStatsJS []byte

//go:embed codebase/config.json
var codebaseConfigJSON []byte

//go:embed codebase/inbound-proxy-tcp.js
var codebaseInboundProxyTCPJS []byte

//go:embed codebase/inbound-recv-http.js
var codebaseInboundRecvHTTPJS []byte

//go:embed codebase/inbound-recv-tcp.js
var codebaseInboundRecvTCPJS []byte

//go:embed codebase/inbound-throttle.js
var codebaseInboundThrottleJS []byte

//go:embed codebase/outbound-breaker.js
var codebaseOutboundBreakerJS []byte

//go:embed codebase/outbound-mux-http.js
var codebaseOutboundMuxHTTPJS []byte

//go:embed codebase/outbound-proxy-tcp.js
var codebaseOutboundProxyTCPJS []byte

//go:embed codebase/outbound-recv-http.js
var codebaseOutboundRecvHTTPJS []byte

//go:embed codebase/dns-main.js
var codebaseDNSMainJS []byte
