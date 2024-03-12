package db

import (
	"database/sql"
	"time"
)

// FormatOptionalString returns the nil value when the string is empty.
func FormatOptionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// FormatOptionalTime returns the nil value when the value is NULL in the db.
func FormatOptionalTime(nt sql.NullTime) *string {
	var res *string
	if nt.Valid {
		f := FormatTime(nt.Time)
		res = &f
	}
	return res
}

// FormatTime returns an empty string when t has the zero value.
func FormatTime(t time.Time) string {
	var ret string
	if !t.IsZero() {
		ret = t.Format(time.RFC3339)
	}
	return ret
}
