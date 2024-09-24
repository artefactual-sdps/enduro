package timerange

import (
	"errors"
	"time"
)

type Range struct {
	Start time.Time
	End   time.Time
}

// New returns a new Range with the given Start and End times. New will return
// an error if the End time is before the Start time.
func New(start, end time.Time) (Range, error) {
	if end.Before(start) {
		return Range{}, errors.New("time range: end cannot be before start")
	}

	return Range{Start: start, End: end}, nil
}

// NewInstant returns a Range where the Start and End times are both set to the
// given time.
func NewInstant(t time.Time) Range {
	return Range{Start: t, End: t}
}

// IsZero returns true when both the Start and End times are zero.
func (r Range) IsZero() bool {
	return r.Start.IsZero() && r.End.IsZero()
}

// IsInstant returns true when the Start an End times are equal.
func (r Range) IsInstant() bool {
	return r.Start == r.End
}
