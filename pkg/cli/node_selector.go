package cli

import (
	"runtime"

	"helm.sh/helm/v3/pkg/chart"
)

// EnsureNodeSelector ensure nodeSelector
func EnsureNodeSelector(chartRequested *chart.Chart) {
	if chartRequested == nil || chartRequested.Values == nil {
		return
	}
	if osmObj, ok := chartRequested.Values["osm"]; ok {
		osmMap := osmObj.(map[string]interface{})
		nodeSelectorObj, exists := osmMap["nodeSelector"]
		if !exists || nodeSelectorObj == nil {
			nodeSelectorObj = make(map[string]interface{})
			osmMap["nodeSelector"] = nodeSelectorObj
		}
		nodeSelectorMap := nodeSelectorObj.(map[string]interface{})
		nodeSelectorMap["arch"] = runtime.GOARCH
	}
}
