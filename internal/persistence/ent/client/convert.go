package client

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
	if sip.Edges.Uploader != nil {
		s.Uploader = convertUser(sip.Edges.Uploader)
	}
	if sip.Edges.Batch != nil {
		s.Batch = convertBatch(sip.Edges.Batch)
	}

	return &s
}

// convertTask converts an entgo `db.Task` representation
// to a `datatypes.Task` representation.
func convertTask(task *db.Task) *datatypes.Task {
	var started sql.NullTime
	if !task.StartedAt.IsZero() {
		started = sql.NullTime{Time: task.StartedAt.UTC(), Valid: true}
	}

	var completed sql.NullTime
	if !task.CompletedAt.IsZero() {
		completed = sql.NullTime{Time: task.CompletedAt.UTC(), Valid: true}
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

// convertBatch converts an entgo `db.Batch` representation to a
// `datatypes.Batch` representation.
func convertBatch(batch *db.Batch) *datatypes.Batch {
	// Convert required fields.
	b := datatypes.Batch{
		ID:          batch.ID,
		UUID:        batch.UUID,
		Identifier:  batch.Identifier,
		SIPSCount:   batch.SipsCount,
		Status:      batch.Status,
		CreatedAt:   batch.CreatedAt,
		StartedAt:   batch.StartedAt,
		CompletedAt: batch.CompletedAt,
	}

	// Convert optional fields.
	if batch.Edges.Uploader != nil {
		b.Uploader = convertUser(batch.Edges.Uploader)
	}

	return &b
}

func convertWorkflow(dbw *db.Workflow) *datatypes.Workflow {
	// ent stores status as int8 constrained by schema; direct cast is safe.
	w := &datatypes.Workflow{
		ID:         dbw.ID,
		UUID:       dbw.UUID,
		TemporalID: dbw.TemporalID,
		Type:       dbw.Type,
		Status:     enums.WorkflowStatus(uint(dbw.Status)), // #nosec G115 -- constrained value.
	}

	if !dbw.StartedAt.IsZero() {
		w.StartedAt = sql.NullTime{Time: dbw.StartedAt.UTC(), Valid: true}
	}
	if !dbw.CompletedAt.IsZero() {
		w.CompletedAt = sql.NullTime{Time: dbw.CompletedAt.UTC(), Valid: true}
	}
	if dbw.Edges.Sip != nil {
		w.SIPUUID = dbw.Edges.Sip.UUID
	}

	// Only populate Tasks if they were loaded, preserving nil vs empty semantics.
	if dbw.Edges.Tasks != nil {
		w.Tasks = make([]*datatypes.Task, 0, len(dbw.Edges.Tasks))
		for _, dbt := range dbw.Edges.Tasks {
			t := convertTask(dbt)
			if t.WorkflowUUID == uuid.Nil {
				t.WorkflowUUID = dbw.UUID
			}
			w.Tasks = append(w.Tasks, t)
		}
	}

	return w
}
