// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudtrust/keycloak-bridge/pkg/health (interfaces: KeycloakClient)

// Package mock is a generated GoMock package.
package mock

import (
	keycloak_client "github.com/cloudtrust/keycloak-client"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// KeycloakClient is a mock of KeycloakClient interface
type KeycloakClient struct {
	ctrl     *gomock.Controller
	recorder *KeycloakClientMockRecorder
}

// KeycloakClientMockRecorder is the mock recorder for KeycloakClient
type KeycloakClientMockRecorder struct {
	mock *KeycloakClient
}

// NewKeycloakClient creates a new mock instance
func NewKeycloakClient(ctrl *gomock.Controller) *KeycloakClient {
	mock := &KeycloakClient{ctrl: ctrl}
	mock.recorder = &KeycloakClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *KeycloakClient) EXPECT() *KeycloakClientMockRecorder {
	return m.recorder
}

// CountUsers mocks base method
func (m *KeycloakClient) CountUsers(arg0, arg1 string) (int, error) {
	ret := m.ctrl.Call(m, "CountUsers", arg0, arg1)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountUsers indicates an expected call of CountUsers
func (mr *KeycloakClientMockRecorder) CountUsers(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountUsers", reflect.TypeOf((*KeycloakClient)(nil).CountUsers), arg0, arg1)
}

// CreateRealm mocks base method
func (m *KeycloakClient) CreateRealm(arg0 string, arg1 keycloak_client.RealmRepresentation) (string, error) {
	ret := m.ctrl.Call(m, "CreateRealm", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateRealm indicates an expected call of CreateRealm
func (mr *KeycloakClientMockRecorder) CreateRealm(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRealm", reflect.TypeOf((*KeycloakClient)(nil).CreateRealm), arg0, arg1)
}

// CreateUser mocks base method
func (m *KeycloakClient) CreateUser(arg0, arg1 string, arg2 keycloak_client.UserRepresentation) (string, error) {
	ret := m.ctrl.Call(m, "CreateUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser
func (mr *KeycloakClientMockRecorder) CreateUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*KeycloakClient)(nil).CreateUser), arg0, arg1, arg2)
}

// DeleteRealm mocks base method
func (m *KeycloakClient) DeleteRealm(arg0, arg1 string) error {
	ret := m.ctrl.Call(m, "DeleteRealm", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRealm indicates an expected call of DeleteRealm
func (mr *KeycloakClientMockRecorder) DeleteRealm(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRealm", reflect.TypeOf((*KeycloakClient)(nil).DeleteRealm), arg0, arg1)
}

// DeleteUser mocks base method
func (m *KeycloakClient) DeleteUser(arg0, arg1, arg2 string) error {
	ret := m.ctrl.Call(m, "DeleteUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser
func (mr *KeycloakClientMockRecorder) DeleteUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*KeycloakClient)(nil).DeleteUser), arg0, arg1, arg2)
}

// GetRealm mocks base method
func (m *KeycloakClient) GetRealm(arg0, arg1 string) (keycloak_client.RealmRepresentation, error) {
	ret := m.ctrl.Call(m, "GetRealm", arg0, arg1)
	ret0, _ := ret[0].(keycloak_client.RealmRepresentation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRealm indicates an expected call of GetRealm
func (mr *KeycloakClientMockRecorder) GetRealm(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRealm", reflect.TypeOf((*KeycloakClient)(nil).GetRealm), arg0, arg1)
}

// GetRealms mocks base method
func (m *KeycloakClient) GetRealms(arg0 string) ([]keycloak_client.RealmRepresentation, error) {
	ret := m.ctrl.Call(m, "GetRealms", arg0)
	ret0, _ := ret[0].([]keycloak_client.RealmRepresentation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRealms indicates an expected call of GetRealms
func (mr *KeycloakClientMockRecorder) GetRealms(arg0 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRealms", reflect.TypeOf((*KeycloakClient)(nil).GetRealms), arg0)
}

// GetUser mocks base method
func (m *KeycloakClient) GetUser(arg0, arg1, arg2 string) (keycloak_client.UserRepresentation, error) {
	ret := m.ctrl.Call(m, "GetUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(keycloak_client.UserRepresentation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser
func (mr *KeycloakClientMockRecorder) GetUser(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*KeycloakClient)(nil).GetUser), arg0, arg1, arg2)
}

// GetUsers mocks base method
func (m *KeycloakClient) GetUsers(arg0, arg1 string, arg2 ...string) ([]keycloak_client.UserRepresentation, error) {
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetUsers", varargs...)
	ret0, _ := ret[0].([]keycloak_client.UserRepresentation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsers indicates an expected call of GetUsers
func (mr *KeycloakClientMockRecorder) GetUsers(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsers", reflect.TypeOf((*KeycloakClient)(nil).GetUsers), varargs...)
}

// UpdateRealm mocks base method
func (m *KeycloakClient) UpdateRealm(arg0, arg1 string, arg2 keycloak_client.RealmRepresentation) error {
	ret := m.ctrl.Call(m, "UpdateRealm", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateRealm indicates an expected call of UpdateRealm
func (mr *KeycloakClientMockRecorder) UpdateRealm(arg0, arg1, arg2 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateRealm", reflect.TypeOf((*KeycloakClient)(nil).UpdateRealm), arg0, arg1, arg2)
}

// UpdateUser mocks base method
func (m *KeycloakClient) UpdateUser(arg0, arg1, arg2 string, arg3 keycloak_client.UserRepresentation) error {
	ret := m.ctrl.Call(m, "UpdateUser", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateUser indicates an expected call of UpdateUser
func (mr *KeycloakClientMockRecorder) UpdateUser(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*KeycloakClient)(nil).UpdateUser), arg0, arg1, arg2, arg3)
}
