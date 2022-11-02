// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/openservicemesh/osm/pkg/multicluster (interfaces: Controller)

// Package multicluster is a generated GoMock package.
package multicluster

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1alpha1 "github.com/openservicemesh/osm/pkg/apis/multicluster/v1alpha1"
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

// GetEndpoints mocks base method.
func (m *MockController) GetEndpoints(arg0 service.MeshService) (*v1.Endpoints, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEndpoints", arg0)
	ret0, _ := ret[0].(*v1.Endpoints)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEndpoints indicates an expected call of GetEndpoints.
func (mr *MockControllerMockRecorder) GetEndpoints(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEndpoints", reflect.TypeOf((*MockController)(nil).GetEndpoints), arg0)
}

// GetExportedRule mocks base method.
func (m *MockController) GetExportedRule(arg0 service.MeshService) (*v1alpha1.ServiceExportRule, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExportedRule", arg0)
	ret0, _ := ret[0].(*v1alpha1.ServiceExportRule)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExportedRule indicates an expected call of GetExportedRule.
func (mr *MockControllerMockRecorder) GetExportedRule(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExportedRule", reflect.TypeOf((*MockController)(nil).GetExportedRule), arg0)
}

// GetIngressControllerServices mocks base method.
func (m *MockController) GetIngressControllerServices() []service.MeshService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIngressControllerServices")
	ret0, _ := ret[0].([]service.MeshService)
	return ret0
}

// GetIngressControllerServices indicates an expected call of GetIngressControllerServices.
func (mr *MockControllerMockRecorder) GetIngressControllerServices() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIngressControllerServices", reflect.TypeOf((*MockController)(nil).GetIngressControllerServices))
}

// GetService mocks base method.
func (m *MockController) GetService(arg0 service.MeshService) *v1.Service {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetService", arg0)
	ret0, _ := ret[0].(*v1.Service)
	return ret0
}

// GetService indicates an expected call of GetService.
func (mr *MockControllerMockRecorder) GetService(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetService", reflect.TypeOf((*MockController)(nil).GetService), arg0)
}

// ListPods mocks base method.
func (m *MockController) ListPods() []*v1.Pod {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPods")
	ret0, _ := ret[0].([]*v1.Pod)
	return ret0
}

// ListPods indicates an expected call of ListPods.
func (mr *MockControllerMockRecorder) ListPods() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPods", reflect.TypeOf((*MockController)(nil).ListPods))
}

// ListServiceAccounts mocks base method.
func (m *MockController) ListServiceAccounts() []*v1.ServiceAccount {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListServiceAccounts")
	ret0, _ := ret[0].([]*v1.ServiceAccount)
	return ret0
}

// ListServiceAccounts indicates an expected call of ListServiceAccounts.
func (mr *MockControllerMockRecorder) ListServiceAccounts() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServiceAccounts", reflect.TypeOf((*MockController)(nil).ListServiceAccounts))
}

// ListServiceIdentitiesForService mocks base method.
func (m *MockController) ListServiceIdentitiesForService(arg0 service.MeshService) ([]identity.K8sServiceAccount, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListServiceIdentitiesForService", arg0)
	ret0, _ := ret[0].([]identity.K8sServiceAccount)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListServiceIdentitiesForService indicates an expected call of ListServiceIdentitiesForService.
func (mr *MockControllerMockRecorder) ListServiceIdentitiesForService(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServiceIdentitiesForService", reflect.TypeOf((*MockController)(nil).ListServiceIdentitiesForService), arg0)
}

// ListServices mocks base method.
func (m *MockController) ListServices() []*v1.Service {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListServices")
	ret0, _ := ret[0].([]*v1.Service)
	return ret0
}

// ListServices indicates an expected call of ListServices.
func (mr *MockControllerMockRecorder) ListServices() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServices", reflect.TypeOf((*MockController)(nil).ListServices))
}