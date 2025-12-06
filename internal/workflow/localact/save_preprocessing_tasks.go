package localact

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/ingest"
	"github.com/artefactual-sdps/enduro/internal/preprocessing"
)

type SavePreprocessingTasksActivityParams struct {
	// Ingestsvc is an ingest service instance.
	Ingestsvc ingest.Service

	// RNG is a random number generator source.
	RNG io.Reader

	// WorkflowUUID is the UUID of the parent Workflow.
	WorkflowUUID uuid.UUID

	// Tasks is a list of preprocessing tasks to save as Tasks.
	Tasks []preprocessing.Task
}

type SavePreprocessingTasksActivityResult struct {
	// Count is the number of saved Tasks.
	Count int
}

func SavePreprocessingTasksActivity(
	ctx context.Context,
	params SavePreprocessingTasksActivityParams,
) (*SavePreprocessingTasksActivityResult, error) {
	var res SavePreprocessingTasksActivityResult

	tasks := make([]*datatypes.Task, 0, len(params.Tasks))
	for _, t := range params.Tasks {
		task := preprocTaskToTask(t)
		task.WorkflowUUID = params.WorkflowUUID
		// TODO: Create deterministic UUIDs and make activities idempotent.
		task.UUID = uuid.Must(uuid.NewRandomFromReader(params.RNG))

		tasks = append(tasks, &task)
		res.Count++
	}

	if err := params.Ingestsvc.CreateTasks(ctx, tasks); err != nil {
		return &res, fmt.Errorf("SavePreprocessingTasksActivity: %v", err)
	}

	return &res, nil
}

func preprocTaskToTask(t preprocessing.Task) datatypes.Task {
	taskOutcomeToTaskStatus := map[enums.PreprocessingTaskOutcome]enums.TaskStatus{
		enums.PreprocessingTaskOutcomeUnspecified:       enums.TaskStatusUnspecified,
		enums.PreprocessingTaskOutcomeSuccess:           enums.TaskStatusDone,
		enums.PreprocessingTaskOutcomeSystemFailure:     enums.TaskStatusError,
		enums.PreprocessingTaskOutcomeValidationFailure: enums.TaskStatusFailed,
	}

	status, found := taskOutcomeToTaskStatus[t.Outcome]
	if !found {
		status = enums.TaskStatusUnspecified
	}

	return datatypes.Task{
		Name:        t.Name,
		Status:      status,
		StartedAt:   timeToNullTime(t.StartedAt),
		CompletedAt: timeToNullTime(t.CompletedAt),
		Note:        t.Message,
	}
}

func timeToNullTime(t time.Time) sql.NullTime {
	var r sql.NullTime
	if !t.IsZero() {
		r = sql.NullTime{Time: t, Valid: true}
	}
	return r
}
