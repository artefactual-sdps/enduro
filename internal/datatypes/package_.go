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

// Package represents a package in the package table.
type Package struct {
	ID         int                 `db:"id"`
	Name       string              `db:"name"`
	WorkflowID string              `db:"workflow_id"`
	RunID      string              `db:"run_id"`
	AIPID      uuid.NullUUID       `db:"aip_id"`      // Nullable.
	LocationID uuid.NullUUID       `db:"location_id"` // Nullable.
	Status     enums.PackageStatus `db:"status"`

	// It defaults to CURRENT_TIMESTAMP(6) so populated as soon as possible.
	CreatedAt time.Time `db:"created_at"`

	// Nullable, populated as soon as processing starts.
	StartedAt sql.NullTime `db:"started_at"`

	// Nullable, populated as soon as ingest completes.
	CompletedAt sql.NullTime `db:"completed_at"`
}

// Goa returns the API representation of the package.
func (p *Package) Goa() *goapackage.EnduroStoredPackage {
	if p == nil {
		return nil
	}

	var id uint
	if p.ID > 0 {
		id = uint(p.ID) // #nosec G115 -- range validated.
	}

	col := goapackage.EnduroStoredPackage{
		ID:          id,
		Name:        db.FormatOptionalString(p.Name),
		WorkflowID:  db.FormatOptionalString(p.WorkflowID),
		RunID:       db.FormatOptionalString(p.RunID),
		Status:      p.Status.String(),
		CreatedAt:   db.FormatTime(p.CreatedAt),
		StartedAt:   db.FormatOptionalTime(p.StartedAt),
		CompletedAt: db.FormatOptionalTime(p.CompletedAt),
	}
	if p.AIPID.Valid {
		col.AipID = ref.New(p.AIPID.UUID.String())
	}
	if p.LocationID.Valid {
		col.LocationID = &p.LocationID.UUID
	}

	return &col
}
