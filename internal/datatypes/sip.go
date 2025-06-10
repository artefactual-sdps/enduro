package datatypes

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

// SIP represents a SIP in the sip table.
type SIP struct {
	ID     int             `db:"id"`
	UUID   uuid.UUID       `db:"uuid"`
	Name   string          `db:"name"`
	AIPID  uuid.NullUUID   `db:"aip_id"` // Nullable.
	Status enums.SIPStatus `db:"status"`

	// It defaults to CURRENT_TIMESTAMP(6) so populated as soon as possible.
	CreatedAt time.Time `db:"created_at"`

	// Nullable, populated as soon as processing starts.
	StartedAt sql.NullTime `db:"started_at"`

	// Nullable, populated as soon as ingest completes.
	CompletedAt sql.NullTime `db:"completed_at"`

	// Set if there is a failure in workflow, it can be empty.
	FailedAs enums.SIPFailedAs `db:"failed_as"`

	// Object key from the failed SIP/PIP in the internal bucket.
	FailedKey string `db:"failed_key"`
}

// Goa returns the API representation of the SIP.
func (s *SIP) Goa() *goaingest.SIP {
	if s == nil {
		return nil
	}

	col := goaingest.SIP{
		UUID:        s.UUID,
		Name:        db.FormatOptionalString(s.Name),
		Status:      s.Status.String(),
		CreatedAt:   db.FormatTime(s.CreatedAt),
		StartedAt:   db.FormatOptionalTime(s.StartedAt),
		CompletedAt: db.FormatOptionalTime(s.CompletedAt),
	}
	if s.AIPID.Valid {
		col.AipID = ref.New(s.AIPID.UUID.String())
	}
	if s.FailedAs != "" {
		col.FailedAs = ref.New(s.FailedAs.String())
	}
	if s.FailedKey != "" {
		col.FailedKey = ref.New(s.FailedKey)
	}

	return &col
}
