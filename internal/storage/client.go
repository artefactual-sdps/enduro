package storage

import (
	"context"
	"io"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

type Client interface {
	CreateAip(context.Context, *goastorage.CreateAipPayload) (*goastorage.AIP, error)
	ListAips(context.Context, *goastorage.ListAipsPayload) (*goastorage.AIPs, error)
	ShowAip(context.Context, *goastorage.ShowAipPayload) (*goastorage.AIP, error)

	SubmitAip(context.Context, *goastorage.SubmitAipPayload) (*goastorage.SubmitAIPResult, error)
	SubmitAipComplete(context.Context, *goastorage.SubmitAipCompletePayload) error

	DownloadAipRequest(
		context.Context,
		*goastorage.DownloadAipRequestPayload,
	) (*goastorage.DownloadAipRequestResult, error)
	DownloadAip(context.Context, *goastorage.DownloadAipPayload) (*goastorage.DownloadAipResult, io.ReadCloser, error)

	MoveAip(context.Context, *goastorage.MoveAipPayload) error
	MoveAipStatus(context.Context, *goastorage.MoveAipStatusPayload) (*goastorage.MoveStatusResult, error)

	RejectAip(context.Context, *goastorage.RejectAipPayload) error

	RequestAipDeletion(context.Context, *goastorage.RequestAipDeletionPayload) error
	ReviewAipDeletion(context.Context, *goastorage.ReviewAipDeletionPayload) error

	CreateLocation(context.Context, *goastorage.CreateLocationPayload) (*goastorage.CreateLocationResult, error)
	ListLocations(context.Context, *goastorage.ListLocationsPayload) (goastorage.LocationCollection, error)
	ShowLocation(context.Context, *goastorage.ShowLocationPayload) (*goastorage.Location, error)
	ListLocationAips(context.Context, *goastorage.ListLocationAipsPayload) (goastorage.AIPCollection, error)

	ListAipWorkflows(context.Context, *goastorage.ListAipWorkflowsPayload) (*goastorage.AIPWorkflows, error)

	MonitorRequest(context.Context, *goastorage.MonitorRequestPayload) (*goastorage.MonitorRequestResult, error)
	Monitor(context.Context, *goastorage.MonitorPayload) (goastorage.MonitorClientStream, error)
}

var _ Client = (*goastorage.Client)(nil)
