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

// Package config defines the constants that are used by multiple other packages within OSM.
package config

const (
	// CNICreatePodURL is the route for cni plugin for creating pod
	CNICreatePodURL = "/v1/cni/create-pod"
	// CNIDeletePodURL is the route for cni plugin for deleting pod
	CNIDeletePodURL = "/v1/cni/delete-pod"
	// CNITransferFdStartURL is the route for cni plugin for transfer fd
	CNITransferFdStartURL = "/v1/cni/transfer-fd"

	// OsmPodFibEbpfMap is the mount point of osm_pod_fib map
	OsmPodFibEbpfMap = "/sys/fs/bpf/tc/globals/osm_pod_fib"
	// OsmNatFibEbpfMap is the mount point of osm_nat_fib map
	OsmNatFibEbpfMap = "/sys/fs/bpf/tc/globals/osm_nat_fib"
)
