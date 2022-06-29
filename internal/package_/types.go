package package_

import (
	"database/sql"
	"time"

	goapackage "github.com/artefactual-labs/enduro/internal/api/gen/package_"
)

// Package represents a package in the package table.
type Package struct {
	ID         uint   `db:"id"`
	Name       string `db:"name"`
	WorkflowID string `db:"workflow_id"`
	RunID      string `db:"run_id"`
	AIPID      string `db:"aip_id"`
	Location   string `db:"location"`
	Status     Status `db:"status"`

	// It defaults to CURRENT_TIMESTAMP(6) so populated as soon as possible.
	CreatedAt time.Time `db:"created_at"`

	// Nullable, populated as soon as processing starts.
	StartedAt sql.NullTime `db:"started_at"`

	// Nullable, populated as soon as ingest completes.
	CompletedAt sql.NullTime `db:"completed_at"`
}

// Goa returns the API representation of the package.
func (c Package) Goa() *goapackage.EnduroStoredPackage {
	col := goapackage.EnduroStoredPackage{
		ID:          c.ID,
		Name:        formatOptionalString(c.Name),
		WorkflowID:  formatOptionalString(c.WorkflowID),
		RunID:       formatOptionalString(c.RunID),
		AipID:       formatOptionalString(c.AIPID),
		Location:    formatOptionalString(c.Location),
		Status:      c.Status.String(),
		CreatedAt:   formatTime(c.CreatedAt),
		StartedAt:   formatOptionalTime(c.StartedAt),
		CompletedAt: formatOptionalTime(c.CompletedAt),
	}

	return &col
}

// formatOptionalString returns the nil value when the string is empty.
func formatOptionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// formatOptionalTime returns the nil value when the value is NULL in the db.
func formatOptionalTime(nt sql.NullTime) *string {
	var res *string
	if nt.Valid {
		f := formatTime(nt.Time)
		res = &f
	}
	return res
}

// formatTime returns an empty string when t has the zero value.
func formatTime(t time.Time) string {
	var ret string
	if !t.IsZero() {
		ret = t.Format(time.RFC3339)
	}
	return ret
}
