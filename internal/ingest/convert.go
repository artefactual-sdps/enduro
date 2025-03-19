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

func sipTogoaingestCreatedEvent(s *datatypes.SIP) *goaingest.SIPCreatedEvent {
	var id uint
	if s.ID > 0 {
		id = uint(s.ID) // #nosec G115 -- range validated.
	}

	return &goaingest.SIPCreatedEvent{
		ID:   id,
		Item: s.Goa(),
	}
}

// workflowToGoa returns the API representation of a workflow.
func workflowToGoa(w *datatypes.Workflow) *goaingest.SIPWorkflow {
	var startedAt string
	if w.StartedAt.Valid {
		startedAt = w.StartedAt.Time.Format(time.RFC3339)
	}

	var id uint
	if w.ID > 0 {
		id = uint(w.ID) // #nosec G115 -- range validated.
	}

	var sipID uint
	if w.SIPID > 0 {
		sipID = uint(w.SIPID) // #nosec G115 -- range validated.
	}

	return &goaingest.SIPWorkflow{
		ID:          id,
		TemporalID:  w.TemporalID,
		Type:        w.Type.String(),
		Status:      w.Status.String(),
		StartedAt:   startedAt,
		CompletedAt: db.FormatOptionalTime(w.CompletedAt),
		SipID:       ref.New(sipID),
	}
}

// taskToGoa returns the API representation of a task.
func taskToGoa(task *datatypes.Task) *goaingest.SIPTask {
	var id uint
	if task.ID > 0 {
		id = uint(task.ID) // #nosec G115 -- range validated.
	}

	var wID uint
	if task.WorkflowID > 0 {
		wID = uint(task.WorkflowID) // #nosec G115 -- range validated.
	}

	return &goaingest.SIPTask{
		ID:     id,
		TaskID: task.TaskID,
		Name:   task.Name,
		Status: task.Status.String(),

		// TODO: Make Goa StartedAt a pointer to a string to avoid having to
		// convert a null time to an empty (zero value) string.
		StartedAt: ref.DerefZero(db.FormatOptionalTime(task.CompletedAt)),

		CompletedAt: db.FormatOptionalTime(task.CompletedAt),
		Note:        &task.Note,
		WorkflowID:  ref.New(wID),
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

	pf := persistence.SIPFilter{
		AIPID:     aipID,
		Name:      payload.Name,
		Status:    status,
		CreatedAt: createdAt,
		Sort:      entfilter.NewSort().AddCol("id", true),
		Page: persistence.Page{
			Limit:  ref.DerefZero(payload.Limit),
			Offset: ref.DerefZero(payload.Offset),
		},
	}

	return &pf, nil
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
