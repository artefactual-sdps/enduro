package entclient

import (
	"database/sql"

	"github.com/google/uuid"

	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence/ent/db"
)

// convertSIP converts an entgo `db.SIP` representation to a
// `datatypes.SIP` representation.
func convertSIP(sip *db.SIP) *datatypes.SIP {
	var started, completed sql.NullTime
	if !sip.StartedAt.IsZero() {
		started = sql.NullTime{Time: sip.StartedAt, Valid: true}
	}
	if !sip.CompletedAt.IsZero() {
		completed = sql.NullTime{Time: sip.CompletedAt, Valid: true}
	}

	var aipID uuid.NullUUID
	if sip.AipID != uuid.Nil {
		aipID = uuid.NullUUID{UUID: sip.AipID, Valid: true}
	}

	var locID uuid.NullUUID
	if sip.LocationID != uuid.Nil {
		locID = uuid.NullUUID{UUID: sip.LocationID, Valid: true}
	}

	var status uint
	if sip.Status > 0 {
		status = uint(sip.Status) // #nosec G115 -- range validated.
	}

	return &datatypes.SIP{
		ID:          sip.ID,
		Name:        sip.Name,
		WorkflowID:  sip.WorkflowID,
		RunID:       sip.RunID.String(),
		AIPID:       aipID,
		LocationID:  locID,
		Status:      enums.SIPStatus(status),
		CreatedAt:   sip.CreatedAt,
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
		ID:          pa.ID,
		WorkflowID:  pa.WorkflowID,
		Type:        enums.PreservationActionType(pa.Type),     // #nosec G115 -- constrained value.
		Status:      enums.PreservationActionStatus(pa.Status), // #nosec G115 -- constrained value.
		StartedAt:   started,
		CompletedAt: completed,
		SIPID:       pa.SipID,
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

	var status uint
	if pt.Status > 0 {
		status = uint(pt.Status) // #nosec G115 -- range validated.
	}

	return &datatypes.PreservationTask{
		ID:                   pt.ID,
		TaskID:               pt.TaskID.String(),
		Name:                 pt.Name,
		Status:               enums.PreservationTaskStatus(status),
		StartedAt:            started,
		CompletedAt:          completed,
		Note:                 pt.Note,
		PreservationActionID: pt.PreservationActionID,
	}
}
