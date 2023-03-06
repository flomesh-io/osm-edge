// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package helpers implements ebpf helpers.
package helpers

import (
	"fmt"

	"github.com/cilium/ebpf"

	"github.com/openservicemesh/osm/pkg/cni/config"
)

var (
	podFibMap *ebpf.Map
	natFibMap *ebpf.Map
)

// InitLoadPinnedMap init, load and pinned mapsß
func InitLoadPinnedMap() error {
	var err error
	podFibMap, err = ebpf.LoadPinnedMap(config.OsmPodFibEbpfMap, &ebpf.LoadPinOptions{})
	if err != nil {
		return fmt.Errorf("load map[%s] error: %v", config.OsmPodFibEbpfMap, err)
	}
	natFibMap, err = ebpf.LoadPinnedMap(config.OsmNatFibEbpfMap, &ebpf.LoadPinOptions{})
	if err != nil {
		return fmt.Errorf("load map[%s] error: %v", err, config.OsmNatFibEbpfMap)
	}
	return nil
}

// GetPodFibMap returns pod fib map
func GetPodFibMap() *ebpf.Map {
	if podFibMap == nil {
		_ = InitLoadPinnedMap()
	}
	return podFibMap
}

// GetNatFibMap returns nat fib map
func GetNatFibMap() *ebpf.Map {
	if natFibMap == nil {
		_ = InitLoadPinnedMap()
	}
	return natFibMap
}
