package persistence

import (
	"github.com/google/uuid"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

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

// Page represents a subset of results within a search result set.
type Page struct {
	// Limit is the maximum number of results per page.
	Limit int

	// Offset is the ordinal position, relative to the start of the unfiltered
	// set, of the first result of the page.
	Offset int

	// Total is the total number of search results before paging.
	Total int
}

func (p *Page) Goa() *goaingest.EnduroPage {
	if p == nil {
		return nil
	}

	return &goaingest.EnduroPage{
		Limit:  p.Limit,
		Offset: p.Offset,
		Total:  p.Total,
	}
}

type SIPFilter struct {
	// Name filters for SIPs whose names contain the given string.
	Name *string

	AIPID      *uuid.UUID
	LocationID *uuid.UUID
	Status     *enums.SIPStatus
	CreatedAt  *timerange.Range

	Sort
	Page
}
