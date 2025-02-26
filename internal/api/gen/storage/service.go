// Code generated by goa v3.15.2, DO NOT EDIT.
//
// storage service
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package storage

import (
	"context"

	storageviews "github.com/artefactual-sdps/enduro/internal/api/gen/storage/views"
	"github.com/google/uuid"
	goa "goa.design/goa/v3/pkg"
	"goa.design/goa/v3/security"
)

// The storage service manages locations and AIPs.
type Service interface {
	// Create a new AIP
	CreateAip(context.Context, *CreateAipPayload) (res *AIP, err error)
	// Start the submission of an AIP
	SubmitAip(context.Context, *SubmitAipPayload) (res *SubmitAIPResult, err error)
	// Signal that an AIP submission is complete
	UpdateAip(context.Context, *UpdateAipPayload) (err error)
	// Download AIP by AIPID
	DownloadAip(context.Context, *DownloadAipPayload) (res []byte, err error)
	// Move an AIP to a permanent storage location
	MoveAip(context.Context, *MoveAipPayload) (err error)
	// Retrieve the status of a permanent storage location move of the AIP
	MoveAipStatus(context.Context, *MoveAipStatusPayload) (res *MoveStatusResult, err error)
	// Reject an AIP
	RejectAip(context.Context, *RejectAipPayload) (err error)
	// Show AIP by AIPID
	ShowAip(context.Context, *ShowAipPayload) (res *AIP, err error)
	// List locations
	ListLocations(context.Context, *ListLocationsPayload) (res LocationCollection, err error)
	// Create a storage location
	CreateLocation(context.Context, *CreateLocationPayload) (res *CreateLocationResult, err error)
	// Show location by UUID
	ShowLocation(context.Context, *ShowLocationPayload) (res *Location, err error)
	// List all the AIPs stored in the location with UUID
	ListLocationAips(context.Context, *ListLocationAipsPayload) (res AIPCollection, err error)
}

// Auther defines the authorization functions to be implemented by the service.
type Auther interface {
	// JWTAuth implements the authorization logic for the JWT security scheme.
	JWTAuth(ctx context.Context, token string, schema *security.JWTScheme) (context.Context, error)
}

// APIName is the name of the API as defined in the design.
const APIName = "enduro"

// APIVersion is the version of the API as defined in the design.
const APIVersion = "0.0.1"

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "storage"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [12]string{"create_aip", "submit_aip", "update_aip", "download_aip", "move_aip", "move_aip_status", "reject_aip", "show_aip", "list_locations", "create_location", "show_location", "list_location_aips"}

// AIP is the result type of the storage service create_aip method.
type AIP struct {
	Name string
	UUID uuid.UUID
	// Status of the AIP
	Status    string
	ObjectKey uuid.UUID
	// Identifier of storage location
	LocationID *uuid.UUID
	// Creation datetime
	CreatedAt string
}

// AIPCollection is the result type of the storage service list_location_aips
// method.
type AIPCollection []*AIP

// AIP not found.
type AIPNotFound struct {
	// Message of error
	Message string
	// Identifier of missing AIP
	UUID uuid.UUID
}

type AMSSConfig struct {
	APIKey   string
	URL      string
	Username string
}

// CreateAipPayload is the payload type of the storage service create_aip
// method.
type CreateAipPayload struct {
	// Identifier of the AIP
	UUID string
	// Name of the AIP
	Name string
	// ObjectKey of the AIP
	ObjectKey string
	// Status of the the AIP
	Status string
	// Identifier of the AIP's storage location
	LocationID *uuid.UUID
	Token      *string
}

// CreateLocationPayload is the payload type of the storage service
// create_location method.
type CreateLocationPayload struct {
	Name        string
	Description *string
	Source      string
	Purpose     string
	Config      interface {
		configVal()
	}
	Token *string
}

// CreateLocationResult is the result type of the storage service
// create_location method.
type CreateLocationResult struct {
	UUID string
}

// DownloadAipPayload is the payload type of the storage service download_aip
// method.
type DownloadAipPayload struct {
	// Identifier of AIP
	UUID  string
	Token *string
}

// ListLocationAipsPayload is the payload type of the storage service
// list_location_aips method.
type ListLocationAipsPayload struct {
	// Identifier of location
	UUID  string
	Token *string
}

// ListLocationsPayload is the payload type of the storage service
// list_locations method.
type ListLocationsPayload struct {
	Token *string
}

// Location is the result type of the storage service show_location method.
type Location struct {
	// Name of location
	Name string
	// Description of the location
	Description *string
	// Data source of the location
	Source string
	// Purpose of the location
	Purpose string
	UUID    uuid.UUID
	Config  interface {
		configVal()
	}
	// Creation datetime
	CreatedAt string
}

// LocationCollection is the result type of the storage service list_locations
// method.
type LocationCollection []*Location

// Storage location not found.
type LocationNotFound struct {
	// Message of error
	Message string
	UUID    uuid.UUID
}

type MonitorPingEvent struct {
	Message *string
}

// MoveAipPayload is the payload type of the storage service move_aip method.
type MoveAipPayload struct {
	// Identifier of AIP
	UUID string
	// Identifier of storage location
	LocationID uuid.UUID
	Token      *string
}

// MoveAipStatusPayload is the payload type of the storage service
// move_aip_status method.
type MoveAipStatusPayload struct {
	// Identifier of AIP
	UUID  string
	Token *string
}

// MoveStatusResult is the result type of the storage service move_aip_status
// method.
type MoveStatusResult struct {
	Done bool
}

// RejectAipPayload is the payload type of the storage service reject_aip
// method.
type RejectAipPayload struct {
	// Identifier of AIP
	UUID  string
	Token *string
}

type S3Config struct {
	Bucket    string
	Region    string
	Endpoint  *string
	PathStyle *bool
	Profile   *string
	Key       *string
	Secret    *string
	Token     *string
}

type SFTPConfig struct {
	Address   string
	Username  string
	Password  string
	Directory string
}

// SIP describes an ingest SIP type.
type SIP struct {
	// Identifier of SIP
	ID uint
	// Name of the SIP
	Name *string
	// Identifier of storage location
	LocationID *uuid.UUID
	// Status of the SIP
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

type SIPCreatedEvent struct {
	// Identifier of SIP
	ID   uint
	Item *SIP
}

type SIPLocationUpdatedEvent struct {
	// Identifier of SIP
	ID uint
	// Identifier of storage location
	LocationID uuid.UUID
}

// SIPPreservationAction describes a preservation action of a SIP.
type SIPPreservationAction struct {
	ID          uint
	WorkflowID  string
	Type        string
	Status      string
	StartedAt   string
	CompletedAt *string
	Tasks       SIPPreservationTaskCollection
	SipID       *uint
}

type SIPPreservationActionCreatedEvent struct {
	// Identifier of preservation action
	ID   uint
	Item *SIPPreservationAction
}

type SIPPreservationActionUpdatedEvent struct {
	// Identifier of preservation action
	ID   uint
	Item *SIPPreservationAction
}

// SIPPreservationTask describes a SIP preservation action task.
type SIPPreservationTask struct {
	ID                   uint
	TaskID               string
	Name                 string
	Status               string
	StartedAt            string
	CompletedAt          *string
	Note                 *string
	PreservationActionID *uint
}

type SIPPreservationTaskCollection []*SIPPreservationTask

type SIPPreservationTaskCreatedEvent struct {
	// Identifier of preservation task
	ID   uint
	Item *SIPPreservationTask
}

type SIPPreservationTaskUpdatedEvent struct {
	// Identifier of preservation task
	ID   uint
	Item *SIPPreservationTask
}

type SIPStatusUpdatedEvent struct {
	// Identifier of SIP
	ID     uint
	Status string
}

type SIPUpdatedEvent struct {
	// Identifier of SIP
	ID   uint
	Item *SIP
}

// ShowAipPayload is the payload type of the storage service show_aip method.
type ShowAipPayload struct {
	// Identifier of AIP
	UUID  string
	Token *string
}

// ShowLocationPayload is the payload type of the storage service show_location
// method.
type ShowLocationPayload struct {
	// Identifier of location
	UUID  string
	Token *string
}

// SubmitAIPResult is the result type of the storage service submit_aip method.
type SubmitAIPResult struct {
	URL string
}

// SubmitAipPayload is the payload type of the storage service submit_aip
// method.
type SubmitAipPayload struct {
	// Identifier of AIP
	UUID  string
	Name  string
	Token *string
}

type URLConfig struct {
	URL string
}

// UpdateAipPayload is the payload type of the storage service update_aip
// method.
type UpdateAipPayload struct {
	// Identifier of AIP
	UUID  string
	Token *string
}

// Forbidden
type Forbidden string

// Unauthorized
type Unauthorized string

// Error returns an error description.
func (e *AIPNotFound) Error() string {
	return "AIP not found."
}

// ErrorName returns "AIPNotFound".
//
// Deprecated: Use GoaErrorName - https://github.com/goadesign/goa/issues/3105
func (e *AIPNotFound) ErrorName() string {
	return e.GoaErrorName()
}

// GoaErrorName returns "AIPNotFound".
func (e *AIPNotFound) GoaErrorName() string {
	return "not_found"
}

// Error returns an error description.
func (e *LocationNotFound) Error() string {
	return "Storage location not found."
}

// ErrorName returns "LocationNotFound".
//
// Deprecated: Use GoaErrorName - https://github.com/goadesign/goa/issues/3105
func (e *LocationNotFound) ErrorName() string {
	return e.GoaErrorName()
}

// GoaErrorName returns "LocationNotFound".
func (e *LocationNotFound) GoaErrorName() string {
	return "not_found"
}

// Error returns an error description.
func (e Forbidden) Error() string {
	return "Forbidden"
}

// ErrorName returns "forbidden".
//
// Deprecated: Use GoaErrorName - https://github.com/goadesign/goa/issues/3105
func (e Forbidden) ErrorName() string {
	return e.GoaErrorName()
}

// GoaErrorName returns "forbidden".
func (e Forbidden) GoaErrorName() string {
	return "forbidden"
}

// Error returns an error description.
func (e Unauthorized) Error() string {
	return "Unauthorized"
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
func (*AMSSConfig) configVal() {}
func (*S3Config) configVal()   {}
func (*SFTPConfig) configVal() {}
func (*URLConfig) configVal()  {}

// MakeNotValid builds a goa.ServiceError from an error.
func MakeNotValid(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "not_valid", false, false, false)
}

// MakeNotAvailable builds a goa.ServiceError from an error.
func MakeNotAvailable(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "not_available", false, false, false)
}

// MakeFailedDependency builds a goa.ServiceError from an error.
func MakeFailedDependency(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "failed_dependency", false, false, false)
}

// NewAIP initializes result type AIP from viewed result type AIP.
func NewAIP(vres *storageviews.AIP) *AIP {
	return newAIP(vres.Projected)
}

// NewViewedAIP initializes viewed result type AIP from result type AIP using
// the given view.
func NewViewedAIP(res *AIP, view string) *storageviews.AIP {
	p := newAIPView(res)
	return &storageviews.AIP{Projected: p, View: "default"}
}

// NewLocationCollection initializes result type LocationCollection from viewed
// result type LocationCollection.
func NewLocationCollection(vres storageviews.LocationCollection) LocationCollection {
	return newLocationCollection(vres.Projected)
}

// NewViewedLocationCollection initializes viewed result type
// LocationCollection from result type LocationCollection using the given view.
func NewViewedLocationCollection(res LocationCollection, view string) storageviews.LocationCollection {
	p := newLocationCollectionView(res)
	return storageviews.LocationCollection{Projected: p, View: "default"}
}

// NewLocation initializes result type Location from viewed result type
// Location.
func NewLocation(vres *storageviews.Location) *Location {
	return newLocation(vres.Projected)
}

// NewViewedLocation initializes viewed result type Location from result type
// Location using the given view.
func NewViewedLocation(res *Location, view string) *storageviews.Location {
	p := newLocationView(res)
	return &storageviews.Location{Projected: p, View: "default"}
}

// NewAIPCollection initializes result type AIPCollection from viewed result
// type AIPCollection.
func NewAIPCollection(vres storageviews.AIPCollection) AIPCollection {
	return newAIPCollection(vres.Projected)
}

// NewViewedAIPCollection initializes viewed result type AIPCollection from
// result type AIPCollection using the given view.
func NewViewedAIPCollection(res AIPCollection, view string) storageviews.AIPCollection {
	p := newAIPCollectionView(res)
	return storageviews.AIPCollection{Projected: p, View: "default"}
}

// newAIP converts projected type AIP to service type AIP.
func newAIP(vres *storageviews.AIPView) *AIP {
	res := &AIP{
		LocationID: vres.LocationID,
	}
	if vres.Name != nil {
		res.Name = *vres.Name
	}
	if vres.UUID != nil {
		res.UUID = *vres.UUID
	}
	if vres.Status != nil {
		res.Status = *vres.Status
	}
	if vres.ObjectKey != nil {
		res.ObjectKey = *vres.ObjectKey
	}
	if vres.CreatedAt != nil {
		res.CreatedAt = *vres.CreatedAt
	}
	if vres.Status == nil {
		res.Status = "unspecified"
	}
	return res
}

// newAIPView projects result type AIP to projected type AIPView using the
// "default" view.
func newAIPView(res *AIP) *storageviews.AIPView {
	vres := &storageviews.AIPView{
		Name:       &res.Name,
		UUID:       &res.UUID,
		Status:     &res.Status,
		ObjectKey:  &res.ObjectKey,
		LocationID: res.LocationID,
		CreatedAt:  &res.CreatedAt,
	}
	return vres
}

// newLocationCollection converts projected type LocationCollection to service
// type LocationCollection.
func newLocationCollection(vres storageviews.LocationCollectionView) LocationCollection {
	res := make(LocationCollection, len(vres))
	for i, n := range vres {
		res[i] = newLocation(n)
	}
	return res
}

// newLocationCollectionView projects result type LocationCollection to
// projected type LocationCollectionView using the "default" view.
func newLocationCollectionView(res LocationCollection) storageviews.LocationCollectionView {
	vres := make(storageviews.LocationCollectionView, len(res))
	for i, n := range res {
		vres[i] = newLocationView(n)
	}
	return vres
}

// newLocation converts projected type Location to service type Location.
func newLocation(vres *storageviews.LocationView) *Location {
	res := &Location{
		Description: vres.Description,
	}
	if vres.Name != nil {
		res.Name = *vres.Name
	}
	if vres.Source != nil {
		res.Source = *vres.Source
	}
	if vres.Purpose != nil {
		res.Purpose = *vres.Purpose
	}
	if vres.UUID != nil {
		res.UUID = *vres.UUID
	}
	if vres.CreatedAt != nil {
		res.CreatedAt = *vres.CreatedAt
	}
	if vres.Source == nil {
		res.Source = "unspecified"
	}
	if vres.Purpose == nil {
		res.Purpose = "unspecified"
	}
	return res
}

// newLocationView projects result type Location to projected type LocationView
// using the "default" view.
func newLocationView(res *Location) *storageviews.LocationView {
	vres := &storageviews.LocationView{
		Name:        &res.Name,
		Description: res.Description,
		Source:      &res.Source,
		Purpose:     &res.Purpose,
		UUID:        &res.UUID,
		CreatedAt:   &res.CreatedAt,
	}
	return vres
}

// newAIPCollection converts projected type AIPCollection to service type
// AIPCollection.
func newAIPCollection(vres storageviews.AIPCollectionView) AIPCollection {
	res := make(AIPCollection, len(vres))
	for i, n := range vres {
		res[i] = newAIP(n)
	}
	return res
}

// newAIPCollectionView projects result type AIPCollection to projected type
// AIPCollectionView using the "default" view.
func newAIPCollectionView(res AIPCollection) storageviews.AIPCollectionView {
	vres := make(storageviews.AIPCollectionView, len(res))
	for i, n := range res {
		vres[i] = newAIPView(n)
	}
	return vres
}
