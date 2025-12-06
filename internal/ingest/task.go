package ingest

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

func (svc *ingestImpl) CreateTask(ctx context.Context, task *datatypes.Task) error {
	err := svc.perSvc.CreateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("task: create: %v", err)
	}

	PublishEvent(ctx, svc.evsvc, &goaingest.SIPTaskCreatedEvent{
		UUID: task.UUID,
		Item: taskToGoa(task),
	})

	return nil
}

func (svc *ingestImpl) CreateTasks(ctx context.Context, tasks []*datatypes.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	err := svc.perSvc.CreateTasks(ctx, tasks)
	if err != nil {
		return fmt.Errorf("tasks: create: %v", err)
	}

	for _, task := range tasks {
		PublishEvent(ctx, svc.evsvc, &goaingest.SIPTaskCreatedEvent{
			UUID: task.UUID,
			Item: taskToGoa(task),
		})
	}

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

	PublishEvent(ctx, svc.evsvc, &goaingest.SIPTaskUpdatedEvent{
		UUID: task.UUID,
		Item: taskToGoa(task),
	})

	return nil
}
