package storage

import (
	"context"
	"io"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

type Client interface {
	ListAips(context.Context, *goastorage.ListAipsPayload) (*goastorage.AIPs, error)
	SubmitAip(context.Context, *goastorage.SubmitAipPayload) (*goastorage.SubmitAIPResult, error)
	CreateAip(context.Context, *goastorage.CreateAipPayload) (*goastorage.AIP, error)
	UpdateAip(context.Context, *goastorage.UpdateAipPayload) error
	DownloadAipRequest(
		context.Context,
		*goastorage.DownloadAipRequestPayload,
	) (*goastorage.DownloadAipRequestResult, error)
	DownloadAip(context.Context, *goastorage.DownloadAipPayload) (*goastorage.DownloadAipResult, io.ReadCloser, error)
	ListLocations(context.Context, *goastorage.ListLocationsPayload) (goastorage.LocationCollection, error)
	CreateLocation(context.Context, *goastorage.CreateLocationPayload) (*goastorage.CreateLocationResult, error)
	MoveAip(context.Context, *goastorage.MoveAipPayload) error
	MoveAipStatus(context.Context, *goastorage.MoveAipStatusPayload) (*goastorage.MoveStatusResult, error)
	RejectAip(context.Context, *goastorage.RejectAipPayload) error
	ShowAip(context.Context, *goastorage.ShowAipPayload) (*goastorage.AIP, error)
	ListAipWorkflows(context.Context, *goastorage.ListAipWorkflowsPayload) (*goastorage.AIPWorkflows, error)
	RequestAipDeletion(context.Context, *goastorage.RequestAipDeletionPayload) error
	ReviewAipDeletion(context.Context, *goastorage.ReviewAipDeletionPayload) error
	ShowLocation(context.Context, *goastorage.ShowLocationPayload) (*goastorage.Location, error)
	ListLocationAips(context.Context, *goastorage.ListLocationAipsPayload) (goastorage.AIPCollection, error)
}

var _ Client = (*goastorage.Client)(nil)
