// Code generated by goa v3.11.3, DO NOT EDIT.
//
// storage views
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package views

import (
	"github.com/google/uuid"
	goa "goa.design/goa/v3/pkg"
)

// LocationCollection is the viewed result type that is projected based on a
// view.
type LocationCollection struct {
	// Type to project
	Projected LocationCollectionView
	// View to render
	View string
}

// Package is the viewed result type that is projected based on a view.
type Package struct {
	// Type to project
	Projected *PackageView
	// View to render
	View string
}

// Location is the viewed result type that is projected based on a view.
type Location struct {
	// Type to project
	Projected *LocationView
	// View to render
	View string
}

// PackageCollection is the viewed result type that is projected based on a
// view.
type PackageCollection struct {
	// Type to project
	Projected PackageCollectionView
	// View to render
	View string
}

// LocationCollectionView is a type that runs validations on a projected type.
type LocationCollectionView []*LocationView

// LocationView is a type that runs validations on a projected type.
type LocationView struct {
	// Name of location
	Name *string
	// Description of the location
	Description *string
	// Data source of the location
	Source *string
	// Purpose of the location
	Purpose *string
	UUID    *uuid.UUID
	Config  interface {
		configVal()
	}
	// Creation datetime
	CreatedAt *string
}

// S3ConfigView is a type that runs validations on a projected type.
type S3ConfigView struct {
	Bucket    *string
	Region    *string
	Endpoint  *string
	PathStyle *bool
	Profile   *string
	Key       *string
	Secret    *string
	Token     *string
}

// SFTPConfigView is a type that runs validations on a projected type.
type SFTPConfigView struct {
	Address   *string
	Username  *string
	Password  *string
	Directory *string
}

// PackageView is a type that runs validations on a projected type.
type PackageView struct {
	Name  *string
	AipID *uuid.UUID
	// Status of the package
	Status    *string
	ObjectKey *uuid.UUID
	// Identifier of storage location
	LocationID *uuid.UUID
	// Creation datetime
	CreatedAt *string
}

// PackageCollectionView is a type that runs validations on a projected type.
type PackageCollectionView []*PackageView

func (*S3ConfigView) configVal()   {}
func (*SFTPConfigView) configVal() {}

var (
	// LocationCollectionMap is a map indexing the attribute names of
	// LocationCollection by view name.
	LocationCollectionMap = map[string][]string{
		"default": {
			"name",
			"description",
			"source",
			"purpose",
			"uuid",
			"created_at",
		},
	}
	// PackageMap is a map indexing the attribute names of Package by view name.
	PackageMap = map[string][]string{
		"default": {
			"name",
			"aip_id",
			"status",
			"object_key",
			"location_id",
			"created_at",
		},
	}
	// LocationMap is a map indexing the attribute names of Location by view name.
	LocationMap = map[string][]string{
		"default": {
			"name",
			"description",
			"source",
			"purpose",
			"uuid",
			"created_at",
		},
	}
	// PackageCollectionMap is a map indexing the attribute names of
	// PackageCollection by view name.
	PackageCollectionMap = map[string][]string{
		"default": {
			"name",
			"aip_id",
			"status",
			"object_key",
			"location_id",
			"created_at",
		},
	}
)

// ValidateLocationCollection runs the validations defined on the viewed result
// type LocationCollection.
func ValidateLocationCollection(result LocationCollection) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateLocationCollectionView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidatePackage runs the validations defined on the viewed result type
// Package.
func ValidatePackage(result *Package) (err error) {
	switch result.View {
	case "default", "":
		err = ValidatePackageView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateLocation runs the validations defined on the viewed result type
// Location.
func ValidateLocation(result *Location) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateLocationView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidatePackageCollection runs the validations defined on the viewed result
// type PackageCollection.
func ValidatePackageCollection(result PackageCollection) (err error) {
	switch result.View {
	case "default", "":
		err = ValidatePackageCollectionView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateLocationCollectionView runs the validations defined on
// LocationCollectionView using the "default" view.
func ValidateLocationCollectionView(result LocationCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateLocationView(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateLocationView runs the validations defined on LocationView using the
// "default" view.
func ValidateLocationView(result *LocationView) (err error) {
	if result.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "result"))
	}
	if result.Source == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("source", "result"))
	}
	if result.Purpose == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("purpose", "result"))
	}
	if result.UUID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("uuid", "result"))
	}
	if result.CreatedAt == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("created_at", "result"))
	}
	if result.Source != nil {
		if !(*result.Source == "unspecified" || *result.Source == "minio" || *result.Source == "sftp") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("result.source", *result.Source, []any{"unspecified", "minio", "sftp"}))
		}
	}
	if result.Purpose != nil {
		if !(*result.Purpose == "unspecified" || *result.Purpose == "aip_store") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("result.purpose", *result.Purpose, []any{"unspecified", "aip_store"}))
		}
	}
	if result.CreatedAt != nil {
		err = goa.MergeErrors(err, goa.ValidateFormat("result.created_at", *result.CreatedAt, goa.FormatDateTime))
	}
	return
}

// ValidateS3ConfigView runs the validations defined on S3ConfigView.
func ValidateS3ConfigView(result *S3ConfigView) (err error) {
	if result.Bucket == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("bucket", "result"))
	}
	if result.Region == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("region", "result"))
	}
	return
}

// ValidateSFTPConfigView runs the validations defined on SFTPConfigView.
func ValidateSFTPConfigView(result *SFTPConfigView) (err error) {
	if result.Address == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("address", "result"))
	}
	if result.Username == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("username", "result"))
	}
	if result.Password == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("password", "result"))
	}
	if result.Directory == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("directory", "result"))
	}
	return
}

// ValidatePackageView runs the validations defined on PackageView using the
// "default" view.
func ValidatePackageView(result *PackageView) (err error) {
	if result.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "result"))
	}
	if result.AipID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("aip_id", "result"))
	}
	if result.Status == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("status", "result"))
	}
	if result.ObjectKey == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("object_key", "result"))
	}
	if result.CreatedAt == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("created_at", "result"))
	}
	if result.Status != nil {
		if !(*result.Status == "unspecified" || *result.Status == "in_review" || *result.Status == "rejected" || *result.Status == "stored" || *result.Status == "moving") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("result.status", *result.Status, []any{"unspecified", "in_review", "rejected", "stored", "moving"}))
		}
	}
	if result.CreatedAt != nil {
		err = goa.MergeErrors(err, goa.ValidateFormat("result.created_at", *result.CreatedAt, goa.FormatDateTime))
	}
	return
}

// ValidatePackageCollectionView runs the validations defined on
// PackageCollectionView using the "default" view.
func ValidatePackageCollectionView(result PackageCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidatePackageView(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}
