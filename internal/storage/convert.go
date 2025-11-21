package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
	"github.com/artefactual-sdps/enduro/internal/auditlog"
	"github.com/artefactual-sdps/enduro/internal/storage/enums"
	"github.com/artefactual-sdps/enduro/internal/storage/persistence"
	"github.com/artefactual-sdps/enduro/internal/storage/types"
)

// workflowToGoa converts a storage Workflow to a Goa AIPWorkflow.
func (svc *serviceImpl) workflowToGoa(w *types.Workflow) *goastorage.AIPWorkflow {
	var startedAt, completedAt *string

	if !w.StartedAt.IsZero() {
		startedAt = ref.New(w.StartedAt.Format(time.RFC3339))
	}

	if !w.CompletedAt.IsZero() {
		completedAt = ref.New(w.CompletedAt.Format(time.RFC3339))
	}

	// Tasks are loaded separately when needed.
	return &goastorage.AIPWorkflow{
		UUID:        w.UUID,
		TemporalID:  w.TemporalID,
		Type:        w.Type.String(),
		Status:      w.Status.String(),
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		AipUUID:     w.AIPUUID,
	}
}

// taskToGoa converts a storage Task to a Goa AIPTask.
func (svc *serviceImpl) taskToGoa(t *types.Task) *goastorage.AIPTask {
	var startedAt, completedAt, note *string

	if !t.StartedAt.IsZero() {
		startedAt = ref.New(t.StartedAt.Format(time.RFC3339))
	}

	if !t.CompletedAt.IsZero() {
		completedAt = ref.New(t.CompletedAt.Format(time.RFC3339))
	}

	if t.Note != "" {
		note = ref.New(t.Note)
	}

	return &goastorage.AIPTask{
		UUID:         t.UUID,
		Name:         t.Name,
		Status:       t.Status.String(),
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
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

func deletionRequestFilterFromPayload(
	payload *goastorage.ListDeletionRequestsPayload,
) (*persistence.DeletionRequestFilter, error) {
	if payload == nil {
		return nil, nil
	}

	aipUUID, err := uuid.Parse(payload.UUID)
	if err != nil {
		return nil, goastorage.MakeNotValid(errors.New("invalid UUID"))
	}

	f := &persistence.DeletionRequestFilter{AIPUUID: &aipUUID}

	if payload.Status != nil {
		status, err := enums.ParseDeletionRequestStatus(*payload.Status)
		if err != nil {
			return nil, goastorage.MakeNotValid(errors.New("invalid status"))
		}
		f.Status = &status
	}

	return f, nil
}

func deletionRequestAsGoa(dr *types.DeletionRequest) *goastorage.DeletionRequest {
	g := &goastorage.DeletionRequest{
		UUID:        dr.UUID,
		AipUUID:     dr.AIPUUID,
		Requester:   dr.Requester,
		RequestedAt: dr.RequestedAt.Format(time.RFC3339),
		Reason:      dr.Reason,
		Status:      dr.Status.String(),
	}

	if dr.Reviewer != "" {
		g.Reviewer = &dr.Reviewer
	}
	if !dr.ReviewedAt.IsZero() {
		g.ReviewedAt = ref.New(dr.ReviewedAt.Format(time.RFC3339))
	}
	if dr.ReportKey != "" {
		g.ReportKey = &dr.ReportKey
	}

	return g
}
