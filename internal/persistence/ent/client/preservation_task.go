package entclient

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

func (c *client) CreatePreservationTask(ctx context.Context, pt *datatypes.PreservationTask) error {
	q, err := createPreservationTaskBuilder(c.ent.PreservationTask, pt)
	if err != nil {
		return err
	}

	r, err := q.Save(ctx)
	if err != nil {
		return newDBErrorWithDetails(err, "create preservation task")
	}

	// Update value of pt with data from DB (e.g. ID).
	*pt = *convertPreservationTask(r)

	return nil
}

func (c *client) CreatePreservationTasks(
	ctx context.Context,
	seq func(yield func(*datatypes.PreservationTask) bool),
) ([]*datatypes.PreservationTask, error) {
	ret := make([]*datatypes.PreservationTask, 0, defaultBatchSize)

	for pts := range batch(seq) {
		builders := make([]*db.PreservationTaskCreate, len(pts))
		for i, pt := range pts {
			op, err := createPreservationTaskBuilder(c.ent.PreservationTask, pt)
			if err != nil {
				return nil, err
			}
			builders[i] = op
		}
		if pts, err := c.ent.PreservationTask.CreateBulk(builders...).Save(ctx); err != nil {
			return nil, err
		} else {
			for _, pt := range pts {
				ret = append(ret, convertPreservationTask(pt))
			}
		}
	}

	return ret, nil
}

func (c *client) UpdatePreservationTask(
	ctx context.Context,
	id int,
	updater persistence.PresTaskUpdater,
) (*datatypes.PreservationTask, error) {
	tx, err := c.ent.BeginTx(ctx, nil)
	if err != nil {
		return nil, newDBErrorWithDetails(err, "update preservation task")
	}

	pt, err := tx.PreservationTask.Get(ctx, id)
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

	q := tx.PreservationTask.UpdateOneID(id).
		SetTaskID(taskID).
		SetName(up.Name).
		SetStatus(int8(up.Status)). // #nosec G115 -- constrained value.
		SetNote(up.Note).
		SetPreservationActionID(int(up.PreservationActionID))

	// Set nullable column values.
	if up.StartedAt.Valid {
		q.SetStartedAt(up.StartedAt.Time)
	}
	if up.CompletedAt.Valid {
		q.SetCompletedAt(up.CompletedAt.Time)
	}

	pt, err = q.Save(ctx)
	if err != nil {
		return nil, rollback(tx, newDBError(err))
	}
	if err = tx.Commit(); err != nil {
		return nil, rollback(tx, newDBError(err))
	}

	return convertPreservationTask(pt), nil
}

func createPreservationTaskBuilder(
	c *db.PreservationTaskClient,
	pt *datatypes.PreservationTask,
) (*db.PreservationTaskCreate, error) {
	// Validate required fields.
	taskID, err := uuid.Parse(pt.TaskID)
	if err != nil {
		return nil, newParseError(err, "TaskID")
	}
	if pt.Name == "" {
		return nil, newRequiredFieldError("Name")
	}
	if pt.PreservationActionID == 0 {
		return nil, newRequiredFieldError("PreservationActionID")
	}

	status := enums.PreservationTaskStatus(uint(pt.Status))
	if !status.IsValid() {
		return nil, newParseError(errors.New("invalid"), "Status")
	}

	// Handle nullable fields.
	var startedAt *time.Time
	if pt.StartedAt.Valid {
		startedAt = &pt.StartedAt.Time
	}
	var completedAt *time.Time
	if pt.CompletedAt.Valid {
		completedAt = &pt.CompletedAt.Time
	}

	op := c.Create().
		SetTaskID(taskID).
		SetName(pt.Name).
		SetStatus(int8(status)). // #nosec G115 -- constrained value.
		SetNillableStartedAt(startedAt).
		SetNillableCompletedAt(completedAt).
		SetNote(pt.Note).
		SetPreservationActionID(int(pt.PreservationActionID))

	return op, nil
}
