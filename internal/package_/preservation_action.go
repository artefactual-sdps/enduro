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

func (w *goaWrapper) PreservationActions(
	ctx context.Context,
	payload *goapackage.PreservationActionsPayload,
) (*goapackage.EnduroPackagePreservationActions, error) {
	goapkg, err := w.Show(ctx, &goapackage.ShowPayload{ID: payload.ID})
	if err != nil {
		return nil, err
	}

	query := "SELECT id, workflow_id, type, status, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at FROM preservation_action WHERE package_id = ? ORDER BY started_at DESC"
	args := []interface{}{goapkg.ID}

	rows, err := w.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying the database: %w", err)
	}
	defer rows.Close()

	preservation_actions := []*goapackage.EnduroPackagePreservationAction{}
	for rows.Next() {
		pa := datatypes.PreservationAction{}
		if err := rows.StructScan(&pa); err != nil {
			return nil, fmt.Errorf("error scanning database result: %w", err)
		}
		goapa := &goapackage.EnduroPackagePreservationAction{
			ID:          pa.ID,
			WorkflowID:  pa.WorkflowID,
			Type:        pa.Type.String(),
			Status:      pa.Status.String(),
			StartedAt:   db.FormatTime(pa.StartedAt.Time),
			CompletedAt: db.FormatOptionalTime(pa.CompletedAt),
		}

		ptQuery := "SELECT id, task_id, name, status, CONVERT_TZ(started_at, @@session.time_zone, '+00:00') AS started_at, CONVERT_TZ(completed_at, @@session.time_zone, '+00:00') AS completed_at, note FROM preservation_task WHERE preservation_action_id = ?"
		ptQueryArgs := []interface{}{pa.ID}

		ptRows, err := w.db.QueryxContext(ctx, ptQuery, ptQueryArgs...)
		if err != nil {
			return nil, fmt.Errorf("error querying the database: %w", err)
		}
		defer ptRows.Close()

		preservation_tasks := []*goapackage.EnduroPackagePreservationTask{}
		for ptRows.Next() {
			pt := datatypes.PreservationTask{}
			if err := ptRows.StructScan(&pt); err != nil {
				return nil, fmt.Errorf("error scanning database result: %w", err)
			}
			goapt := &goapackage.EnduroPackagePreservationTask{
				ID:          pt.ID,
				TaskID:      pt.TaskID,
				Name:        pt.Name,
				Status:      pt.Status.String(),
				StartedAt:   db.FormatTime(pt.StartedAt.Time),
				CompletedAt: db.FormatOptionalTime(pt.CompletedAt),
				Note:        &pt.Note,
			}
			preservation_tasks = append(preservation_tasks, goapt)
		}

		goapa.Tasks = preservation_tasks
		preservation_actions = append(preservation_actions, goapa)
	}

	result := &goapackage.EnduroPackagePreservationActions{
		Actions: preservation_actions,
	}

	return result, nil
}

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
