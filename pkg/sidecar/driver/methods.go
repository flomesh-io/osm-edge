package driver

import "time"

// NewHealthProbe is the new method of HealthProbe
func NewHealthProbe(path string, port int32, http bool, timeout time.Duration, tcpSocket bool) *HealthProbe {
	return &HealthProbe{
		path:      path,
		port:      port,
		timeout:   timeout,
		http:      http,
		tcpSocket: tcpSocket,
	}
}

// NewHealthProbes is the new method of HealthProbes
func NewHealthProbes(liveness, readiness, startup *HealthProbe) *HealthProbes {
	return &HealthProbes{
		liveness:  liveness,
		readiness: readiness,
		startup:   startup,
	}
}

// GetPath is the getter of HealthProbe
func (h *HealthProbe) GetPath() string {
	return h.path
}

// SetPath is the setter of HealthProbe
func (h *HealthProbe) SetPath(path string) {
	h.path = path
}

// GetPort is the getter of HealthProbe
func (h *HealthProbe) GetPort() int32 {
	return h.port
}

// SetPort is the setter of HealthProbe
func (h *HealthProbe) SetPort(port int32) {
	h.port = port
}

// GetTimeout is the getter of HealthProbe
func (h *HealthProbe) GetTimeout() time.Duration {
	return h.timeout
}

// SetTimeout is the setter of HealthProbe
func (h *HealthProbe) SetTimeout(timeout time.Duration) {
	h.timeout = timeout
}

// IsHTTP corresponds to an httpGet probe with a scheme of HTTP or undefined.
// This helps inform what kind of Sidecar config to add to the pod.
func (h *HealthProbe) IsHTTP() bool {
	return h.http
}

// SetHTTP is the setter of HealthProbe
func (h *HealthProbe) SetHTTP(http bool) {
	h.http = http
}

// IsTCPSocket indicates if the probe defines a TCPSocketAction.
func (h *HealthProbe) IsTCPSocket() bool {
	return h.tcpSocket
}

// SetTCPSocket is the setter of HealthProbe
func (h *HealthProbe) SetTCPSocket(tcpSocket bool) {
	h.tcpSocket = tcpSocket
}

// GetLiveness is the getter of HealthProbes
func (hps *HealthProbes) GetLiveness() *HealthProbe {
	return hps.liveness
}

// SetLiveness is the setter of HealthProbes
func (hps *HealthProbes) SetLiveness(liveness *HealthProbe) {
	hps.liveness = liveness
}

// GetReadiness is the getter of HealthProbes
func (hps *HealthProbes) GetReadiness() *HealthProbe {
	return hps.readiness
}

// SetReadiness is the setter of HealthProbes
func (hps *HealthProbes) SetReadiness(readiness *HealthProbe) {
	hps.readiness = readiness
}

// GetStartup is the getter of HealthProbes
func (hps *HealthProbes) GetStartup() *HealthProbe {
	return hps.startup
}

// SetStartup is the setter of HealthProbes
func (hps *HealthProbes) SetStartup(startup *HealthProbe) {
	hps.startup = startup
}

// Deadline returns the time when work done on behalf of this context
// should be canceled. Deadline returns ok==false when no deadline is
// set. Successive calls to Deadline return the same results.
func (c *InjectorContext) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled. Done may return nil if this context can
// never be canceled. Successive calls to Done return the same value.
// The close of the Done channel may happen asynchronously,
// after the cancel function returns.
func (c *InjectorContext) Done() <-chan struct{} {
	return nil
}

// Err returns nil, if Done is not yet closed,
// If Done is closed, Err returns a non-nil error explaining why:
// Canceled if the context was canceled
// or DeadlineExceeded if the context's deadline passed.
// After Err returns a non-nil error, successive calls to Err return the same error.
func (c *InjectorContext) Err() error {
	return nil
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *InjectorContext) Value(key interface{}) interface{} {
	if key == &InjectorCtxKey {
		return c
	}
	return nil
}

// Deadline returns the time when work done on behalf of this context
// should be canceled. Deadline returns ok==false when no deadline is
// set. Successive calls to Deadline return the same results.
func (c *ControllerContext) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done returns a channel that's closed when work done on behalf of this
// context should be canceled. Done may return nil if this context can
// never be canceled. Successive calls to Done return the same value.
// The close of the Done channel may happen asynchronously,
// after the cancel function returns.
func (c *ControllerContext) Done() <-chan struct{} {
	return nil
}

// Err returns nil, if Done is not yet closed,
// If Done is closed, Err returns a non-nil error explaining why:
// Canceled if the context was canceled
// or DeadlineExceeded if the context's deadline passed.
// After Err returns a non-nil error, successive calls to Err return the same error.
func (c *ControllerContext) Err() error {
	return nil
}

// Value returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Value with
// the same key returns the same result.
func (c *ControllerContext) Value(key interface{}) interface{} {
	if key == &ControllerCtxKey {
		return c
	}
	return nil
}
