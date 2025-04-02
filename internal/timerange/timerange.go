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
	return r.Start.Equal(r.End)
}

// Parse returns a new Range with the given Start and End strings.
// Parse will return a nil Range if both strings are nil, if only
// Start is nil it will use an arbitrary time far in the past, and
// if only End is nil it will use the current time. Parse will return
// an error if the End time is before the Start time or if the strings
// cannot be parsed as RFC3339 time format.
func Parse(start, end *string) (*Range, error) {
	var s, e time.Time
	var err error

	if start == nil && end == nil {
		return nil, nil
	}

	if start == nil {
		s = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		s, err = time.Parse(time.RFC3339, *start)
		if err != nil {
			return nil, errors.New("time range: cannot parse start time")
		}
	}

	if end == nil {
		e = time.Now()
	} else {
		e, err = time.Parse(time.RFC3339, *end)
		if err != nil {
			return nil, errors.New("time range: cannot parse end time")
		}
	}

	r, err := New(s, e)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
