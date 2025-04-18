package entfilter_test

import (
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"go.artefactual.dev/tools/ref"
	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/entfilter"
	"github.com/artefactual-sdps/enduro/internal/enums"
	"github.com/artefactual-sdps/enduro/internal/timerange"
)

type pred func(*sql.Selector)

type orderOpt func(*sql.Selector)

// query is a querier like *db.SIPQuery for testing.
type query struct {
	table  string
	limit  int
	offset int
	order  []string
	where  string
	args   []any
}

func (q *query) Where(preds ...pred) *query {
	sel := sql.Select().From(sql.Table(q.table))
	for _, pred := range preds {
		pred(sel)
	}
	_, q.args = sel.Query()
	if p := sel.P(); p != nil {
		q.where = sel.P().String()
	}
	return q
}

func (q *query) Limit(l int) *query {
	q.limit = l
	return q
}

func (q *query) Offset(o int) *query {
	q.offset = o
	return q
}

func (q *query) Order(fn ...orderOpt) *query {
	sel := sql.Select().From(sql.Table(q.table))
	for _, f := range fn {
		f(sel)
	}
	q.order = sel.OrderColumns()
	return q
}

func (q query) Clone() *query {
	return &query{
		table:  q.table,
		limit:  q.limit,
		offset: q.offset,
		order:  append([]string{}, q.order...),
		where:  q.where,
		args:   append([]any{}, q.args...),
	}
}

func newSortableFields(fields ...string) entfilter.SortableFields {
	sf := map[string]entfilter.SortableField{}
	for i, name := range fields {
		sf[name] = entfilter.SortableField{Name: name, Default: i == 0}
	}

	return sf
}

func TestFilter(t *testing.T) {
	t.Parallel()

	t.Run("Sorts allowed fields", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id", "name"),
		)

		f.OrderBy(entfilter.NewSort().
			AddCol("id", false).
			AddCol("name", false),
		)
		page, whole := f.Apply()

		assert.DeepEqual(
			t,
			page,
			&query{
				table: "data",
				limit: entfilter.DefaultPageSize,
				order: []string{"`data`.`id`", "`data`.`name`"},
				args:  []any{},
			},
			cmp.AllowUnexported(query{}),
		)
		assert.DeepEqual(
			t,
			whole,
			&query{
				table: "data",
				order: []string{"`data`.`id`", "`data`.`name`"},
			},
			cmp.AllowUnexported(query{}),
		)
	})

	t.Run("Sorts allowed fields in descending order", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id", "name"),
		)
		f.OrderBy(entfilter.NewSort().AddCol("name", true))
		page, whole := f.Apply()

		assert.DeepEqual(
			t,
			page,
			&query{
				table: "data",
				limit: entfilter.DefaultPageSize,
				order: []string{"`data`.`name` DESC"},
				args:  []any{},
			},
			cmp.AllowUnexported(query{}),
		)
		assert.DeepEqual(
			t,
			whole,
			&query{
				table: "data",
				order: []string{"`data`.`name` DESC"},
			},
			cmp.AllowUnexported(query{}),
		)
	})

	t.Run("Sorts by default sort column", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			map[string]entfilter.SortableField{
				"id":   {Name: "id"},
				"name": {Name: "name", Default: true},
			},
		)
		page, whole := f.Apply()

		assert.DeepEqual(
			t,
			page,
			&query{
				table: "data",
				limit: entfilter.DefaultPageSize,
				order: []string{"`data`.`name`"},
				args:  []any{},
			},
			cmp.AllowUnexported(query{}),
		)
		assert.DeepEqual(
			t,
			whole,
			&query{
				table: "data",
				order: []string{"`data`.`name`"},
			},
			cmp.AllowUnexported(query{}),
		)
	})

	t.Run("Panics when sortableFields is empty", func(t *testing.T) {
		t.Parallel()

		q := &query{table: "data"}

		defer func() {
			r := recover()
			assert.Equal(t, r.(string), "sortableFields is empty")
		}()

		entfilter.NewFilter(q, nil)
	})

	t.Run("Panics when no default sortableField is set", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			map[string]entfilter.SortableField{
				"id": {Name: "id"},
			},
		)

		defer func() {
			r := recover()
			assert.Equal(t, r.(string), "no default sort field specified")
		}()

		_, _ = f.Apply()
	})

	t.Run("Handles unknown sorting field, defaults to first known field", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id", "age"),
		)
		f.OrderBy(entfilter.NewSort().AddCol("count", false))
		page, whole := f.Apply()
		assert.DeepEqual(
			t,
			page,
			&query{
				table: "data",
				limit: entfilter.DefaultPageSize,
				order: []string{"`data`.`id`"},
				args:  []any{},
			},
			cmp.AllowUnexported(query{}),
		)
		assert.DeepEqual(
			t,
			whole,
			&query{
				table: "data",
				order: []string{"`data`.`id`"},
			},
			cmp.AllowUnexported(query{}),
		)
	})

	t.Run("Sorts by the final sort on a field", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id", "name"),
		)
		f.OrderBy(entfilter.NewSort().
			AddCol("name", true).
			AddCol("name", false),
		)
		page, whole := f.Apply()

		assert.DeepEqual(t, page.order, []string{"`data`.`name`"})
		assert.DeepEqual(t, whole.order, []string{"`data`.`name`"})
	})

	t.Run("Default sort when given an empty order param", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id", "name"),
		)
		f.OrderBy(entfilter.NewSort())
		page, whole := f.Apply()

		assert.DeepEqual(t, page.order, []string{"`data`.`id`"})
		assert.DeepEqual(t, whole.order, []string{"`data`.`id`"})
	})

	t.Run("Sets page limit and offset", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id", "name"),
		)
		f.Page(50, 100)
		f.OrderBy(entfilter.NewSort().AddCol("name", false))
		page, whole := f.Apply()

		assert.Equal(t, page.limit, 50)
		assert.Equal(t, page.offset, 100)
		assert.DeepEqual(t, page.order, []string{"`data`.`name`"})

		assert.Equal(t, whole.limit, 0)
		assert.Equal(t, whole.offset, 0)
		assert.DeepEqual(t, whole.order, []string{"`data`.`name`"})
	})

	t.Run("Page size defaults to default value", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)
		page, whole := f.Apply()

		assert.Equal(t, page.limit, entfilter.DefaultPageSize)
		assert.Equal(t, whole.limit, 0)
	})

	t.Run("Passing a zero page limit uses the default page limit", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)
		f.Page(0, 0)
		page, whole := f.Apply()

		assert.Equal(t, page.limit, entfilter.DefaultPageSize)
		assert.Equal(t, whole.limit, 0)
	})

	t.Run("Page size limited to max page size", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)
		f.Page(100_000, 0)
		page, whole := f.Apply()

		assert.Equal(t, page.limit, entfilter.MaxPageSize)
		assert.Equal(t, whole.limit, 0)
	})

	t.Run("Page size is max page size when set to a negative number", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)
		f.Page(-100, 0)
		page, whole := f.Apply()

		assert.Equal(t, page.limit, entfilter.MaxPageSize)
		assert.Equal(t, whole.limit, 0)
	})

	t.Run("Adds an equality filter", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)

		id := 1234
		name := "Joe"
		aipID := uuid.New()

		f.Equals("id", &id)                 // Filter on an *int value.
		f.Equals("id2", id)                 // Ignore a non-pointer (int) value.
		f.Equals("name", &name)             // Filter on a *string value.
		f.Equals("aip_id", &aipID)          // Filter on a *uuid.UUID value.
		f.Equals("address", (*string)(nil)) // Ignore a typed nil value.
		f.Equals("address2", nil)           // Ignore (interface{})(nil).
		_, whole := f.Apply()

		assert.Equal(t, whole.where, "(`data`.`id` = ? AND `data`.`name` = ?) AND `data`.`aip_id` = ?")
		assert.DeepEqual(t, whole.args, []any{&id, &name, &aipID})
	})

	t.Run("Filters enums", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)

		// Add a string enum entfilter.
		taskOutcome := enums.PreprocessingTaskOutcomeSuccess
		f.Equals("outcome", &taskOutcome)

		// Add an integer enum entfilter.
		sipStatus := enums.SIPStatusIngested
		f.Equals("status", &sipStatus)

		// Omit invalid enum values.
		f.Equals("outcome2", ref.New(enums.PreprocessingTaskOutcome("invalid")))

		// Omit nil enum pointers.
		f.Equals("status2", (*enums.SIPStatus)(nil))

		_, whole := f.Apply()

		assert.Equal(t, whole.where, "`data`.`outcome` = ? AND `data`.`status` = ?")
		assert.DeepEqual(t, whole.args, []any{&taskOutcome, &sipStatus})
	})

	t.Run("Filters on a list of strings", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)
		f.In("name", []any{"foo", "bar", ""})
		f.In("empty", []any{})    // Ignore an empty slice.
		f.In("nil", ([]any)(nil)) // Ignore a nil slice.
		_, whole := f.Apply()

		assert.Equal(t, whole.where, "`data`.`name` IN (?, ?, ?)")
		assert.DeepEqual(t, whole.args, []any{"foo", "bar", ""})
	})

	t.Run("Filters on a list of enums", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)
		f.In("status", []any{
			enums.SIPStatusProcessing,
			enums.SIPStatusIngested,
			enums.SIPStatus("invalid"), // Ignore an invalid enum.
		})
		_, whole := f.Apply()

		assert.Equal(t, whole.where, "`data`.`status` IN (?, ?)")
		assert.DeepEqual(t, whole.args, []any{
			enums.SIPStatusProcessing,
			enums.SIPStatusIngested,
		})
	})

	t.Run("Filters on a list of UUIDs", func(t *testing.T) {
		t.Parallel()

		uuid0 := uuid.New()
		uuid1 := uuid.New()
		var uuid2 uuid.UUID

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)
		f.In("aip_id", []any{
			uuid0,
			uuid1,
			uuid2, // Ignore a nil UUID.
		})
		_, whole := f.Apply()

		assert.Equal(t, whole.where, "`data`.`aip_id` IN (?, ?)")
		assert.DeepEqual(t, whole.args, []any{uuid0, uuid1})
	})

	t.Run("Filters on a date range", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)

		r, err := timerange.New(
			time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
		)
		assert.NilError(t, err)

		f.AddDateRange("created_at", &r)
		_, whole := f.Apply()

		assert.Equal(t, whole.where, "`created_at` >= ? AND `created_at` < ?")
		assert.DeepEqual(t, whole.args, []any{
			time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
		})
	})

	t.Run("Filters on an exact time", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)

		r := timerange.NewInstant(time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC))
		f.AddDateRange("created_at", &r)
		_, whole := f.Apply()

		assert.Equal(t, whole.where, "`created_at` = ?")
		assert.DeepEqual(t, whole.args, []any{
			time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
		})
	})

	t.Run("No filter added when date range is zero", func(t *testing.T) {
		t.Parallel()

		f := entfilter.NewFilter(
			&query{table: "data"},
			newSortableFields("id"),
		)

		var r timerange.Range
		f.AddDateRange("created_at", &r)
		_, whole := f.Apply()

		assert.Equal(t, whole.where, "")
		assert.Assert(t, whole.args == nil)
	})
}
