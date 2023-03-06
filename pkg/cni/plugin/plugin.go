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

package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	cniv1 "github.com/containernetworking/cni/pkg/types/100"
	log "github.com/sirupsen/logrus"

	"github.com/openservicemesh/osm/pkg/cni/config"
)

const (
	// SidecarInjectionAnnotation is the annotation used for sidecar injection
	SidecarInjectionAnnotation = "openservicemesh.io/sidecar-injection"
)

// K8sArgs is the valid CNI_ARGS used for Kubernetes
// The field names need to match exact keys in kubelet args for unmarshalling
type K8sArgs struct {
	types.CommonArgs
	IP                         net.IP
	K8S_POD_NAME               types.UnmarshallableString // nolint: revive, stylecheck
	K8S_POD_NAMESPACE          types.UnmarshallableString // nolint: revive, stylecheck
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString // nolint: revive, stylecheck
}

type podInfo struct {
	Containers        []string
	InitContainers    map[string]struct{}
	Labels            map[string]string
	Annotations       map[string]string
	ProxyEnvironments map[string]string
}

func ignore(conf *Config, k8sArgs *K8sArgs) bool {
	ns := string(k8sArgs.K8S_POD_NAMESPACE)
	name := string(k8sArgs.K8S_POD_NAME)
	if ns != "" && name != "" {
		for _, excludeNs := range conf.Kubernetes.ExcludeNamespaces {
			if ns == excludeNs {
				log.Infof("Pod %s/%s excluded", ns, name)
				return true
			}
		}
		client, err := newKubeClient(*conf)
		if err != nil {
			log.Error(err)
			return true
		}
		pi := &podInfo{}
		for attempt := 1; attempt <= podRetrievalMaxRetries; attempt++ {
			pi, err = getKubePodInfo(client, name, ns)
			if err == nil {
				break
			}
			log.Debugf("Failed to get %s/%s pod info: %v", ns, name, err)
			time.Sleep(podRetrievalInterval)
		}
		if err != nil {
			log.Errorf("Failed to get %s/%s pod info: %v", ns, name, err)
			return true
		}

		return ignoreMeshlessPod(ns, name, pi)
	}
	log.Debugf("Not a kubernetes pod")
	return true
}

func ignoreMeshlessPod(namespace, name string, pod *podInfo) bool {
	if len(pod.Containers) > 1 {
		// Check if the pod is annotated for injection
		if podInjectAnnotationExists, injectEnabled, err := isAnnotatedForInjection(pod.Annotations); err != nil {
			log.Warnf("Pod %s/%s error determining sidecar-injection annotation", namespace, name)
			return true
		} else if podInjectAnnotationExists && !injectEnabled {
			log.Infof("Pod %s/%s excluded due to sidecar-injection annotation", namespace, name)
			return true
		}

		sidecarExists := false
		for _, container := range pod.Containers {
			if container == `sidecar` {
				sidecarExists = true
				break
			}
		}
		if !sidecarExists {
			log.Infof("Pod %s/%s excluded due to not existing sidecar", namespace, name)
			return true
		}
		return false
	}
	log.Infof("Pod %s/%s excluded because it only has %d containers", namespace, name, len(pod.Containers))
	return true
}

func isAnnotatedForInjection(annotations map[string]string) (exists bool, enabled bool, err error) {
	inject, ok := annotations[SidecarInjectionAnnotation]
	if !ok {
		return
	}
	exists = true
	switch strings.ToLower(inject) {
	case "enabled", "yes", "true":
		enabled = true
	case "disabled", "no", "false":
		enabled = false
	default:
		err = fmt.Errorf("invalid annotation value for key %q: %s", SidecarInjectionAnnotation, inject)
	}
	return
}

// CmdAdd is the implementation of the cmdAdd interface of CNI plugin
func CmdAdd(args *skel.CmdArgs) (err error) {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		log.Errorf("osm-cni cmdAdd failed to parse config %v %v", string(args.StdinData), err)
		return err
	}
	k8sArgs := K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		return err
	}

	if !ignore(conf, &k8sArgs) {
		httpc := http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", "/var/run/osm-cni.sock")
				},
			},
		}
		bs, _ := json.Marshal(args)
		body := bytes.NewReader(bs)
		_, err = httpc.Post("http://osm-cni"+config.CNICreatePodURL, "application/json", body)
		if err != nil {
			return err
		}
	}

	var result *cniv1.Result
	if conf.PrevResult == nil {
		result = &cniv1.Result{
			CNIVersion: cniv1.ImplementedSpecVersion,
		}
	} else {
		// Pass through the result for the next plugin
		result = conf.PrevResult
	}
	return types.PrintResult(result, conf.CNIVersion)
}

// CmdCheck is the implementation of the cmdCheck interface of CNI plugin
func CmdCheck(*skel.CmdArgs) (err error) {
	return err
}

// CmdDelete is the implementation of the cmdDelete interface of CNI plugin
func CmdDelete(args *skel.CmdArgs) (err error) {
	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/osm-cni.sock")
			},
		},
	}
	bs, _ := json.Marshal(args)
	body := bytes.NewReader(bs)
	_, err = httpc.Post("http://osm-cni"+config.CNIDeletePodURL, "application/json", body)
	return err
}
