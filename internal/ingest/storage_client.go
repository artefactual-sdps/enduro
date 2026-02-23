package ingest

import (
	"context"
	"io"
	"net/http/httptrace"

	"github.com/hashicorp/go-cleanhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	goahttp "goa.design/goa/v3/http"

	goahttpstorage "github.com/artefactual-sdps/enduro/internal/api/gen/http/storage/client"
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// StorageClient represents the ingest domain dependency on the storage public API.
type StorageClient interface {
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

	AipDeletionAuto(context.Context, *goastorage.AipDeletionAutoPayload) error
	RequestAipDeletion(context.Context, *goastorage.RequestAipDeletionPayload) error
	ReviewAipDeletion(context.Context, *goastorage.ReviewAipDeletionPayload) error
	CancelAipDeletion(context.Context, *goastorage.CancelAipDeletionPayload) error

	AipDeletionReportRequest(
		context.Context,
		*goastorage.AipDeletionReportRequestPayload,
	) (*goastorage.AipDeletionReportRequestResult, error)
	AipDeletionReport(
		context.Context,
		*goastorage.AipDeletionReportPayload,
	) (*goastorage.AipDeletionReportResult, io.ReadCloser, error)

	CreateLocation(context.Context, *goastorage.CreateLocationPayload) (*goastorage.CreateLocationResult, error)
	ListLocations(context.Context, *goastorage.ListLocationsPayload) (goastorage.LocationCollection, error)
	ShowLocation(context.Context, *goastorage.ShowLocationPayload) (*goastorage.Location, error)
	ListLocationAips(context.Context, *goastorage.ListLocationAipsPayload) (goastorage.AIPCollection, error)

	ListAipWorkflows(context.Context, *goastorage.ListAipWorkflowsPayload) (*goastorage.AIPWorkflows, error)

	MonitorRequest(context.Context, *goastorage.MonitorRequestPayload) (*goastorage.MonitorRequestResult, error)
	Monitor(context.Context, *goastorage.MonitorPayload) (goastorage.MonitorClientStream, error)
}

var _ StorageClient = (*goastorage.Client)(nil)

func NewStorageClient(
	ctx context.Context,
	tp trace.TracerProvider,
	cfg StorageConfig,
) (StorageClient, error) {
	httpClient := cleanhttp.DefaultPooledClient()
	transport := httpClient.Transport
	if cfg.OIDC.Enabled {
		tokenProvider, err := NewOIDCAccessTokenProvider(ctx, cfg.OIDC)
		if err != nil {
			return nil, err
		}
		transport = NewBearerTransport(transport, tokenProvider)
	}
	httpClient.Transport = otelhttp.NewTransport(
		transport,
		otelhttp.WithTracerProvider(tp),
		otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
			return otelhttptrace.NewClientTrace(ctx)
		}),
	)

	storageHTTPClient := goahttpstorage.NewClient(
		"http",
		cfg.Address,
		httpClient,
		goahttp.RequestEncoder,
		goahttp.ResponseDecoder,
		false,
		nil,
		nil,
	)

	return goastorage.NewClient(
		storageHTTPClient.MonitorRequest(),
		storageHTTPClient.Monitor(),
		storageHTTPClient.ListAips(),
		storageHTTPClient.CreateAip(),
		storageHTTPClient.SubmitAip(),
		storageHTTPClient.SubmitAipComplete(),
		storageHTTPClient.DownloadAipRequest(),
		storageHTTPClient.DownloadAip(),
		storageHTTPClient.MoveAip(),
		storageHTTPClient.MoveAipStatus(),
		storageHTTPClient.RejectAip(),
		storageHTTPClient.ShowAip(),
		storageHTTPClient.ListAipWorkflows(),
		storageHTTPClient.AipDeletionAuto(),
		storageHTTPClient.RequestAipDeletion(),
		storageHTTPClient.ReviewAipDeletion(),
		storageHTTPClient.CancelAipDeletion(),
		storageHTTPClient.AipDeletionReportRequest(),
		storageHTTPClient.AipDeletionReport(),
		storageHTTPClient.ListLocations(),
		storageHTTPClient.CreateLocation(),
		storageHTTPClient.ShowLocation(),
		storageHTTPClient.ListLocationAips(),
	), nil
}
