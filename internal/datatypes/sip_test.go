package datatypes

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gotest.tools/v3/assert"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

func TestSIPGoa(t *testing.T) {
	t.Parallel()

	sipUUID := uuid.New()
	aipUUID := uuid.New()
	uploaderUUID := uuid.New()
	batchUUID := uuid.New()
	createdAt := time.Date(2025, 6, 23, 14, 57, 12, 0, time.UTC)
	startedAt := time.Date(2025, 6, 23, 15, 0, 0, 0, time.UTC)
	completedAt := time.Date(2025, 6, 23, 15, 5, 0, 0, time.UTC)

	type test struct {
		name string
		sip  *SIP
		want *goaingest.SIP
	}

	for _, tt := range []test{
		{
			name: "Converts nil SIP to nil Goa SIP",
		},
		{
			name: "Converts SIP with zero optional times",
			sip: &SIP{
				UUID:      sipUUID,
				Name:      "transfer-1",
				Status:    enums.SIPStatusQueued,
				CreatedAt: createdAt,
			},
			want: &goaingest.SIP{
				UUID:      sipUUID,
				Name:      new("transfer-1"),
				Status:    enums.SIPStatusQueued.String(),
				CreatedAt: "2025-06-23T14:57:12Z",
			},
		},
		{
			name: "Converts SIP with all optional fields",
			sip: &SIP{
				UUID:        sipUUID,
				Name:        "transfer-1",
				AIPID:       uuid.NullUUID{Valid: true, UUID: aipUUID},
				Status:      enums.SIPStatusIngested,
				CreatedAt:   createdAt,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				FailedAs:    enums.SIPFailedAsPIP,
				FailedKey:   "failed/pip.7z",
				Uploader: &User{
					UUID:  uploaderUUID,
					Email: "nobody@example.com",
					Name:  "Test User",
				},
				Batch: &Batch{
					UUID:       batchUUID,
					Identifier: "batch-1",
					Status:     enums.BatchStatusPending,
				},
			},
			want: &goaingest.SIP{
				UUID:            sipUUID,
				Name:            new("transfer-1"),
				Status:          enums.SIPStatusIngested.String(),
				AipUUID:         new(aipUUID.String()),
				CreatedAt:       "2025-06-23T14:57:12Z",
				StartedAt:       new("2025-06-23T15:00:00Z"),
				CompletedAt:     new("2025-06-23T15:05:00Z"),
				FailedAs:        new(enums.SIPFailedAsPIP.String()),
				FailedKey:       new("failed/pip.7z"),
				UploaderUUID:    new(uploaderUUID),
				UploaderEmail:   new("nobody@example.com"),
				UploaderName:    new("Test User"),
				BatchUUID:       new(batchUUID),
				BatchIdentifier: new("batch-1"),
				BatchStatus:     new(enums.BatchStatusPending.String()),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.sip.Goa()
			assert.DeepEqual(t, got, tt.want)
		})
	}
}
