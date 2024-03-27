package enums

import (
	"encoding/json"
	"strings"
)

type PreservationActionType uint

const (
	PreservationActionTypeUnspecified PreservationActionType = iota
	PreservationActionTypeCreateAIP
	PreservationActionTypeCreateAndReviewAIP
	PreservationActionTypeMovePackage
)

func NewPreservationActionType(status string) PreservationActionType {
	var s PreservationActionType

	switch strings.ToLower(status) {
	case "create-aip":
		s = PreservationActionTypeCreateAIP
	case "create-and-review-aip":
		s = PreservationActionTypeCreateAndReviewAIP
	case "move-package":
		s = PreservationActionTypeMovePackage
	default:
		s = PreservationActionTypeUnspecified
	}

	return s
}

func (p PreservationActionType) String() string {
	switch p {
	case PreservationActionTypeCreateAIP:
		return "create-aip"
	case PreservationActionTypeCreateAndReviewAIP:
		return "create-and-review-aip"
	case PreservationActionTypeMovePackage:
		return "move-package"
	}

	return "unspecified"
}

func (p PreservationActionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PreservationActionType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*p = NewPreservationActionType(s)

	return nil
}

type PreservationActionStatus uint

const (
	PreservationActionStatusUnspecified PreservationActionStatus = iota
	PreservationActionStatusInProgress
	PreservationActionStatusDone
	PreservationActionStatusError
	PreservationActionStatusQueued
	PreservationActionStatusPending
)

func NewPreservationActionStatus(status string) PreservationActionStatus {
	var s PreservationActionStatus

	switch strings.ToLower(status) {
	case "in progress":
		s = PreservationActionStatusInProgress
	case "done":
		s = PreservationActionStatusDone
	case "error":
		s = PreservationActionStatusError
	case "queued":
		s = PreservationActionStatusQueued
	case "pending":
		s = PreservationActionStatusPending
	default:
		s = PreservationActionStatusUnspecified
	}

	return s
}

func (p PreservationActionStatus) String() string {
	switch p {
	case PreservationActionStatusInProgress:
		return "in progress"
	case PreservationActionStatusDone:
		return "done"
	case PreservationActionStatusError:
		return "error"
	case PreservationActionStatusQueued:
		return "queued"
	case PreservationActionStatusPending:
		return "pending"
	}

	return "unspecified"
}

func (p PreservationActionStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PreservationActionStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*p = NewPreservationActionStatus(s)

	return nil
}
