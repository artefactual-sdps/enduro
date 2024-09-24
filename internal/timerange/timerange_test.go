package timerange_test

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/timerange"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("Returns a time range", func(t *testing.T) {
		t.Parallel()

		r, err := timerange.New(
			time.Date(2024, 9, 17, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 9, 18, 0, 0, 0, 0, time.UTC),
		)
		assert.NilError(t, err)
		assert.DeepEqual(t, r, timerange.Range{
			Start: time.Date(2024, 9, 17, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 9, 18, 0, 0, 0, 0, time.UTC),
		})
	})

	t.Run("Errors when end time is before start time", func(t *testing.T) {
		t.Parallel()

		_, err := timerange.New(
			time.Date(2024, 9, 17, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 9, 16, 0, 0, 0, 0, time.UTC),
		)
		assert.Error(t, err, "time range: end cannot be before start")
	})
}

func TestNewInstant(t *testing.T) {
	t.Parallel()

	t.Run("Returns an instant time", func(t *testing.T) {
		t.Parallel()

		r := timerange.NewInstant(time.Date(2024, 9, 17, 0, 0, 0, 0, time.UTC))
		assert.DeepEqual(t, r, timerange.Range{
			Start: time.Date(2024, 9, 17, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 9, 17, 0, 0, 0, 0, time.UTC),
		})
	})
}

func TestIsZero(t *testing.T) {
	t.Parallel()

	t.Run("Returns true when start and end times are zero", func(t *testing.T) {
		t.Parallel()

		var r timerange.Range
		assert.Assert(t, r.IsZero())
	})

	t.Run("Returns false when the start or end time is not zero", func(t *testing.T) {
		t.Parallel()

		var t0 time.Time
		r, err := timerange.New(t0, time.Now())
		assert.NilError(t, err)
		assert.Assert(t, !r.IsZero())
	})
}

func TestIsInstant(t *testing.T) {
	t.Parallel()

	t.Run("Returns true when start and end time are equal", func(t *testing.T) {
		t.Parallel()

		n := time.Now()
		r, err := timerange.New(n, n)
		assert.NilError(t, err)
		assert.Assert(t, r.IsInstant())
	})

	t.Run("Returns false when start and end time are not equal", func(t *testing.T) {
		t.Parallel()

		r, err := timerange.New(time.Now(), time.Now())
		assert.NilError(t, err)
		assert.Assert(t, !r.IsInstant())
	})
}
