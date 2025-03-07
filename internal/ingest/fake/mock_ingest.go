// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/artefactual-sdps/enduro/internal/ingest (interfaces: Service)
//
// Generated by this command:
//
//	mockgen -typed -destination=./internal/ingest/fake/mock_ingest.go -package=fake github.com/artefactual-sdps/enduro/internal/ingest Service
//

// Package fake is a generated GoMock package.
package fake

import (
	context "context"
	reflect "reflect"
	time "time"

	ingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	datatypes "github.com/artefactual-sdps/enduro/internal/datatypes"
	enums "github.com/artefactual-sdps/enduro/internal/enums"
	uuid "github.com/google/uuid"
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

// CompletePreservationAction mocks base method.
func (m *MockService) CompletePreservationAction(arg0 context.Context, arg1 int, arg2 enums.PreservationActionStatus, arg3 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompletePreservationAction", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// CompletePreservationAction indicates an expected call of CompletePreservationAction.
func (mr *MockServiceMockRecorder) CompletePreservationAction(arg0, arg1, arg2, arg3 any) *MockServiceCompletePreservationActionCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompletePreservationAction", reflect.TypeOf((*MockService)(nil).CompletePreservationAction), arg0, arg1, arg2, arg3)
	return &MockServiceCompletePreservationActionCall{Call: call}
}

// MockServiceCompletePreservationActionCall wrap *gomock.Call
type MockServiceCompletePreservationActionCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceCompletePreservationActionCall) Return(arg0 error) *MockServiceCompletePreservationActionCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceCompletePreservationActionCall) Do(f func(context.Context, int, enums.PreservationActionStatus, time.Time) error) *MockServiceCompletePreservationActionCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceCompletePreservationActionCall) DoAndReturn(f func(context.Context, int, enums.PreservationActionStatus, time.Time) error) *MockServiceCompletePreservationActionCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// CompletePreservationTask mocks base method.
func (m *MockService) CompletePreservationTask(arg0 context.Context, arg1 int, arg2 enums.PreservationTaskStatus, arg3 time.Time, arg4 *string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CompletePreservationTask", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// CompletePreservationTask indicates an expected call of CompletePreservationTask.
func (mr *MockServiceMockRecorder) CompletePreservationTask(arg0, arg1, arg2, arg3, arg4 any) *MockServiceCompletePreservationTaskCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CompletePreservationTask", reflect.TypeOf((*MockService)(nil).CompletePreservationTask), arg0, arg1, arg2, arg3, arg4)
	return &MockServiceCompletePreservationTaskCall{Call: call}
}

// MockServiceCompletePreservationTaskCall wrap *gomock.Call
type MockServiceCompletePreservationTaskCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceCompletePreservationTaskCall) Return(arg0 error) *MockServiceCompletePreservationTaskCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceCompletePreservationTaskCall) Do(f func(context.Context, int, enums.PreservationTaskStatus, time.Time, *string) error) *MockServiceCompletePreservationTaskCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceCompletePreservationTaskCall) DoAndReturn(f func(context.Context, int, enums.PreservationTaskStatus, time.Time, *string) error) *MockServiceCompletePreservationTaskCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Create mocks base method.
func (m *MockService) Create(arg0 context.Context, arg1 *datatypes.SIP) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockServiceMockRecorder) Create(arg0, arg1 any) *MockServiceCreateCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockService)(nil).Create), arg0, arg1)
	return &MockServiceCreateCall{Call: call}
}

// MockServiceCreateCall wrap *gomock.Call
type MockServiceCreateCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceCreateCall) Return(arg0 error) *MockServiceCreateCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceCreateCall) Do(f func(context.Context, *datatypes.SIP) error) *MockServiceCreateCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceCreateCall) DoAndReturn(f func(context.Context, *datatypes.SIP) error) *MockServiceCreateCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// CreatePreservationAction mocks base method.
func (m *MockService) CreatePreservationAction(arg0 context.Context, arg1 *datatypes.PreservationAction) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePreservationAction", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreatePreservationAction indicates an expected call of CreatePreservationAction.
func (mr *MockServiceMockRecorder) CreatePreservationAction(arg0, arg1 any) *MockServiceCreatePreservationActionCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePreservationAction", reflect.TypeOf((*MockService)(nil).CreatePreservationAction), arg0, arg1)
	return &MockServiceCreatePreservationActionCall{Call: call}
}

// MockServiceCreatePreservationActionCall wrap *gomock.Call
type MockServiceCreatePreservationActionCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceCreatePreservationActionCall) Return(arg0 error) *MockServiceCreatePreservationActionCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceCreatePreservationActionCall) Do(f func(context.Context, *datatypes.PreservationAction) error) *MockServiceCreatePreservationActionCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceCreatePreservationActionCall) DoAndReturn(f func(context.Context, *datatypes.PreservationAction) error) *MockServiceCreatePreservationActionCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// CreatePreservationTask mocks base method.
func (m *MockService) CreatePreservationTask(arg0 context.Context, arg1 *datatypes.PreservationTask) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePreservationTask", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreatePreservationTask indicates an expected call of CreatePreservationTask.
func (mr *MockServiceMockRecorder) CreatePreservationTask(arg0, arg1 any) *MockServiceCreatePreservationTaskCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePreservationTask", reflect.TypeOf((*MockService)(nil).CreatePreservationTask), arg0, arg1)
	return &MockServiceCreatePreservationTaskCall{Call: call}
}

// MockServiceCreatePreservationTaskCall wrap *gomock.Call
type MockServiceCreatePreservationTaskCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceCreatePreservationTaskCall) Return(arg0 error) *MockServiceCreatePreservationTaskCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceCreatePreservationTaskCall) Do(f func(context.Context, *datatypes.PreservationTask) error) *MockServiceCreatePreservationTaskCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceCreatePreservationTaskCall) DoAndReturn(f func(context.Context, *datatypes.PreservationTask) error) *MockServiceCreatePreservationTaskCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Goa mocks base method.
func (m *MockService) Goa() ingest.Service {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Goa")
	ret0, _ := ret[0].(ingest.Service)
	return ret0
}

// Goa indicates an expected call of Goa.
func (mr *MockServiceMockRecorder) Goa() *MockServiceGoaCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Goa", reflect.TypeOf((*MockService)(nil).Goa))
	return &MockServiceGoaCall{Call: call}
}

// MockServiceGoaCall wrap *gomock.Call
type MockServiceGoaCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceGoaCall) Return(arg0 ingest.Service) *MockServiceGoaCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceGoaCall) Do(f func() ingest.Service) *MockServiceGoaCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceGoaCall) DoAndReturn(f func() ingest.Service) *MockServiceGoaCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SetLocationID mocks base method.
func (m *MockService) SetLocationID(arg0 context.Context, arg1 int, arg2 uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetLocationID", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetLocationID indicates an expected call of SetLocationID.
func (mr *MockServiceMockRecorder) SetLocationID(arg0, arg1, arg2 any) *MockServiceSetLocationIDCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLocationID", reflect.TypeOf((*MockService)(nil).SetLocationID), arg0, arg1, arg2)
	return &MockServiceSetLocationIDCall{Call: call}
}

// MockServiceSetLocationIDCall wrap *gomock.Call
type MockServiceSetLocationIDCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceSetLocationIDCall) Return(arg0 error) *MockServiceSetLocationIDCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceSetLocationIDCall) Do(f func(context.Context, int, uuid.UUID) error) *MockServiceSetLocationIDCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceSetLocationIDCall) DoAndReturn(f func(context.Context, int, uuid.UUID) error) *MockServiceSetLocationIDCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SetPreservationActionStatus mocks base method.
func (m *MockService) SetPreservationActionStatus(arg0 context.Context, arg1 int, arg2 enums.PreservationActionStatus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetPreservationActionStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetPreservationActionStatus indicates an expected call of SetPreservationActionStatus.
func (mr *MockServiceMockRecorder) SetPreservationActionStatus(arg0, arg1, arg2 any) *MockServiceSetPreservationActionStatusCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPreservationActionStatus", reflect.TypeOf((*MockService)(nil).SetPreservationActionStatus), arg0, arg1, arg2)
	return &MockServiceSetPreservationActionStatusCall{Call: call}
}

// MockServiceSetPreservationActionStatusCall wrap *gomock.Call
type MockServiceSetPreservationActionStatusCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceSetPreservationActionStatusCall) Return(arg0 error) *MockServiceSetPreservationActionStatusCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceSetPreservationActionStatusCall) Do(f func(context.Context, int, enums.PreservationActionStatus) error) *MockServiceSetPreservationActionStatusCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceSetPreservationActionStatusCall) DoAndReturn(f func(context.Context, int, enums.PreservationActionStatus) error) *MockServiceSetPreservationActionStatusCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SetStatus mocks base method.
func (m *MockService) SetStatus(arg0 context.Context, arg1 int, arg2 enums.SIPStatus) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetStatus indicates an expected call of SetStatus.
func (mr *MockServiceMockRecorder) SetStatus(arg0, arg1, arg2 any) *MockServiceSetStatusCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStatus", reflect.TypeOf((*MockService)(nil).SetStatus), arg0, arg1, arg2)
	return &MockServiceSetStatusCall{Call: call}
}

// MockServiceSetStatusCall wrap *gomock.Call
type MockServiceSetStatusCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceSetStatusCall) Return(arg0 error) *MockServiceSetStatusCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceSetStatusCall) Do(f func(context.Context, int, enums.SIPStatus) error) *MockServiceSetStatusCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceSetStatusCall) DoAndReturn(f func(context.Context, int, enums.SIPStatus) error) *MockServiceSetStatusCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SetStatusInProgress mocks base method.
func (m *MockService) SetStatusInProgress(arg0 context.Context, arg1 int, arg2 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetStatusInProgress", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetStatusInProgress indicates an expected call of SetStatusInProgress.
func (mr *MockServiceMockRecorder) SetStatusInProgress(arg0, arg1, arg2 any) *MockServiceSetStatusInProgressCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStatusInProgress", reflect.TypeOf((*MockService)(nil).SetStatusInProgress), arg0, arg1, arg2)
	return &MockServiceSetStatusInProgressCall{Call: call}
}

// MockServiceSetStatusInProgressCall wrap *gomock.Call
type MockServiceSetStatusInProgressCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceSetStatusInProgressCall) Return(arg0 error) *MockServiceSetStatusInProgressCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceSetStatusInProgressCall) Do(f func(context.Context, int, time.Time) error) *MockServiceSetStatusInProgressCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceSetStatusInProgressCall) DoAndReturn(f func(context.Context, int, time.Time) error) *MockServiceSetStatusInProgressCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SetStatusPending mocks base method.
func (m *MockService) SetStatusPending(arg0 context.Context, arg1 int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetStatusPending", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetStatusPending indicates an expected call of SetStatusPending.
func (mr *MockServiceMockRecorder) SetStatusPending(arg0, arg1 any) *MockServiceSetStatusPendingCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStatusPending", reflect.TypeOf((*MockService)(nil).SetStatusPending), arg0, arg1)
	return &MockServiceSetStatusPendingCall{Call: call}
}

// MockServiceSetStatusPendingCall wrap *gomock.Call
type MockServiceSetStatusPendingCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceSetStatusPendingCall) Return(arg0 error) *MockServiceSetStatusPendingCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceSetStatusPendingCall) Do(f func(context.Context, int) error) *MockServiceSetStatusPendingCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceSetStatusPendingCall) DoAndReturn(f func(context.Context, int) error) *MockServiceSetStatusPendingCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// UpdateWorkflowStatus mocks base method.
func (m *MockService) UpdateWorkflowStatus(arg0 context.Context, arg1 int, arg2, arg3, arg4, arg5 string, arg6 enums.SIPStatus, arg7 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateWorkflowStatus", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateWorkflowStatus indicates an expected call of UpdateWorkflowStatus.
func (mr *MockServiceMockRecorder) UpdateWorkflowStatus(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 any) *MockServiceUpdateWorkflowStatusCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateWorkflowStatus", reflect.TypeOf((*MockService)(nil).UpdateWorkflowStatus), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	return &MockServiceUpdateWorkflowStatusCall{Call: call}
}

// MockServiceUpdateWorkflowStatusCall wrap *gomock.Call
type MockServiceUpdateWorkflowStatusCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockServiceUpdateWorkflowStatusCall) Return(arg0 error) *MockServiceUpdateWorkflowStatusCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockServiceUpdateWorkflowStatusCall) Do(f func(context.Context, int, string, string, string, string, enums.SIPStatus, time.Time) error) *MockServiceUpdateWorkflowStatusCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockServiceUpdateWorkflowStatusCall) DoAndReturn(f func(context.Context, int, string, string, string, string, enums.SIPStatus, time.Time) error) *MockServiceUpdateWorkflowStatusCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
