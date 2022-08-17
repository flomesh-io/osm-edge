// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/openservicemesh/osm/pkg/catalog (interfaces: MeshCataloger)

// Package catalog is a generated GoMock package.
package catalog

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	endpoint "github.com/openservicemesh/osm/pkg/endpoint"
	identity "github.com/openservicemesh/osm/pkg/identity"
	k8s "github.com/openservicemesh/osm/pkg/k8s"
	service "github.com/openservicemesh/osm/pkg/service"
	trafficpolicy "github.com/openservicemesh/osm/pkg/trafficpolicy"
)

// MockMeshCataloger is a mock of MeshCataloger interface.
type MockMeshCataloger struct {
	ctrl     *gomock.Controller
	recorder *MockMeshCatalogerMockRecorder
}

// MockMeshCatalogerMockRecorder is the mock recorder for MockMeshCataloger.
type MockMeshCatalogerMockRecorder struct {
	mock *MockMeshCataloger
}

// NewMockMeshCataloger creates a new mock instance.
func NewMockMeshCataloger(ctrl *gomock.Controller) *MockMeshCataloger {
	mock := &MockMeshCataloger{ctrl: ctrl}
	mock.recorder = &MockMeshCatalogerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMeshCataloger) EXPECT() *MockMeshCatalogerMockRecorder {
	return m.recorder
}

// GetEgressTrafficPolicy mocks base method.
func (m *MockMeshCataloger) GetEgressTrafficPolicy(arg0 identity.ServiceIdentity) (*trafficpolicy.EgressTrafficPolicy, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEgressTrafficPolicy", arg0)
	ret0, _ := ret[0].(*trafficpolicy.EgressTrafficPolicy)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEgressTrafficPolicy indicates an expected call of GetEgressTrafficPolicy.
func (mr *MockMeshCatalogerMockRecorder) GetEgressTrafficPolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEgressTrafficPolicy", reflect.TypeOf((*MockMeshCataloger)(nil).GetEgressTrafficPolicy), arg0)
}

// GetInboundMeshTrafficPolicy mocks base method.
func (m *MockMeshCataloger) GetInboundMeshTrafficPolicy(arg0 identity.ServiceIdentity, arg1 []service.MeshService) *trafficpolicy.InboundMeshTrafficPolicy {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInboundMeshTrafficPolicy", arg0, arg1)
	ret0, _ := ret[0].(*trafficpolicy.InboundMeshTrafficPolicy)
	return ret0
}

// GetInboundMeshTrafficPolicy indicates an expected call of GetInboundMeshTrafficPolicy.
func (mr *MockMeshCatalogerMockRecorder) GetInboundMeshTrafficPolicy(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInboundMeshTrafficPolicy", reflect.TypeOf((*MockMeshCataloger)(nil).GetInboundMeshTrafficPolicy), arg0, arg1)
}

// GetIngressTrafficPolicy mocks base method.
func (m *MockMeshCataloger) GetIngressTrafficPolicy(arg0 service.MeshService) (*trafficpolicy.IngressTrafficPolicy, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIngressTrafficPolicy", arg0)
	ret0, _ := ret[0].(*trafficpolicy.IngressTrafficPolicy)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIngressTrafficPolicy indicates an expected call of GetIngressTrafficPolicy.
func (mr *MockMeshCatalogerMockRecorder) GetIngressTrafficPolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIngressTrafficPolicy", reflect.TypeOf((*MockMeshCataloger)(nil).GetIngressTrafficPolicy), arg0)
}

// GetKubeController mocks base method.
func (m *MockMeshCataloger) GetKubeController() k8s.Controller {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKubeController")
	ret0, _ := ret[0].(k8s.Controller)
	return ret0
}

// GetKubeController indicates an expected call of GetKubeController.
func (mr *MockMeshCatalogerMockRecorder) GetKubeController() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKubeController", reflect.TypeOf((*MockMeshCataloger)(nil).GetKubeController))
}

// GetOutboundMeshTrafficPolicy mocks base method.
func (m *MockMeshCataloger) GetOutboundMeshTrafficPolicy(arg0 identity.ServiceIdentity) *trafficpolicy.OutboundMeshTrafficPolicy {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOutboundMeshTrafficPolicy", arg0)
	ret0, _ := ret[0].(*trafficpolicy.OutboundMeshTrafficPolicy)
	return ret0
}

// GetOutboundMeshTrafficPolicy indicates an expected call of GetOutboundMeshTrafficPolicy.
func (mr *MockMeshCatalogerMockRecorder) GetOutboundMeshTrafficPolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOutboundMeshTrafficPolicy", reflect.TypeOf((*MockMeshCataloger)(nil).GetOutboundMeshTrafficPolicy), arg0)
}

// ListAllowedUpstreamEndpointsForService mocks base method.
func (m *MockMeshCataloger) ListAllowedUpstreamEndpointsForService(arg0 identity.ServiceIdentity, arg1 service.MeshService) []endpoint.Endpoint {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAllowedUpstreamEndpointsForService", arg0, arg1)
	ret0, _ := ret[0].([]endpoint.Endpoint)
	return ret0
}

// ListAllowedUpstreamEndpointsForService indicates an expected call of ListAllowedUpstreamEndpointsForService.
func (mr *MockMeshCatalogerMockRecorder) ListAllowedUpstreamEndpointsForService(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAllowedUpstreamEndpointsForService", reflect.TypeOf((*MockMeshCataloger)(nil).ListAllowedUpstreamEndpointsForService), arg0, arg1)
}

// ListInboundServiceIdentities mocks base method.
func (m *MockMeshCataloger) ListInboundServiceIdentities(arg0 identity.ServiceIdentity) []identity.ServiceIdentity {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListInboundServiceIdentities", arg0)
	ret0, _ := ret[0].([]identity.ServiceIdentity)
	return ret0
}

// ListInboundServiceIdentities indicates an expected call of ListInboundServiceIdentities.
func (mr *MockMeshCatalogerMockRecorder) ListInboundServiceIdentities(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListInboundServiceIdentities", reflect.TypeOf((*MockMeshCataloger)(nil).ListInboundServiceIdentities), arg0)
}

// ListInboundTrafficTargetsWithRoutes mocks base method.
func (m *MockMeshCataloger) ListInboundTrafficTargetsWithRoutes(arg0 identity.ServiceIdentity) ([]trafficpolicy.TrafficTargetWithRoutes, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListInboundTrafficTargetsWithRoutes", arg0)
	ret0, _ := ret[0].([]trafficpolicy.TrafficTargetWithRoutes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListInboundTrafficTargetsWithRoutes indicates an expected call of ListInboundTrafficTargetsWithRoutes.
func (mr *MockMeshCatalogerMockRecorder) ListInboundTrafficTargetsWithRoutes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListInboundTrafficTargetsWithRoutes", reflect.TypeOf((*MockMeshCataloger)(nil).ListInboundTrafficTargetsWithRoutes), arg0)
}

// ListOutboundServiceIdentities mocks base method.
func (m *MockMeshCataloger) ListOutboundServiceIdentities(arg0 identity.ServiceIdentity) []identity.ServiceIdentity {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListOutboundServiceIdentities", arg0)
	ret0, _ := ret[0].([]identity.ServiceIdentity)
	return ret0
}

// ListOutboundServiceIdentities indicates an expected call of ListOutboundServiceIdentities.
func (mr *MockMeshCatalogerMockRecorder) ListOutboundServiceIdentities(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOutboundServiceIdentities", reflect.TypeOf((*MockMeshCataloger)(nil).ListOutboundServiceIdentities), arg0)
}

// ListOutboundServicesForIdentity mocks base method.
func (m *MockMeshCataloger) ListOutboundServicesForIdentity(arg0 identity.ServiceIdentity) []service.MeshService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListOutboundServicesForIdentity", arg0)
	ret0, _ := ret[0].([]service.MeshService)
	return ret0
}

// ListOutboundServicesForIdentity indicates an expected call of ListOutboundServicesForIdentity.
func (mr *MockMeshCatalogerMockRecorder) ListOutboundServicesForIdentity(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOutboundServicesForIdentity", reflect.TypeOf((*MockMeshCataloger)(nil).ListOutboundServicesForIdentity), arg0)
}

// ListServiceIdentitiesForService mocks base method.
func (m *MockMeshCataloger) ListServiceIdentitiesForService(arg0 service.MeshService) []identity.ServiceIdentity {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListServiceIdentitiesForService", arg0)
	ret0, _ := ret[0].([]identity.ServiceIdentity)
	return ret0
}

// ListServiceIdentitiesForService indicates an expected call of ListServiceIdentitiesForService.
func (mr *MockMeshCatalogerMockRecorder) ListServiceIdentitiesForService(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListServiceIdentitiesForService", reflect.TypeOf((*MockMeshCataloger)(nil).ListServiceIdentitiesForService), arg0)
}
