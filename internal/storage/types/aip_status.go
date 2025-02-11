package types

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

type AIPStatus uint

const (
	AIPStatusUnspecified AIPStatus = iota
	AIPStatusInReview
	AIPStatusRejected
	AIPStatusStored
	AIPStatusMoving
)

func NewAIPStatus(status string) AIPStatus {
	var s AIPStatus

	switch strings.ToLower(status) {
	case "stored":
		s = AIPStatusStored
	case "rejected":
		s = AIPStatusRejected
	case "in_review":
		s = AIPStatusInReview
	case "moving":
		s = AIPStatusMoving
	default:
		s = AIPStatusUnspecified
	}

	return s
}

func (a AIPStatus) String() string {
	switch a {
	case AIPStatusStored:
		return "stored"
	case AIPStatusRejected:
		return "rejected"
	case AIPStatusInReview:
		return "in_review"
	case AIPStatusMoving:
		return "moving"
	}
	return "unspecified"
}

func (a AIPStatus) Values() []string {
	return []string{
		AIPStatusUnspecified.String(),
		AIPStatusInReview.String(),
		AIPStatusRejected.String(),
		AIPStatusStored.String(),
		AIPStatusMoving.String(),
	}
}

// Value provides the DB a string from int.
func (a AIPStatus) Value() (driver.Value, error) {
	return a.String(), nil
}

// Scan tells our code how to read the enum into our type.
func (a *AIPStatus) Scan(val interface{}) error {
	var s string
	switch v := val.(type) {
	case nil:
		return nil
	case string:
		s = v
	case []uint8:
		s = string(v)
	}

	*a = NewAIPStatus(s)

	return nil
}

func (a AIPStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *AIPStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*a = NewAIPStatus(s)

	return nil
}
