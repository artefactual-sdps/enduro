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
	// Convert required fields.
	s := datatypes.SIP{
		ID:        sip.ID,
		UUID:      sip.UUID,
		Name:      sip.Name,
		Status:    sip.Status,
		CreatedAt: sip.CreatedAt,
		FailedAs:  sip.FailedAs,
		FailedKey: sip.FailedKey,
	}

	// Convert optional fields.
	if !sip.StartedAt.IsZero() {
		s.StartedAt = sql.NullTime{Time: sip.StartedAt, Valid: true}
	}
	if !sip.CompletedAt.IsZero() {
		s.CompletedAt = sql.NullTime{Time: sip.CompletedAt, Valid: true}
	}
	if sip.AipID != uuid.Nil {
		s.AIPID = uuid.NullUUID{UUID: sip.AipID, Valid: true}
	}
	if sip.UploaderID != 0 && sip.Edges.User != nil {
		s.Uploader = &datatypes.Uploader{
			UUID:  sip.Edges.User.UUID,
			Email: sip.Edges.User.Email,
			Name:  sip.Edges.User.Name,
		}
	}

	return &s
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

	var wUUID uuid.UUID
	if task.Edges.Workflow != nil {
		wUUID = task.Edges.Workflow.UUID
	}

	return &datatypes.Task{
		ID:           task.ID,
		UUID:         task.UUID,
		Name:         task.Name,
		Status:       enums.TaskStatus(status),
		StartedAt:    started,
		CompletedAt:  completed,
		Note:         task.Note,
		WorkflowUUID: wUUID,
	}
}

// convertUser converts an entgo `db.User` representation to a
// `datatypes.User` representation.
func convertUser(dbu *db.User) *datatypes.User {
	return &datatypes.User{
		UUID:      dbu.UUID,
		CreatedAt: dbu.CreatedAt,
		Email:     dbu.Email,
		Name:      dbu.Name,
		OIDCIss:   dbu.OidcIss,
		OIDCSub:   dbu.OidcSub,
	}
}
