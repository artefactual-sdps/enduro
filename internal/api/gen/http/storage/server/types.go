// Code generated by goa v3.8.3, DO NOT EDIT.
//
// storage HTTP server types
//
// Command:
// $ goa-v3.8.3 gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package server

import (
	"encoding/json"

	storage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	storageviews "github.com/artefactual-sdps/enduro/internal/api/gen/storage/views"
	"github.com/google/uuid"
	goa "goa.design/goa/v3/pkg"
)

// SubmitRequestBody is the type of the "storage" service "submit" endpoint
// HTTP request body.
type SubmitRequestBody struct {
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
}

// AddLocationRequestBody is the type of the "storage" service "add_location"
// endpoint HTTP request body.
type AddLocationRequestBody struct {
	Name        *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
	Source      *string `form:"source,omitempty" json:"source,omitempty" xml:"source,omitempty"`
	Purpose     *string `form:"purpose,omitempty" json:"purpose,omitempty" xml:"purpose,omitempty"`
	Config      *struct {
		// Union type name, one of:
		// - "s3"
		Type *string `form:"Type" json:"Type" xml:"Type"`
		// JSON formatted union value
		Value *string `form:"Value" json:"Value" xml:"Value"`
	} `form:"config,omitempty" json:"config,omitempty" xml:"config,omitempty"`
}

// MoveRequestBody is the type of the "storage" service "move" endpoint HTTP
// request body.
type MoveRequestBody struct {
	LocationID *uuid.UUID `form:"location_id,omitempty" json:"location_id,omitempty" xml:"location_id,omitempty"`
}

// SubmitResponseBody is the type of the "storage" service "submit" endpoint
// HTTP response body.
type SubmitResponseBody struct {
	URL string `form:"url" json:"url" xml:"url"`
}

// StoredLocationResponseCollection is the type of the "storage" service
// "locations" endpoint HTTP response body.
type StoredLocationResponseCollection []*StoredLocationResponse

// AddLocationResponseBody is the type of the "storage" service "add_location"
// endpoint HTTP response body.
type AddLocationResponseBody struct {
	UUID string `form:"uuid" json:"uuid" xml:"uuid"`
}

// MoveStatusResponseBody is the type of the "storage" service "move_status"
// endpoint HTTP response body.
type MoveStatusResponseBody struct {
	Done bool `form:"done" json:"done" xml:"done"`
}

// ShowResponseBody is the type of the "storage" service "show" endpoint HTTP
// response body.
type ShowResponseBody struct {
	Name  string `form:"name" json:"name" xml:"name"`
	AipID string `form:"aip_id" json:"aip_id" xml:"aip_id"`
	// Status of the package
	Status     string     `form:"status" json:"status" xml:"status"`
	ObjectKey  uuid.UUID  `form:"object_key" json:"object_key" xml:"object_key"`
	LocationID *uuid.UUID `form:"location_id,omitempty" json:"location_id,omitempty" xml:"location_id,omitempty"`
}

// ShowLocationResponseBody is the type of the "storage" service
// "show-location" endpoint HTTP response body.
type ShowLocationResponseBody struct {
	// Name of location
	Name string `form:"name" json:"name" xml:"name"`
	// Description of the location
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
	// Data source of the location
	Source string `form:"source" json:"source" xml:"source"`
	// Purpose of the location
	Purpose string     `form:"purpose" json:"purpose" xml:"purpose"`
	UUID    *uuid.UUID `form:"uuid,omitempty" json:"uuid,omitempty" xml:"uuid,omitempty"`
}

// SubmitNotAvailableResponseBody is the type of the "storage" service "submit"
// endpoint HTTP response body for the "not_available" error.
type SubmitNotAvailableResponseBody struct {
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

// SubmitNotValidResponseBody is the type of the "storage" service "submit"
// endpoint HTTP response body for the "not_valid" error.
type SubmitNotValidResponseBody struct {
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

// UpdateNotAvailableResponseBody is the type of the "storage" service "update"
// endpoint HTTP response body for the "not_available" error.
type UpdateNotAvailableResponseBody struct {
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

// UpdateNotValidResponseBody is the type of the "storage" service "update"
// endpoint HTTP response body for the "not_valid" error.
type UpdateNotValidResponseBody struct {
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

// DownloadNotFoundResponseBody is the type of the "storage" service "download"
// endpoint HTTP response body for the "not_found" error.
type DownloadNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	AipID string `form:"aip_id" json:"aip_id" xml:"aip_id"`
}

// AddLocationNotValidResponseBody is the type of the "storage" service
// "add_location" endpoint HTTP response body for the "not_valid" error.
type AddLocationNotValidResponseBody struct {
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

// MoveNotAvailableResponseBody is the type of the "storage" service "move"
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

// MoveNotValidResponseBody is the type of the "storage" service "move"
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

// MoveNotFoundResponseBody is the type of the "storage" service "move"
// endpoint HTTP response body for the "not_found" error.
type MoveNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	AipID string `form:"aip_id" json:"aip_id" xml:"aip_id"`
}

// MoveStatusFailedDependencyResponseBody is the type of the "storage" service
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

// MoveStatusNotFoundResponseBody is the type of the "storage" service
// "move_status" endpoint HTTP response body for the "not_found" error.
type MoveStatusNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	AipID string `form:"aip_id" json:"aip_id" xml:"aip_id"`
}

// RejectNotAvailableResponseBody is the type of the "storage" service "reject"
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

// RejectNotValidResponseBody is the type of the "storage" service "reject"
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

// RejectNotFoundResponseBody is the type of the "storage" service "reject"
// endpoint HTTP response body for the "not_found" error.
type RejectNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	AipID string `form:"aip_id" json:"aip_id" xml:"aip_id"`
}

// ShowNotFoundResponseBody is the type of the "storage" service "show"
// endpoint HTTP response body for the "not_found" error.
type ShowNotFoundResponseBody struct {
	// Message of error
	Message string `form:"message" json:"message" xml:"message"`
	// Identifier of missing package
	AipID string `form:"aip_id" json:"aip_id" xml:"aip_id"`
}

// ShowLocationNotFoundResponseBody is the type of the "storage" service
// "show-location" endpoint HTTP response body for the "not_found" error.
type ShowLocationNotFoundResponseBody struct {
	// Message of error
	Message string    `form:"message" json:"message" xml:"message"`
	UUID    uuid.UUID `form:"uuid" json:"uuid" xml:"uuid"`
}

// StoredLocationResponse is used to define fields on response body types.
type StoredLocationResponse struct {
	// Name of location
	Name string `form:"name" json:"name" xml:"name"`
	// Description of the location
	Description *string `form:"description,omitempty" json:"description,omitempty" xml:"description,omitempty"`
	// Data source of the location
	Source string `form:"source" json:"source" xml:"source"`
	// Purpose of the location
	Purpose string     `form:"purpose" json:"purpose" xml:"purpose"`
	UUID    *uuid.UUID `form:"uuid,omitempty" json:"uuid,omitempty" xml:"uuid,omitempty"`
}

// NewSubmitResponseBody builds the HTTP response body from the result of the
// "submit" endpoint of the "storage" service.
func NewSubmitResponseBody(res *storage.SubmitResult) *SubmitResponseBody {
	body := &SubmitResponseBody{
		URL: res.URL,
	}
	return body
}

// NewStoredLocationResponseCollection builds the HTTP response body from the
// result of the "locations" endpoint of the "storage" service.
func NewStoredLocationResponseCollection(res storageviews.StoredLocationCollectionView) StoredLocationResponseCollection {
	body := make([]*StoredLocationResponse, len(res))
	for i, val := range res {
		body[i] = marshalStorageviewsStoredLocationViewToStoredLocationResponse(val)
	}
	return body
}

// NewAddLocationResponseBody builds the HTTP response body from the result of
// the "add_location" endpoint of the "storage" service.
func NewAddLocationResponseBody(res *storage.AddLocationResult) *AddLocationResponseBody {
	body := &AddLocationResponseBody{
		UUID: res.UUID,
	}
	return body
}

// NewMoveStatusResponseBody builds the HTTP response body from the result of
// the "move_status" endpoint of the "storage" service.
func NewMoveStatusResponseBody(res *storage.MoveStatusResult) *MoveStatusResponseBody {
	body := &MoveStatusResponseBody{
		Done: res.Done,
	}
	return body
}

// NewShowResponseBody builds the HTTP response body from the result of the
// "show" endpoint of the "storage" service.
func NewShowResponseBody(res *storageviews.StoredStoragePackageView) *ShowResponseBody {
	body := &ShowResponseBody{
		Name:       *res.Name,
		AipID:      *res.AipID,
		Status:     *res.Status,
		ObjectKey:  *res.ObjectKey,
		LocationID: res.LocationID,
	}
	return body
}

// NewShowLocationResponseBody builds the HTTP response body from the result of
// the "show-location" endpoint of the "storage" service.
func NewShowLocationResponseBody(res *storageviews.StoredLocationView) *ShowLocationResponseBody {
	body := &ShowLocationResponseBody{
		Name:        *res.Name,
		Description: res.Description,
		Source:      *res.Source,
		Purpose:     *res.Purpose,
		UUID:        res.UUID,
	}
	return body
}

// NewSubmitNotAvailableResponseBody builds the HTTP response body from the
// result of the "submit" endpoint of the "storage" service.
func NewSubmitNotAvailableResponseBody(res *goa.ServiceError) *SubmitNotAvailableResponseBody {
	body := &SubmitNotAvailableResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewSubmitNotValidResponseBody builds the HTTP response body from the result
// of the "submit" endpoint of the "storage" service.
func NewSubmitNotValidResponseBody(res *goa.ServiceError) *SubmitNotValidResponseBody {
	body := &SubmitNotValidResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewUpdateNotAvailableResponseBody builds the HTTP response body from the
// result of the "update" endpoint of the "storage" service.
func NewUpdateNotAvailableResponseBody(res *goa.ServiceError) *UpdateNotAvailableResponseBody {
	body := &UpdateNotAvailableResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewUpdateNotValidResponseBody builds the HTTP response body from the result
// of the "update" endpoint of the "storage" service.
func NewUpdateNotValidResponseBody(res *goa.ServiceError) *UpdateNotValidResponseBody {
	body := &UpdateNotValidResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewDownloadNotFoundResponseBody builds the HTTP response body from the
// result of the "download" endpoint of the "storage" service.
func NewDownloadNotFoundResponseBody(res *storage.StoragePackageNotfound) *DownloadNotFoundResponseBody {
	body := &DownloadNotFoundResponseBody{
		Message: res.Message,
		AipID:   res.AipID,
	}
	return body
}

// NewAddLocationNotValidResponseBody builds the HTTP response body from the
// result of the "add_location" endpoint of the "storage" service.
func NewAddLocationNotValidResponseBody(res *goa.ServiceError) *AddLocationNotValidResponseBody {
	body := &AddLocationNotValidResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewMoveNotAvailableResponseBody builds the HTTP response body from the
// result of the "move" endpoint of the "storage" service.
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
// the "move" endpoint of the "storage" service.
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
// the "move" endpoint of the "storage" service.
func NewMoveNotFoundResponseBody(res *storage.StoragePackageNotfound) *MoveNotFoundResponseBody {
	body := &MoveNotFoundResponseBody{
		Message: res.Message,
		AipID:   res.AipID,
	}
	return body
}

// NewMoveStatusFailedDependencyResponseBody builds the HTTP response body from
// the result of the "move_status" endpoint of the "storage" service.
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
// result of the "move_status" endpoint of the "storage" service.
func NewMoveStatusNotFoundResponseBody(res *storage.StoragePackageNotfound) *MoveStatusNotFoundResponseBody {
	body := &MoveStatusNotFoundResponseBody{
		Message: res.Message,
		AipID:   res.AipID,
	}
	return body
}

// NewRejectNotAvailableResponseBody builds the HTTP response body from the
// result of the "reject" endpoint of the "storage" service.
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
// of the "reject" endpoint of the "storage" service.
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
// of the "reject" endpoint of the "storage" service.
func NewRejectNotFoundResponseBody(res *storage.StoragePackageNotfound) *RejectNotFoundResponseBody {
	body := &RejectNotFoundResponseBody{
		Message: res.Message,
		AipID:   res.AipID,
	}
	return body
}

// NewShowNotFoundResponseBody builds the HTTP response body from the result of
// the "show" endpoint of the "storage" service.
func NewShowNotFoundResponseBody(res *storage.StoragePackageNotfound) *ShowNotFoundResponseBody {
	body := &ShowNotFoundResponseBody{
		Message: res.Message,
		AipID:   res.AipID,
	}
	return body
}

// NewShowLocationNotFoundResponseBody builds the HTTP response body from the
// result of the "show-location" endpoint of the "storage" service.
func NewShowLocationNotFoundResponseBody(res *storage.StorageLocationNotfound) *ShowLocationNotFoundResponseBody {
	body := &ShowLocationNotFoundResponseBody{
		Message: res.Message,
		UUID:    res.UUID,
	}
	return body
}

// NewSubmitPayload builds a storage service submit endpoint payload.
func NewSubmitPayload(body *SubmitRequestBody, aipID string) *storage.SubmitPayload {
	v := &storage.SubmitPayload{
		Name: *body.Name,
	}
	v.AipID = aipID

	return v
}

// NewUpdatePayload builds a storage service update endpoint payload.
func NewUpdatePayload(aipID string) *storage.UpdatePayload {
	v := &storage.UpdatePayload{}
	v.AipID = aipID

	return v
}

// NewDownloadPayload builds a storage service download endpoint payload.
func NewDownloadPayload(aipID string) *storage.DownloadPayload {
	v := &storage.DownloadPayload{}
	v.AipID = aipID

	return v
}

// NewAddLocationPayload builds a storage service add_location endpoint payload.
func NewAddLocationPayload(body *AddLocationRequestBody) *storage.AddLocationPayload {
	v := &storage.AddLocationPayload{
		Name:        *body.Name,
		Description: body.Description,
		Source:      *body.Source,
		Purpose:     *body.Purpose,
	}
	if body.Config != nil {
		switch *body.Config.Type {
		case "s3":
			var val *storage.S3Config
			json.Unmarshal([]byte(*body.Config.Value), &val)
			v.Config = val
		}
	}

	return v
}

// NewMovePayload builds a storage service move endpoint payload.
func NewMovePayload(body *MoveRequestBody, aipID string) *storage.MovePayload {
	v := &storage.MovePayload{
		LocationID: *body.LocationID,
	}
	v.AipID = aipID

	return v
}

// NewMoveStatusPayload builds a storage service move_status endpoint payload.
func NewMoveStatusPayload(aipID string) *storage.MoveStatusPayload {
	v := &storage.MoveStatusPayload{}
	v.AipID = aipID

	return v
}

// NewRejectPayload builds a storage service reject endpoint payload.
func NewRejectPayload(aipID string) *storage.RejectPayload {
	v := &storage.RejectPayload{}
	v.AipID = aipID

	return v
}

// NewShowPayload builds a storage service show endpoint payload.
func NewShowPayload(aipID string) *storage.ShowPayload {
	v := &storage.ShowPayload{}
	v.AipID = aipID

	return v
}

// NewShowLocationPayload builds a storage service show-location endpoint
// payload.
func NewShowLocationPayload(uuid string) *storage.ShowLocationPayload {
	v := &storage.ShowLocationPayload{}
	v.UUID = uuid

	return v
}

// ValidateSubmitRequestBody runs the validations defined on SubmitRequestBody
func ValidateSubmitRequestBody(body *SubmitRequestBody) (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	return
}

// ValidateAddLocationRequestBody runs the validations defined on
// add_location_request_body
func ValidateAddLocationRequestBody(body *AddLocationRequestBody) (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.Source == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("source", "body"))
	}
	if body.Purpose == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("purpose", "body"))
	}
	if body.Source != nil {
		if !(*body.Source == "unspecified" || *body.Source == "minio") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("body.source", *body.Source, []interface{}{"unspecified", "minio"}))
		}
	}
	if body.Purpose != nil {
		if !(*body.Purpose == "unspecified" || *body.Purpose == "aip_store") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("body.purpose", *body.Purpose, []interface{}{"unspecified", "aip_store"}))
		}
	}
	if body.Config != nil {
		if body.Config.Type == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("Type", "body.config"))
		}
		if body.Config.Value == nil {
			err = goa.MergeErrors(err, goa.MissingFieldError("Value", "body.config"))
		}
		if body.Config.Type != nil {
			if !(*body.Config.Type == "s3") {
				err = goa.MergeErrors(err, goa.InvalidEnumValueError("body.config.Type", *body.Config.Type, []interface{}{"s3"}))
			}
		}
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
