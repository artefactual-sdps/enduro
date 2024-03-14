// Code generated by goa v3.15.2, DO NOT EDIT.
//
// package client
//
// Command:
// $ goa gen github.com/artefactual-sdps/enduro/internal/api/design -o
// internal/api

package package_

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// Client is the "package" service client.
type Client struct {
	MonitorRequestEndpoint      goa.Endpoint
	MonitorEndpoint             goa.Endpoint
	ListEndpoint                goa.Endpoint
	ShowEndpoint                goa.Endpoint
	PreservationActionsEndpoint goa.Endpoint
	ConfirmEndpoint             goa.Endpoint
	RejectEndpoint              goa.Endpoint
	MoveEndpoint                goa.Endpoint
	MoveStatusEndpoint          goa.Endpoint
}

// NewClient initializes a "package" service client given the endpoints.
func NewClient(monitorRequest, monitor, list, show, preservationActions, confirm, reject, move, moveStatus goa.Endpoint) *Client {
	return &Client{
		MonitorRequestEndpoint:      monitorRequest,
		MonitorEndpoint:             monitor,
		ListEndpoint:                list,
		ShowEndpoint:                show,
		PreservationActionsEndpoint: preservationActions,
		ConfirmEndpoint:             confirm,
		RejectEndpoint:              reject,
		MoveEndpoint:                move,
		MoveStatusEndpoint:          moveStatus,
	}
}

// MonitorRequest calls the "monitor_request" endpoint of the "package" service.
// MonitorRequest may return the following errors:
//   - "not_available" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) MonitorRequest(ctx context.Context, p *MonitorRequestPayload) (res *MonitorRequestResult, err error) {
	var ires any
	ires, err = c.MonitorRequestEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*MonitorRequestResult), nil
}

// Monitor calls the "monitor" endpoint of the "package" service.
// Monitor may return the following errors:
//   - "not_available" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Monitor(ctx context.Context, p *MonitorPayload) (res MonitorClientStream, err error) {
	var ires any
	ires, err = c.MonitorEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(MonitorClientStream), nil
}

// List calls the "list" endpoint of the "package" service.
// List may return the following errors:
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) List(ctx context.Context, p *ListPayload) (res *ListResult, err error) {
	var ires any
	ires, err = c.ListEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*ListResult), nil
}

// Show calls the "show" endpoint of the "package" service.
// Show may return the following errors:
//   - "not_found" (type *PackageNotFound): Package not found
//   - "not_available" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Show(ctx context.Context, p *ShowPayload) (res *EnduroStoredPackage, err error) {
	var ires any
	ires, err = c.ShowEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*EnduroStoredPackage), nil
}

// PreservationActions calls the "preservation_actions" endpoint of the
// "package" service.
// PreservationActions may return the following errors:
//   - "not_found" (type *PackageNotFound): Package not found
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) PreservationActions(ctx context.Context, p *PreservationActionsPayload) (res *EnduroPackagePreservationActions, err error) {
	var ires any
	ires, err = c.PreservationActionsEndpoint(ctx, p)
	if err != nil {
		return
	}
	return ires.(*EnduroPackagePreservationActions), nil
}

// Confirm calls the "confirm" endpoint of the "package" service.
// Confirm may return the following errors:
//   - "not_found" (type *PackageNotFound): Package not found
//   - "not_available" (type *goa.ServiceError)
//   - "not_valid" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Confirm(ctx context.Context, p *ConfirmPayload) (err error) {
	_, err = c.ConfirmEndpoint(ctx, p)
	return
}

// Reject calls the "reject" endpoint of the "package" service.
// Reject may return the following errors:
//   - "not_found" (type *PackageNotFound): Package not found
//   - "not_available" (type *goa.ServiceError)
//   - "not_valid" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Reject(ctx context.Context, p *RejectPayload) (err error) {
	_, err = c.RejectEndpoint(ctx, p)
	return
}

// Move calls the "move" endpoint of the "package" service.
// Move may return the following errors:
//   - "not_found" (type *PackageNotFound): Package not found
//   - "not_available" (type *goa.ServiceError)
//   - "not_valid" (type *goa.ServiceError)
//   - "unauthorized" (type Unauthorized)
//   - error: internal error
func (c *Client) Move(ctx context.Context, p *MovePayload) (err error) {
	_, err = c.MoveEndpoint(ctx, p)
	return
}

// MoveStatus calls the "move_status" endpoint of the "package" service.
// MoveStatus may return the following errors:
//   - "not_found" (type *PackageNotFound): Package not found
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
