package ingest

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/datatypes"
	"github.com/artefactual-sdps/enduro/internal/db"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

// sipToCreatedEvent returns the API representation of a SIP created event.
func sipToCreatedEvent(s *datatypes.SIP) *goaingest.SIPCreatedEvent {
	return &goaingest.SIPCreatedEvent{
		UUID: s.UUID,
		Item: s.Goa(),
	}
}

// workflowToGoa returns the API representation of a workflow.
func workflowToGoa(w *datatypes.Workflow) *goaingest.SIPWorkflow {
	var startedAt string
	if w.StartedAt.Valid {
		startedAt = w.StartedAt.Time.Format(time.RFC3339)
	}

	return &goaingest.SIPWorkflow{
		UUID:        w.UUID,
		TemporalID:  w.TemporalID,
		Type:        w.Type.String(),
		Status:      w.Status.String(),
		StartedAt:   startedAt,
		CompletedAt: db.FormatOptionalTime(w.CompletedAt),
		SipUUID:     w.SIPUUID,
	}
}

// taskToGoa returns the API representation of a task.
func taskToGoa(task *datatypes.Task) *goaingest.SIPTask {
	return &goaingest.SIPTask{
		UUID:   task.UUID,
		Name:   task.Name,
		Status: task.Status.String(),

		// TODO: Make Goa StartedAt a pointer to a string to avoid having to
		// convert a null time to an empty (zero value) string.
		StartedAt: ref.DerefZero(db.FormatOptionalTime(task.CompletedAt)),

		CompletedAt:  db.FormatOptionalTime(task.CompletedAt),
		Note:         &task.Note,
		WorkflowUUID: task.WorkflowUUID,
	}
}

func listSipsPayloadToSIPFilter(payload *goaingest.ListSipsPayload) (*persistence.SIPFilter, error) {
	aipID, err := stringToUUIDPtr(payload.AipID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("aip_id: invalid UUID"))
	}

	var status *enums.SIPStatus
	if payload.Status != nil {
		s, err := enums.ParseSIPStatus(*payload.Status)
		if err != nil {
			return nil, goaingest.MakeNotValid(errors.New("status: invalid value"))
		}
		status = &s
	}

	createdAt, err := timerange.Parse(payload.EarliestCreatedTime, payload.LatestCreatedTime)
	if err != nil {
		return nil, goaingest.MakeNotValid(fmt.Errorf("created at: %v", err))
	}

	var uploaderID *uuid.UUID
	if payload.UploaderID != nil {
		id, err := uuid.Parse(*payload.UploaderID)
		if err != nil {
			return nil, goaingest.MakeNotValid(errors.New("uploader_id: invalid UUID"))
		}
		uploaderID = &id
	}

	pf := persistence.SIPFilter{
		AIPID:      aipID,
		Name:       payload.Name,
		Status:     status,
		CreatedAt:  createdAt,
		UploaderID: uploaderID,
		Sort:       entfilter.NewSort().AddCol("id", true),
		Page: persistence.Page{
			Limit:  ref.DerefZero(payload.Limit),
			Offset: ref.DerefZero(payload.Offset),
		},
	}

	return &pf, nil
}

func listUsersPayloadToUserFilter(payload *goaingest.ListUsersPayload) (*persistence.UserFilter, error) {
	if payload == nil {
		return nil, nil
	}

	email, err := validateStringPtr(payload.Email, 255)
	if err != nil {
		return nil, goaingest.MakeNotValid(fmt.Errorf("email: %w", err))
	}

	name, err := validateStringPtr(payload.Name, 255)
	if err != nil {
		return nil, goaingest.MakeNotValid(fmt.Errorf("name: %w", err))
	}

	f := persistence.UserFilter{
		Email: email,
		Name:  name,
		Page: persistence.Page{
			Limit:  ref.DerefZero(payload.Limit),
			Offset: ref.DerefZero(payload.Offset),
		},
	}

	return &f, nil
}

func stringToUUIDPtr(s *string) (*uuid.UUID, error) {
	if s == nil {
		return nil, nil
	}

	u, err := uuid.Parse(*s)
	if err != nil {
		return nil, errors.New("invalid UUID")
	}

	return &u, nil
}

func validateStringPtr(s *string, maxLength int) (*string, error) {
	if s == nil {
		return nil, nil
	}

	if len(*s) > maxLength {
		return nil, fmt.Errorf("exceeds maximum length of %d", maxLength)
	}

	return s, nil
}
