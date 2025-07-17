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
	"github.com/jmoiron/sqlx"
	temporalsdk_client "go.temporal.io/sdk/client"
	"gocloud.dev/blob"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
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
	// Goa returns an implementation of the goaingest Service.
	Goa() goaingest.Service
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
}

type ingestImpl struct {
	logger          logr.Logger
	db              *sqlx.DB
	tc              temporalsdk_client.Client
	evsvc           event.EventService
	perSvc          persistence.Service
	tokenVerifier   auth.TokenVerifier
	ticketProvider  auth.TicketProvider
	taskQueue       string
	internalStorage *blob.Bucket
	uploadMaxSize   int64
	rander          io.Reader
	sipSource       sipsource.SIPSource
}

var _ Service = (*ingestImpl)(nil)

func NewService(
	logger logr.Logger,
	db *sql.DB,
	tc temporalsdk_client.Client,
	evsvc event.EventService,
	psvc persistence.Service,
	tokenVerifier auth.TokenVerifier,
	ticketProvider auth.TicketProvider,
	taskQueue string,
	internalBucket *blob.Bucket,
	uploadMaxSize int64,
	rander io.Reader,
	sipSource sipsource.SIPSource,
) *ingestImpl {
	return &ingestImpl{
		logger:          logger,
		db:              sqlx.NewDb(db, "mysql"),
		tc:              tc,
		evsvc:           evsvc,
		perSvc:          psvc,
		tokenVerifier:   tokenVerifier,
		ticketProvider:  ticketProvider,
		taskQueue:       taskQueue,
		internalStorage: internalBucket,
		uploadMaxSize:   uploadMaxSize,
		rander:          rander,
		sipSource:       sipSource,
	}
}

func (svc *ingestImpl) Goa() goaingest.Service {
	return &goaWrapper{
		ingestImpl: svc,
	}
}

// CreateSIP persists s to the data store then updates it from the data store,
// adding generated data (e.g. ID, CreatedAt).
func (svc *ingestImpl) CreateSIP(ctx context.Context, s *datatypes.SIP) error {
	err := svc.perSvc.CreateSIP(ctx, s)
	if err != nil {
		return fmt.Errorf("ingest: create SIP: %v", err)
	}

	event.PublishEvent(ctx, svc.evsvc, sipToCreatedEvent(s))

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
	event.PublishEvent(ctx, svc.evsvc, ev)

	return s, nil
}

func (svc *ingestImpl) SetStatus(ctx context.Context, id uuid.UUID, status enums.SIPStatus) error {
	query := `UPDATE sip SET status = ? WHERE uuid = ?`
	args := []any{
		status,
		id.String(),
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	ev := &goaingest.SIPStatusUpdatedEvent{UUID: id, Status: status.String()}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *ingestImpl) SetStatusInProgress(ctx context.Context, id uuid.UUID, startedAt time.Time) error {
	var query string
	args := []any{enums.SIPStatusProcessing}

	if !startedAt.IsZero() {
		query = `UPDATE sip SET status = ?, started_at = ? WHERE uuid = ?`
		args = append(args, startedAt, id.String())
	} else {
		query = `UPDATE sip SET status = ? WHERE uuid = ?`
		args = append(args, id.String())
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	event.PublishEvent(ctx, svc.evsvc, &goaingest.SIPStatusUpdatedEvent{
		UUID:   id,
		Status: enums.SIPStatusProcessing.String(),
	})

	return nil
}

func (svc *ingestImpl) updateRow(ctx context.Context, query string, args []any) error {
	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating SIP: %v", err)
	}

	return nil
}
