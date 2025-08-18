package ingest_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/event"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

func setupIngestSvc(t *testing.T, evsvc event.Service[*goaingest.IngestEvent]) ingest.Service {
	t.Helper()

	ingestSvc := ingest.NewService(
		logr.Discard(), // logger
		nil,            // db
		nil,            // temporal client
		evsvc,          // event service
		nil,            // persistence service
		nil,            // token verifier
		nil,            // ticket provider
		"test",         // taskQueue
		nil,            // internal bucket
		1024,           // upload max size (1 KiB)
		nil,            // random number generator
		nil,            // sipSource
	)

	return ingestSvc
}

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

func TestSIPCreatedLogger(t *testing.T) {
	t.Parallel()

	var (
		sipID  = uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		userID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	)

	type test struct {
		name   string
		evsvc  event.Service[*goaingest.IngestEvent]
		event  *goaingest.SIPCreatedEvent
		logger func(w io.Writer) *slog.Logger
		want   string
	}
	for _, tt := range []test{
		{
			name:  "Log SIP created event",
			evsvc: event.NewServiceInMem[*goaingest.IngestEvent](),
			event: &goaingest.SIPCreatedEvent{
				UUID: sipID,
				Item: &goaingest.SIP{
					UUID:         sipID,
					UploaderUUID: &userID,
				},
			},
			logger: func(w io.Writer) *slog.Logger { return testLogger(w) },
			want: `{"time":"2025-01-01T00:00:00Z","level":"INFO","msg":"SIP deposited","type":"SIP.deposit","objectID":"` + sipID.String() + `","userID":"` + userID.String() + `"}
`,
		},
		{
			name:  "Log SIP created event with no uploader ID",
			evsvc: event.NewServiceInMem[*goaingest.IngestEvent](),
			event: &goaingest.SIPCreatedEvent{
				UUID: sipID,
				Item: &goaingest.SIP{UUID: sipID},
			},
			logger: func(w io.Writer) *slog.Logger { return testLogger(w) },
			want: `{"time":"2025-01-01T00:00:00Z","level":"INFO","msg":"SIP deposited","type":"SIP.deposit","objectID":"` + sipID.String() + `","userID":""}
`,
		},
		{
			name:  "Logger disabled",
			evsvc: event.NewServiceInMem[*goaingest.IngestEvent](),
			event: &goaingest.SIPCreatedEvent{
				UUID: sipID,
				Item: &goaingest.SIP{UUID: sipID},
			},
			logger: func(w io.Writer) *slog.Logger { return nil },
			want:   "",
		},
		{
			name:  "Can't subscribe to event service",
			evsvc: event.NewServiceNop[*goaingest.IngestEvent](),
			event: &goaingest.SIPCreatedEvent{
				UUID: sipID,
				Item: &goaingest.SIP{UUID: sipID},
			},
			logger: func(w io.Writer) *slog.Logger { return testLogger(w) },
			want:   "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Use a pipe as a thread-safe read/write buffer.
			r, w := io.Pipe()

			ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
			defer cancel()

			ingestSvc := setupIngestSvc(t, tt.evsvc)
			ingestSvc.WithAuditLogger(ctx, tt.logger(w))
			ingest.PublishEvent(ctx, tt.evsvc, tt.event)

			// Wait 10ms for the event to publish, then close the writer end of
			// the pipe to send an EOF to the reader end.
			time.AfterFunc(10*time.Millisecond, func() {
				w.Close()
			})

			b, err := io.ReadAll(r)
			assert.NilError(t, err)
			assert.Equal(t, string(b), tt.want)
		})
	}
}
