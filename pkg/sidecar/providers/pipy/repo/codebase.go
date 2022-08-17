package repo

import (
	_ "embed"
)

//go:embed codebase_main.js
var codebaseMainJS []byte

//go:embed codebase_f_codes.js
var codebaseCodesJS []byte

//go:embed codebase_f_breaker.js
var codebaseBreakerJS []byte

//go:embed codebase_f_config.js
var codebaseConfigJS []byte

//go:embed codebase_f_metrics.js
var codebaseMetricsJS []byte

//go:embed codebase_p_gather.js
var codebaseGatherJS []byte

//go:embed codebase_p_stats.js
var codebaseStatsJS []byte

//go:embed codebase_pipy.json
var codebasePipyJSON []byte
