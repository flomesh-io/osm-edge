package sidecar

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"

	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/configurator"
	"github.com/openservicemesh/osm/pkg/constants"
	"github.com/openservicemesh/osm/pkg/health"
	"github.com/openservicemesh/osm/pkg/identity"
	"github.com/openservicemesh/osm/pkg/sidecar/driver"
)

var (
	driversMutex sync.RWMutex
	drivers      = make(map[string]driver.Driver)
	engineDriver driver.Driver
)

// NewCertCommonName returns a newly generated CommonName for a certificate of the form: <ProxyUUID>.<kind>.<serviceAccount>.<namespace>
func NewCertCommonName(proxyUUID uuid.UUID, kind ProxyKind, serviceAccount, namespace string) certificate.CommonName {
	return certificate.CommonName(fmt.Sprintf("%s.%s.%s.%s.%s", proxyUUID.String(), kind, serviceAccount, namespace, identity.ClusterLocalTrustDomain))
}

// InitDriver is to serve as an indication of the using sidecar driver
func InitDriver(driverName string) error {
	driversMutex.Lock()
	defer driversMutex.Unlock()
	registeredDriver, ok := drivers[driverName]
	if !ok {
		return fmt.Errorf("sidecar: unknown driver %q (forgotten import?)", driverName)
	}
	engineDriver = registeredDriver
	return nil
}

// Register makes a sidecar driver available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver driver.Driver) {
	driversMutex.Lock()
	defer driversMutex.Unlock()
	if driver == nil {
		panic("sidecar: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("sidecar: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// Patch is an adapter method for InjectorDriver.Patch
func Patch(ctx context.Context, pod *corev1.Pod) ([]*corev1.Secret, error) {
	driversMutex.RLock()
	defer driversMutex.RUnlock()
	if engineDriver == nil {
		return nil, errors.New("sidecar: unknown driver (forgotten init?)")
	}
	return engineDriver.Patch(ctx, pod)
}

// Start is an adapter method for ControllerDriver.Start
func Start(ctx context.Context, port int, cert *certificate.Certificate) (health.Probes, error) {
	driversMutex.RLock()
	defer driversMutex.RUnlock()
	if engineDriver == nil {
		return nil, errors.New("sidecar: unknown driver (forgotten init?)")
	}
	return engineDriver.Start(ctx, port, cert)
}

// GetPlatformSpecificSpecComponents return the Platform Spec with SecurityContext and sidecarContainer
func GetPlatformSpecificSpecComponents(cfg configurator.Configurator, podOS string) (podSecurityContext *corev1.SecurityContext, sidecarContainer string) {
	if strings.EqualFold(podOS, constants.OSWindows) {
		podSecurityContext = &corev1.SecurityContext{
			WindowsOptions: &corev1.WindowsSecurityContextOptions{
				RunAsUserName: func() *string {
					userName := constants.SidecarWindowsUser
					return &userName
				}(),
			},
		}
		sidecarContainer = cfg.GetSidecarWindowsImage()
	} else {
		podSecurityContext = &corev1.SecurityContext{
			RunAsUser: func() *int64 {
				uid := constants.SidecarUID
				return &uid
			}(),
		}
		sidecarContainer = cfg.GetSidecarImage()
	}
	return
}
