package entclient

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
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
