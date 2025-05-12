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

	return &datatypes.SIP{
		ID:          sip.ID,
		UUID:        sip.UUID,
		Name:        sip.Name,
		AIPID:       aipID,
		Status:      sip.Status,
		CreatedAt:   sip.CreatedAt,
		StartedAt:   started,
		CompletedAt: completed,
	}
}

// convertTask converts an entgo `db.Task` representation
// to a `datatypes.Task` representation.
func convertTask(task *db.Task) *datatypes.Task {
	var started sql.NullTime
	if !task.StartedAt.IsZero() {
		started = sql.NullTime{Time: task.StartedAt, Valid: true}
	}

	var completed sql.NullTime
	if !task.CompletedAt.IsZero() {
		completed = sql.NullTime{Time: task.CompletedAt, Valid: true}
	}

	var status uint
	if task.Status > 0 {
		status = uint(task.Status) // #nosec G115 -- range validated.
	}

	return &datatypes.Task{
		ID:          task.ID,
		TaskID:      task.TaskID.String(),
		Name:        task.Name,
		Status:      enums.TaskStatus(status),
		StartedAt:   started,
		CompletedAt: completed,
		Note:        task.Note,
		WorkflowID:  task.WorkflowID,
	}
}
