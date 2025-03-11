// Code generated by goa v3.15.2, DO NOT EDIT.
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

// AIPs is the viewed result type that is projected based on a view.
type AIPs struct {
	// Type to project
	Projected *AIPsView
	// View to render
	View string
}

// AIP is the viewed result type that is projected based on a view.
type AIP struct {
	// Type to project
	Projected *AIPView
	// View to render
	View string
}

// LocationCollection is the viewed result type that is projected based on a
// view.
type LocationCollection struct {
	// Type to project
	Projected LocationCollectionView
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

// AIPCollection is the viewed result type that is projected based on a view.
type AIPCollection struct {
	// Type to project
	Projected AIPCollectionView
	// View to render
	View string
}

// AIPsView is a type that runs validations on a projected type.
type AIPsView struct {
	Items AIPCollectionView
	Page  *EnduroPageView
}

// AIPCollectionView is a type that runs validations on a projected type.
type AIPCollectionView []*AIPView

// AIPView is a type that runs validations on a projected type.
type AIPView struct {
	Name *string
	UUID *uuid.UUID
	// Status of the AIP
	Status    *string
	ObjectKey *uuid.UUID
	// Identifier of storage location
	LocationID *uuid.UUID
	// Creation datetime
	CreatedAt *string
}

// EnduroPageView is a type that runs validations on a projected type.
type EnduroPageView struct {
	// Maximum items per page
	Limit *int
	// Offset from first result to start of page
	Offset *int
	// Total result count before paging
	Total *int
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

// AMSSConfigView is a type that runs validations on a projected type.
type AMSSConfigView struct {
	APIKey   *string
	URL      *string
	Username *string
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

// URLConfigView is a type that runs validations on a projected type.
type URLConfigView struct {
	URL *string
}

func (*AMSSConfigView) configVal() {}
func (*S3ConfigView) configVal()   {}
func (*SFTPConfigView) configVal() {}
func (*URLConfigView) configVal()  {}

var (
	// AIPsMap is a map indexing the attribute names of AIPs by view name.
	AIPsMap = map[string][]string{
		"default": {
			"items",
			"page",
		},
	}
	// AIPMap is a map indexing the attribute names of AIP by view name.
	AIPMap = map[string][]string{
		"default": {
			"name",
			"uuid",
			"status",
			"object_key",
			"location_id",
			"created_at",
		},
	}
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
	// AIPCollectionMap is a map indexing the attribute names of AIPCollection by
	// view name.
	AIPCollectionMap = map[string][]string{
		"default": {
			"name",
			"uuid",
			"status",
			"object_key",
			"location_id",
			"created_at",
		},
	}
	// EnduroPageMap is a map indexing the attribute names of EnduroPage by view
	// name.
	EnduroPageMap = map[string][]string{
		"default": {
			"limit",
			"offset",
			"total",
		},
	}
)

// ValidateAIPs runs the validations defined on the viewed result type AIPs.
func ValidateAIPs(result *AIPs) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateAIPsView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateAIP runs the validations defined on the viewed result type AIP.
func ValidateAIP(result *AIP) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateAIPView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

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

// ValidateAIPCollection runs the validations defined on the viewed result type
// AIPCollection.
func ValidateAIPCollection(result AIPCollection) (err error) {
	switch result.View {
	case "default", "":
		err = ValidateAIPCollectionView(result.Projected)
	default:
		err = goa.InvalidEnumValueError("view", result.View, []any{"default"})
	}
	return
}

// ValidateAIPsView runs the validations defined on AIPsView using the
// "default" view.
func ValidateAIPsView(result *AIPsView) (err error) {

	if result.Items != nil {
		if err2 := ValidateAIPCollectionView(result.Items); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	if result.Page != nil {
		if err2 := ValidateEnduroPageView(result.Page); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateAIPCollectionView runs the validations defined on AIPCollectionView
// using the "default" view.
func ValidateAIPCollectionView(result AIPCollectionView) (err error) {
	for _, item := range result {
		if err2 := ValidateAIPView(item); err2 != nil {
			err = goa.MergeErrors(err, err2)
		}
	}
	return
}

// ValidateAIPView runs the validations defined on AIPView using the "default"
// view.
func ValidateAIPView(result *AIPView) (err error) {
	if result.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "result"))
	}
	if result.UUID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("uuid", "result"))
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

// ValidateEnduroPageView runs the validations defined on EnduroPageView using
// the "default" view.
func ValidateEnduroPageView(result *EnduroPageView) (err error) {
	if result.Limit == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("limit", "result"))
	}
	if result.Offset == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("offset", "result"))
	}
	if result.Total == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("total", "result"))
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
		if !(*result.Source == "unspecified" || *result.Source == "minio" || *result.Source == "sftp" || *result.Source == "amss") {
			err = goa.MergeErrors(err, goa.InvalidEnumValueError("result.source", *result.Source, []any{"unspecified", "minio", "sftp", "amss"}))
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

// ValidateAMSSConfigView runs the validations defined on AMSSConfigView.
func ValidateAMSSConfigView(result *AMSSConfigView) (err error) {
	if result.APIKey == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("api_key", "result"))
	}
	if result.URL == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("url", "result"))
	}
	if result.Username == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("username", "result"))
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

// ValidateURLConfigView runs the validations defined on URLConfigView.
func ValidateURLConfigView(result *URLConfigView) (err error) {
	if result.URL == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("url", "result"))
	}
	return
}
