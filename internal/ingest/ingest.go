package ingest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	temporalsdk_client "go.temporal.io/sdk/client"
	"gocloud.dev/blob"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/sipsource"
)

const (
	// Prefix used to store the SIP in the internal bucket after upload.
	SIPPrefix = "SIP_"
	// Prefix used to store the SIP in the internal bucket after failure.
	FailedSIPPrefix = "Failed_SIP_"
	// Prefix used to store the PIP in the internal bucket after failure.
	FailedPIPPrefix = "Failed_PIP_"
)

var ErrInvalid = errors.New("invalid")

type Service interface {
	goaingest.Service

	CreateSIP(context.Context, *datatypes.SIP) error
	UpdateSIP(context.Context, uuid.UUID, persistence.SIPUpdater) (*datatypes.SIP, error)
	SetStatus(ctx context.Context, id uuid.UUID, status enums.SIPStatus) error
	SetStatusInProgress(ctx context.Context, id uuid.UUID, startedAt time.Time) error
	CreateWorkflow(ctx context.Context, w *datatypes.Workflow) error
	SetWorkflowStatus(ctx context.Context, ID int, status enums.WorkflowStatus) error
	CompleteWorkflow(
		ctx context.Context,
		ID int,
		status enums.WorkflowStatus,
		completedAt time.Time,
	) error
	CreateTask(ctx context.Context, task *datatypes.Task) error
	CompleteTask(
		ctx context.Context,
		ID int,
		status enums.TaskStatus,
		completedAt time.Time,
		note *string,
	) error
	UpdateBatch(context.Context, uuid.UUID, persistence.BatchUpdater) (*datatypes.Batch, error)
}

type ingestImpl struct {
	logger                logr.Logger
	tc                    temporalsdk_client.Client
	evsvc                 event.Service[*goaingest.IngestEvent]
	perSvc                persistence.Service
	tokenVerifier         auth.TokenVerifier
	ticketProvider        auth.TicketProvider
	taskQueue             string
	internalStorage       *blob.Bucket
	uploadMaxSize         int64
	uploadRetentionPeriod time.Duration
	rander                io.Reader
	sipSource             sipsource.SIPSource
	auditLogger           *auditlog.Logger
}

var _ Service = (*ingestImpl)(nil)

type ServiceParams struct {
	Logger                logr.Logger
	DB                    *sql.DB
	TemporalClient        temporalsdk_client.Client
	EventService          event.Service[*goaingest.IngestEvent]
	PersistenceService    persistence.Service
	TokenVerifier         auth.TokenVerifier
	TicketProvider        auth.TicketProvider
	TaskQueue             string
	InternalStorage       *blob.Bucket
	UploadMaxSize         int64
	UploadRetentionPeriod time.Duration
	Rander                io.Reader
	SIPSource             sipsource.SIPSource
	AuditLogger           *auditlog.Logger
}

func NewService(params ServiceParams) *ingestImpl {
	return &ingestImpl{
		logger:          params.Logger,
		tc:              params.TemporalClient,
		evsvc:           params.EventService,
		perSvc:          params.PersistenceService,
		tokenVerifier:   params.TokenVerifier,
		ticketProvider:  params.TicketProvider,
		taskQueue:       params.TaskQueue,
		internalStorage: params.InternalStorage,
		uploadMaxSize:   params.UploadMaxSize,
		rander:          params.Rander,
		sipSource:       params.SIPSource,
		auditLogger:     params.AuditLogger,
	}
}

// CreateSIP persists s to the data store then updates it from the data store,
// adding generated data (e.g. ID, CreatedAt).
func (svc *ingestImpl) CreateSIP(ctx context.Context, s *datatypes.SIP) error {
	err := svc.perSvc.CreateSIP(ctx, s)
	if err != nil {
		return fmt.Errorf("ingest: create SIP: %v", err)
	}

	PublishEvent(ctx, svc.evsvc, sipToCreatedEvent(s))
	svc.auditLogger.Log(ctx, sipIngestAuditEvent(s))

	return nil
}

func (svc *ingestImpl) UpdateSIP(
	ctx context.Context,
	id uuid.UUID,
	upd persistence.SIPUpdater,
) (*datatypes.SIP, error) {
	s, err := svc.perSvc.UpdateSIP(ctx, id, upd)
	if err != nil {
		return nil, fmt.Errorf("ingest: update SIP: %v", err)
	}

	ev := &goaingest.SIPUpdatedEvent{UUID: id, Item: s.Goa()}
	PublishEvent(ctx, svc.evsvc, ev)

	return s, nil
}

func (svc *ingestImpl) SetStatus(ctx context.Context, id uuid.UUID, status enums.SIPStatus) error {
	_, err := svc.perSvc.UpdateSIP(ctx, id, func(s *datatypes.SIP) (*datatypes.SIP, error) {
		s.Status = status
		return s, nil
	})
	if err != nil {
		return fmt.Errorf("error updating SIP: %v", err)
	}

	ev := &goaingest.SIPStatusUpdatedEvent{UUID: id, Status: status.String()}
	PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *ingestImpl) SetStatusInProgress(ctx context.Context, id uuid.UUID, startedAt time.Time) error {
	_, err := svc.perSvc.UpdateSIP(ctx, id, func(s *datatypes.SIP) (*datatypes.SIP, error) {
		s.Status = enums.SIPStatusProcessing
		if !startedAt.IsZero() {
			s.StartedAt = sql.NullTime{
				Time:  startedAt,
				Valid: true,
			}
		}
		return s, nil
	})
	if err != nil {
		return fmt.Errorf("error updating SIP: %v", err)
	}

	PublishEvent(ctx, svc.evsvc, &goaingest.SIPStatusUpdatedEvent{
		UUID:   id,
		Status: enums.SIPStatusProcessing.String(),
	})

	return nil
}

func (svc *ingestImpl) UpdateBatch(
	ctx context.Context,
	id uuid.UUID,
	upd persistence.BatchUpdater,
) (*datatypes.Batch, error) {
	b, err := svc.perSvc.UpdateBatch(ctx, id, upd)
	if err != nil {
		return nil, fmt.Errorf("ingest: update Batch: %v", err)
	}

	ev := &goaingest.BatchUpdatedEvent{UUID: id, Item: b.Goa()}
	PublishEvent(ctx, svc.evsvc, ev)

	return b, nil
}
