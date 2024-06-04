package package_

import (
	"context"
	"fmt"
	"time"

	"go.artefactual.dev/tools/ref"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func (svc *packageImpl) CreatePreservationAction(
	ctx context.Context,
	pa *datatypes.PreservationAction,
) error {
	err := svc.perSvc.CreatePreservationAction(ctx, pa)
	if err != nil {
		return fmt.Errorf("preservation action: create: %v", err)
	}

	ev := &goapackage.PreservationActionCreatedEvent{
		ID:   pa.ID,
		Item: preservationActionToGoa(pa),
	}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *packageImpl) SetPreservationActionStatus(
	ctx context.Context,
	ID uint,
	status enums.PreservationActionStatus,
) error {
	query := `UPDATE preservation_action SET status = ? WHERE id = ?`
	args := []interface{}{
		status,
		ID,
	}

	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating preservation action: %w", err)
	}

	if item, err := svc.readPreservationAction(ctx, ID); err == nil {
		ev := &goapackage.PreservationActionUpdatedEvent{ID: ID, Item: item}
		event.PublishEvent(ctx, svc.evsvc, ev)
	}

	return nil
}

func (svc *packageImpl) CompletePreservationAction(
	ctx context.Context,
	ID uint,
	status enums.PreservationActionStatus,
	completedAt time.Time,
) error {
	query := `UPDATE preservation_action SET status = ?, completed_at = ? WHERE id = ?`
	args := []interface{}{
		status,
		completedAt,
		ID,
	}

	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating preservation action: %w", err)
	}

	if item, err := svc.readPreservationAction(ctx, ID); err == nil {
		ev := &goapackage.PreservationActionUpdatedEvent{ID: ID, Item: item}
		event.PublishEvent(ctx, svc.evsvc, ev)
	}

	return nil
}

func (svc *packageImpl) readPreservationAction(
	ctx context.Context,
	ID uint,
) (*goapackage.EnduroPackagePreservationAction, error) {
	query := `
		SELECT
			preservation_action.id,
			preservation_action.workflow_id,
			preservation_action.type,
			preservation_action.status,
			CONVERT_TZ(preservation_action.started_at, @@session.time_zone, '+00:00') AS started_at,
			CONVERT_TZ(preservation_action.completed_at, @@session.time_zone, '+00:00') AS completed_at,
			preservation_action.package_id
		FROM preservation_action
		LEFT JOIN package ON (preservation_action.package_id = package.id)
		WHERE preservation_action.id = ?
	`
	args := []interface{}{ID}
	dbItem := datatypes.PreservationAction{}

	if err := svc.db.GetContext(ctx, &dbItem, query, args...); err != nil {
		return nil, err
	}

	item := goapackage.EnduroPackagePreservationAction{
		ID:          dbItem.ID,
		WorkflowID:  dbItem.WorkflowID,
		Type:        dbItem.Type.String(),
		Status:      dbItem.Status.String(),
		StartedAt:   ref.DerefZero(db.FormatOptionalTime(dbItem.StartedAt)),
		CompletedAt: db.FormatOptionalTime(dbItem.CompletedAt),
		PackageID:   ref.New(dbItem.PackageID),
	}

	return &item, nil
}
