// Code generated by goa v3.15.2, DO NOT EDIT.
//
// storage client
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package storage

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// Client is the "storage" service client.
type Client struct {
	SubmitEndpoint           goa.Endpoint
	UpdateEndpoint           goa.Endpoint
	DownloadEndpoint         goa.Endpoint
	LocationsEndpoint        goa.Endpoint
	AddLocationEndpoint      goa.Endpoint
	MoveEndpoint             goa.Endpoint
	MoveStatusEndpoint       goa.Endpoint
	RejectEndpoint           goa.Endpoint
	ShowEndpoint             goa.Endpoint
	ShowLocationEndpoint     goa.Endpoint
	LocationPackagesEndpoint goa.Endpoint
}

// NewClient initializes a "storage" service client given the endpoints.
func NewClient(submit, update, download, locations, addLocation, move, moveStatus, reject, show, showLocation, locationPackages goa.Endpoint) *Client {
	return &Client{
		SubmitEndpoint:           submit,
		UpdateEndpoint:           update,
		DownloadEndpoint:         download,
		LocationsEndpoint:        locations,
		AddLocationEndpoint:      addLocation,
		MoveEndpoint:             move,
		MoveStatusEndpoint:       moveStatus,
		RejectEndpoint:           reject,
		ShowEndpoint:             show,
		ShowLocationEndpoint:     showLocation,
		LocationPackagesEndpoint: locationPackages,
	}
}

// Submit calls the "submit" endpoint of the "storage" service.
// Submit may return the following errors:
//   - "not_found" (type *PackageNotFound): Storage package not found
//   - "not_available" (type *goa.ServiceError)
//   - "not_valid" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Submit(ctx context.Context, p *SubmitPayload) (res *SubmitResult, err error) {
	var ires any
	ires, err = c.SubmitEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*SubmitResult), nil
}

// Update calls the "update" endpoint of the "storage" service.
// Update may return the following errors:
//   - "not_found" (type *PackageNotFound): Storage package not found
//   - "not_available" (type *goa.ServiceError)
//   - "not_valid" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Update(ctx context.Context, p *UpdatePayload) (err error) {
	_, err = c.UpdateEndpoint(ctx, p)
	return
}

// Download calls the "download" endpoint of the "storage" service.
// Download may return the following errors:
//   - "not_found" (type *PackageNotFound): Storage package not found
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Download(ctx context.Context, p *DownloadPayload) (res []byte, err error) {
	var ires any
	ires, err = c.DownloadEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.([]byte), nil
}

// Locations calls the "locations" endpoint of the "storage" service.
// Locations may return the following errors:
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Locations(ctx context.Context, p *LocationsPayload) (res LocationCollection, err error) {
	var ires any
	ires, err = c.LocationsEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(LocationCollection), nil
}

// AddLocation calls the "add_location" endpoint of the "storage" service.
// AddLocation may return the following errors:
//   - "not_valid" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) AddLocation(ctx context.Context, p *AddLocationPayload) (res *AddLocationResult, err error) {
	var ires any
	ires, err = c.AddLocationEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*AddLocationResult), nil
}

// Move calls the "move" endpoint of the "storage" service.
// Move may return the following errors:
//   - "not_found" (type *PackageNotFound): Storage package not found
//   - "not_available" (type *goa.ServiceError)
//   - "not_valid" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Move(ctx context.Context, p *MovePayload) (err error) {
	_, err = c.MoveEndpoint(ctx, p)
	return
}

// MoveStatus calls the "move_status" endpoint of the "storage" service.
// MoveStatus may return the following errors:
//   - "not_found" (type *PackageNotFound): Storage package not found
//   - "failed_dependency" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) MoveStatus(ctx context.Context, p *MoveStatusPayload) (res *MoveStatusResult, err error) {
	var ires any
	ires, err = c.MoveStatusEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*MoveStatusResult), nil
}

// Reject calls the "reject" endpoint of the "storage" service.
// Reject may return the following errors:
//   - "not_found" (type *PackageNotFound): Storage package not found
//   - "not_available" (type *goa.ServiceError)
//   - "not_valid" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Reject(ctx context.Context, p *RejectPayload) (err error) {
	_, err = c.RejectEndpoint(ctx, p)
	return
}

// Show calls the "show" endpoint of the "storage" service.
// Show may return the following errors:
//   - "not_found" (type *PackageNotFound): Storage package not found
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Show(ctx context.Context, p *ShowPayload) (res *Package, err error) {
	var ires any
	ires, err = c.ShowEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*Package), nil
}

// ShowLocation calls the "show_location" endpoint of the "storage" service.
// ShowLocation may return the following errors:
//   - "not_found" (type *LocationNotFound): Storage location not found
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) ShowLocation(ctx context.Context, p *ShowLocationPayload) (res *Location, err error) {
	var ires any
	ires, err = c.ShowLocationEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*Location), nil
}

// LocationPackages calls the "location_packages" endpoint of the "storage"
// service.
// LocationPackages may return the following errors:
//   - "not_found" (type *LocationNotFound): Storage location not found
//   - "not_valid" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) LocationPackages(ctx context.Context, p *LocationPackagesPayload) (res PackageCollection, err error) {
	var ires any
	ires, err = c.LocationPackagesEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(PackageCollection), nil
}
