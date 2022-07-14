package storage

import (
	"encoding/json"
	"strings"

	goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

type PackageStatus uint

const (
	StatusUnspecified PackageStatus = iota
	StatusInReview
	StatusRejected
	StatusStored
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
	}
	return "unspecified"
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

// Package represents a package in the storage_package table.
type Package struct {
	ID        uint          `db:"id"`
	Name      string        `db:"name"`
	AIPID     string        `db:"aip_id"`
	Status    PackageStatus `db:"status"`
	ObjectKey string        `db:"object_key"`
	Location  string        `db:"location"`
}

// Goa returns the API representation of the package.
func (p Package) Goa() *goastorage.StoredStoragePackage {
	return &goastorage.StoredStoragePackage{
		ID:        p.ID,
		Name:      p.Name,
		AipID:     p.AIPID,
		Status:    p.Status.String(),
		ObjectKey: p.ObjectKey,
		Location:  &p.Location,
	}
}
