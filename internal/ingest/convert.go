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

// preservationActionToGoa returns the API representation of a preservation task.
func preservationActionToGoa(pa *datatypes.PreservationAction) *goaingest.SIPPreservationAction {
	var startedAt string
	if pa.StartedAt.Valid {
		startedAt = pa.StartedAt.Time.Format(time.RFC3339)
	}

	var id uint
	if pa.ID > 0 {
		id = uint(pa.ID) // #nosec G115 -- range validated.
	}

	var sipID uint
	if pa.SIPID > 0 {
		sipID = uint(pa.SIPID) // #nosec G115 -- range validated.
	}

	return &goaingest.SIPPreservationAction{
		ID:          id,
		WorkflowID:  pa.WorkflowID,
		Type:        pa.Type.String(),
		Status:      pa.Status.String(),
		StartedAt:   startedAt,
		CompletedAt: db.FormatOptionalTime(pa.CompletedAt),
		SipID:       ref.New(sipID),
	}
}

// preservationTaskToGoa returns the API representation of a preservation task.
func preservationTaskToGoa(pt *datatypes.PreservationTask) *goaingest.SIPPreservationTask {
	var id uint
	if pt.ID > 0 {
		id = uint(pt.ID) // #nosec G115 -- range validated.
	}

	var paID uint
	if pt.PreservationActionID > 0 {
		paID = uint(pt.PreservationActionID) // #nosec G115 -- range validated.
	}

	return &goaingest.SIPPreservationTask{
		ID:     id,
		TaskID: pt.TaskID,
		Name:   pt.Name,
		Status: pt.Status.String(),

		// TODO: Make Goa StartedAt a pointer to a string to avoid having to
		// convert a null time to an empty (zero value) string.
		StartedAt: ref.DerefZero(db.FormatOptionalTime(pt.CompletedAt)),

		CompletedAt:          db.FormatOptionalTime(pt.CompletedAt),
		Note:                 &pt.Note,
		PreservationActionID: ref.New(paID),
	}
}

func listSipsPayloadToSIPFilter(payload *goaingest.ListSipsPayload) (*persistence.SIPFilter, error) {
	aipID, err := stringToUUIDPtr(payload.AipID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("aip_id: invalid UUID"))
	}

	locID, err := stringToUUIDPtr(payload.LocationID)
	if err != nil {
		return nil, goaingest.MakeNotValid(errors.New("location_id: invalid UUID"))
	}

	var status *enums.SIPStatus
	if payload.Status != nil {
		s, err := enums.ParseSIPStatus(*payload.Status)
		if err != nil {
			return nil, goaingest.MakeNotValid(errors.New("status: invalid value"))
		}
		status = &s
	}

	createdAt, err := parseCreatedAtRange(payload.EarliestCreatedTime, payload.LatestCreatedTime)
	if err != nil {
		return nil, goaingest.MakeNotValid(err)
	}

	pf := persistence.SIPFilter{
		AIPID:      aipID,
		Name:       payload.Name,
		LocationID: locID,
		Status:     status,
		CreatedAt:  createdAt,
		Sort:       persistence.NewSort().AddCol("id", true),
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

func parseCreatedAtRange(start, end *string) (*timerange.Range, error) {
	var s, e time.Time
	var err error

	if start == nil && end == nil {
		return nil, nil
	}

	if start == nil {
		// Make start date an arbitrary time far in the past.
		s = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		s, err = parseTime(start)
		if err != nil {
			return nil, fmt.Errorf("earliest_created_time: %v", err)
		}
	}

	if end == nil {
		e = time.Now()
	} else {
		e, err = parseTime(end)
		if err != nil {
			return nil, fmt.Errorf("latest_created_time: %v", err)
		}
	}

	r, err := timerange.New(s, e)
	if err != nil {
		return nil, fmt.Errorf("created at: %v", err)
	}

	return &r, nil
}

func parseTime(value *string) (time.Time, error) {
	if value == nil {
		return time.Time{}, nil
	}

	t, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return time.Time{}, errors.New("invalid time")
	}

	return t, nil
}
