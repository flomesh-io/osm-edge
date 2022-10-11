// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/openservicemesh/osm/pkg/policy (interfaces: Controller)

// Package policy is a generated GoMock package.
package policy

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/openservicemesh/osm/pkg/apis/policy/v1alpha1"
	identity "github.com/openservicemesh/osm/pkg/identity"
	service "github.com/openservicemesh/osm/pkg/service"
	v1 "k8s.io/api/core/v1"
)

// MockController is a mock of Controller interface.
type MockController struct {
	ctrl     *gomock.Controller
	recorder *MockControllerMockRecorder
}

// MockControllerMockRecorder is the mock recorder for MockController.
type MockControllerMockRecorder struct {
	mock *MockController
}

// NewMockController creates a new mock instance.
func NewMockController(ctrl *gomock.Controller) *MockController {
	mock := &MockController{ctrl: ctrl}
	mock.recorder = &MockControllerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockController) EXPECT() *MockControllerMockRecorder {
	return m.recorder
}

// GetAccessControlPolicy mocks base method.
func (m *MockController) GetAccessControlPolicy(arg0 service.MeshService) *v1alpha1.AccessControl {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccessControlPolicy", arg0)
	ret0, _ := ret[0].(*v1alpha1.AccessControl)
	return ret0
}

// GetAccessControlPolicy indicates an expected call of GetAccessControlPolicy.
func (mr *MockControllerMockRecorder) GetAccessControlPolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccessControlPolicy", reflect.TypeOf((*MockController)(nil).GetAccessControlPolicy), arg0)
}

// GetEgressSourceSecret mocks base method.
func (m *MockController) GetEgressSourceSecret(arg0 v1.SecretReference) (*v1.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEgressSourceSecret", arg0)
	ret0, _ := ret[0].(*v1.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEgressSourceSecret indicates an expected call of GetEgressSourceSecret.
func (mr *MockControllerMockRecorder) GetEgressSourceSecret(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEgressSourceSecret", reflect.TypeOf((*MockController)(nil).GetEgressSourceSecret), arg0)
}

// GetIngressBackendPolicy mocks base method.
func (m *MockController) GetIngressBackendPolicy(arg0 service.MeshService) *v1alpha1.IngressBackend {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIngressBackendPolicy", arg0)
	ret0, _ := ret[0].(*v1alpha1.IngressBackend)
	return ret0
}

// GetIngressBackendPolicy indicates an expected call of GetIngressBackendPolicy.
func (mr *MockControllerMockRecorder) GetIngressBackendPolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIngressBackendPolicy", reflect.TypeOf((*MockController)(nil).GetIngressBackendPolicy), arg0)
}

// GetUpstreamTrafficSetting mocks base method.
func (m *MockController) GetUpstreamTrafficSetting(arg0 UpstreamTrafficSettingGetOpt) *v1alpha1.UpstreamTrafficSetting {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpstreamTrafficSetting", arg0)
	ret0, _ := ret[0].(*v1alpha1.UpstreamTrafficSetting)
	return ret0
}

// GetUpstreamTrafficSetting indicates an expected call of GetUpstreamTrafficSetting.
func (mr *MockControllerMockRecorder) GetUpstreamTrafficSetting(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpstreamTrafficSetting", reflect.TypeOf((*MockController)(nil).GetUpstreamTrafficSetting), arg0)
}

// ListEgressGateways mocks base method.
func (m *MockController) ListEgressGateways() []*v1alpha1.EgressGateway {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEgressGateways")
	ret0, _ := ret[0].([]*v1alpha1.EgressGateway)
	return ret0
}

// ListEgressGateways indicates an expected call of ListEgressGateways.
func (mr *MockControllerMockRecorder) ListEgressGateways() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEgressGateways", reflect.TypeOf((*MockController)(nil).ListEgressGateways))
}

// ListEgressPoliciesForSourceIdentity mocks base method.
func (m *MockController) ListEgressPoliciesForSourceIdentity(arg0 identity.K8sServiceAccount) []*v1alpha1.Egress {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEgressPoliciesForSourceIdentity", arg0)
	ret0, _ := ret[0].([]*v1alpha1.Egress)
	return ret0
}

// ListEgressPoliciesForSourceIdentity indicates an expected call of ListEgressPoliciesForSourceIdentity.
func (mr *MockControllerMockRecorder) ListEgressPoliciesForSourceIdentity(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEgressPoliciesForSourceIdentity", reflect.TypeOf((*MockController)(nil).ListEgressPoliciesForSourceIdentity), arg0)
}

// ListRetryPolicies mocks base method.
func (m *MockController) ListRetryPolicies(arg0 identity.K8sServiceAccount) []*v1alpha1.Retry {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRetryPolicies", arg0)
	ret0, _ := ret[0].([]*v1alpha1.Retry)
	return ret0
}

// ListRetryPolicies indicates an expected call of ListRetryPolicies.
func (mr *MockControllerMockRecorder) ListRetryPolicies(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRetryPolicies", reflect.TypeOf((*MockController)(nil).ListRetryPolicies), arg0)
}
