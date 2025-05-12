package ingest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
)

var ErrInvalid = errors.New("invalid")

type Service interface {
	// Goa returns an implementation of the goaingest Service.
	Goa() goaingest.Service
	CreateSIP(context.Context, *datatypes.SIP) error
	UpdateSIP(
		ctx context.Context,
		id uuid.UUID,
		name, aipID string,
		status enums.SIPStatus,
		completedAt time.Time,
	) error
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
	logger         logr.Logger
	db             *sqlx.DB
	tc             temporalsdk_client.Client
	evsvc          event.EventService
	perSvc         persistence.Service
	tokenVerifier  auth.TokenVerifier
	ticketProvider *auth.TicketProvider
	taskQueue      string
	uploadBucket   *blob.Bucket
	uploadMaxSize  int64
}

var _ Service = (*ingestImpl)(nil)

func NewService(
	logger logr.Logger,
	db *sql.DB,
	tc temporalsdk_client.Client,
	evsvc event.EventService,
	psvc persistence.Service,
	tokenVerifier auth.TokenVerifier,
	ticketProvider *auth.TicketProvider,
	taskQueue string,
	uploadBucket *blob.Bucket,
	uploadMaxSize int64,
) *ingestImpl {
	return &ingestImpl{
		logger:         logger,
		db:             sqlx.NewDb(db, "mysql"),
		tc:             tc,
		evsvc:          evsvc,
		perSvc:         psvc,
		tokenVerifier:  tokenVerifier,
		ticketProvider: ticketProvider,
		taskQueue:      taskQueue,
		uploadBucket:   uploadBucket,
		uploadMaxSize:  uploadMaxSize,
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
	name, aipID string,
	status enums.SIPStatus,
	completedAt time.Time,
) error {
	// Ensure that completedAt is reset during retries.
	compAt := &completedAt
	if status == enums.SIPStatusProcessing {
		compAt = nil
	}
	if compAt != nil && compAt.IsZero() {
		compAt = nil
	}

	query := `UPDATE sip SET name = ?, aip_id = ?, status = ?, completed_at = ? WHERE uuid = ?`
	args := []interface{}{
		name,
		aipID,
		status,
		compAt,
		id.String(),
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	if s, err := svc.Goa().ShowSip(ctx, &goaingest.ShowSipPayload{UUID: id.String()}); err == nil {
		ev := &goaingest.SIPUpdatedEvent{UUID: id, Item: s}
		event.PublishEvent(ctx, svc.evsvc, ev)
	}

	return nil
}

func (svc *ingestImpl) SetStatus(ctx context.Context, id uuid.UUID, status enums.SIPStatus) error {
	query := `UPDATE sip SET status = ? WHERE uuid = ?`
	args := []interface{}{
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
	args := []interface{}{enums.SIPStatusProcessing}

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

func (svc *ingestImpl) updateRow(ctx context.Context, query string, args []interface{}) error {
	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating SIP: %v", err)
	}

	return nil
}

func (svc *ingestImpl) read(ctx context.Context, id string) (*datatypes.SIP, error) {
	query := "SELECT id, uuid, name, aip_id, status, CONVERT_TZ(created_at, @@session.time_zone, '+00:00') AS created_at, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at FROM sip WHERE uuid = ?"
	args := []interface{}{id}
	c := datatypes.SIP{}

	if err := svc.db.GetContext(ctx, &c, query, args...); err != nil {
		return nil, err
	}

	return &c, nil
}
