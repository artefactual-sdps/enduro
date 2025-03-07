package entfilter

type (
	// Sort determines how the filtered results are sorted by specifying a
	// slice of sort columns.  The first SortColumn has the highest sort
	// precedence, and the last SortColumn the lowest precedence.
	Sort []SortColumn

	// SortColumn specifies a column name on which to sort results, and the
	// direction of the sort (ascending or descending).
	SortColumn struct {
		// Name of the column on which to sort the results.
		Name string

		// Desc is true if the sort order is descending.
		Desc bool
	}
)

// NewSort returns a new sort instance.
func NewSort() Sort {
	return Sort{}
}

// AddCol adds a SortColumn to a Sort then returns the updated Sort.
func (s Sort) AddCol(name string, desc bool) Sort {
	s = append(s, SortColumn{Name: name, Desc: desc})
	return s
}
