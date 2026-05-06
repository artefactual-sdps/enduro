package childwf_test

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/pkg/childwf"
)

func TestTask(t *testing.T) {
	t.Parallel()

	var (
		started   = time.Date(2024, 6, 6, 14, 48, 12, 0, time.UTC)
		completed = time.Date(2024, 6, 6, 14, 48, 13, 0, time.UTC)
	)

	t.Run("Task succeeds", func(t *testing.T) {
		t.Parallel()

		task := childwf.NewTask(started, "test task")
		task.Succeed(
			completed,
			"completed at %s",
			completed.Format(time.RFC3339),
		)
		assert.DeepEqual(t, task, &childwf.Task{
			Name:        "test task",
			Message:     "completed at 2024-06-06T14:48:13Z",
			Outcome:     childwf.TaskOutcomeSuccess,
			StartedAt:   started,
			CompletedAt: completed,
		})
		assert.Equal(t, task.IsSuccess(), true)
	})

	t.Run("Task outcome is validation failure", func(t *testing.T) {
		t.Parallel()

		task := childwf.NewTask(started, "test task")
		task.Complete(
			completed,
			childwf.TaskOutcomeValidationFailure,
			"Content error: metadata validation has failed",
		)
		assert.DeepEqual(t, task, &childwf.Task{
			Name:        "test task",
			Message:     "Content error: metadata validation has failed",
			Outcome:     childwf.TaskOutcomeValidationFailure,
			StartedAt:   started,
			CompletedAt: completed,
		})
		assert.Equal(t, task.IsSuccess(), false)
	})
}
