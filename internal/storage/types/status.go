package types

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

type PackageStatus uint

const (
	StatusUnspecified PackageStatus = iota
	StatusInReview
	StatusRejected
	StatusStored
	StatusMoving
)

func NewPackageStatus(status string) PackageStatus {
	var s PackageStatus

	switch strings.ToLower(status) {
	case "stored":
		s = StatusStored
	case "rejected":
		s = StatusRejected
	case "in_review":
		s = StatusInReview
	case "moving":
		s = StatusMoving
	default:
		s = StatusUnspecified
	}

	return s
}

func (p PackageStatus) String() string {
	switch p {
	case StatusStored:
		return "stored"
	case StatusRejected:
		return "rejected"
	case StatusInReview:
		return "in_review"
	case StatusMoving:
		return "moving"
	}
	return "unspecified"
}

func (p PackageStatus) Values() []string {
	return []string{
		StatusUnspecified.String(),
		StatusInReview.String(),
		StatusRejected.String(),
		StatusStored.String(),
		StatusMoving.String(),
	}
}

// Value provides the DB a string from int.
func (p PackageStatus) Value() (driver.Value, error) {
	return p.String(), nil
}

// Scan tells our code how to read the enum into our type.
func (p *PackageStatus) Scan(val interface{}) error {
	var s string
	switch v := val.(type) {
	case nil:
		return nil
	case string:
		s = v
	case []uint8:
		s = string(v)
	}

	*p = NewPackageStatus(s)

	return nil
}

func (p PackageStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PackageStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*p = NewPackageStatus(s)

	return nil
}
