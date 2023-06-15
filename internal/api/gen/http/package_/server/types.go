// Code generated by goa v3.11.3, DO NOT EDIT.
//
// package HTTP server types
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package server

import (
	package_ "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	package_views "github.com/artefactual-sdps/enduro/internal/api/gen/package_/views"
	"github.com/google/uuid"
	goa "goa.design/goa/v3/pkg"
)

// ConfirmRequestBody is the type of the "package" service "confirm" endpoint
// HTTP request body.
type ConfirmRequestBody struct {
	// Identifier of storage location
	LocationID *uuid.UUID `form:"location_id,omitempty" json:"location_id,omitempty" xml:"location_id,omitempty"`
}

// MoveRequestBody is the type of the "package" service "move" endpoint HTTP
// request body.
type MoveRequestBody struct {
	// Identifier of storage location
	LocationID *uuid.UUID `form:"location_id,omitempty" json:"location_id,omitempty" xml:"location_id,omitempty"`
}

// MonitorResponseBody is the type of the "package" service "monitor" endpoint
// HTTP response body.
type MonitorResponseBody struct {
	MonitorPingEvent               *EnduroMonitorPingEventResponseBody               `form:"monitor_ping_event,omitempty" json:"monitor_ping_event,omitempty" xml:"monitor_ping_event,omitempty"`
	PackageCreatedEvent            *EnduroPackageCreatedEventResponseBody            `form:"package_created_event,omitempty" json:"package_created_event,omitempty" xml:"package_created_event,omitempty"`
	PackageUpdatedEvent            *EnduroPackageUpdatedEventResponseBody            `form:"package_updated_event,omitempty" json:"package_updated_event,omitempty" xml:"package_updated_event,omitempty"`
	PackageStatusUpdatedEvent      *EnduroPackageStatusUpdatedEventResponseBody      `form:"package_status_updated_event,omitempty" json:"package_status_updated_event,omitempty" xml:"package_status_updated_event,omitempty"`
	PackageLocationUpdatedEvent    *EnduroPackageLocationUpdatedEventResponseBody    `form:"package_location_updated_event,omitempty" json:"package_location_updated_event,omitempty" xml:"package_location_updated_event,omitempty"`
	PreservationActionCreatedEvent *EnduroPreservationActionCreatedEventResponseBody `form:"preservation_action_created_event,omitempty" json:"preservation_action_created_event,omitempty" xml:"preservation_action_created_event,omitempty"`
	PreservationActionUpdatedEvent *EnduroPreservationActionUpdatedEventResponseBody `form:"preservation_action_updated_event,omitempty" json:"preservation_action_updated_event,omitempty" xml:"preservation_action_updated_event,omitempty"`
	PreservationTaskCreatedEvent   *EnduroPreservationTaskCreatedEventResponseBody   `form:"preservation_task_created_event,omitempty" json:"preservation_task_created_event,omitempty" xml:"preservation_task_created_event,omitempty"`
	PreservationTaskUpdatedEvent   *EnduroPreservationTaskUpdatedEventResponseBody   `form:"preservation_task_updated_event,omitempty" json:"preservation_task_updated_event,omitempty" xml:"preservation_task_updated_event,omitempty"`
}

// ListResponseBody is the type of the "package" service "list" endpoint HTTP
// response body.
type ListResponseBody struct {
	Items      EnduroStoredPackageCollectionResponseBody `form:"items" json:"items" xml:"items"`
	NextCursor *string                                   `form:"next_cursor,omitempty" json:"next_cursor,omitempty" xml:"next_cursor,omitempty"`
}

// ShowResponseBody is the type of the "package" service "show" endpoint HTTP
// response body.
type ShowResponseBody struct {
	// Identifier of package
	ID uint `form:"id" json:"id" xml:"id"`
	// Name of the package
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Identifier of storage location
	LocationID *uuid.UUID `form:"location_id,omitempty" json:"location_id,omitempty" xml:"location_id,omitempty"`
	// Status of the package
	Status string `form:"status" json:"status" xml:"status"`
	// Identifier of processing workflow
	WorkflowID *string `form:"workflow_id,omitempty" json:"workflow_id,omitempty" xml:"workflow_id,omitempty"`
	// Identifier of latest processing workflow run
	RunID *string `form:"run_id,omitempty" json:"run_id,omitempty" xml:"run_id,omitempty"`
	// Identifier of AIP
	AipID *string `form:"aip_id,omitempty" json:"aip_id,omitempty" xml:"aip_id,omitempty"`
	// Creation datetime
	CreatedAt string `form:"created_at" json:"created_at" xml:"created_at"`
	// Start datetime
	StartedAt *string `form:"started_at,omitempty" json:"started_at,omitempty" xml:"started_at,omitempty"`
	// Completion datetime
	CompletedAt *string `form:"completed_at,omitempty" json:"completed_at,omitempty" xml:"completed_at,omitempty"`
}

// PreservationActionsResponseBody is the type of the "package" service
// "preservation_actions" endpoint HTTP response body.
type PreservationActionsResponseBody struct {
	Actions EnduroPackagePreservationActionResponseBodyCollection `form:"actions,omitempty" json:"actions,omitempty" xml:"actions,omitempty"`
}

// MoveStatusResponseBody is the type of the "package" service "move_status"
// endpoint HTTP response body.
type MoveStatusResponseBody struct {
	Done bool `form:"done" json:"done" xml:"done"`
}

// MonitorRequestNotAvailableResponseBody is the type of the "package" service
// "monitor_request" endpoint HTTP response body for the "not_available" error.
type MonitorRequestNotAvailableResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// MonitorNotAvailableResponseBody is the type of the "package" service
// "monitor" endpoint HTTP response body for the "not_available" error.
type MonitorNotAvailableResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// ShowNotAvailableResponseBody is the type of the "package" service "show"
// endpoint HTTP response body for the "not_available" error.
type ShowNotAvailableResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// ShowNotFoundResponseBody is the type of the "package" service "show"
// endpoint HTTP response body for the "not_found" error.
type ShowNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	ID uint `form:"id" json:"id" xml:"id"`
}

// PreservationActionsNotFoundResponseBody is the type of the "package" service
// "preservation_actions" endpoint HTTP response body for the "not_found" error.
type PreservationActionsNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	ID uint `form:"id" json:"id" xml:"id"`
}

// ConfirmNotAvailableResponseBody is the type of the "package" service
// "confirm" endpoint HTTP response body for the "not_available" error.
type ConfirmNotAvailableResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// ConfirmNotValidResponseBody is the type of the "package" service "confirm"
// endpoint HTTP response body for the "not_valid" error.
type ConfirmNotValidResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// ConfirmNotFoundResponseBody is the type of the "package" service "confirm"
// endpoint HTTP response body for the "not_found" error.
type ConfirmNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	ID uint `form:"id" json:"id" xml:"id"`
}

// RejectNotAvailableResponseBody is the type of the "package" service "reject"
// endpoint HTTP response body for the "not_available" error.
type RejectNotAvailableResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// RejectNotValidResponseBody is the type of the "package" service "reject"
// endpoint HTTP response body for the "not_valid" error.
type RejectNotValidResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// RejectNotFoundResponseBody is the type of the "package" service "reject"
// endpoint HTTP response body for the "not_found" error.
type RejectNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	ID uint `form:"id" json:"id" xml:"id"`
}

// MoveNotAvailableResponseBody is the type of the "package" service "move"
// endpoint HTTP response body for the "not_available" error.
type MoveNotAvailableResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// MoveNotValidResponseBody is the type of the "package" service "move"
// endpoint HTTP response body for the "not_valid" error.
type MoveNotValidResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// MoveNotFoundResponseBody is the type of the "package" service "move"
// endpoint HTTP response body for the "not_found" error.
type MoveNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	ID uint `form:"id" json:"id" xml:"id"`
}

// MoveStatusFailedDependencyResponseBody is the type of the "package" service
// "move_status" endpoint HTTP response body for the "failed_dependency" error.
type MoveStatusFailedDependencyResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// MoveStatusNotFoundResponseBody is the type of the "package" service
// "move_status" endpoint HTTP response body for the "not_found" error.
type MoveStatusNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	ID uint `form:"id" json:"id" xml:"id"`
}

// EnduroMonitorPingEventResponseBody is used to define fields on response body
// types.
type EnduroMonitorPingEventResponseBody struct {
	Message *string `form:"message,omitempty" json:"message,omitempty" xml:"message,omitempty"`
}

// EnduroPackageCreatedEventResponseBody is used to define fields on response
// body types.
type EnduroPackageCreatedEventResponseBody struct {
	// Identifier of package
	ID   uint                             `form:"id" json:"id" xml:"id"`
	Item *EnduroStoredPackageResponseBody `form:"item" json:"item" xml:"item"`
}

// EnduroStoredPackageResponseBody is used to define fields on response body
// types.
type EnduroStoredPackageResponseBody struct {
	// Identifier of package
	ID uint `form:"id" json:"id" xml:"id"`
	// Name of the package
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Identifier of storage location
	LocationID *uuid.UUID `form:"location_id,omitempty" json:"location_id,omitempty" xml:"location_id,omitempty"`
	// Status of the package
	Status string `form:"status" json:"status" xml:"status"`
	// Identifier of processing workflow
	WorkflowID *string `form:"workflow_id,omitempty" json:"workflow_id,omitempty" xml:"workflow_id,omitempty"`
	// Identifier of latest processing workflow run
	RunID *string `form:"run_id,omitempty" json:"run_id,omitempty" xml:"run_id,omitempty"`
	// Identifier of AIP
	AipID *string `form:"aip_id,omitempty" json:"aip_id,omitempty" xml:"aip_id,omitempty"`
	// Creation datetime
	CreatedAt string `form:"created_at" json:"created_at" xml:"created_at"`
	// Start datetime
	StartedAt *string `form:"started_at,omitempty" json:"started_at,omitempty" xml:"started_at,omitempty"`
	// Completion datetime
	CompletedAt *string `form:"completed_at,omitempty" json:"completed_at,omitempty" xml:"completed_at,omitempty"`
}

// EnduroPackageUpdatedEventResponseBody is used to define fields on response
// body types.
type EnduroPackageUpdatedEventResponseBody struct {
	// Identifier of package
	ID   uint                             `form:"id" json:"id" xml:"id"`
	Item *EnduroStoredPackageResponseBody `form:"item" json:"item" xml:"item"`
}

// EnduroPackageStatusUpdatedEventResponseBody is used to define fields on
// response body types.
type EnduroPackageStatusUpdatedEventResponseBody struct {
	// Identifier of package
	ID     uint   `form:"id" json:"id" xml:"id"`
	Status string `form:"status" json:"status" xml:"status"`
}

// EnduroPackageLocationUpdatedEventResponseBody is used to define fields on
// response body types.
type EnduroPackageLocationUpdatedEventResponseBody struct {
	// Identifier of package
	ID uint `form:"id" json:"id" xml:"id"`
	// Identifier of storage location
	LocationID uuid.UUID `form:"location_id" json:"location_id" xml:"location_id"`
}

// EnduroPreservationActionCreatedEventResponseBody is used to define fields on
// response body types.
type EnduroPreservationActionCreatedEventResponseBody struct {
	// Identifier of preservation action
	ID   uint                                               `form:"id" json:"id" xml:"id"`
	Item *EnduroPackagePreservationActionResponseBodySimple `form:"item" json:"item" xml:"item"`
}

// EnduroPackagePreservationActionResponseBodySimple is used to define fields
// on response body types.
type EnduroPackagePreservationActionResponseBodySimple struct {
	ID          uint    `form:"id" json:"id" xml:"id"`
	WorkflowID  string  `form:"workflow_id" json:"workflow_id" xml:"workflow_id"`
	Type        string  `form:"type" json:"type" xml:"type"`
	Status      string  `form:"status" json:"status" xml:"status"`
	StartedAt   string  `form:"started_at" json:"started_at" xml:"started_at"`
	CompletedAt *string `form:"completed_at,omitempty" json:"completed_at,omitempty" xml:"completed_at,omitempty"`
	PackageID   *uint   `form:"package_id,omitempty" json:"package_id,omitempty" xml:"package_id,omitempty"`
}

// EnduroPreservationActionUpdatedEventResponseBody is used to define fields on
// response body types.
type EnduroPreservationActionUpdatedEventResponseBody struct {
	// Identifier of preservation action
	ID   uint                                               `form:"id" json:"id" xml:"id"`
	Item *EnduroPackagePreservationActionResponseBodySimple `form:"item" json:"item" xml:"item"`
}

// EnduroPreservationTaskCreatedEventResponseBody is used to define fields on
// response body types.
type EnduroPreservationTaskCreatedEventResponseBody struct {
	// Identifier of preservation task
	ID   uint                                       `form:"id" json:"id" xml:"id"`
	Item *EnduroPackagePreservationTaskResponseBody `form:"item" json:"item" xml:"item"`
}

// EnduroPackagePreservationTaskResponseBody is used to define fields on
// response body types.
type EnduroPackagePreservationTaskResponseBody struct {
	ID                   uint    `form:"id" json:"id" xml:"id"`
	TaskID               string  `form:"task_id" json:"task_id" xml:"task_id"`
	Name                 string  `form:"name" json:"name" xml:"name"`
	Status               string  `form:"status" json:"status" xml:"status"`
	StartedAt            string  `form:"started_at" json:"started_at" xml:"started_at"`
	CompletedAt          *string `form:"completed_at,omitempty" json:"completed_at,omitempty" xml:"completed_at,omitempty"`
	Note                 *string `form:"note,omitempty" json:"note,omitempty" xml:"note,omitempty"`
	PreservationActionID *uint   `form:"preservation_action_id,omitempty" json:"preservation_action_id,omitempty" xml:"preservation_action_id,omitempty"`
}

// EnduroPreservationTaskUpdatedEventResponseBody is used to define fields on
// response body types.
type EnduroPreservationTaskUpdatedEventResponseBody struct {
	// Identifier of preservation task
	ID   uint                                       `form:"id" json:"id" xml:"id"`
	Item *EnduroPackagePreservationTaskResponseBody `form:"item" json:"item" xml:"item"`
}

// EnduroStoredPackageCollectionResponseBody is used to define fields on
// response body types.
type EnduroStoredPackageCollectionResponseBody []*EnduroStoredPackageResponseBody

// EnduroPackagePreservationActionResponseBodyCollection is used to define
// fields on response body types.
type EnduroPackagePreservationActionResponseBodyCollection []*EnduroPackagePreservationActionResponseBody

// EnduroPackagePreservationActionResponseBody is used to define fields on
// response body types.
type EnduroPackagePreservationActionResponseBody struct {
	ID          uint                                                `form:"id" json:"id" xml:"id"`
	WorkflowID  string                                              `form:"workflow_id" json:"workflow_id" xml:"workflow_id"`
	Type        string                                              `form:"type" json:"type" xml:"type"`
	Status      string                                              `form:"status" json:"status" xml:"status"`
	StartedAt   string                                              `form:"started_at" json:"started_at" xml:"started_at"`
	CompletedAt *string                                             `form:"completed_at,omitempty" json:"completed_at,omitempty" xml:"completed_at,omitempty"`
	Tasks       EnduroPackagePreservationTaskResponseBodyCollection `form:"tasks,omitempty" json:"tasks,omitempty" xml:"tasks,omitempty"`
	PackageID   *uint                                               `form:"package_id,omitempty" json:"package_id,omitempty" xml:"package_id,omitempty"`
}

// EnduroPackagePreservationTaskResponseBodyCollection is used to define fields
// on response body types.
type EnduroPackagePreservationTaskResponseBodyCollection []*EnduroPackagePreservationTaskResponseBody

// NewMonitorResponseBody builds the HTTP response body from the result of the
// "monitor" endpoint of the "package" service.
func NewMonitorResponseBody(res *package_views.EnduroMonitorEventView) *MonitorResponseBody {
	body := &MonitorResponseBody{}
	if res.MonitorPingEvent != nil {
		body.MonitorPingEvent = marshalPackageViewsEnduroMonitorPingEventViewToEnduroMonitorPingEventResponseBody(res.MonitorPingEvent)
	}
	if res.PackageCreatedEvent != nil {
		body.PackageCreatedEvent = marshalPackageViewsEnduroPackageCreatedEventViewToEnduroPackageCreatedEventResponseBody(res.PackageCreatedEvent)
	}
	if res.PackageUpdatedEvent != nil {
		body.PackageUpdatedEvent = marshalPackageViewsEnduroPackageUpdatedEventViewToEnduroPackageUpdatedEventResponseBody(res.PackageUpdatedEvent)
	}
	if res.PackageStatusUpdatedEvent != nil {
		body.PackageStatusUpdatedEvent = marshalPackageViewsEnduroPackageStatusUpdatedEventViewToEnduroPackageStatusUpdatedEventResponseBody(res.PackageStatusUpdatedEvent)
	}
	if res.PackageLocationUpdatedEvent != nil {
		body.PackageLocationUpdatedEvent = marshalPackageViewsEnduroPackageLocationUpdatedEventViewToEnduroPackageLocationUpdatedEventResponseBody(res.PackageLocationUpdatedEvent)
	}
	if res.PreservationActionCreatedEvent != nil {
		body.PreservationActionCreatedEvent = marshalPackageViewsEnduroPreservationActionCreatedEventViewToEnduroPreservationActionCreatedEventResponseBody(res.PreservationActionCreatedEvent)
	}
	if res.PreservationActionUpdatedEvent != nil {
		body.PreservationActionUpdatedEvent = marshalPackageViewsEnduroPreservationActionUpdatedEventViewToEnduroPreservationActionUpdatedEventResponseBody(res.PreservationActionUpdatedEvent)
	}
	if res.PreservationTaskCreatedEvent != nil {
		body.PreservationTaskCreatedEvent = marshalPackageViewsEnduroPreservationTaskCreatedEventViewToEnduroPreservationTaskCreatedEventResponseBody(res.PreservationTaskCreatedEvent)
	}
	if res.PreservationTaskUpdatedEvent != nil {
		body.PreservationTaskUpdatedEvent = marshalPackageViewsEnduroPreservationTaskUpdatedEventViewToEnduroPreservationTaskUpdatedEventResponseBody(res.PreservationTaskUpdatedEvent)
	}
	return body
}

// NewListResponseBody builds the HTTP response body from the result of the
// "list" endpoint of the "package" service.
func NewListResponseBody(res *package_.ListResult) *ListResponseBody {
	body := &ListResponseBody{
		NextCursor: res.NextCursor,
	}
	if res.Items != nil {
		body.Items = make([]*EnduroStoredPackageResponseBody, len(res.Items))
		for i, val := range res.Items {
			body.Items[i] = marshalPackageEnduroStoredPackageToEnduroStoredPackageResponseBody(val)
		}
	}
	return body
}

// NewShowResponseBody builds the HTTP response body from the result of the
// "show" endpoint of the "package" service.
func NewShowResponseBody(res *package_views.EnduroStoredPackageView) *ShowResponseBody {
	body := &ShowResponseBody{
		ID:          *res.ID,
		Name:        res.Name,
		LocationID:  res.LocationID,
		Status:      *res.Status,
		WorkflowID:  res.WorkflowID,
		RunID:       res.RunID,
		AipID:       res.AipID,
		CreatedAt:   *res.CreatedAt,
		StartedAt:   res.StartedAt,
		CompletedAt: res.CompletedAt,
	}
	return body
}

// NewPreservationActionsResponseBody builds the HTTP response body from the
// result of the "preservation_actions" endpoint of the "package" service.
func NewPreservationActionsResponseBody(res *package_views.EnduroPackagePreservationActionsView) *PreservationActionsResponseBody {
	body := &PreservationActionsResponseBody{}
	if res.Actions != nil {
		body.Actions = make([]*EnduroPackagePreservationActionResponseBody, len(res.Actions))
		for i, val := range res.Actions {
			body.Actions[i] = marshalPackageViewsEnduroPackagePreservationActionViewToEnduroPackagePreservationActionResponseBody(val)
		}
	}
	return body
}

// NewMoveStatusResponseBody builds the HTTP response body from the result of
// the "move_status" endpoint of the "package" service.
func NewMoveStatusResponseBody(res *package_.MoveStatusResult) *MoveStatusResponseBody {
	body := &MoveStatusResponseBody{
		Done: res.Done,
	}
	return body
}

// NewMonitorRequestNotAvailableResponseBody builds the HTTP response body from
// the result of the "monitor_request" endpoint of the "package" service.
func NewMonitorRequestNotAvailableResponseBody(res *goa.ServiceError) *MonitorRequestNotAvailableResponseBody {
	body := &MonitorRequestNotAvailableResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewMonitorNotAvailableResponseBody builds the HTTP response body from the
// result of the "monitor" endpoint of the "package" service.
func NewMonitorNotAvailableResponseBody(res *goa.ServiceError) *MonitorNotAvailableResponseBody {
	body := &MonitorNotAvailableResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewShowNotAvailableResponseBody builds the HTTP response body from the
// result of the "show" endpoint of the "package" service.
func NewShowNotAvailableResponseBody(res *goa.ServiceError) *ShowNotAvailableResponseBody {
	body := &ShowNotAvailableResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewShowNotFoundResponseBody builds the HTTP response body from the result of
// the "show" endpoint of the "package" service.
func NewShowNotFoundResponseBody(res *package_.PackageNotFound) *ShowNotFoundResponseBody {
	body := &ShowNotFoundResponseBody{
		Message: res.Message,
		ID:      res.ID,
	}
	return body
}

// NewPreservationActionsNotFoundResponseBody builds the HTTP response body
// from the result of the "preservation_actions" endpoint of the "package"
// service.
func NewPreservationActionsNotFoundResponseBody(res *package_.PackageNotFound) *PreservationActionsNotFoundResponseBody {
	body := &PreservationActionsNotFoundResponseBody{
		Message: res.Message,
		ID:      res.ID,
	}
	return body
}

// NewConfirmNotAvailableResponseBody builds the HTTP response body from the
// result of the "confirm" endpoint of the "package" service.
func NewConfirmNotAvailableResponseBody(res *goa.ServiceError) *ConfirmNotAvailableResponseBody {
	body := &ConfirmNotAvailableResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewConfirmNotValidResponseBody builds the HTTP response body from the result
// of the "confirm" endpoint of the "package" service.
func NewConfirmNotValidResponseBody(res *goa.ServiceError) *ConfirmNotValidResponseBody {
	body := &ConfirmNotValidResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewConfirmNotFoundResponseBody builds the HTTP response body from the result
// of the "confirm" endpoint of the "package" service.
func NewConfirmNotFoundResponseBody(res *package_.PackageNotFound) *ConfirmNotFoundResponseBody {
	body := &ConfirmNotFoundResponseBody{
		Message: res.Message,
		ID:      res.ID,
	}
	return body
}

// NewRejectNotAvailableResponseBody builds the HTTP response body from the
// result of the "reject" endpoint of the "package" service.
func NewRejectNotAvailableResponseBody(res *goa.ServiceError) *RejectNotAvailableResponseBody {
	body := &RejectNotAvailableResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewRejectNotValidResponseBody builds the HTTP response body from the result
// of the "reject" endpoint of the "package" service.
func NewRejectNotValidResponseBody(res *goa.ServiceError) *RejectNotValidResponseBody {
	body := &RejectNotValidResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewRejectNotFoundResponseBody builds the HTTP response body from the result
// of the "reject" endpoint of the "package" service.
func NewRejectNotFoundResponseBody(res *package_.PackageNotFound) *RejectNotFoundResponseBody {
	body := &RejectNotFoundResponseBody{
		Message: res.Message,
		ID:      res.ID,
	}
	return body
}

// NewMoveNotAvailableResponseBody builds the HTTP response body from the
// result of the "move" endpoint of the "package" service.
func NewMoveNotAvailableResponseBody(res *goa.ServiceError) *MoveNotAvailableResponseBody {
	body := &MoveNotAvailableResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewMoveNotValidResponseBody builds the HTTP response body from the result of
// the "move" endpoint of the "package" service.
func NewMoveNotValidResponseBody(res *goa.ServiceError) *MoveNotValidResponseBody {
	body := &MoveNotValidResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewMoveNotFoundResponseBody builds the HTTP response body from the result of
// the "move" endpoint of the "package" service.
func NewMoveNotFoundResponseBody(res *package_.PackageNotFound) *MoveNotFoundResponseBody {
	body := &MoveNotFoundResponseBody{
		Message: res.Message,
		ID:      res.ID,
	}
	return body
}

// NewMoveStatusFailedDependencyResponseBody builds the HTTP response body from
// the result of the "move_status" endpoint of the "package" service.
func NewMoveStatusFailedDependencyResponseBody(res *goa.ServiceError) *MoveStatusFailedDependencyResponseBody {
	body := &MoveStatusFailedDependencyResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewMoveStatusNotFoundResponseBody builds the HTTP response body from the
// result of the "move_status" endpoint of the "package" service.
func NewMoveStatusNotFoundResponseBody(res *package_.PackageNotFound) *MoveStatusNotFoundResponseBody {
	body := &MoveStatusNotFoundResponseBody{
		Message: res.Message,
		ID:      res.ID,
	}
	return body
}

// NewMonitorRequestPayload builds a package service monitor_request endpoint
// payload.
func NewMonitorRequestPayload(oauthToken *string) *package_.MonitorRequestPayload {
	v := &package_.MonitorRequestPayload{}
	v.OauthToken = oauthToken

	return v
}

// NewMonitorPayload builds a package service monitor endpoint payload.
func NewMonitorPayload(ticket *string) *package_.MonitorPayload {
	v := &package_.MonitorPayload{}
	v.Ticket = ticket

	return v
}

// NewListPayload builds a package service list endpoint payload.
func NewListPayload(name *string, aipID *string, earliestCreatedTime *string, latestCreatedTime *string, locationID *string, status *string, cursor *string, oauthToken *string) *package_.ListPayload {
	v := &package_.ListPayload{}
	v.Name = name
	v.AipID = aipID
	v.EarliestCreatedTime = earliestCreatedTime
	v.LatestCreatedTime = latestCreatedTime
	v.LocationID = locationID
	v.Status = status
	v.Cursor = cursor
	v.OauthToken = oauthToken

	return v
}

// NewShowPayload builds a package service show endpoint payload.
func NewShowPayload(id uint, oauthToken *string) *package_.ShowPayload {
	v := &package_.ShowPayload{}
	v.ID = id
	v.OauthToken = oauthToken

	return v
}

// NewPreservationActionsPayload builds a package service preservation_actions
// endpoint payload.
func NewPreservationActionsPayload(id uint, oauthToken *string) *package_.PreservationActionsPayload {
	v := &package_.PreservationActionsPayload{}
	v.ID = id
	v.OauthToken = oauthToken

	return v
}

// NewConfirmPayload builds a package service confirm endpoint payload.
func NewConfirmPayload(body *ConfirmRequestBody, id uint, oauthToken *string) *package_.ConfirmPayload {
	v := &package_.ConfirmPayload{
		LocationID: *body.LocationID,
	}
	v.ID = id
	v.OauthToken = oauthToken

	return v
}

// NewRejectPayload builds a package service reject endpoint payload.
func NewRejectPayload(id uint, oauthToken *string) *package_.RejectPayload {
	v := &package_.RejectPayload{}
	v.ID = id
	v.OauthToken = oauthToken

	return v
}

// NewMovePayload builds a package service move endpoint payload.
func NewMovePayload(body *MoveRequestBody, id uint, oauthToken *string) *package_.MovePayload {
	v := &package_.MovePayload{
		LocationID: *body.LocationID,
	}
	v.ID = id
	v.OauthToken = oauthToken

	return v
}

// NewMoveStatusPayload builds a package service move_status endpoint payload.
func NewMoveStatusPayload(id uint, oauthToken *string) *package_.MoveStatusPayload {
	v := &package_.MoveStatusPayload{}
	v.ID = id
	v.OauthToken = oauthToken

	return v
}

// ValidateConfirmRequestBody runs the validations defined on ConfirmRequestBody
func ValidateConfirmRequestBody(body *ConfirmRequestBody) (err error) {
	if body.LocationID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("location_id", "body"))
	}
	return
}

// ValidateMoveRequestBody runs the validations defined on MoveRequestBody
func ValidateMoveRequestBody(body *MoveRequestBody) (err error) {
	if body.LocationID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("location_id", "body"))
	}
	return
}
