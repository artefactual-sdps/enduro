package childwf_test

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/pkg/childwf"
)

func TestPreprocessingResultHelpers(t *testing.T) {
	t.Parallel()

	var (
		started   = time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC)
		completed = time.Date(2026, 5, 7, 9, 1, 0, 0, time.UTC)
	)

	res := childwf.PreprocessingResult{}
	validationTask := res.NewTask(started, "Validate content")
	res.ValidationError(completed, validationTask, "missing manifest", "invalid checksum")
	assert.Equal(t, res.Outcome, childwf.OutcomeContentError)
	assert.DeepEqual(t, res.Tasks, []*childwf.Task{
		{
			Name:        "Validate content",
			Message:     "Content error: missing manifest\n\ninvalid checksum",
			Outcome:     childwf.TaskOutcomeValidationFailure,
			StartedAt:   started,
			CompletedAt: completed,
		},
	})

	res = childwf.PreprocessingResult{}
	systemTask := res.NewTask(started, "Read content")
	res.SystemError(completed, systemTask, "cannot read file")
	assert.Equal(t, res.Outcome, childwf.OutcomeSystemError)
	assert.DeepEqual(t, res.Tasks, []*childwf.Task{
		{
			Name:        "Read content",
			Message:     "System error: cannot read file",
			Outcome:     childwf.TaskOutcomeSystemFailure,
			StartedAt:   started,
			CompletedAt: completed,
		},
	})
}
