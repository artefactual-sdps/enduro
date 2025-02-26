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
	Create(context.Context, *datatypes.SIP) error
	UpdateWorkflowStatus(
		ctx context.Context,
		ID int,
		name, workflowID, runID, aipID string,
		status enums.SIPStatus,
		storedAt time.Time,
	) error
	SetStatus(ctx context.Context, ID int, status enums.SIPStatus) error
	SetStatusInProgress(ctx context.Context, ID int, startedAt time.Time) error
	SetStatusPending(ctx context.Context, ID int) error
	SetLocationID(ctx context.Context, ID int, locationID uuid.UUID) error
	CreatePreservationAction(ctx context.Context, pa *datatypes.PreservationAction) error
	SetPreservationActionStatus(ctx context.Context, ID int, status enums.PreservationActionStatus) error
	CompletePreservationAction(
		ctx context.Context,
		ID int,
		status enums.PreservationActionStatus,
		completedAt time.Time,
	) error
	CreatePreservationTask(ctx context.Context, pt *datatypes.PreservationTask) error
	CompletePreservationTask(
		ctx context.Context,
		ID int,
		status enums.PreservationTaskStatus,
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

// Create persists s to the data store then updates it from the data store,
// adding generated data (e.g. ID, CreatedAt).
func (svc *ingestImpl) Create(ctx context.Context, s *datatypes.SIP) error {
	err := svc.perSvc.CreateSIP(ctx, s)
	if err != nil {
		return fmt.Errorf("ingest: create SIP: %v", err)
	}

	event.PublishEvent(ctx, svc.evsvc, sipTogoaingestCreatedEvent(s))

	return nil
}

func (svc *ingestImpl) UpdateWorkflowStatus(
	ctx context.Context,
	ID int,
	name, workflowID, runID, aipID string,
	status enums.SIPStatus,
	storedAt time.Time,
) error {
	// Ensure that storedAt is reset during retries.
	completedAt := &storedAt
	if status == enums.SIPStatusInProgress {
		completedAt = nil
	}
	if completedAt != nil && completedAt.IsZero() {
		completedAt = nil
	}

	if ID < 0 {
		return fmt.Errorf("%w: ID", ErrInvalid)
	}
	id := uint(ID) // #nosec G115 -- range validated.

	query := `UPDATE sip SET name = ?, workflow_id = ?, run_id = ?, aip_id = ?, status = ?, completed_at = ? WHERE id = ?`
	args := []interface{}{
		name,
		workflowID,
		runID,
		aipID,
		status,
		completedAt,
		ID,
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	if s, err := svc.Goa().ShowSip(ctx, &goaingest.ShowSipPayload{ID: id}); err == nil {
		ev := &goaingest.SIPUpdatedEvent{ID: s.ID, Item: s}
		event.PublishEvent(ctx, svc.evsvc, ev)
	}

	return nil
}

func (svc *ingestImpl) SetStatus(ctx context.Context, ID int, status enums.SIPStatus) error {
	if ID < 0 {
		return fmt.Errorf("%w: ID", ErrInvalid)
	}
	id := uint(ID) // #nosec G115 -- range validated.

	query := `UPDATE sip SET status = ? WHERE id = ?`
	args := []interface{}{
		status,
		ID,
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	ev := &goaingest.SIPStatusUpdatedEvent{ID: id, Status: status.String()}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *ingestImpl) SetStatusInProgress(ctx context.Context, ID int, startedAt time.Time) error {
	var query string
	args := []interface{}{enums.SIPStatusInProgress}

	if ID < 0 {
		return fmt.Errorf("%w: ID", ErrInvalid)
	}
	id := uint(ID) // #nosec G115 -- range validated.

	if !startedAt.IsZero() {
		query = `UPDATE sip SET status = ?, started_at = ? WHERE id = ?`
		args = append(args, startedAt, ID)
	} else {
		query = `UPDATE sip SET status = ? WHERE id = ?`
		args = append(args, ID)
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	event.PublishEvent(ctx, svc.evsvc, &goaingest.SIPStatusUpdatedEvent{
		ID:     id,
		Status: enums.SIPStatusInProgress.String(),
	})

	return nil
}

func (svc *ingestImpl) SetStatusPending(ctx context.Context, ID int) error {
	query := `UPDATE sip SET status = ?, WHERE id = ?`
	args := []interface{}{
		enums.SIPStatusPending,
		ID,
	}

	if ID < 0 {
		return fmt.Errorf("%w: ID", ErrInvalid)
	}
	id := uint(ID) // #nosec G115 -- range validated.

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	event.PublishEvent(ctx, svc.evsvc, &goaingest.SIPStatusUpdatedEvent{
		ID:     id,
		Status: enums.SIPStatusPending.String(),
	})

	return nil
}

func (svc *ingestImpl) SetLocationID(ctx context.Context, ID int, locationID uuid.UUID) error {
	if ID < 0 {
		return fmt.Errorf("%w: ID", ErrInvalid)
	}
	id := uint(ID) // #nosec G115 -- range validated.

	query := `UPDATE sip SET location_id = ? WHERE id = ?`
	args := []interface{}{
		locationID,
		ID,
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	event.PublishEvent(ctx, svc.evsvc, &goaingest.SIPLocationUpdatedEvent{
		ID:         id,
		LocationID: locationID,
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

func (svc *ingestImpl) read(ctx context.Context, ID uint) (*datatypes.SIP, error) {
	query := "SELECT id, name, workflow_id, run_id, aip_id, location_id, status, CONVERT_TZ(created_at, @@session.time_zone, '+00:00') AS created_at, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at FROM sip WHERE id = ?"
	args := []interface{}{ID}
	c := datatypes.SIP{}

	if err := svc.db.GetContext(ctx, &c, query, args...); err != nil {
		return nil, err
	}

	return &c, nil
}
