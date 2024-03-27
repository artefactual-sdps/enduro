package enums

import (
	"encoding/json"
	"strings"
)

type PreservationTaskStatus uint

const (
	PreservationTaskStatusUnspecified PreservationTaskStatus = iota
	PreservationTaskStatusInProgress
	PreservationTaskStatusDone
	PreservationTaskStatusError
	PreservationTaskStatusQueued
	PreservationTaskStatusPending
)

func NewPreservationTaskStatus(status string) PreservationTaskStatus {
	var s PreservationTaskStatus

	switch strings.ToLower(status) {
	case "in progress":
		s = PreservationTaskStatusInProgress
	case "done":
		s = PreservationTaskStatusDone
	case "error":
		s = PreservationTaskStatusError
	case "queued":
		s = PreservationTaskStatusQueued
	case "pending":
		s = PreservationTaskStatusPending
	default:
		s = PreservationTaskStatusUnspecified
	}

	return s
}

func (p PreservationTaskStatus) String() string {
	switch p {
	case PreservationTaskStatusInProgress:
		return "in progress"
	case PreservationTaskStatusDone:
		return "done"
	case PreservationTaskStatusError:
		return "error"
	case PreservationTaskStatusQueued:
		return "queued"
	case PreservationTaskStatusPending:
		return "pending"
	}

	return "unspecified"
}

func (p PreservationTaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PreservationTaskStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*p = NewPreservationTaskStatus(s)

	return nil
}
