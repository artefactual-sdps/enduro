package entclient

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

// convertPkgToPackage converts an entgo `db.Pkg` package representation to a
// `datatypes.Package` representation.
func convertPkgToPackage(pkg *db.Pkg) *datatypes.Package {
	var started, completed sql.NullTime
	if !pkg.StartedAt.IsZero() {
		started = sql.NullTime{Time: pkg.StartedAt, Valid: true}
	}
	if !pkg.CompletedAt.IsZero() {
		completed = sql.NullTime{Time: pkg.CompletedAt, Valid: true}
	}

	var aipID uuid.NullUUID
	if pkg.AipID != uuid.Nil {
		aipID = uuid.NullUUID{UUID: pkg.AipID, Valid: true}
	}

	var locID uuid.NullUUID
	if pkg.LocationID != uuid.Nil {
		locID = uuid.NullUUID{UUID: pkg.LocationID, Valid: true}
	}

	return &datatypes.Package{
		ID:          uint(pkg.ID),
		Name:        pkg.Name,
		WorkflowID:  pkg.WorkflowID,
		RunID:       pkg.RunID.String(),
		AIPID:       aipID,
		LocationID:  locID,
		Status:      enums.PackageStatus(pkg.Status),
		CreatedAt:   pkg.CreatedAt,
		StartedAt:   started,
		CompletedAt: completed,
	}
}

// convertPreservationAction converts an entgo `db.PreservationAction`
// representation to a `datatypes.PreservationAction` representation.
func convertPreservationAction(pa *db.PreservationAction) *datatypes.PreservationAction {
	var started sql.NullTime
	if !pa.StartedAt.IsZero() {
		started = sql.NullTime{Time: pa.StartedAt, Valid: true}
	}

	var completed sql.NullTime
	if !pa.CompletedAt.IsZero() {
		completed = sql.NullTime{Time: pa.CompletedAt, Valid: true}
	}

	return &datatypes.PreservationAction{
		ID:          uint(pa.ID),
		WorkflowID:  pa.WorkflowID,
		Type:        enums.PreservationActionType(pa.Type),
		Status:      enums.PreservationActionStatus(pa.Status),
		StartedAt:   started,
		CompletedAt: completed,
		PackageID:   uint(pa.PackageID),
	}
}

// convertPreservationTask converts an entgo `db.PreservationTask` representation
// to a `datatypes.PreservationTask` representation.
func convertPreservationTask(pt *db.PreservationTask) *datatypes.PreservationTask {
	var started sql.NullTime
	if !pt.StartedAt.IsZero() {
		started = sql.NullTime{Time: pt.StartedAt, Valid: true}
	}

	var completed sql.NullTime
	if !pt.CompletedAt.IsZero() {
		completed = sql.NullTime{Time: pt.CompletedAt, Valid: true}
	}

	return &datatypes.PreservationTask{
		ID:                   uint(pt.ID),
		TaskID:               pt.TaskID.String(),
		Name:                 pt.Name,
		Status:               enums.PreservationTaskStatus(pt.Status),
		StartedAt:            started,
		CompletedAt:          completed,
		Note:                 pt.Note,
		PreservationActionID: uint(pt.PreservationActionID),
	}
}
