// Code generated by goa v3.15.2, DO NOT EDIT.
//
// package service
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package package_

import (
	"context"

	package_views "github.com/artefactual-sdps/enduro/internal/api/gen/package_/views"
	"github.com/google/uuid"
	goa "goa.design/goa/v3/pkg"
	"goa.design/goa/v3/security"
)

// The package service manages packages being transferred to a3m.
type Service interface {
	// Request access to the /monitor WebSocket.
	MonitorRequest(context.Context, *MonitorRequestPayload) (res *MonitorRequestResult, err error)
	// Monitor implements monitor.
	Monitor(context.Context, *MonitorPayload, MonitorServerStream) (err error)
	// List all stored packages
	List(context.Context, *ListPayload) (res *ListResult, err error)
	// Show package by ID
	Show(context.Context, *ShowPayload) (res *EnduroStoredPackage, err error)
	// List all preservation actions by ID
	PreservationActions(context.Context, *PreservationActionsPayload) (res *EnduroPackagePreservationActions, err error)
	// Signal the package has been reviewed and accepted
	Confirm(context.Context, *ConfirmPayload) (err error)
	// Signal the package has been reviewed and rejected
	Reject(context.Context, *RejectPayload) (err error)
	// Move a package to a permanent storage location
	Move(context.Context, *MovePayload) (err error)
	// Retrieve the status of a permanent storage location move of the package
	MoveStatus(context.Context, *MoveStatusPayload) (res *MoveStatusResult, err error)
}

// Auther defines the authorization functions to be implemented by the service.
type Auther interface {
	// OAuth2Auth implements the authorization logic for the OAuth2 security scheme.
	OAuth2Auth(ctx context.Context, token string, schema *security.OAuth2Scheme) (context.Context, error)
}

// APIName is the name of the API as defined in the design.
const APIName = "enduro"

// APIVersion is the version of the API as defined in the design.
const APIVersion = "0.0.1"

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "package"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [9]string{"monitor_request", "monitor", "list", "show", "preservation_actions", "confirm", "reject", "move", "move_status"}

// MonitorServerStream is the interface a "monitor" endpoint server stream must
// satisfy.
type MonitorServerStream interface {
	// Send streams instances of "MonitorEvent".
	Send(*MonitorEvent) error
	// Close closes the stream.
	Close() error
}

// MonitorClientStream is the interface a "monitor" endpoint client stream must
// satisfy.
type MonitorClientStream interface {
	// Recv reads instances of "MonitorEvent" from the stream.
	Recv() (*MonitorEvent, error)
}

// ConfirmPayload is the payload type of the package service confirm method.
type ConfirmPayload struct {
	// Identifier of package to look up
	ID uint
	// Identifier of storage location
	LocationID uuid.UUID
	OauthToken *string
}

// PreservationAction describes a preservation action.
type EnduroPackagePreservationAction struct {
	ID          uint
	WorkflowID  string
	Type        string
	Status      string
	StartedAt   string
	CompletedAt *string
	Tasks       EnduroPackagePreservationTaskCollection
	PackageID   *uint
}

type EnduroPackagePreservationActionCollection []*EnduroPackagePreservationAction

// EnduroPackagePreservationActions is the result type of the package service
// preservation_actions method.
type EnduroPackagePreservationActions struct {
	Actions EnduroPackagePreservationActionCollection
}

// PreservationTask describes a preservation action task.
type EnduroPackagePreservationTask struct {
	ID                   uint
	TaskID               string
	Name                 string
	Status               string
	StartedAt            string
	CompletedAt          *string
	Note                 *string
	PreservationActionID *uint
}

type EnduroPackagePreservationTaskCollection []*EnduroPackagePreservationTask

// EnduroStoredPackage is the result type of the package service show method.
type EnduroStoredPackage struct {
	// Identifier of package
	ID uint
	// Name of the package
	Name *string
	// Identifier of storage location
	LocationID *uuid.UUID
	// Status of the package
	Status string
	// Identifier of processing workflow
	WorkflowID *string
	// Identifier of latest processing workflow run
	RunID *string
	// Identifier of AIP
	AipID *string
	// Creation datetime
	CreatedAt string
	// Start datetime
	StartedAt *string
	// Completion datetime
	CompletedAt *string
}

type EnduroStoredPackageCollection []*EnduroStoredPackage

// ListPayload is the payload type of the package service list method.
type ListPayload struct {
	Name *string
	// Identifier of AIP
	AipID               *string
	EarliestCreatedTime *string
	LatestCreatedTime   *string
	// Identifier of storage location
	LocationID *string
	Status     *string
	// Pagination cursor
	Cursor     *string
	OauthToken *string
}

// ListResult is the result type of the package service list method.
type ListResult struct {
	Items      EnduroStoredPackageCollection
	NextCursor *string
}

// MonitorEvent is the result type of the package service monitor method.
type MonitorEvent struct {
	Event interface {
		eventVal()
	}
}

// MonitorPayload is the payload type of the package service monitor method.
type MonitorPayload struct {
	Ticket *string
}

type MonitorPingEvent struct {
	Message *string
}

// MonitorRequestPayload is the payload type of the package service
// monitor_request method.
type MonitorRequestPayload struct {
	OauthToken *string
}

// MonitorRequestResult is the result type of the package service
// monitor_request method.
type MonitorRequestResult struct {
	Ticket *string
}

// MovePayload is the payload type of the package service move method.
type MovePayload struct {
	// Identifier of package to move
	ID uint
	// Identifier of storage location
	LocationID uuid.UUID
	OauthToken *string
}

// MoveStatusPayload is the payload type of the package service move_status
// method.
type MoveStatusPayload struct {
	// Identifier of package to move
	ID         uint
	OauthToken *string
}

// MoveStatusResult is the result type of the package service move_status
// method.
type MoveStatusResult struct {
	Done bool
}

type PackageCreatedEvent struct {
	// Identifier of package
	ID   uint
	Item *EnduroStoredPackage
}

type PackageLocationUpdatedEvent struct {
	// Identifier of package
	ID uint
	// Identifier of storage location
	LocationID uuid.UUID
}

// Package not found.
type PackageNotFound struct {
	// Message of error
	Message string
	// Identifier of missing package
	ID uint
}

type PackageStatusUpdatedEvent struct {
	// Identifier of package
	ID     uint
	Status string
}

type PackageUpdatedEvent struct {
	// Identifier of package
	ID   uint
	Item *EnduroStoredPackage
}

type PreservationActionCreatedEvent struct {
	// Identifier of preservation action
	ID   uint
	Item *EnduroPackagePreservationAction
}

type PreservationActionUpdatedEvent struct {
	// Identifier of preservation action
	ID   uint
	Item *EnduroPackagePreservationAction
}

// PreservationActionsPayload is the payload type of the package service
// preservation_actions method.
type PreservationActionsPayload struct {
	// Identifier of package to look up
	ID         uint
	OauthToken *string
}

type PreservationTaskCreatedEvent struct {
	// Identifier of preservation task
	ID   uint
	Item *EnduroPackagePreservationTask
}

type PreservationTaskUpdatedEvent struct {
	// Identifier of preservation task
	ID   uint
	Item *EnduroPackagePreservationTask
}

// RejectPayload is the payload type of the package service reject method.
type RejectPayload struct {
	// Identifier of package to look up
	ID         uint
	OauthToken *string
}

// ShowPayload is the payload type of the package service show method.
type ShowPayload struct {
	// Identifier of package to show
	ID         uint
	OauthToken *string
}

// Invalid token
type Unauthorized string

// Error returns an error description.
func (e *PackageNotFound) Error() string {
	return "Package not found."
}

// ErrorName returns "PackageNotFound".
//
// Deprecated: Use GoaErrorName - https://github.com/goadesign/goa/issues/3105
func (e *PackageNotFound) ErrorName() string {
	return e.GoaErrorName()
}

// GoaErrorName returns "PackageNotFound".
func (e *PackageNotFound) GoaErrorName() string {
	return "not_found"
}

// Error returns an error description.
func (e Unauthorized) Error() string {
	return "Invalid token"
}

// ErrorName returns "unauthorized".
//
// Deprecated: Use GoaErrorName - https://github.com/goadesign/goa/issues/3105
func (e Unauthorized) ErrorName() string {
	return e.GoaErrorName()
}

// GoaErrorName returns "unauthorized".
func (e Unauthorized) GoaErrorName() string {
	return "unauthorized"
}
func (*MonitorPingEvent) eventVal()               {}
func (*PackageCreatedEvent) eventVal()            {}
func (*PackageLocationUpdatedEvent) eventVal()    {}
func (*PackageStatusUpdatedEvent) eventVal()      {}
func (*PackageUpdatedEvent) eventVal()            {}
func (*PreservationActionCreatedEvent) eventVal() {}
func (*PreservationActionUpdatedEvent) eventVal() {}
func (*PreservationTaskCreatedEvent) eventVal()   {}
func (*PreservationTaskUpdatedEvent) eventVal()   {}

// MakeNotAvailable builds a goa.ServiceError from an error.
func MakeNotAvailable(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "not_available", false, false, false)
}

// MakeNotValid builds a goa.ServiceError from an error.
func MakeNotValid(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "not_valid", false, false, false)
}

// MakeFailedDependency builds a goa.ServiceError from an error.
func MakeFailedDependency(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "failed_dependency", false, false, false)
}

// NewEnduroStoredPackage initializes result type EnduroStoredPackage from
// viewed result type EnduroStoredPackage.
func NewEnduroStoredPackage(vres *package_views.EnduroStoredPackage) *EnduroStoredPackage {
	return newEnduroStoredPackage(vres.Projected)
}

// NewViewedEnduroStoredPackage initializes viewed result type
// EnduroStoredPackage from result type EnduroStoredPackage using the given
// view.
func NewViewedEnduroStoredPackage(res *EnduroStoredPackage, view string) *package_views.EnduroStoredPackage {
	p := newEnduroStoredPackageView(res)
	return &package_views.EnduroStoredPackage{Projected: p, View: "default"}
}

// NewEnduroPackagePreservationActions initializes result type
// EnduroPackagePreservationActions from viewed result type
// EnduroPackagePreservationActions.
func NewEnduroPackagePreservationActions(vres *package_views.EnduroPackagePreservationActions) *EnduroPackagePreservationActions {
	return newEnduroPackagePreservationActions(vres.Projected)
}

// NewViewedEnduroPackagePreservationActions initializes viewed result type
// EnduroPackagePreservationActions from result type
// EnduroPackagePreservationActions using the given view.
func NewViewedEnduroPackagePreservationActions(res *EnduroPackagePreservationActions, view string) *package_views.EnduroPackagePreservationActions {
	p := newEnduroPackagePreservationActionsView(res)
	return &package_views.EnduroPackagePreservationActions{Projected: p, View: "default"}
}

// newEnduroStoredPackage converts projected type EnduroStoredPackage to
// service type EnduroStoredPackage.
func newEnduroStoredPackage(vres *package_views.EnduroStoredPackageView) *EnduroStoredPackage {
	res := &EnduroStoredPackage{
		Name:        vres.Name,
		LocationID:  vres.LocationID,
		WorkflowID:  vres.WorkflowID,
		RunID:       vres.RunID,
		AipID:       vres.AipID,
		StartedAt:   vres.StartedAt,
		CompletedAt: vres.CompletedAt,
	}
	if vres.ID != nil {
		res.ID = *vres.ID
	}
	if vres.Status != nil {
		res.Status = *vres.Status
	}
	if vres.CreatedAt != nil {
		res.CreatedAt = *vres.CreatedAt
	}
	if vres.Status == nil {
		res.Status = "new"
	}
	return res
}

// newEnduroStoredPackageView projects result type EnduroStoredPackage to
// projected type EnduroStoredPackageView using the "default" view.
func newEnduroStoredPackageView(res *EnduroStoredPackage) *package_views.EnduroStoredPackageView {
	vres := &package_views.EnduroStoredPackageView{
		ID:          &res.ID,
		Name:        res.Name,
		LocationID:  res.LocationID,
		Status:      &res.Status,
		WorkflowID:  res.WorkflowID,
		RunID:       res.RunID,
		AipID:       res.AipID,
		CreatedAt:   &res.CreatedAt,
		StartedAt:   res.StartedAt,
		CompletedAt: res.CompletedAt,
	}
	return vres
}

// newEnduroPackagePreservationActions converts projected type
// EnduroPackagePreservationActions to service type
// EnduroPackagePreservationActions.
func newEnduroPackagePreservationActions(vres *package_views.EnduroPackagePreservationActionsView) *EnduroPackagePreservationActions {
	res := &EnduroPackagePreservationActions{}
	if vres.Actions != nil {
		res.Actions = newEnduroPackagePreservationActionCollection(vres.Actions)
	}
	return res
}

// newEnduroPackagePreservationActionsView projects result type
// EnduroPackagePreservationActions to projected type
// EnduroPackagePreservationActionsView using the "default" view.
func newEnduroPackagePreservationActionsView(res *EnduroPackagePreservationActions) *package_views.EnduroPackagePreservationActionsView {
	vres := &package_views.EnduroPackagePreservationActionsView{}
	if res.Actions != nil {
		vres.Actions = newEnduroPackagePreservationActionCollectionView(res.Actions)
	}
	return vres
}

// newEnduroPackagePreservationActionCollectionSimple converts projected type
// EnduroPackagePreservationActionCollection to service type
// EnduroPackagePreservationActionCollection.
func newEnduroPackagePreservationActionCollectionSimple(vres package_views.EnduroPackagePreservationActionCollectionView) EnduroPackagePreservationActionCollection {
	res := make(EnduroPackagePreservationActionCollection, len(vres))
	for i, n := range vres {
		res[i] = newEnduroPackagePreservationActionSimple(n)
	}
	return res
}

// newEnduroPackagePreservationActionCollection converts projected type
// EnduroPackagePreservationActionCollection to service type
// EnduroPackagePreservationActionCollection.
func newEnduroPackagePreservationActionCollection(vres package_views.EnduroPackagePreservationActionCollectionView) EnduroPackagePreservationActionCollection {
	res := make(EnduroPackagePreservationActionCollection, len(vres))
	for i, n := range vres {
		res[i] = newEnduroPackagePreservationAction(n)
	}
	return res
}

// newEnduroPackagePreservationActionCollectionViewSimple projects result type
// EnduroPackagePreservationActionCollection to projected type
// EnduroPackagePreservationActionCollectionView using the "simple" view.
func newEnduroPackagePreservationActionCollectionViewSimple(res EnduroPackagePreservationActionCollection) package_views.EnduroPackagePreservationActionCollectionView {
	vres := make(package_views.EnduroPackagePreservationActionCollectionView, len(res))
	for i, n := range res {
		vres[i] = newEnduroPackagePreservationActionViewSimple(n)
	}
	return vres
}

// newEnduroPackagePreservationActionCollectionView projects result type
// EnduroPackagePreservationActionCollection to projected type
// EnduroPackagePreservationActionCollectionView using the "default" view.
func newEnduroPackagePreservationActionCollectionView(res EnduroPackagePreservationActionCollection) package_views.EnduroPackagePreservationActionCollectionView {
	vres := make(package_views.EnduroPackagePreservationActionCollectionView, len(res))
	for i, n := range res {
		vres[i] = newEnduroPackagePreservationActionView(n)
	}
	return vres
}

// newEnduroPackagePreservationActionSimple converts projected type
// EnduroPackagePreservationAction to service type
// EnduroPackagePreservationAction.
func newEnduroPackagePreservationActionSimple(vres *package_views.EnduroPackagePreservationActionView) *EnduroPackagePreservationAction {
	res := &EnduroPackagePreservationAction{
		CompletedAt: vres.CompletedAt,
		PackageID:   vres.PackageID,
	}
	if vres.ID != nil {
		res.ID = *vres.ID
	}
	if vres.WorkflowID != nil {
		res.WorkflowID = *vres.WorkflowID
	}
	if vres.Type != nil {
		res.Type = *vres.Type
	}
	if vres.Status != nil {
		res.Status = *vres.Status
	}
	if vres.StartedAt != nil {
		res.StartedAt = *vres.StartedAt
	}
	if vres.Tasks != nil {
		res.Tasks = newEnduroPackagePreservationTaskCollection(vres.Tasks)
	}
	return res
}

// newEnduroPackagePreservationAction converts projected type
// EnduroPackagePreservationAction to service type
// EnduroPackagePreservationAction.
func newEnduroPackagePreservationAction(vres *package_views.EnduroPackagePreservationActionView) *EnduroPackagePreservationAction {
	res := &EnduroPackagePreservationAction{
		CompletedAt: vres.CompletedAt,
		PackageID:   vres.PackageID,
	}
	if vres.ID != nil {
		res.ID = *vres.ID
	}
	if vres.WorkflowID != nil {
		res.WorkflowID = *vres.WorkflowID
	}
	if vres.Type != nil {
		res.Type = *vres.Type
	}
	if vres.Status != nil {
		res.Status = *vres.Status
	}
	if vres.StartedAt != nil {
		res.StartedAt = *vres.StartedAt
	}
	if vres.Tasks != nil {
		res.Tasks = newEnduroPackagePreservationTaskCollection(vres.Tasks)
	}
	return res
}

// newEnduroPackagePreservationActionViewSimple projects result type
// EnduroPackagePreservationAction to projected type
// EnduroPackagePreservationActionView using the "simple" view.
func newEnduroPackagePreservationActionViewSimple(res *EnduroPackagePreservationAction) *package_views.EnduroPackagePreservationActionView {
	vres := &package_views.EnduroPackagePreservationActionView{
		ID:          &res.ID,
		WorkflowID:  &res.WorkflowID,
		Type:        &res.Type,
		Status:      &res.Status,
		StartedAt:   &res.StartedAt,
		CompletedAt: res.CompletedAt,
		PackageID:   res.PackageID,
	}
	return vres
}

// newEnduroPackagePreservationActionView projects result type
// EnduroPackagePreservationAction to projected type
// EnduroPackagePreservationActionView using the "default" view.
func newEnduroPackagePreservationActionView(res *EnduroPackagePreservationAction) *package_views.EnduroPackagePreservationActionView {
	vres := &package_views.EnduroPackagePreservationActionView{
		ID:          &res.ID,
		WorkflowID:  &res.WorkflowID,
		Type:        &res.Type,
		Status:      &res.Status,
		StartedAt:   &res.StartedAt,
		CompletedAt: res.CompletedAt,
		PackageID:   res.PackageID,
	}
	if res.Tasks != nil {
		vres.Tasks = newEnduroPackagePreservationTaskCollectionView(res.Tasks)
	}
	return vres
}

// newEnduroPackagePreservationTaskCollection converts projected type
// EnduroPackagePreservationTaskCollection to service type
// EnduroPackagePreservationTaskCollection.
func newEnduroPackagePreservationTaskCollection(vres package_views.EnduroPackagePreservationTaskCollectionView) EnduroPackagePreservationTaskCollection {
	res := make(EnduroPackagePreservationTaskCollection, len(vres))
	for i, n := range vres {
		res[i] = newEnduroPackagePreservationTask(n)
	}
	return res
}

// newEnduroPackagePreservationTaskCollectionView projects result type
// EnduroPackagePreservationTaskCollection to projected type
// EnduroPackagePreservationTaskCollectionView using the "default" view.
func newEnduroPackagePreservationTaskCollectionView(res EnduroPackagePreservationTaskCollection) package_views.EnduroPackagePreservationTaskCollectionView {
	vres := make(package_views.EnduroPackagePreservationTaskCollectionView, len(res))
	for i, n := range res {
		vres[i] = newEnduroPackagePreservationTaskView(n)
	}
	return vres
}

// newEnduroPackagePreservationTask converts projected type
// EnduroPackagePreservationTask to service type EnduroPackagePreservationTask.
func newEnduroPackagePreservationTask(vres *package_views.EnduroPackagePreservationTaskView) *EnduroPackagePreservationTask {
	res := &EnduroPackagePreservationTask{
		CompletedAt:          vres.CompletedAt,
		Note:                 vres.Note,
		PreservationActionID: vres.PreservationActionID,
	}
	if vres.ID != nil {
		res.ID = *vres.ID
	}
	if vres.TaskID != nil {
		res.TaskID = *vres.TaskID
	}
	if vres.Name != nil {
		res.Name = *vres.Name
	}
	if vres.Status != nil {
		res.Status = *vres.Status
	}
	if vres.StartedAt != nil {
		res.StartedAt = *vres.StartedAt
	}
	return res
}

// newEnduroPackagePreservationTaskView projects result type
// EnduroPackagePreservationTask to projected type
// EnduroPackagePreservationTaskView using the "default" view.
func newEnduroPackagePreservationTaskView(res *EnduroPackagePreservationTask) *package_views.EnduroPackagePreservationTaskView {
	vres := &package_views.EnduroPackagePreservationTaskView{
		ID:                   &res.ID,
		TaskID:               &res.TaskID,
		Name:                 &res.Name,
		Status:               &res.Status,
		StartedAt:            &res.StartedAt,
		CompletedAt:          res.CompletedAt,
		Note:                 res.Note,
		PreservationActionID: res.PreservationActionID,
	}
	return vres
}
