package childwf

import (
	"strings"
	"time"
)

type result struct {
	outcome *Outcome
	tasks   *[]*Task
}

func (r result) newTask(t time.Time, name string) *Task {
	task := NewTask(t, name)
	*r.tasks = append(*r.tasks, task)

	return task
}

func (r result) validationError(t time.Time, task *Task, msg ...string) {
	*r.outcome = OutcomeContentError
	task.Complete(
		t,
		TaskOutcomeValidationFailure,
		"Content error: %s",
		strings.Join(msg, "\n\n"),
	)
}

func (r result) systemError(t time.Time, task *Task, msg ...string) {
	*r.outcome = OutcomeSystemError
	task.Complete(
		t,
		TaskOutcomeSystemFailure,
		"System error: %s",
		strings.Join(msg, "\n\n"),
	)
}
