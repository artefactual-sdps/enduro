// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/artefactual-sdps/enduro/internal/sftp (interfaces: Service)
//
// Generated by this command:
//
//	mockgen -typed -destination=./internal/sftp/fake/mock_sftp.go -package=fake github.com/artefactual-sdps/enduro/internal/sftp Service
//
// Package fake is a generated GoMock package.
package fake

import (
	io "io"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// Upload mocks base method.
func (m *MockService) Upload(arg0 io.Reader, arg1 string) (int64, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", arg0, arg1)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Upload indicates an expected call of Upload.
func (mr *MockServiceMockRecorder) Upload(arg0, arg1 any) *ServiceUploadCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockService)(nil).Upload), arg0, arg1)
	return &ServiceUploadCall{Call: call}
}

// ServiceUploadCall wrap *gomock.Call
type ServiceUploadCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *ServiceUploadCall) Return(arg0 int64, arg1 string, arg2 error) *ServiceUploadCall {
	c.Call = c.Call.Return(arg0, arg1, arg2)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *ServiceUploadCall) Do(f func(io.Reader, string) (int64, string, error)) *ServiceUploadCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *ServiceUploadCall) DoAndReturn(f func(io.Reader, string) (int64, string, error)) *ServiceUploadCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}