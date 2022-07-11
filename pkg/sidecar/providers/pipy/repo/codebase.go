package repo

import (
	_ "embed"
)

//go:embed codebase_main.js
var codebaseMainJS []byte

//go:embed codebase_config.js
var codebaseConfigJS []byte

//go:embed codebase_metrics.js
var codebaseMetricsJS []byte

//go:embed codebase_pipy.json
var codebasePipyJSON []byte
