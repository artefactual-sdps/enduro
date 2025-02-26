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

	// PreservationActionID is the primary key of the parent PreservationAction.
	PreservationActionID int

	// Tasks is a list of preprocessing tasks to save as PreservationTasks.
	Tasks []preprocessing.Task
}

type SavePreprocessingTasksActivityResult struct {
	// Count is the number of saved PreservationTasks.
	Count int
}

func SavePreprocessingTasksActivity(
	ctx context.Context,
	params SavePreprocessingTasksActivityParams,
) (*SavePreprocessingTasksActivityResult, error) {
	var res SavePreprocessingTasksActivityResult
	for _, t := range params.Tasks {
		pt := preprocTaskToPresTask(t)
		pt.PreservationActionID = params.PreservationActionID

		u, err := uuid.NewRandomFromReader(params.RNG)
		if err != nil {
			return &res, fmt.Errorf("SavePreprocessingTasksActivity: generate UUID: %v", err)
		}
		pt.TaskID = u.String()

		if err := params.Ingestsvc.CreatePreservationTask(ctx, &pt); err != nil {
			return &res, fmt.Errorf("SavePreprocessingTasksActivity: %v", err)
		}
		res.Count++
	}

	return &res, nil
}

func preprocTaskToPresTask(t preprocessing.Task) datatypes.PreservationTask {
	taskOutcomeToPresTaskStatus := map[enums.PreprocessingTaskOutcome]enums.PreservationTaskStatus{
		enums.PreprocessingTaskOutcomeUnspecified:       enums.PreservationTaskStatusUnspecified,
		enums.PreprocessingTaskOutcomeSuccess:           enums.PreservationTaskStatusDone,
		enums.PreprocessingTaskOutcomeSystemFailure:     enums.PreservationTaskStatusError,
		enums.PreprocessingTaskOutcomeValidationFailure: enums.PreservationTaskStatusError,
	}

	status, found := taskOutcomeToPresTaskStatus[t.Outcome]
	if !found {
		status = enums.PreservationTaskStatusUnspecified
	}

	return datatypes.PreservationTask{
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
