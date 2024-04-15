package db_test

import (
	"database/sql"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/db"
)

func TestFormatOptionalString(t *testing.T) {
	t.Parallel()

	t.Run("Returns nil pointer for an empty string", func(t *testing.T) {
		t.Parallel()
		got := db.FormatOptionalString("")
		assert.Assert(t, got == nil)
	})

	t.Run("Returns a pointer to a string", func(t *testing.T) {
		t.Parallel()
		got := db.FormatOptionalString("foo")
		assert.Equal(t, *got, "foo")
	})
}

func TestFormatOptionalTime(t *testing.T) {
	t.Parallel()

	t.Run("Returns nil pointer for null time", func(t *testing.T) {
		t.Parallel()
		got := db.FormatOptionalTime(sql.NullTime{})
		assert.Assert(t, got == nil)
	})

	t.Run("Returns an RFC3339 time string", func(t *testing.T) {
		t.Parallel()
		got := db.FormatOptionalTime(sql.NullTime{
			Time:  time.Date(2024, 3, 6, 11, 57, 17, 115, time.UTC),
			Valid: true,
		})
		assert.Equal(t, *got, "2024-03-06T11:57:17Z")
	})
}
