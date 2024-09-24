package entclient

import (
	"slices"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"

	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/persistence"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

const (
	DefaultPageSize int = 20
	MaxPageSize     int = 50_000
)

// Predicate (P) is the constraint for all Ent predicates, e.g. predicate.Batch,
// predicate.Transfer and so on.
type Predicate interface {
	~func(s *sql.Selector)
}

// OrderOption (O) is the constraint for all Ent ordering options, e.g.
// batch.OrderOption, transfer.OrderOption and so on.
type OrderOption interface {
	~func(s *sql.Selector)
}

// Querier (Q) wraps queriers methods provided by Ent queries.
type Querier[P Predicate, O OrderOption, Q any] interface {
	Where(ps ...P) Q
	Limit(int) Q
	Offset(int) Q
	Order(...O) Q
	Clone() Q
}

type columnFilter[P Predicate] struct {
	column    string
	predicate P
}

type orderOption[O OrderOption] struct {
	column string
	option O
}

// Filter provides a mechanism to filter, order and paginate using Ent queries.
// Invoke the Apply method last to apply the remaining filters.
type Filter[Q Querier[P, O, Q], O OrderOption, P Predicate] struct {
	query          Q
	filters        []columnFilter[P]
	sortableFields SortableFields
	orderBy        []orderOption[O]
	limit          int
	offset         int
}

// NewFilter returns a new Filter. It panics if orderingFields is empty.
func NewFilter[Q Querier[P, O, Q], O OrderOption, P Predicate](query Q, sf SortableFields) *Filter[Q, O, P] {
	if len(sf) == 0 {
		panic("sortableFields is empty")
	}

	f := &Filter[Q, O, P]{
		query:          query,
		filters:        []columnFilter[P]{},
		sortableFields: sf,
		orderBy:        []orderOption[O]{},
		limit:          DefaultPageSize,
	}

	return f
}

// OrderBy sets the query sort order.
func (f *Filter[Q, O, P]) OrderBy(sort persistence.Sort) {
	if len(sort) == 0 {
		return
	}

	for _, c := range sort {
		f.addOrderOpt(c.Name, c.Desc)
	}
}

func (f *Filter[Q, O, P]) addOrderOpt(field string, dsc bool) {
	// Check that field is an allowed sortableField.
	if !slices.Contains(f.sortableFields.Columns(), field) {
		return
	}

	opt := orderOption[O]{
		column: field,
		option: orderFunc(field, dsc),
	}

	// Check if we've already sorted on this field.
	i := slices.IndexFunc(f.orderBy, func(o orderOption[O]) bool {
		return o.column == field
	})

	switch {
	case i < 0:
		f.orderBy = append(f.orderBy, opt)
	default:
		// Replace any previous sort on this field.
		f.orderBy[i] = opt
	}
}

func (f *Filter[Q, O, P]) setDefaultOrderBy(sf SortableFields) {
	d := sf.Default()
	f.addOrderOpt(d.Name, false)
}

// orderFunc is called by the ent query builder to convert a selector
// OrderOption to a MySQL "order by" clause.
func orderFunc(field string, desc bool) func(sel *sql.Selector) {
	return func(sel *sql.Selector) {
		s := sel.C(field)
		if desc {
			s += " DESC"
		}
		sel.OrderBy(s)
	}
}

// Page sets the query limit and offset criteria.
//
// The actual query limit will be set to a value between one and MaxPageSize
// based on the input limit (x) as follows:
// (x < 0)               -> MaxPageSize
// (x == 0)              -> DefaultPageSize
// (0 < x < MaxPageSize) -> x
// (x >= MaxPageSize)    -> MaxPageSize
//
// The actual query offset will be set based on the input offset (y) as follows:
// (y <= 0) -> 0
// (y > 0)  -> y
func (f *Filter[Q, O, P]) Page(limit, offset int) {
	f.addLimit(limit)

	if offset > 0 {
		f.offset = offset
	}
}

// addLimit adds the page limit l to a filter.
func (f *Filter[Q, O, P]) addLimit(l int) {
	switch {
	case l == 0:
		l = DefaultPageSize
	case l < 0:
		l = MaxPageSize
	case l > MaxPageSize:
		l = MaxPageSize
	}

	f.limit = l
}

// addFilter adds a new selector for column. Any existing filters on column will
// be retained to allow multiple criteria for the same column (e.g. name="foo"
// or name="bar").
func (f *Filter[Q, O, P]) addFilter(column string, selector func(s *sql.Selector)) {
	f.filters = append(f.filters, columnFilter[P]{column, selector})
}

// validPtrValue returns true if the given pointer ptr is not nil, and the
// underlying value is valid.
//
// Validating pointers is complicated because ptr has an interface{} type. The
// conditional `ptr == nil` doesn't evaluate true when ptr is a typed nil like
// (*enums.PackageStatus)(nil). A type switch case on the validator interface
// can then assign the nil *enums.PackageStatus to the validator interface and
// calling `t.IsValid()` causes a panic from trying to call `IsValid()` on a
// nil pointer.
func validPtrValue(ptr any) bool {
	if ptr == nil {
		return false
	}

	switch t := ptr.(type) {
	case *enums.PackageStatus:
		return t != nil && t.IsValid()
	case *enums.PreprocessingTaskOutcome:
		return t != nil && t.IsValid()
	case *int:
		return t != nil
	case *string:
		return t != nil
	case *uuid.UUID:
		return t != nil && *t != uuid.Nil
	default:
		// Return false when v's type is unknown.
		return false
	}
}

// Equals adds a filter on column being equal to value. If value implements the
// validator interface, value is validated before the filter is added.
func (f *Filter[Q, O, P]) Equals(column string, value any) {
	// The current code always calls this function with a pointer value (e.g.
	// *string, *enums.PackageStatus). If we need to pass value types (e.g.
	// (string, enums.PackageStatus) in the future we'll have to combine the
	// validPtrValue() & validValue() type switch cases.
	if !validPtrValue(value) {
		return
	}

	f.addFilter(column, func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(column), value))
	})
}

// Validator is a simple validation interface. Validator is currently used for
// enums, but it could represent any type that implements validation.
type validator interface {
	IsValid() bool
}

func validValue(v any) bool {
	switch t := v.(type) {
	case validator:
		return t.IsValid()
	case uuid.UUID:
		return t != uuid.Nil
	default:
		// Return true for all types that can't be validated. This allows
		// filtering for empty values (e.g. the empty string "").
		return true
	}
}

// In adds a filter on column being equal to one of the given values. Each
// element in values that implements validator is validated before being added
// to the list of filter values.
func (f *Filter[Q, O, P]) In(column string, values []any) {
	if len(values) == 0 {
		return
	}

	validated := make([]any, 0, len(values))
	for _, val := range values {
		// I can't see any reason we'd want to pass pointers as elements in the
		// values slice. We can and do pass ([]any)(nil) but doing so skips this
		// loop altogether.
		if !validValue(val) {
			continue
		}
		validated = append(validated, val)
	}

	if len(validated) == 0 {
		return
	}

	f.addFilter(column, func(s *sql.Selector) {
		s.Where(sql.In(s.C(column), validated...))
	})
}

// dateRangeSelector returns a predicate matching rows within a date range
// (range.Start <= date < range.End).
func dateRangeSelector(column string, r *timerange.Range) func(*sql.Selector) {
	return func(s *sql.Selector) {
		var p *sql.Predicate

		switch {
		case r.IsInstant():
			p = sql.EQ(column, r.Start)
		default:
			p = sql.And(
				sql.GTE(column, r.Start),
				sql.LT(column, r.End),
			)
		}

		s.Where(p)
	}
}

func (f *Filter[Q, O, P]) AddDateRange(column string, r *timerange.Range) {
	if r == nil || r.IsZero() {
		return
	}

	f.addFilter(column, dateRangeSelector(column, r))
}

// Apply filters, returning queriers of the filtered subset and the page.
func (f *Filter[Q, O, P]) Apply() (page, whole Q) {
	whole = f.query.Clone()

	ps := []P{}
	for _, cf := range f.filters {
		ps = append(ps, cf.predicate)
	}
	whole.Where(ps...)

	if len(f.orderBy) == 0 {
		f.setDefaultOrderBy(f.sortableFields)
	}

	opts := []O{}
	for _, ob := range f.orderBy {
		opts = append(opts, ob.option)
	}
	whole.Order(opts...)

	page = whole.Clone()
	page.Limit(f.limit)
	page.Offset(f.offset)

	return page, whole
}

type SortableField struct {
	Name    string
	Default bool
}

// SortableFields maps column names to Ent type field names.
// Usage examples: batchOrderingFields, transferOrderingFields...
type SortableFields map[string]SortableField

// Default returns the default sort field.
func (sf SortableFields) Default() SortableField {
	for _, f := range sf {
		if f.Default {
			return f
		}
	}

	panic("no default sort field specified")
}

func (sf SortableFields) Columns() []string {
	return maps.Keys(sf)
}
