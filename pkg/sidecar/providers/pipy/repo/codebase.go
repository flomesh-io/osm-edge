package repo

import (
	_ "embed"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/client"
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

var osmCodebaseItems = []client.BatchItem{
	{
		Filename: "main.js",
		Content:  codebaseMainJS,
	},
	{
		Filename: "config.js",
		Content:  codebaseConfigJS,
	},
	{
		Filename: "metrics.js",
		Content:  codebaseMetricsJS,
	},
	{
		Filename: osmCodebaseConfig,
		Content:  codebaseConfigJSON,
	},
	{
		Filename: "codes.js",
		Content:  codebaseCodesJS,
	},
	{
		Filename: "breaker.js",
		Content:  codebaseBreakerJS,
	},
	{
		Filename: "gather.js",
		Content:  codebaseGatherJS,
	},
	{
		Filename: "stats.js",
		Content:  codebaseStatsJS,
	},
	{
		Filename: "inbound-proxy-tcp.js",
		Content:  codebaseInboundProxyTCPJS,
	},
	{
		Filename: "inbound-recv-http.js",
		Content:  codebaseInboundRecvHTTPJS,
	},
	{
		Filename: "inbound-recv-tcp.js",
		Content:  codebaseInboundRecvTCPJS,
	},
	{
		Filename: "inbound-throttle.js",
		Content:  codebaseInboundThrottleJS,
	},
	{
		Filename: "outbound-breaker.js",
		Content:  codebaseOutboundBreakerJS,
	},
	{
		Filename: "outbound-mux-http.js",
		Content:  codebaseOutboundMuxHTTPJS,
	},
	{
		Filename: "outbound-proxy-tcp.js",
		Content:  codebaseOutboundProxyTCPJS,
	},
	{
		Filename: "outbound-recv-http.js",
		Content:  codebaseOutboundRecvHTTPJS,
	},
	{
		Filename: "dns-main.js",
		Content:  codebaseDNSMainJS,
	},
}
