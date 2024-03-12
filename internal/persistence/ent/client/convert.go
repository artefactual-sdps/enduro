package entclient

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

// convertPkgToPackage converts an ent `db.Pkg` package representation to a
// `datatypes.Package` representation.
func convertPkgToPackage(pkg *db.Pkg) *datatypes.Package {
	var started, completed sql.NullTime
	if !pkg.StartedAt.IsZero() {
		started = sql.NullTime{Time: pkg.StartedAt, Valid: true}
	}
	if !pkg.CompletedAt.IsZero() {
		completed = sql.NullTime{Time: pkg.CompletedAt, Valid: true}
	}

	var locID uuid.NullUUID
	if pkg.LocationID != uuid.Nil {
		locID = uuid.NullUUID{UUID: pkg.LocationID, Valid: true}
	}

	return &datatypes.Package{
		ID:          uint(pkg.ID),
		Name:        pkg.Name,
		LocationID:  locID,
		Status:      enums.PackageStatus(pkg.Status),
		WorkflowID:  pkg.WorkflowID,
		RunID:       pkg.RunID.String(),
		AIPID:       pkg.AipID.String(),
		CreatedAt:   pkg.CreatedAt,
		StartedAt:   started,
		CompletedAt: completed,
	}
}
