// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/artefactual-sdps/enduro/internal/storage (interfaces: Client)
//
// Generated by this command:
//
//	mockgen -typed -destination=./internal/storage/fake/mock_client.go -package=fake github.com/artefactual-sdps/enduro/internal/storage Client
//

// Package fake is a generated GoMock package.
package fake

import (
	context "context"
	io "io"
	reflect "reflect"

	storage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	gomock "go.uber.org/mock/gomock"
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

// CreateAip mocks base method.
func (m *MockClient) CreateAip(arg0 context.Context, arg1 *storage.CreateAipPayload) (*storage.AIP, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAip", arg0, arg1)
	ret0, _ := ret[0].(*storage.AIP)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateAip indicates an expected call of CreateAip.
func (mr *MockClientMockRecorder) CreateAip(arg0, arg1 any) *MockClientCreateAipCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAip", reflect.TypeOf((*MockClient)(nil).CreateAip), arg0, arg1)
	return &MockClientCreateAipCall{Call: call}
}

// MockClientCreateAipCall wrap *gomock.Call
type MockClientCreateAipCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientCreateAipCall) Return(arg0 *storage.AIP, arg1 error) *MockClientCreateAipCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientCreateAipCall) Do(f func(context.Context, *storage.CreateAipPayload) (*storage.AIP, error)) *MockClientCreateAipCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientCreateAipCall) DoAndReturn(f func(context.Context, *storage.CreateAipPayload) (*storage.AIP, error)) *MockClientCreateAipCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// CreateLocation mocks base method.
func (m *MockClient) CreateLocation(arg0 context.Context, arg1 *storage.CreateLocationPayload) (*storage.CreateLocationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateLocation", arg0, arg1)
	ret0, _ := ret[0].(*storage.CreateLocationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateLocation indicates an expected call of CreateLocation.
func (mr *MockClientMockRecorder) CreateLocation(arg0, arg1 any) *MockClientCreateLocationCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateLocation", reflect.TypeOf((*MockClient)(nil).CreateLocation), arg0, arg1)
	return &MockClientCreateLocationCall{Call: call}
}

// MockClientCreateLocationCall wrap *gomock.Call
type MockClientCreateLocationCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientCreateLocationCall) Return(arg0 *storage.CreateLocationResult, arg1 error) *MockClientCreateLocationCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientCreateLocationCall) Do(f func(context.Context, *storage.CreateLocationPayload) (*storage.CreateLocationResult, error)) *MockClientCreateLocationCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientCreateLocationCall) DoAndReturn(f func(context.Context, *storage.CreateLocationPayload) (*storage.CreateLocationResult, error)) *MockClientCreateLocationCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// DownloadAip mocks base method.
func (m *MockClient) DownloadAip(arg0 context.Context, arg1 *storage.DownloadAipPayload) (*storage.DownloadAipResult, io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadAip", arg0, arg1)
	ret0, _ := ret[0].(*storage.DownloadAipResult)
	ret1, _ := ret[1].(io.ReadCloser)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// DownloadAip indicates an expected call of DownloadAip.
func (mr *MockClientMockRecorder) DownloadAip(arg0, arg1 any) *MockClientDownloadAipCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadAip", reflect.TypeOf((*MockClient)(nil).DownloadAip), arg0, arg1)
	return &MockClientDownloadAipCall{Call: call}
}

// MockClientDownloadAipCall wrap *gomock.Call
type MockClientDownloadAipCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientDownloadAipCall) Return(arg0 *storage.DownloadAipResult, arg1 io.ReadCloser, arg2 error) *MockClientDownloadAipCall {
	c.Call = c.Call.Return(arg0, arg1, arg2)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientDownloadAipCall) Do(f func(context.Context, *storage.DownloadAipPayload) (*storage.DownloadAipResult, io.ReadCloser, error)) *MockClientDownloadAipCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientDownloadAipCall) DoAndReturn(f func(context.Context, *storage.DownloadAipPayload) (*storage.DownloadAipResult, io.ReadCloser, error)) *MockClientDownloadAipCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// DownloadAipRequest mocks base method.
func (m *MockClient) DownloadAipRequest(arg0 context.Context, arg1 *storage.DownloadAipRequestPayload) (*storage.DownloadAipRequestResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadAipRequest", arg0, arg1)
	ret0, _ := ret[0].(*storage.DownloadAipRequestResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownloadAipRequest indicates an expected call of DownloadAipRequest.
func (mr *MockClientMockRecorder) DownloadAipRequest(arg0, arg1 any) *MockClientDownloadAipRequestCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadAipRequest", reflect.TypeOf((*MockClient)(nil).DownloadAipRequest), arg0, arg1)
	return &MockClientDownloadAipRequestCall{Call: call}
}

// MockClientDownloadAipRequestCall wrap *gomock.Call
type MockClientDownloadAipRequestCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientDownloadAipRequestCall) Return(arg0 *storage.DownloadAipRequestResult, arg1 error) *MockClientDownloadAipRequestCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientDownloadAipRequestCall) Do(f func(context.Context, *storage.DownloadAipRequestPayload) (*storage.DownloadAipRequestResult, error)) *MockClientDownloadAipRequestCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientDownloadAipRequestCall) DoAndReturn(f func(context.Context, *storage.DownloadAipRequestPayload) (*storage.DownloadAipRequestResult, error)) *MockClientDownloadAipRequestCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ListAipWorkflows mocks base method.
func (m *MockClient) ListAipWorkflows(arg0 context.Context, arg1 *storage.ListAipWorkflowsPayload) (*storage.AIPWorkflows, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAipWorkflows", arg0, arg1)
	ret0, _ := ret[0].(*storage.AIPWorkflows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAipWorkflows indicates an expected call of ListAipWorkflows.
func (mr *MockClientMockRecorder) ListAipWorkflows(arg0, arg1 any) *MockClientListAipWorkflowsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAipWorkflows", reflect.TypeOf((*MockClient)(nil).ListAipWorkflows), arg0, arg1)
	return &MockClientListAipWorkflowsCall{Call: call}
}

// MockClientListAipWorkflowsCall wrap *gomock.Call
type MockClientListAipWorkflowsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientListAipWorkflowsCall) Return(arg0 *storage.AIPWorkflows, arg1 error) *MockClientListAipWorkflowsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientListAipWorkflowsCall) Do(f func(context.Context, *storage.ListAipWorkflowsPayload) (*storage.AIPWorkflows, error)) *MockClientListAipWorkflowsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientListAipWorkflowsCall) DoAndReturn(f func(context.Context, *storage.ListAipWorkflowsPayload) (*storage.AIPWorkflows, error)) *MockClientListAipWorkflowsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ListAips mocks base method.
func (m *MockClient) ListAips(arg0 context.Context, arg1 *storage.ListAipsPayload) (*storage.AIPs, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAips", arg0, arg1)
	ret0, _ := ret[0].(*storage.AIPs)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAips indicates an expected call of ListAips.
func (mr *MockClientMockRecorder) ListAips(arg0, arg1 any) *MockClientListAipsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAips", reflect.TypeOf((*MockClient)(nil).ListAips), arg0, arg1)
	return &MockClientListAipsCall{Call: call}
}

// MockClientListAipsCall wrap *gomock.Call
type MockClientListAipsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientListAipsCall) Return(arg0 *storage.AIPs, arg1 error) *MockClientListAipsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientListAipsCall) Do(f func(context.Context, *storage.ListAipsPayload) (*storage.AIPs, error)) *MockClientListAipsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientListAipsCall) DoAndReturn(f func(context.Context, *storage.ListAipsPayload) (*storage.AIPs, error)) *MockClientListAipsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ListLocationAips mocks base method.
func (m *MockClient) ListLocationAips(arg0 context.Context, arg1 *storage.ListLocationAipsPayload) (storage.AIPCollection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListLocationAips", arg0, arg1)
	ret0, _ := ret[0].(storage.AIPCollection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListLocationAips indicates an expected call of ListLocationAips.
func (mr *MockClientMockRecorder) ListLocationAips(arg0, arg1 any) *MockClientListLocationAipsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListLocationAips", reflect.TypeOf((*MockClient)(nil).ListLocationAips), arg0, arg1)
	return &MockClientListLocationAipsCall{Call: call}
}

// MockClientListLocationAipsCall wrap *gomock.Call
type MockClientListLocationAipsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientListLocationAipsCall) Return(arg0 storage.AIPCollection, arg1 error) *MockClientListLocationAipsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientListLocationAipsCall) Do(f func(context.Context, *storage.ListLocationAipsPayload) (storage.AIPCollection, error)) *MockClientListLocationAipsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientListLocationAipsCall) DoAndReturn(f func(context.Context, *storage.ListLocationAipsPayload) (storage.AIPCollection, error)) *MockClientListLocationAipsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ListLocations mocks base method.
func (m *MockClient) ListLocations(arg0 context.Context, arg1 *storage.ListLocationsPayload) (storage.LocationCollection, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListLocations", arg0, arg1)
	ret0, _ := ret[0].(storage.LocationCollection)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListLocations indicates an expected call of ListLocations.
func (mr *MockClientMockRecorder) ListLocations(arg0, arg1 any) *MockClientListLocationsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListLocations", reflect.TypeOf((*MockClient)(nil).ListLocations), arg0, arg1)
	return &MockClientListLocationsCall{Call: call}
}

// MockClientListLocationsCall wrap *gomock.Call
type MockClientListLocationsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientListLocationsCall) Return(arg0 storage.LocationCollection, arg1 error) *MockClientListLocationsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientListLocationsCall) Do(f func(context.Context, *storage.ListLocationsPayload) (storage.LocationCollection, error)) *MockClientListLocationsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientListLocationsCall) DoAndReturn(f func(context.Context, *storage.ListLocationsPayload) (storage.LocationCollection, error)) *MockClientListLocationsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Monitor mocks base method.
func (m *MockClient) Monitor(arg0 context.Context, arg1 *storage.MonitorPayload) (storage.MonitorClientStream, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Monitor", arg0, arg1)
	ret0, _ := ret[0].(storage.MonitorClientStream)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Monitor indicates an expected call of Monitor.
func (mr *MockClientMockRecorder) Monitor(arg0, arg1 any) *MockClientMonitorCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Monitor", reflect.TypeOf((*MockClient)(nil).Monitor), arg0, arg1)
	return &MockClientMonitorCall{Call: call}
}

// MockClientMonitorCall wrap *gomock.Call
type MockClientMonitorCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientMonitorCall) Return(arg0 storage.MonitorClientStream, arg1 error) *MockClientMonitorCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientMonitorCall) Do(f func(context.Context, *storage.MonitorPayload) (storage.MonitorClientStream, error)) *MockClientMonitorCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientMonitorCall) DoAndReturn(f func(context.Context, *storage.MonitorPayload) (storage.MonitorClientStream, error)) *MockClientMonitorCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MonitorRequest mocks base method.
func (m *MockClient) MonitorRequest(arg0 context.Context, arg1 *storage.MonitorRequestPayload) (*storage.MonitorRequestResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MonitorRequest", arg0, arg1)
	ret0, _ := ret[0].(*storage.MonitorRequestResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MonitorRequest indicates an expected call of MonitorRequest.
func (mr *MockClientMockRecorder) MonitorRequest(arg0, arg1 any) *MockClientMonitorRequestCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MonitorRequest", reflect.TypeOf((*MockClient)(nil).MonitorRequest), arg0, arg1)
	return &MockClientMonitorRequestCall{Call: call}
}

// MockClientMonitorRequestCall wrap *gomock.Call
type MockClientMonitorRequestCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientMonitorRequestCall) Return(arg0 *storage.MonitorRequestResult, arg1 error) *MockClientMonitorRequestCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientMonitorRequestCall) Do(f func(context.Context, *storage.MonitorRequestPayload) (*storage.MonitorRequestResult, error)) *MockClientMonitorRequestCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientMonitorRequestCall) DoAndReturn(f func(context.Context, *storage.MonitorRequestPayload) (*storage.MonitorRequestResult, error)) *MockClientMonitorRequestCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MoveAip mocks base method.
func (m *MockClient) MoveAip(arg0 context.Context, arg1 *storage.MoveAipPayload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MoveAip", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// MoveAip indicates an expected call of MoveAip.
func (mr *MockClientMockRecorder) MoveAip(arg0, arg1 any) *MockClientMoveAipCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MoveAip", reflect.TypeOf((*MockClient)(nil).MoveAip), arg0, arg1)
	return &MockClientMoveAipCall{Call: call}
}

// MockClientMoveAipCall wrap *gomock.Call
type MockClientMoveAipCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientMoveAipCall) Return(arg0 error) *MockClientMoveAipCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientMoveAipCall) Do(f func(context.Context, *storage.MoveAipPayload) error) *MockClientMoveAipCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientMoveAipCall) DoAndReturn(f func(context.Context, *storage.MoveAipPayload) error) *MockClientMoveAipCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MoveAipStatus mocks base method.
func (m *MockClient) MoveAipStatus(arg0 context.Context, arg1 *storage.MoveAipStatusPayload) (*storage.MoveStatusResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MoveAipStatus", arg0, arg1)
	ret0, _ := ret[0].(*storage.MoveStatusResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MoveAipStatus indicates an expected call of MoveAipStatus.
func (mr *MockClientMockRecorder) MoveAipStatus(arg0, arg1 any) *MockClientMoveAipStatusCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MoveAipStatus", reflect.TypeOf((*MockClient)(nil).MoveAipStatus), arg0, arg1)
	return &MockClientMoveAipStatusCall{Call: call}
}

// MockClientMoveAipStatusCall wrap *gomock.Call
type MockClientMoveAipStatusCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientMoveAipStatusCall) Return(arg0 *storage.MoveStatusResult, arg1 error) *MockClientMoveAipStatusCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientMoveAipStatusCall) Do(f func(context.Context, *storage.MoveAipStatusPayload) (*storage.MoveStatusResult, error)) *MockClientMoveAipStatusCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientMoveAipStatusCall) DoAndReturn(f func(context.Context, *storage.MoveAipStatusPayload) (*storage.MoveStatusResult, error)) *MockClientMoveAipStatusCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RejectAip mocks base method.
func (m *MockClient) RejectAip(arg0 context.Context, arg1 *storage.RejectAipPayload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RejectAip", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RejectAip indicates an expected call of RejectAip.
func (mr *MockClientMockRecorder) RejectAip(arg0, arg1 any) *MockClientRejectAipCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RejectAip", reflect.TypeOf((*MockClient)(nil).RejectAip), arg0, arg1)
	return &MockClientRejectAipCall{Call: call}
}

// MockClientRejectAipCall wrap *gomock.Call
type MockClientRejectAipCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientRejectAipCall) Return(arg0 error) *MockClientRejectAipCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientRejectAipCall) Do(f func(context.Context, *storage.RejectAipPayload) error) *MockClientRejectAipCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientRejectAipCall) DoAndReturn(f func(context.Context, *storage.RejectAipPayload) error) *MockClientRejectAipCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RequestAipDeletion mocks base method.
func (m *MockClient) RequestAipDeletion(arg0 context.Context, arg1 *storage.RequestAipDeletionPayload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestAipDeletion", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RequestAipDeletion indicates an expected call of RequestAipDeletion.
func (mr *MockClientMockRecorder) RequestAipDeletion(arg0, arg1 any) *MockClientRequestAipDeletionCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestAipDeletion", reflect.TypeOf((*MockClient)(nil).RequestAipDeletion), arg0, arg1)
	return &MockClientRequestAipDeletionCall{Call: call}
}

// MockClientRequestAipDeletionCall wrap *gomock.Call
type MockClientRequestAipDeletionCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientRequestAipDeletionCall) Return(arg0 error) *MockClientRequestAipDeletionCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientRequestAipDeletionCall) Do(f func(context.Context, *storage.RequestAipDeletionPayload) error) *MockClientRequestAipDeletionCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientRequestAipDeletionCall) DoAndReturn(f func(context.Context, *storage.RequestAipDeletionPayload) error) *MockClientRequestAipDeletionCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ReviewAipDeletion mocks base method.
func (m *MockClient) ReviewAipDeletion(arg0 context.Context, arg1 *storage.ReviewAipDeletionPayload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReviewAipDeletion", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ReviewAipDeletion indicates an expected call of ReviewAipDeletion.
func (mr *MockClientMockRecorder) ReviewAipDeletion(arg0, arg1 any) *MockClientReviewAipDeletionCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReviewAipDeletion", reflect.TypeOf((*MockClient)(nil).ReviewAipDeletion), arg0, arg1)
	return &MockClientReviewAipDeletionCall{Call: call}
}

// MockClientReviewAipDeletionCall wrap *gomock.Call
type MockClientReviewAipDeletionCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientReviewAipDeletionCall) Return(arg0 error) *MockClientReviewAipDeletionCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientReviewAipDeletionCall) Do(f func(context.Context, *storage.ReviewAipDeletionPayload) error) *MockClientReviewAipDeletionCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientReviewAipDeletionCall) DoAndReturn(f func(context.Context, *storage.ReviewAipDeletionPayload) error) *MockClientReviewAipDeletionCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ShowAip mocks base method.
func (m *MockClient) ShowAip(arg0 context.Context, arg1 *storage.ShowAipPayload) (*storage.AIP, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShowAip", arg0, arg1)
	ret0, _ := ret[0].(*storage.AIP)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShowAip indicates an expected call of ShowAip.
func (mr *MockClientMockRecorder) ShowAip(arg0, arg1 any) *MockClientShowAipCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShowAip", reflect.TypeOf((*MockClient)(nil).ShowAip), arg0, arg1)
	return &MockClientShowAipCall{Call: call}
}

// MockClientShowAipCall wrap *gomock.Call
type MockClientShowAipCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientShowAipCall) Return(arg0 *storage.AIP, arg1 error) *MockClientShowAipCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientShowAipCall) Do(f func(context.Context, *storage.ShowAipPayload) (*storage.AIP, error)) *MockClientShowAipCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientShowAipCall) DoAndReturn(f func(context.Context, *storage.ShowAipPayload) (*storage.AIP, error)) *MockClientShowAipCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ShowLocation mocks base method.
func (m *MockClient) ShowLocation(arg0 context.Context, arg1 *storage.ShowLocationPayload) (*storage.Location, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShowLocation", arg0, arg1)
	ret0, _ := ret[0].(*storage.Location)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShowLocation indicates an expected call of ShowLocation.
func (mr *MockClientMockRecorder) ShowLocation(arg0, arg1 any) *MockClientShowLocationCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShowLocation", reflect.TypeOf((*MockClient)(nil).ShowLocation), arg0, arg1)
	return &MockClientShowLocationCall{Call: call}
}

// MockClientShowLocationCall wrap *gomock.Call
type MockClientShowLocationCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientShowLocationCall) Return(arg0 *storage.Location, arg1 error) *MockClientShowLocationCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientShowLocationCall) Do(f func(context.Context, *storage.ShowLocationPayload) (*storage.Location, error)) *MockClientShowLocationCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientShowLocationCall) DoAndReturn(f func(context.Context, *storage.ShowLocationPayload) (*storage.Location, error)) *MockClientShowLocationCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SubmitAip mocks base method.
func (m *MockClient) SubmitAip(arg0 context.Context, arg1 *storage.SubmitAipPayload) (*storage.SubmitAIPResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitAip", arg0, arg1)
	ret0, _ := ret[0].(*storage.SubmitAIPResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubmitAip indicates an expected call of SubmitAip.
func (mr *MockClientMockRecorder) SubmitAip(arg0, arg1 any) *MockClientSubmitAipCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitAip", reflect.TypeOf((*MockClient)(nil).SubmitAip), arg0, arg1)
	return &MockClientSubmitAipCall{Call: call}
}

// MockClientSubmitAipCall wrap *gomock.Call
type MockClientSubmitAipCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientSubmitAipCall) Return(arg0 *storage.SubmitAIPResult, arg1 error) *MockClientSubmitAipCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientSubmitAipCall) Do(f func(context.Context, *storage.SubmitAipPayload) (*storage.SubmitAIPResult, error)) *MockClientSubmitAipCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientSubmitAipCall) DoAndReturn(f func(context.Context, *storage.SubmitAipPayload) (*storage.SubmitAIPResult, error)) *MockClientSubmitAipCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// UpdateAip mocks base method.
func (m *MockClient) UpdateAip(arg0 context.Context, arg1 *storage.UpdateAipPayload) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAip", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAip indicates an expected call of UpdateAip.
func (mr *MockClientMockRecorder) UpdateAip(arg0, arg1 any) *MockClientUpdateAipCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAip", reflect.TypeOf((*MockClient)(nil).UpdateAip), arg0, arg1)
	return &MockClientUpdateAipCall{Call: call}
}

// MockClientUpdateAipCall wrap *gomock.Call
type MockClientUpdateAipCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockClientUpdateAipCall) Return(arg0 error) *MockClientUpdateAipCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockClientUpdateAipCall) Do(f func(context.Context, *storage.UpdateAipPayload) error) *MockClientUpdateAipCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockClientUpdateAipCall) DoAndReturn(f func(context.Context, *storage.UpdateAipPayload) error) *MockClientUpdateAipCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
