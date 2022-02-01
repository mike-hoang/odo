// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/init/interface.go

// Package init is a generated GoMock package.
package init

import (
	reflect "reflect"

	parser "github.com/devfile/library/pkg/devfile/parser"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// DownloadDirect mocks base method.
func (m *MockClient) DownloadDirect(URL, dest string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadDirect", URL, dest)
	ret0, _ := ret[0].(error)
	return ret0
}

// DownloadDirect indicates an expected call of DownloadDirect.
func (mr *MockClientMockRecorder) DownloadDirect(URL, dest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadDirect", reflect.TypeOf((*MockClient)(nil).DownloadDirect), URL, dest)
}

// DownloadFromRegistry mocks base method.
func (m *MockClient) DownloadFromRegistry(registryName, devfile, dest string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadFromRegistry", registryName, devfile, dest)
	ret0, _ := ret[0].(error)
	return ret0
}

// DownloadFromRegistry indicates an expected call of DownloadFromRegistry.
func (mr *MockClientMockRecorder) DownloadFromRegistry(registryName, devfile, dest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadFromRegistry", reflect.TypeOf((*MockClient)(nil).DownloadFromRegistry), registryName, devfile, dest)
}

// DownloadStarterProject mocks base method.
func (m *MockClient) DownloadStarterProject(devfile parser.DevfileObj, project, dest string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadStarterProject", devfile, project, dest)
	ret0, _ := ret[0].(error)
	return ret0
}

// DownloadStarterProject indicates an expected call of DownloadStarterProject.
func (mr *MockClientMockRecorder) DownloadStarterProject(devfile, project, dest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadStarterProject", reflect.TypeOf((*MockClient)(nil).DownloadStarterProject), devfile, project, dest)
}
