package localact

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/childwf"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

type SaveChildwfTasksActivityParams struct {
	// Ingestsvc is an ingest service instance.
	Ingestsvc ingest.Service

	// RNG is a random number generator source.
	RNG io.Reader

	// WorkflowUUID is the UUID of the parent Workflow.
	WorkflowUUID uuid.UUID

	// Tasks is a list of child workflow tasks to save as Tasks.
	Tasks []childwf.Task
}

type SaveChildwfTasksActivityResult struct {
	// Count is the number of saved Tasks.
	Count int
}

func SaveChildwfTasksActivity(
	ctx context.Context,
	params SaveChildwfTasksActivityParams,
) (*SaveChildwfTasksActivityResult, error) {
	var res SaveChildwfTasksActivityResult

	tasks := make([]*datatypes.Task, 0, len(params.Tasks))
	for _, t := range params.Tasks {
		task := childwfTaskToTask(t)
		task.WorkflowUUID = params.WorkflowUUID
		// TODO: Create deterministic UUIDs and make activities idempotent.
		task.UUID = uuid.Must(uuid.NewRandomFromReader(params.RNG))

		tasks = append(tasks, &task)
		res.Count++
	}

	if err := params.Ingestsvc.CreateTasks(ctx, tasks); err != nil {
		return &res, fmt.Errorf("SaveChildwfTasksActivity: %v", err)
	}

	return &res, nil
}

func childwfTaskToTask(t childwf.Task) datatypes.Task {
	taskOutcomeToTaskStatus := map[enums.ChildwfTaskOutcome]enums.TaskStatus{
		enums.ChildwfTaskOutcomeUnspecified:       enums.TaskStatusUnspecified,
		enums.ChildwfTaskOutcomeSuccess:           enums.TaskStatusDone,
		enums.ChildwfTaskOutcomeSystemFailure:     enums.TaskStatusError,
		enums.ChildwfTaskOutcomeValidationFailure: enums.TaskStatusFailed,
	}

	status, found := taskOutcomeToTaskStatus[t.Outcome]
	if !found {
		status = enums.TaskStatusUnspecified
	}

	return datatypes.Task{
		Name:        t.Name,
		Status:      status,
		StartedAt:   t.StartedAt,
		CompletedAt: t.CompletedAt,
		Note:        t.Message,
	}
}
