package ingest_test

import (
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/ingest"
)

func TestHandleAuditEvent(t *testing.T) {
	t.Parallel()

	sipID := uuid.MustParse("6592dbe2-a2db-4cfb-a73b-0e620da1053f")
	userID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	type test struct {
		name  string
		event *goaingest.IngestEvent
		want  *auditlog.Event
	}
	for _, tc := range []test{
		{
			name: "Handle a SIP created event",
			event: &goaingest.IngestEvent{
				IngestValue: &goaingest.SIPCreatedEvent{
					UUID: sipID,
					Item: &goaingest.SIP{UUID: sipID, UploaderUUID: &userID},
				},
			},
			want: &auditlog.Event{
				Level:    slog.LevelInfo,
				Msg:      "SIP created",
				Type:     "sip.created",
				ObjectID: "6592dbe2-a2db-4cfb-a73b-0e620da1053f",
				UserID:   userID.String(),
			},
		},
		{
			name: "Handle a SIP created event with no uploader",
			event: &goaingest.IngestEvent{
				IngestValue: &goaingest.SIPCreatedEvent{
					UUID: sipID,
					Item: &goaingest.SIP{UUID: sipID},
				},
			},
			want: &auditlog.Event{
				Level:    slog.LevelInfo,
				Msg:      "SIP created",
				Type:     "sip.created",
				ObjectID: "6592dbe2-a2db-4cfb-a73b-0e620da1053f",
				UserID:   "",
			},
		},
		{
			name: "Ignore unsupported event types",
			event: &goaingest.IngestEvent{
				IngestValue: &goaingest.IngestPingEvent{Message: ref.New("ping")},
			},
			want: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ingest.HandleAuditEvent(tc.event)
			assert.DeepEqual(t, got, tc.want)
		})
	}
}
