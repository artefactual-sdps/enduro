package storage

import (
	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

// workflowToGoa converts a storage Workflow to a Goa AIPWorkflow.
func (svc *serviceImpl) workflowToGoa(w *types.Workflow) *goastorage.AIPWorkflow {
	// Tasks are loaded separately when needed.
	return &goastorage.AIPWorkflow{
		UUID:        w.UUID,
		TemporalID:  w.TemporalID,
		Type:        w.Type.String(),
		Status:      w.Status.String(),
		StartedAt:   db.FormatOptionalZeroTime(w.StartedAt),
		CompletedAt: db.FormatOptionalZeroTime(w.CompletedAt),
		AipUUID:     w.AIPUUID,
	}
}

// taskToGoa converts a storage Task to a Goa AIPTask.
func (svc *serviceImpl) taskToGoa(t *types.Task) *goastorage.AIPTask {
	var note *string

	if t.Note != "" {
		note = new(t.Note)
	}

	return &goastorage.AIPTask{
		UUID:         t.UUID,
		Name:         t.Name,
		Status:       t.Status.String(),
		StartedAt:    db.FormatOptionalZeroTime(t.StartedAt),
		CompletedAt:  db.FormatOptionalZeroTime(t.CompletedAt),
		Note:         note,
		WorkflowUUID: t.WorkflowUUID,
	}
}

func deletionRequestAuditEvent(dr *types.DeletionRequest) *auditlog.Event {
	ev := auditlog.Event{
		Level:      auditlog.LevelInfo,
		Type:       "AIP.deletion.request",
		ResourceID: dr.AIPUUID.String(),
	}

	switch dr.Status {
	case enums.DeletionRequestStatusPending:
		ev.Msg = "AIP deletion requested"
		ev.User = dr.Requester
	case enums.DeletionRequestStatusCanceled:
		ev.Msg = "AIP deletion request canceled"
		ev.User = dr.Requester
	case enums.DeletionRequestStatusApproved:
		ev.Msg = "AIP deletion request approved"
		ev.User = dr.Reviewer
	case enums.DeletionRequestStatusRejected:
		ev.Msg = "AIP deletion request rejected"
		ev.User = dr.Reviewer
	default:
		return nil
	}

	return &ev
}
