package entclient

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

func (c *client) CreatePreservationTask(ctx context.Context, pt *datatypes.PreservationTask) error {
	// Validate required fields.
	taskID, err := uuid.Parse(pt.TaskID)
	if err != nil {
		return newParseError(err, "TaskID")
	}
	if pt.Name == "" {
		return newRequiredFieldError("Name")
	}
	if pt.PreservationActionID == 0 {
		return newRequiredFieldError("PreservationActionID")
	}
	// TODO: Validate Status.

	// Handle nullable fields.
	var startedAt *time.Time
	if pt.StartedAt.Valid {
		startedAt = &pt.StartedAt.Time
	}

	var completedAt *time.Time
	if pt.CompletedAt.Valid {
		completedAt = &pt.CompletedAt.Time
	}

	q := c.ent.PreservationTask.Create().
		SetTaskID(taskID).
		SetName(pt.Name).
		SetStatus(int8(pt.Status)).
		SetNillableStartedAt(startedAt).
		SetNillableCompletedAt(completedAt).
		SetNote(pt.Note).
		SetPreservationActionID(int(pt.PreservationActionID))

	r, err := q.Save(ctx)
	if err != nil {
		return newDBErrorWithDetails(err, "create preservation task")
	}

	// Update value of pt with data from DB (e.g. ID).
	*pt = *convertPreservationTask(r)

	return nil
}

func (c *client) UpdatePreservationTask(
	ctx context.Context,
	id uint,
	updater persistence.PresTaskUpdater,
) (*datatypes.PreservationTask, error) {
	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return nil, newDBErrorWithDetails(err, "update preservation task")
	}

	pt, err := tx.PreservationTask.Get(ctx, int(id))
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	up, err := updater(convertPreservationTask(pt))
	if err != nil {
		return nil, rollback(tx, newUpdaterError(err))
	}

	// Set required column values.
	taskID, err := uuid.Parse(up.TaskID)
	if err != nil {
		return nil, rollback(tx, newParseError(err, "TaskID"))
	}

	q := tx.PreservationTask.UpdateOneID(int(id)).
		SetTaskID(taskID).
		SetName(up.Name).
		SetStatus(int8(up.Status)).
		SetNote(up.Note).
		SetPreservationActionID(int(up.PreservationActionID))

	// Set nullable column values.
	if up.StartedAt.Valid {
		q.SetStartedAt(up.StartedAt.Time)
	}
	if up.CompletedAt.Valid {
		q.SetCompletedAt(up.CompletedAt.Time)
	}

	// Save changes.
	pt, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	return convertPreservationTask(pt), nil
}
