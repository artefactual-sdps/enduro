// Code generated by goa v3.11.3, DO NOT EDIT.
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

// The storage service manages the storage of packages.
type Service interface {
	// Start the submission of a package
	Submit(context.Context, *SubmitPayload) (res *SubmitResult, err error)
	// Signal the storage service that an upload is complete
	Update(context.Context, *UpdatePayload) (err error)
	// Download package by AIPID
	Download(context.Context, *DownloadPayload) (res []byte, err error)
	// List locations
	Locations(context.Context, *LocationsPayload) (res LocationCollection, err error)
	// Add a storage location
	AddLocation(context.Context, *AddLocationPayload) (res *AddLocationResult, err error)
	// Move a package to a permanent storage location
	Move(context.Context, *MovePayload) (err error)
	// Retrieve the status of a permanent storage location move of the package
	MoveStatus(context.Context, *MoveStatusPayload) (res *MoveStatusResult, err error)
	// Reject a package
	Reject(context.Context, *RejectPayload) (err error)
	// Show package by AIPID
	Show(context.Context, *ShowPayload) (res *Package, err error)
	// Show location by UUID
	ShowLocation(context.Context, *ShowLocationPayload) (res *Location, err error)
	// List all the packages stored in the location with UUID
	LocationPackages(context.Context, *LocationPackagesPayload) (res PackageCollection, err error)
}

// Auther defines the authorization functions to be implemented by the service.
type Auther interface {
	// OAuth2Auth implements the authorization logic for the OAuth2 security scheme.
	OAuth2Auth(ctx context.Context, token string, schema *security.OAuth2Scheme) (context.Context, error)
}

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "storage"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [11]string{"submit", "update", "download", "locations", "add_location", "move", "move_status", "reject", "show", "show_location", "location_packages"}

// AddLocationPayload is the payload type of the storage service add_location
// method.
type AddLocationPayload struct {
	Name        string
	Description *string
	Source      string
	Purpose     string
	Config      interface {
		configVal()
	}
	OauthToken *string
}

// AddLocationResult is the result type of the storage service add_location
// method.
type AddLocationResult struct {
	UUID string
}

// DownloadPayload is the payload type of the storage service download method.
type DownloadPayload struct {
	// Identifier of AIP
	AipID      string
	OauthToken *string
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

// LocationCollection is the result type of the storage service locations
// method.
type LocationCollection []*Location

// Storage location not found.
type LocationNotFound struct {
	// Message of error
	Message string
	UUID    uuid.UUID
}

// LocationPackagesPayload is the payload type of the storage service
// location_packages method.
type LocationPackagesPayload struct {
	// Identifier of location
	UUID       string
	OauthToken *string
}

// LocationsPayload is the payload type of the storage service locations method.
type LocationsPayload struct {
	OauthToken *string
}

// MovePayload is the payload type of the storage service move method.
type MovePayload struct {
	// Identifier of AIP
	AipID string
	// Identifier of storage location
	LocationID uuid.UUID
	OauthToken *string
}

// MoveStatusPayload is the payload type of the storage service move_status
// method.
type MoveStatusPayload struct {
	// Identifier of AIP
	AipID      string
	OauthToken *string
}

// MoveStatusResult is the result type of the storage service move_status
// method.
type MoveStatusResult struct {
	Done bool
}

// Package is the result type of the storage service show method.
type Package struct {
	Name  string
	AipID uuid.UUID
	// Status of the package
	Status    string
	ObjectKey uuid.UUID
	// Identifier of storage location
	LocationID *uuid.UUID
	// Creation datetime
	CreatedAt string
}

// PackageCollection is the result type of the storage service
// location_packages method.
type PackageCollection []*Package

// Storage package not found.
type PackageNotFound struct {
	// Message of error
	Message string
	// Identifier of missing package
	AipID uuid.UUID
}

// RejectPayload is the payload type of the storage service reject method.
type RejectPayload struct {
	// Identifier of AIP
	AipID      string
	OauthToken *string
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

// ShowLocationPayload is the payload type of the storage service show_location
// method.
type ShowLocationPayload struct {
	// Identifier of location
	UUID       string
	OauthToken *string
}

// ShowPayload is the payload type of the storage service show method.
type ShowPayload struct {
	// Identifier of AIP
	AipID      string
	OauthToken *string
}

// SubmitPayload is the payload type of the storage service submit method.
type SubmitPayload struct {
	// Identifier of AIP
	AipID      string
	Name       string
	OauthToken *string
}

// SubmitResult is the result type of the storage service submit method.
type SubmitResult struct {
	URL string
}

type URLConfig struct {
	URL string
}

// UpdatePayload is the payload type of the storage service update method.
type UpdatePayload struct {
	// Identifier of AIP
	AipID      string
	OauthToken *string
}

// Invalid token
type Unauthorized string

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
func (e *PackageNotFound) Error() string {
	return "Storage package not found."
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
func (*S3Config) configVal()   {}
func (*SFTPConfig) configVal() {}
func (*URLConfig) configVal()  {}

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

// NewPackage initializes result type Package from viewed result type Package.
func NewPackage(vres *storageviews.Package) *Package {
	return newPackage(vres.Projected)
}

// NewViewedPackage initializes viewed result type Package from result type
// Package using the given view.
func NewViewedPackage(res *Package, view string) *storageviews.Package {
	p := newPackageView(res)
	return &storageviews.Package{Projected: p, View: "default"}
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

// NewPackageCollection initializes result type PackageCollection from viewed
// result type PackageCollection.
func NewPackageCollection(vres storageviews.PackageCollection) PackageCollection {
	return newPackageCollection(vres.Projected)
}

// NewViewedPackageCollection initializes viewed result type PackageCollection
// from result type PackageCollection using the given view.
func NewViewedPackageCollection(res PackageCollection, view string) storageviews.PackageCollection {
	p := newPackageCollectionView(res)
	return storageviews.PackageCollection{Projected: p, View: "default"}
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

// newPackage converts projected type Package to service type Package.
func newPackage(vres *storageviews.PackageView) *Package {
	res := &Package{
		LocationID: vres.LocationID,
	}
	if vres.Name != nil {
		res.Name = *vres.Name
	}
	if vres.AipID != nil {
		res.AipID = *vres.AipID
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

// newPackageView projects result type Package to projected type PackageView
// using the "default" view.
func newPackageView(res *Package) *storageviews.PackageView {
	vres := &storageviews.PackageView{
		Name:       &res.Name,
		AipID:      &res.AipID,
		Status:     &res.Status,
		ObjectKey:  &res.ObjectKey,
		LocationID: res.LocationID,
		CreatedAt:  &res.CreatedAt,
	}
	return vres
}

// newPackageCollection converts projected type PackageCollection to service
// type PackageCollection.
func newPackageCollection(vres storageviews.PackageCollectionView) PackageCollection {
	res := make(PackageCollection, len(vres))
	for i, n := range vres {
		res[i] = newPackage(n)
	}
	return res
}

// newPackageCollectionView projects result type PackageCollection to projected
// type PackageCollectionView using the "default" view.
func newPackageCollectionView(res PackageCollection) storageviews.PackageCollectionView {
	vres := make(storageviews.PackageCollectionView, len(res))
	for i, n := range res {
		vres[i] = newPackageView(n)
	}
	return vres
}
