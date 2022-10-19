package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

var (
	refreshFrequency = time.Second * 2
)

type watcher struct {
	osmApp            string
	podName           string
	spinner           *spinner.Spinner
	containerCnt      int
	readyContainerCnt int
	restartCnt        int
	ready             bool
}

func (w *watcher) refresh(clientSet kubernetes.Interface, osmNamespace string) {
	if w.ready {
		return
	}
	fieldSelector := fields.OneTermEqualSelector("metadata.name", w.podName).String()
	pods, err := clientSet.CoreV1().Pods(osmNamespace).List(context.Background(), metav1.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		return
	}
	if len(pods.Items) == 0 {
		w.ready = true
		w.spinner.Stop()
		return
	}
	pod := pods.Items[0]
	phase := pod.Status.Phase
	containers := pod.Status.ContainerStatuses
	w.containerCnt = len(containers)
	w.readyContainerCnt = 0
	w.restartCnt = 0
	for _, c := range containers {
		if c.Ready {
			w.readyContainerCnt++
		}
		w.restartCnt = w.restartCnt + int(c.RestartCount)
	}
	if w.containerCnt == w.readyContainerCnt || v1.PodSucceeded == phase {
		w.ready = true
		w.spinner.Stop()
	} else {
		w.spinner.Suffix = fmt.Sprintf("%s[%s] READY:%d/%d STATUS:%s RESTARTS:%d",
			w.podName, osmNamespace, w.readyContainerCnt, w.containerCnt, phase, w.restartCnt)
	}
}

// Spinner indicator to osm install
type Spinner struct {
	osmNamespace string
	clientSet    kubernetes.Interface
	watchers     map[string]*watcher
	err          error
	quit         chan bool

	deployPrometheus bool
	deployGrafana    bool
	deployJaeger     bool
}

// Init instance of Spinner with the supplied options
func (s *Spinner) Init(clientSet kubernetes.Interface, osmNamespace string, vals map[string]interface{}) {
	s.clientSet = clientSet
	s.osmNamespace = osmNamespace
	s.quit = make(chan bool, 1)
	s.watchers = make(map[string]*watcher)
	if osm, exists := vals["osm"]; exists {
		osmVals := osm.(map[string]interface{})
		if v, has := osmVals["deployPrometheus"]; has {
			s.deployPrometheus = v.(bool)
		}
		if v, has := osmVals["deployGrafana"]; has {
			s.deployGrafana = v.(bool)
		}
		if v, has := osmVals["deployJaeger"]; has {
			s.deployJaeger = v.(bool)
		}
	}
}

func (s *Spinner) done() bool {
	if len(s.watchers) >= 3 {
		doneApps := map[string]bool{
			"osm-bootstrap":  false,
			"osm-injector":   false,
			"osm-controller": false,
		}
		if s.deployPrometheus {
			doneApps["osm-prometheus"] = false
		}
		if s.deployGrafana {
			doneApps["osm-grafana"] = false
		}
		if s.deployJaeger {
			doneApps["osm-jaeger"] = false
		}
		for _, w := range s.watchers {
			if !w.ready {
				return false
			}
			doneApps[w.osmApp] = true
		}
		for _, done := range doneApps {
			if !done {
				return false
			}
		}
		return true
	}
	return false
}

func (s *Spinner) refreshOsmApps() {
	apps, err := s.clientSet.CoreV1().Pods(s.osmNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		s.err = err
		s.quit <- true
		return
	}
	if len(apps.Items) == 0 {
		return
	}
	for _, app := range apps.Items {
		_, exists := s.watchers[app.Name]
		if !exists {
			w := new(watcher)
			w.podName = app.Name
			w.osmApp = app.Labels["app"]
			if len(w.osmApp) == 0 {
				parts := strings.Split(w.podName, "-")
				if len(parts) > 1 {
					w.osmApp = fmt.Sprintf("%s-%s", parts[0], parts[1])
				}
			}
			w.spinner = spinner.New(spinner.CharSets[35], time.Millisecond*500)
			_ = w.spinner.Color("bgBlue", "bold", "fgGreen")

			if len(w.osmApp) > 0 {
				w.spinner.Suffix = fmt.Sprintf("%s[%s] is being deployed ...", w.osmApp, w.podName)
				w.spinner.FinalMSG = fmt.Sprintf("%s[%s] Done\n", w.osmApp, w.podName)
			} else {
				w.spinner.Suffix = w.podName
				w.spinner.FinalMSG = fmt.Sprintf("%s Done\n", w.podName)
			}
			w.spinner.Start()
			s.watchers[app.Name] = w
		}
	}
	for _, w := range s.watchers {
		w.refresh(s.clientSet, s.osmNamespace)
	}
}

// Run starts spinner indicator
func (s *Spinner) Run(job func() error) error {
	updateChan := make(chan interface{}, 1)

	slidingTimer := time.NewTimer(refreshFrequency)
	defer slidingTimer.Stop()

	go func() {
		if err := job(); err != nil {
			s.err = err
			s.quit <- true
		}
	}()

	for {
		select {
		case <-s.quit:
			return s.err
		case <-updateChan:
			slidingTimer.Reset(refreshFrequency)
		case <-slidingTimer.C:
			s.refreshOsmApps()
			if !s.done() {
				updateChan <- new(interface{})
			} else {
				s.quit <- true
			}
		}
	}
}
