package storage_test

import (
	"testing"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/storage"
)

func TestHandleAuditEvent(t *testing.T) {
	t.Parallel()

	reqID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	type test struct {
		name string
		ev   *goastorage.StorageEvent
		want *auditlog.Event
	}
	for _, tc := range []test{
		{
			name: "Handles a deletion request created event",
			ev: &goastorage.StorageEvent{
				StorageValue: &goastorage.AIPDeletionRequestCreatedEvent{
					UUID: reqID,
					Item: &goastorage.AIPDeletionRequest{
						UUID:      reqID,
						Requester: "user123",
					},
				},
			},
			want: &auditlog.Event{
				Msg:      "AIP deletion request created",
				Type:     "AIP.deletion.request",
				ObjectID: reqID.String(),
				User:     "user123",
			},
		},
		{
			name: "Ignores unsupported event types",
			ev: &goastorage.StorageEvent{
				StorageValue: &goastorage.StoragePingEvent{
					Message: ref.New("ping"),
				},
			},
			want: nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := storage.HandleAuditEvent(tc.ev)
			assert.DeepEqual(t, tc.want, got)
		})
	}
}
