package design

import (
	"goa.design/goa/v3/dsl"
)

func PaginatedCollectionOf(v interface{}, adsl ...func()) interface{} {
	return func() {
		dsl.Attribute("items", dsl.CollectionOf(v, adsl...))
		dsl.Attribute("next_cursor", dsl.String)
		dsl.Required("items")
	}
}

var Page = dsl.ResultType("application/vnd.enduro.page", func() {
	dsl.Description("Page represents a subset of search results.")
	dsl.Attribute("limit", dsl.Int, "Maximum items per page")
	dsl.Attribute("offset", dsl.Int, "Offset from first result to start of page")
	dsl.Attribute("total", dsl.Int, "Total result count before paging")
	dsl.Required("limit", "offset", "total")
})
