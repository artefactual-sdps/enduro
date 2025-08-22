package auditlog_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/event"
)

func testLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Set time to a deterministic value for testing.
			if a.Key == slog.TimeKey {
				a.Value = slog.TimeValue(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
				return a
			}
			return a
		},
	}))
}

func TestAuditLog(t *testing.T) {
	t.Parallel()

	type test struct {
		name    string
		evsvc   event.Service[*auditlog.Event]
		handler auditlog.EventHandler[*auditlog.Event]
		want    string
		wantErr string
	}
	for _, tc := range []test{
		{
			name:  "Logs an event",
			evsvc: event.NewServiceInMem[*auditlog.Event](),
			handler: func(e *auditlog.Event) *auditlog.Event {
				return e
			},
			want: `{"time":"2025-01-01T00:00:00Z","level":"INFO","msg":"SIP created","type":"sip.created","objectID":"sip-123","userID":"user-456"}
`,
		},
		{
			name:  "Returns error when subscription fails",
			evsvc: event.NewServiceNop[*auditlog.Event](),
			handler: func(e *auditlog.Event) *auditlog.Event {
				return e
			},
			wantErr: "audit log: subscribe: Subscribe not supported by nop service",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Use a pipe as a thread-safe read/write buffer.
			r, w := io.Pipe()
			al := auditlog.New(testLogger(w), tc.handler)

			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()

			// Start the audit log listener.
			err := al.Listen(ctx, tc.evsvc)
			if tc.wantErr != "" {
				assert.Error(t, err, tc.wantErr)
				return
			}
			defer al.Close()

			tc.evsvc.PublishEvent(ctx, &auditlog.Event{
				Level:    0,
				Msg:      "SIP created",
				Type:     "sip.created",
				ObjectID: "sip-123",
				UserID:   "user-456",
			})

			// Wait 10ms for the event to publish, then close the writer end of
			// the pipe to send an EOF to the reader end.
			time.AfterFunc(10*time.Millisecond, func() {
				w.Close()
			})

			b, err := io.ReadAll(r)
			assert.NilError(t, err)
			assert.Equal(t, string(b), tc.want)
		})
	}
}
