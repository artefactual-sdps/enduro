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

func (svc *packageImpl) CreatePreservationTask(ctx context.Context, pt *datatypes.PreservationTask) error {
	err := svc.perSvc.CreatePreservationTask(ctx, pt)
	if err != nil {
		return fmt.Errorf("preservation task: create: %v", err)
	}

	ev := &goapackage.PreservationTaskCreatedEvent{
		ID:   pt.ID,
		Item: preservationTaskToGoa(pt),
	}
	event.PublishEvent(ctx, svc.evsvc, ev)

	return nil
}

func (svc *packageImpl) CompletePreservationTask(
	ctx context.Context,
	ID uint,
	status enums.PreservationTaskStatus,
	completedAt time.Time,
	note *string,
) error {
	var query string
	args := []interface{}{}

	if note != nil {
		query = `UPDATE preservation_task SET note = ?, status = ?, completed_at = ? WHERE id = ?`
		args = append(args, note, status, completedAt, ID)
	} else {
		query = `UPDATE preservation_task SET status = ?, completed_at = ? WHERE id = ?`
		args = append(args, status, completedAt, ID)
	}

	_, err := svc.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error updating preservation task: %w", err)
	}

	if item, err := svc.readPreservationTask(ctx, ID); err == nil {
		ev := &goapackage.PreservationTaskUpdatedEvent{ID: ID, Item: item}
		event.PublishEvent(ctx, svc.evsvc, ev)
	}

	return nil
}

func (svc *packageImpl) readPreservationTask(
	ctx context.Context,
	ID uint,
) (*goapackage.EnduroPackagePreservationTask, error) {
	query := `
		SELECT
			preservation_task.id,
			preservation_task.task_id,
			preservation_task.name,
			preservation_task.status,
			CONVERT_TZ(preservation_task.started_at, @@session.time_zone, '+00:00') AS started_at,
			CONVERT_TZ(preservation_task.completed_at, @@session.time_zone, '+00:00') AS completed_at,
			preservation_task.note,
			preservation_task.preservation_action_id
		FROM preservation_task
		LEFT JOIN preservation_action ON (preservation_task.preservation_action_id = preservation_action.id)
		WHERE preservation_task.id = ?
	`
	args := []interface{}{ID}
	dbItem := datatypes.PreservationTask{}

	if err := svc.db.GetContext(ctx, &dbItem, query, args...); err != nil {
		return nil, err
	}

	item := goapackage.EnduroPackagePreservationTask{
		ID:                   dbItem.ID,
		TaskID:               dbItem.TaskID,
		Name:                 dbItem.Name,
		Status:               dbItem.Status.String(),
		StartedAt:            ref.DerefZero(db.FormatOptionalTime(dbItem.StartedAt)),
		CompletedAt:          db.FormatOptionalTime(dbItem.CompletedAt),
		Note:                 ref.New(dbItem.Note),
		PreservationActionID: ref.New(dbItem.PreservationActionID),
	}

	return &item, nil
}
