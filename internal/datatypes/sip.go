package datatypes

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"

	goapackage "github.com/artefactual-sdps/enduro/internal/api/gen/package_"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/enums"
)

// SIP represents a SIP in the sip table.
type SIP struct {
	ID         int             `db:"id"`
	Name       string          `db:"name"`
	WorkflowID string          `db:"workflow_id"`
	RunID      string          `db:"run_id"`
	AIPID      uuid.NullUUID   `db:"aip_id"`      // Nullable.
	LocationID uuid.NullUUID   `db:"location_id"` // Nullable.
	Status     enums.SIPStatus `db:"status"`

	// It defaults to CURRENT_TIMESTAMP(6) so populated as soon as possible.
	CreatedAt time.Time `db:"created_at"`

	// Nullable, populated as soon as processing starts.
	StartedAt sql.NullTime `db:"started_at"`

	// Nullable, populated as soon as ingest completes.
	CompletedAt sql.NullTime `db:"completed_at"`
}

// Goa returns the API representation of the SIP.
func (s *SIP) Goa() *goapackage.EnduroStoredPackage {
	if s == nil {
		return nil
	}

	var id uint
	if s.ID > 0 {
		id = uint(s.ID) // #nosec G115 -- range validated.
	}

	col := goapackage.EnduroStoredPackage{
		ID:          id,
		Name:        db.FormatOptionalString(s.Name),
		WorkflowID:  db.FormatOptionalString(s.WorkflowID),
		RunID:       db.FormatOptionalString(s.RunID),
		Status:      s.Status.String(),
		CreatedAt:   db.FormatTime(s.CreatedAt),
		StartedAt:   db.FormatOptionalTime(s.StartedAt),
		CompletedAt: db.FormatOptionalTime(s.CompletedAt),
	}
	if s.AIPID.Valid {
		col.AipID = ref.New(s.AIPID.UUID.String())
	}
	if s.LocationID.Valid {
		col.LocationID = &s.LocationID.UUID
	}

	return &col
}
