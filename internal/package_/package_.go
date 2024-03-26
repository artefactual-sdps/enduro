package package_

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	temporalsdk_client "go.temporal.io/sdk/client"

	"github.com/artefactual-sdps/enduro/internal/api/auth"
	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

type Service interface {
	// Goa returns an implementation of the goapackage Service.
	Goa() goapackage.Service
	Create(context.Context, *datatypes.Package) error
	UpdateWorkflowStatus(ctx context.Context, ID uint, name, workflowID, runID, aipID string, status enums.PackageStatus, storedAt time.Time) error
	SetStatus(ctx context.Context, ID uint, status enums.PackageStatus) error
	SetStatusInProgress(ctx context.Context, ID uint, startedAt time.Time) error
	SetStatusPending(ctx context.Context, ID uint) error
	SetLocationID(ctx context.Context, ID uint, locationID uuid.UUID) error
	CreatePreservationAction(ctx context.Context, pa *PreservationAction) error
	SetPreservationActionStatus(ctx context.Context, ID uint, status PreservationActionStatus) error
	CompletePreservationAction(ctx context.Context, ID uint, status PreservationActionStatus, completedAt time.Time) error
	CreatePreservationTask(ctx context.Context, pt *PreservationTask) error
	CompletePreservationTask(ctx context.Context, ID uint, status PreservationTaskStatus, completedAt time.Time, note *string) error
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
	}
}

func (svc *packageImpl) Goa() goapackage.Service {
	return &goaWrapper{
		packageImpl: svc,
	}
}

// Create persists pkg to the data store then updates it from the data store,
// adding generated data (e.g. ID, CreatedAt).
func (svc *packageImpl) Create(ctx context.Context, pkg *datatypes.Package) error {
	err := svc.perSvc.CreatePackage(ctx, pkg)
	if err != nil {
		return fmt.Errorf("package: create: %v", err)
	}

	event.PublishEvent(
		ctx,
		svc.evsvc,
		&goapackage.PackageCreatedEvent{ID: uint(pkg.ID), Item: pkg.Goa()},
	)

	return nil
}

func (svc *packageImpl) UpdateWorkflowStatus(ctx context.Context, ID uint, name, workflowID, runID, aipID string, status enums.PackageStatus, storedAt time.Time) error {
	// Ensure that storedAt is reset during retries.
	completedAt := &storedAt
	if status == enums.PackageStatusInProgress {
		completedAt = nil
	}
	if completedAt != nil && completedAt.IsZero() {
		completedAt = nil
	}

	query := `UPDATE package SET name = ?, workflow_id = ?, run_id = ?, aip_id = ?, status = ?, completed_at = ? WHERE id = ?`
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

	if pkg, err := svc.Goa().Show(ctx, &goapackage.ShowPayload{ID: ID}); err == nil {
		ev := &goapackage.PackageUpdatedEvent{ID: uint(ID), Item: pkg}
		event.PublishEvent(ctx, svc.evsvc, ev)
	}

	return nil
}

func (svc *packageImpl) SetStatus(ctx context.Context, ID uint, status enums.PackageStatus) error {
	query := `UPDATE package SET status = ? WHERE id = ?`
	args := []interface{}{
		status,
		ID,
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	ev := &goapackage.PackageStatusUpdatedEvent{ID: uint(ID), Status: status.String()}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *packageImpl) SetStatusInProgress(ctx context.Context, ID uint, startedAt time.Time) error {
	var query string
	args := []interface{}{enums.PackageStatusInProgress}

	if !startedAt.IsZero() {
		query = `UPDATE package SET status = ?, started_at = ? WHERE id = ?`
		args = append(args, startedAt, ID)
	} else {
		query = `UPDATE package SET status = ? WHERE id = ?`
		args = append(args, ID)
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	ev := &goapackage.PackageStatusUpdatedEvent{ID: uint(ID), Status: enums.PackageStatusInProgress.String()}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *packageImpl) SetStatusPending(ctx context.Context, ID uint) error {
	query := `UPDATE package SET status = ?, WHERE id = ?`
	args := []interface{}{
		enums.PackageStatusPending,
		ID,
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	ev := &goapackage.PackageStatusUpdatedEvent{ID: uint(ID), Status: enums.PackageStatusPending.String()}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *packageImpl) SetLocationID(ctx context.Context, ID uint, locationID uuid.UUID) error {
	query := `UPDATE package SET location_id = ? WHERE id = ?`
	args := []interface{}{
		locationID,
		ID,
	}

	if err := svc.updateRow(ctx, query, args); err != nil {
		return err
	}

	ev := &goapackage.PackageLocationUpdatedEvent{ID: uint(ID), LocationID: locationID}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *packageImpl) updateRow(ctx context.Context, query string, args []interface{}) error {
	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating package: %v", err)
	}

	return nil
}

func (svc *packageImpl) read(ctx context.Context, ID uint) (*datatypes.Package, error) {
	query := "SELECT id, name, workflow_id, run_id, aip_id, location_id, status, CONVERT_TZ(created_at, @@session.time_zone, '+00:00') AS created_at, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at FROM package WHERE id = ?"
	args := []interface{}{ID}
	c := datatypes.Package{}

	if err := svc.db.GetContext(ctx, &c, query, args...); err != nil {
		return nil, err
	}

	return &c, nil
}
