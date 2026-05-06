package childwf_test

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/pkg/childwf"
)

func TestPostStorageResultHelpers(t *testing.T) {
	t.Parallel()

	var (
		started   = time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC)
		completed = time.Date(2026, 5, 7, 9, 1, 0, 0, time.UTC)
	)

	res := childwf.PostStorageResult{}
	validationTask := res.NewTask(started, "Validate AIP")
	res.ValidationError(completed, validationTask, "AIP metadata rejected")
	assert.Equal(t, res.Outcome, childwf.OutcomeContentError)
	assert.DeepEqual(t, res.Tasks, []*childwf.Task{
		{
			Name:        "Validate AIP",
			Message:     "Content error: AIP metadata rejected",
			Outcome:     childwf.TaskOutcomeValidationFailure,
			StartedAt:   started,
			CompletedAt: completed,
		},
	})

	res = childwf.PostStorageResult{}
	systemTask := res.NewTask(started, "Notify storage")
	res.SystemError(completed, systemTask, "storage service unavailable")
	assert.Equal(t, res.Outcome, childwf.OutcomeSystemError)
	assert.DeepEqual(t, res.Tasks, []*childwf.Task{
		{
			Name:        "Notify storage",
			Message:     "System error: storage service unavailable",
			Outcome:     childwf.TaskOutcomeSystemFailure,
			StartedAt:   started,
			CompletedAt: completed,
		},
	})
}
