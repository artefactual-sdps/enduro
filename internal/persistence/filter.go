package persistence

import (
	"github.com/google/uuid"

	goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

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
	Status     *enums.SIPStatus
	CreatedAt  *timerange.Range
	UploaderID *uuid.UUID

	entfilter.Sort
	Page
}
