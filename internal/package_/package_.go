package package_

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
	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

var ErrInvalid = errors.New("invalid")

type Service interface {
	// Goa returns an implementation of the goapackage Service.
	Goa() goapackage.Service
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

type packageImpl struct {
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

var _ Service = (*packageImpl)(nil)

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
) *packageImpl {
	return &packageImpl{
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

func (svc *packageImpl) Goa() goapackage.Service {
	return &goaWrapper{
		packageImpl: svc,
	}
}

// Create persists pkg to the data store then updates it from the data store,
// adding generated data (e.g. ID, CreatedAt).
func (svc *packageImpl) Create(ctx context.Context, pkg *datatypes.SIP) error {
	err := svc.perSvc.CreateSIP(ctx, pkg)
	if err != nil {
		return fmt.Errorf("package: create: %v", err)
	}

	event.PublishEvent(ctx, svc.evsvc, sipToGoaPackageCreatedEvent(pkg))

	return nil
}

func (svc *packageImpl) UpdateWorkflowStatus(
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

	if pkg, err := svc.Goa().Show(ctx, &goapackage.ShowPayload{ID: id}); err == nil {
		ev := &goapackage.PackageUpdatedEvent{ID: pkg.ID, Item: pkg}
		event.PublishEvent(ctx, svc.evsvc, ev)
	}

	return nil
}

func (svc *packageImpl) SetStatus(ctx context.Context, ID int, status enums.SIPStatus) error {
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

	ev := &goapackage.PackageStatusUpdatedEvent{ID: id, Status: status.String()}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *packageImpl) SetStatusInProgress(ctx context.Context, ID int, startedAt time.Time) error {
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

	event.PublishEvent(ctx, svc.evsvc, &goapackage.PackageStatusUpdatedEvent{
		ID:     id,
		Status: enums.SIPStatusInProgress.String(),
	})

	return nil
}

func (svc *packageImpl) SetStatusPending(ctx context.Context, ID int) error {
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

	event.PublishEvent(ctx, svc.evsvc, &goapackage.PackageStatusUpdatedEvent{
		ID:     id,
		Status: enums.SIPStatusPending.String(),
	})

	return nil
}

func (svc *packageImpl) SetLocationID(ctx context.Context, ID int, locationID uuid.UUID) error {
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

	event.PublishEvent(ctx, svc.evsvc, &goapackage.PackageLocationUpdatedEvent{
		ID:         id,
		LocationID: locationID,
	})

	return nil
}

func (svc *packageImpl) updateRow(ctx context.Context, query string, args []interface{}) error {
	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating package: %v", err)
	}

	return nil
}

func (svc *packageImpl) read(ctx context.Context, ID uint) (*datatypes.SIP, error) {
	query := "SELECT id, name, workflow_id, run_id, aip_id, location_id, status, CONVERT_TZ(created_at, @@session.time_zone, '+00:00') AS created_at, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at FROM sip WHERE id = ?"
	args := []interface{}{ID}
	c := datatypes.SIP{}

	if err := svc.db.GetContext(ctx, &c, query, args...); err != nil {
		return nil, err
	}

	return &c, nil
}
