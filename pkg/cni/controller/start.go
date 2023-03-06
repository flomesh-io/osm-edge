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

package controller

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/openservicemesh/osm/pkg/cni/config"
	"github.com/openservicemesh/osm/pkg/cni/kube"
)

var (
	disableWatch = false
)

// Run start to run controller to watch
func Run(disableWatcher, skip bool, cniReady chan struct{}, stop chan struct{}) error {
	var err error
	var client kubernetes.Interface

	// create and check start up configuration
	err = NewOptions()
	if err != nil {
		return fmt.Errorf("create options error: %v", err)
	}

	// get default kubernetes client
	client, err = kube.GetKubernetesClientWithFile(config.KubeConfig, config.Context)
	if err != nil {
		return fmt.Errorf("create client error: %v", err)
	}

	disableWatch = disableWatcher

	// run local ip controller
	if err = runLocalPodController(skip, client, stop); err != nil {
		return fmt.Errorf("run local ip controller error: %v", err)
	}

	return nil
}
