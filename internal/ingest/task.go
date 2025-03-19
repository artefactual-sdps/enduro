package ingest

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func (svc *ingestImpl) CreateTask(ctx context.Context, task *datatypes.Task) error {
	err := svc.perSvc.CreateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("task: create: %v", err)
	}

	event.PublishEvent(ctx, svc.evsvc, &goaingest.SIPTaskCreatedEvent{
		ID:   uint(task.ID), // #nosec G115 -- constrained value.
		Item: taskToGoa(task),
	})

	return nil
}

func (svc *ingestImpl) CompleteTask(
	ctx context.Context,
	id int,
	status enums.TaskStatus,
	completedAt time.Time,
	note *string,
) error {
	if id < 0 {
		return fmt.Errorf("%w: ID", ErrInvalid)
	}

	task, err := svc.perSvc.UpdateTask(
		ctx,
		id,
		func(task *datatypes.Task) (*datatypes.Task, error) {
			task.Status = status
			task.CompletedAt = sql.NullTime{
				Time:  completedAt,
				Valid: true,
			}
			if note != nil {
				task.Note = *note
			}

			return task, nil
		},
	)
	if err != nil {
		return fmt.Errorf("error updating task: %v", err)
	}

	event.PublishEvent(ctx, svc.evsvc, &goaingest.SIPTaskUpdatedEvent{
		ID:   uint(id), // #nosec G115 -- range validated.
		Item: taskToGoa(task),
	})

	return nil
}
