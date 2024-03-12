package enums

import (
	"encoding/json"
	"strings"
)

// See https://gist.github.com/sevein/dd36c2af23fd0d9e2e2438d8eb091314.
type PackageStatus uint

const (
	PackageStatusNew        PackageStatus = iota // Unused!
	PackageStatusInProgress                      // Undergoing work.
	PackageStatusDone                            // Work has completed.
	PackageStatusError                           // Processing failed.
	PackageStatusUnknown                         // Unused!
	PackageStatusQueued                          // Awaiting resource allocation.
	PackageStatusAbandoned                       // User abandoned processing.
	PackageStatusPending                         // Awaiting user decision.
)

func NewPackageStatus(status string) PackageStatus {
	var s PackageStatus

	switch strings.ToLower(status) {
	case "new":
		s = PackageStatusNew
	case "in progress":
		s = PackageStatusInProgress
	case "done":
		s = PackageStatusDone
	case "error":
		s = PackageStatusError
	case "queued":
		s = PackageStatusQueued
	case "abandoned":
		s = PackageStatusAbandoned
	case "pending":
		s = PackageStatusPending
	default:
		s = PackageStatusUnknown
	}

	return s
}

func (p PackageStatus) String() string {
	switch p {
	case PackageStatusNew:
		return "new"
	case PackageStatusInProgress:
		return "in progress"
	case PackageStatusDone:
		return "done"
	case PackageStatusError:
		return "error"
	case PackageStatusQueued:
		return "queued"
	case PackageStatusAbandoned:
		return "abandoned"
	case PackageStatusPending:
		return "pending"
	}
	return "unknown"
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
