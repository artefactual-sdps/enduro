package auditlog_test

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"

	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func TestNewFromConfig(t *testing.T) {
	t.Parallel()

	td := fs.NewDir(t, "auditlog_test")
	al := auditlog.NewFromConfig(auditlog.Config{
		Filepath:  td.Join("test.log"),
		MaxSize:   5, // 5 MB
		Compress:  true,
		Verbosity: 0, // INFO
	}, func(e *auditlog.Event) *auditlog.Event {
		return e
	})

	evsvc := event.NewServiceInMem[*auditlog.Event]()
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	// Start the audit log listener.
	err := al.Listen(ctx, evsvc)
	assert.NilError(t, err)

	evsvc.PublishEvent(ctx, &auditlog.Event{
		Level:    0,
		Msg:      "SIP created",
		Type:     "sip.created",
		ObjectID: "sip-123",
		UserID:   "user-456",
	})

	// Wait 10ms for the event to publish, then check that the log file was
	// created. We can't check the contents of the log because the time field
	// is non-deterministic.
	time.AfterFunc(10*time.Millisecond, func() {
		assert.Assert(t, fs.Equal(
			td.Path(),
			fs.Expected(t, fs.WithMode(0o755),
				fs.WithFile("test.log", "", fs.MatchAnyFileContent, fs.WithMode(0o600)),
			),
		))
		al.Close()
	})
}
