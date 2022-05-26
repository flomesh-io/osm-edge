package e2e

import (
	"runtime"
	"testing"

	"helm.sh/helm/v3/pkg/chart"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ginkgo e2e tests")
}

func ensureNodeSelector(chartRequested *chart.Chart) {
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
