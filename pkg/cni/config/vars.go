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

package config

var (
	// Debug indicates debug feature of/off
	Debug = false
	// Skip indicates skip feature of/off
	Skip = false
	// DisableWatcher indicates DisableWatcher feature of/off
	DisableWatcher = false
	// EnableCNI indicates CNI feature enable/disable
	EnableCNI = false
	// IsKind indicates Kubernetes running in Docker
	IsKind = false
	// HostProc defines HostProc volume
	HostProc string
	// CNIBinDir defines CNIBIN volume
	CNIBinDir string
	// CNIConfigDir defines CNIConfig volume
	CNIConfigDir string
	// HostVarRun defines HostVar volume
	HostVarRun string
	// KubeConfig defines kube config
	KubeConfig string
	// Context defines kube context
	Context string
	// EnableHotRestart indicates HotRestart feature enable/disable
	EnableHotRestart = false
)
