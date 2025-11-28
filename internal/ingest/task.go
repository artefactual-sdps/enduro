package ingest

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"time"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
)

func (svc *ingestImpl) CreateTask(ctx context.Context, task *datatypes.Task) error {
	tasks := []*datatypes.Task{task}
	return svc.CreateTasks(ctx, persistence.TaskSequence(slices.Values(tasks)))
}

func (svc *ingestImpl) CreateTasks(
	ctx context.Context,
	seq persistence.TaskSequence,
) error {
	// Tee the sequence to capture task pointers while forwarding them to the
	// persistence layer. This allows us to access the tasks (with their
	// database-generated IDs) after persistence completes, so we can publish
	// events without consuming the sequence twice or changing the persistence
	// API.
	var tasks []*datatypes.Task
	tee := persistence.TaskSequence(func(yield func(*datatypes.Task) bool) {
		seq(func(task *datatypes.Task) bool {
			tasks = append(tasks, task)
			return yield(task)
		})
	})

	if err := svc.perSvc.CreateTasks(ctx, tee); err != nil {
		return fmt.Errorf("task: create: %v", err)
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
