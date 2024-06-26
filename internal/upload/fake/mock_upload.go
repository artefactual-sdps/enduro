// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/artefactual-sdps/enduro/internal/upload (interfaces: Service)
//
// Generated by this command:
//
//	mockgen -typed -destination=./internal/upload/fake/mock_upload.go -package=fake github.com/artefactual-sdps/enduro/internal/upload Service
//

// Package fake is a generated GoMock package.
package fake

import (
	context "context"
	io "io"
	reflect "reflect"

	upload "github.com/artefactual-sdps/enduro/internal/api/gen/upload"
	gomock "go.uber.org/mock/gomock"
	blob "gocloud.dev/blob"
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

// Bucket mocks base method.
func (m *MockService) Bucket() *blob.Bucket {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Bucket")
	ret0, _ := ret[0].(*blob.Bucket)
	return ret0
}

// Bucket indicates an expected call of Bucket.
func (mr *MockServiceMockRecorder) Bucket() *MockServiceBucketCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Bucket", reflect.TypeOf((*MockService)(nil).Bucket))
	return &MockServiceBucketCall{Call: call}
}

// MockServiceBucketCall wrap *gomock.Call
type MockServiceBucketCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceBucketCall) Return(arg0 *blob.Bucket) *MockServiceBucketCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceBucketCall) Do(f func() *blob.Bucket) *MockServiceBucketCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceBucketCall) DoAndReturn(f func() *blob.Bucket) *MockServiceBucketCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Close mocks base method.
func (m *MockService) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockServiceMockRecorder) Close() *MockServiceCloseCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockService)(nil).Close))
	return &MockServiceCloseCall{Call: call}
}

// MockServiceCloseCall wrap *gomock.Call
type MockServiceCloseCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceCloseCall) Return(arg0 error) *MockServiceCloseCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceCloseCall) Do(f func() error) *MockServiceCloseCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceCloseCall) DoAndReturn(f func() error) *MockServiceCloseCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Upload mocks base method.
func (m *MockService) Upload(arg0 context.Context, arg1 *upload.UploadPayload, arg2 io.ReadCloser) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upload indicates an expected call of Upload.
func (mr *MockServiceMockRecorder) Upload(arg0, arg1, arg2 any) *MockServiceUploadCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockService)(nil).Upload), arg0, arg1, arg2)
	return &MockServiceUploadCall{Call: call}
}

// MockServiceUploadCall wrap *gomock.Call
type MockServiceUploadCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceUploadCall) Return(arg0 error) *MockServiceUploadCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceUploadCall) Do(f func(context.Context, *upload.UploadPayload, io.ReadCloser) error) *MockServiceUploadCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceUploadCall) DoAndReturn(f func(context.Context, *upload.UploadPayload, io.ReadCloser) error) *MockServiceUploadCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
