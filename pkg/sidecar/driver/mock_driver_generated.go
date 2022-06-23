// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/openservicemesh/osm/pkg/sidecar/driver (interfaces: ProxyDebugger)

// Package driver is a generated GoMock package.
package driver

import (
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockProxyDebugger is a mock of ProxyDebugger interface.
type MockProxyDebugger struct {
	ctrl     *gomock.Controller
	recorder *MockProxyDebuggerMockRecorder
}

// MockProxyDebuggerMockRecorder is the mock recorder for MockProxyDebugger.
type MockProxyDebuggerMockRecorder struct {
	mock *MockProxyDebugger
}

// NewMockProxyDebugger creates a new mock instance.
func NewMockProxyDebugger(ctrl *gomock.Controller) *MockProxyDebugger {
	mock := &MockProxyDebugger{ctrl: ctrl}
	mock.recorder = &MockProxyDebuggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProxyDebugger) EXPECT() *MockProxyDebuggerMockRecorder {
	return m.recorder
}

// GetDebugHandlers mocks base method.
func (m *MockProxyDebugger) GetDebugHandlers() map[string]http.Handler {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDebugHandlers")
	ret0, _ := ret[0].(map[string]http.Handler)
	return ret0
}

// GetDebugHandlers indicates an expected call of GetDebugHandlers.
func (mr *MockProxyDebuggerMockRecorder) GetDebugHandlers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDebugHandlers", reflect.TypeOf((*MockProxyDebugger)(nil).GetDebugHandlers))
}
