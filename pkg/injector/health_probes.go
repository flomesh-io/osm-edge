package injector

import (
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/openservicemesh/osm/pkg/sidecar/driver"
)

const (
	livenessProbePort  = int32(15901)
	readinessProbePort = int32(15902)
	startupProbePort   = int32(15903)
	healthcheckPort    = int32(15904)

	// LivenessProbePath is used to know when to restart osm's services
	LivenessProbePath = "/osm-liveness-probe"
	// ReadinessProbePath is used to know when osm's services are ready to start
	ReadinessProbePath = "/osm-readiness-probe"
	// StartupProbePath is used to know when osm's services are started
	StartupProbePath = "/osm-startup-probe"
	// HealthcheckPath is used to know whether  osm's services are healthy
	HealthcheckPath = "/osm-healthcheck"
)

var errNoMatchingPort = errors.New("no matching port")

func rewriteHealthProbes(pod *corev1.Pod) driver.HealthProbes {
	probes := driver.HealthProbes{}
	for idx := range pod.Spec.Containers {
		if probe := rewriteLiveness(&pod.Spec.Containers[idx]); probe != nil {
			probes.SetLiveness(probe)
		}
		if probe := rewriteReadiness(&pod.Spec.Containers[idx]); probe != nil {
			probes.SetReadiness(probe)
		}
		if probe := rewriteStartup(&pod.Spec.Containers[idx]); probe != nil {
			probes.SetStartup(probe)
		}
	}
	return probes
}

func rewriteLiveness(container *corev1.Container) *driver.HealthProbe {
	return rewriteProbe(container.LivenessProbe, "liveness", LivenessProbePath, livenessProbePort, &container.Ports)
}

func rewriteReadiness(container *corev1.Container) *driver.HealthProbe {
	return rewriteProbe(container.ReadinessProbe, "readiness", ReadinessProbePath, readinessProbePort, &container.Ports)
}

func rewriteStartup(container *corev1.Container) *driver.HealthProbe {
	return rewriteProbe(container.StartupProbe, "startup", StartupProbePath, startupProbePort, &container.Ports)
}

func rewriteProbe(probe *corev1.Probe, probeType, path string, port int32, containerPorts *[]corev1.ContainerPort) *driver.HealthProbe {
	if probe == nil {
		return nil
	}

	originalProbe := &driver.HealthProbe{}
	var newPath string
	var definedPort *intstr.IntOrString
	if probe.HTTPGet != nil {
		definedPort = &probe.HTTPGet.Port
		originalProbe.SetHTTP(len(probe.HTTPGet.Scheme) == 0 || probe.HTTPGet.Scheme == corev1.URISchemeHTTP)
		originalProbe.SetPath(probe.HTTPGet.Path)
		if originalProbe.IsHTTP() {
			probe.HTTPGet.Path = path
			newPath = probe.HTTPGet.Path
		}
	} else if probe.TCPSocket != nil {
		// Transform the TCPSocket probe into a HttpGet probe
		originalProbe.SetTCPSocket(true)
		probe.HTTPGet = &corev1.HTTPGetAction{
			Port:        probe.TCPSocket.Port,
			Path:        HealthcheckPath,
			HTTPHeaders: []corev1.HTTPHeader{},
		}
		newPath = probe.HTTPGet.Path
		definedPort = &probe.HTTPGet.Port
		port = healthcheckPort
		probe.TCPSocket = nil
	} else {
		return nil
	}

	probePort, err := getPort(*definedPort, containerPorts)
	originalProbe.SetPort(probePort)
	if err != nil {
		log.Error().Err(err).Msgf("Error finding a matching port for %+v on container %+v", *definedPort, containerPorts)
	}
	if originalProbe.IsTCPSocket() {
		probe.HTTPGet.HTTPHeaders = append(probe.HTTPGet.HTTPHeaders, corev1.HTTPHeader{Name: "Original-Tcp-Port", Value: fmt.Sprint(originalProbe.GetPort())})
	}
	*definedPort = intstr.IntOrString{Type: intstr.Int, IntVal: port}
	originalProbe.SetTimeout(time.Duration(probe.TimeoutSeconds) * time.Second)

	log.Debug().Msgf(
		"Rewriting %s probe (:%d%s) to :%d%s",
		probeType,
		originalProbe.GetPort(), originalProbe.GetPath(),
		port, newPath,
	)

	return originalProbe
}

// getPort returns the int32 of an IntOrString port; It looks for port's name matches in the full list of container ports
func getPort(namedPort intstr.IntOrString, containerPorts *[]corev1.ContainerPort) (int32, error) {
	// Maybe this is not a named port
	intPort := int32(namedPort.IntValue())
	if intPort != 0 {
		return intPort, nil
	}

	if containerPorts == nil {
		return 0, errNoMatchingPort
	}

	// Find an integer match for the name of the port in the list of container ports
	portName := namedPort.String()
	for _, p := range *containerPorts {
		if p.Name != "" && p.Name == portName {
			return p.ContainerPort, nil
		}
	}

	return 0, errNoMatchingPort
}
