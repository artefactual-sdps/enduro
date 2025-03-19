package entclient

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

func (c *client) CreateTask(ctx context.Context, task *datatypes.Task) error {
	// Validate required fields.
	taskID, err := uuid.Parse(task.TaskID)
	if err != nil {
		return newParseError(err, "TaskID")
	}
	if task.Name == "" {
		return newRequiredFieldError("Name")
	}
	if task.WorkflowID == 0 {
		return newRequiredFieldError("WorkflowID")
	}
	// TODO: Validate Status.

	// Handle nullable fields.
	var startedAt *time.Time
	if task.StartedAt.Valid {
		startedAt = &task.StartedAt.Time
	}

	var completedAt *time.Time
	if task.CompletedAt.Valid {
		completedAt = &task.CompletedAt.Time
	}

	q := c.ent.Task.Create().
		SetTaskID(taskID).
		SetName(task.Name).
		SetStatus(int8(task.Status)). // #nosec G115 -- constrained value.
		SetNillableStartedAt(startedAt).
		SetNillableCompletedAt(completedAt).
		SetNote(task.Note).
		SetWorkflowID(int(task.WorkflowID))

	r, err := q.Save(ctx)
	if err != nil {
		return newDBErrorWithDetails(err, "create task")
	}

	// Update value of task with data from DB (e.g. ID).
	*task = *convertTask(r)

	return nil
}

func (c *client) UpdateTask(
	ctx context.Context,
	id int,
	updater persistence.TaskUpdater,
) (*datatypes.Task, error) {
	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return nil, newDBErrorWithDetails(err, "update task")
	}

	task, err := tx.Task.Get(ctx, id)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	up, err := updater(convertTask(task))
	if err != nil {
		return nil, rollback(tx, newUpdaterError(err))
	}

	// Set required column values.
	taskID, err := uuid.Parse(up.TaskID)
	if err != nil {
		return nil, rollback(tx, newParseError(err, "TaskID"))
	}

	q := tx.Task.UpdateOneID(id).
		SetTaskID(taskID).
		SetName(up.Name).
		SetStatus(int8(up.Status)). // #nosec G115 -- constrained value.
		SetNote(up.Note).
		SetWorkflowID(int(up.WorkflowID))

	// Set nullable column values.
	if up.StartedAt.Valid {
		q.SetStartedAt(up.StartedAt.Time)
	}
	if up.CompletedAt.Valid {
		q.SetCompletedAt(up.CompletedAt.Time)
	}

	task, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	return convertTask(task), nil
}
