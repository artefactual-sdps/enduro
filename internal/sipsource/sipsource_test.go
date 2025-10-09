package sipsource_test

import (
	"slices"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/sipsource"
)

func TestSort(t *testing.T) {
	t.Parallel()

	type test = struct {
		name string
		objs []*sipsource.Object
		sort *sipsource.Sort
		want []*sipsource.Object
	}
	for _, tt := range []test{
		{
			name: "No sort returns the order unchanged",
			objs: []*sipsource.Object{
				{Key: "b"},
				{Key: "a"},
				{Key: "c"},
			},
			want: []*sipsource.Object{
				{Key: "b"},
				{Key: "a"},
				{Key: "c"},
			},
		},
		{
			name: "Sorts by key ascending",
			objs: []*sipsource.Object{
				{Key: "b"},
				{Key: "a"},
				{Key: "c"},
			},
			sort: sipsource.SortByKey().Asc(),
			want: []*sipsource.Object{
				{Key: "a"},
				{Key: "b"},
				{Key: "c"},
			},
		},
		{
			name: "Sorts by key descending",
			objs: []*sipsource.Object{
				{Key: "b"},
				{Key: "a"},
				{Key: "c"},
			},
			sort: sipsource.SortByKey().Desc(),
			want: []*sipsource.Object{
				{Key: "c"},
				{Key: "b"},
				{Key: "a"},
			},
		},
		{
			name: "Sorts by ModTime ascending",
			objs: []*sipsource.Object{
				{Key: "a", ModTime: time.Date(2025, 10, 15, 12, 30, 0, 0, time.UTC)},
				{Key: "b", ModTime: time.Date(2025, 10, 15, 12, 29, 0, 0, time.UTC)},
				{Key: "c", ModTime: time.Date(2025, 10, 15, 12, 29, 0, 0, time.UTC)},
				{Key: "d", ModTime: time.Date(2025, 10, 15, 12, 31, 0, 0, time.UTC)},
			},
			sort: sipsource.SortByModTime().Asc(),
			want: []*sipsource.Object{
				{Key: "b", ModTime: time.Date(2025, 10, 15, 12, 29, 0, 0, time.UTC)},
				{Key: "c", ModTime: time.Date(2025, 10, 15, 12, 29, 0, 0, time.UTC)},
				{Key: "a", ModTime: time.Date(2025, 10, 15, 12, 30, 0, 0, time.UTC)},
				{Key: "d", ModTime: time.Date(2025, 10, 15, 12, 31, 0, 0, time.UTC)},
			},
		},
		{
			name: "Sorts by ModTime descending",
			objs: []*sipsource.Object{
				{ModTime: time.Date(2025, 10, 15, 12, 30, 0, 0, time.UTC)},
				{ModTime: time.Date(2025, 10, 15, 12, 29, 0, 0, time.UTC)},
				{ModTime: time.Date(2025, 10, 15, 12, 31, 0, 0, time.UTC)},
			},
			sort: sipsource.SortByModTime().Desc(),
			want: []*sipsource.Object{
				{ModTime: time.Date(2025, 10, 15, 12, 31, 0, 0, time.UTC)},
				{ModTime: time.Date(2025, 10, 15, 12, 30, 0, 0, time.UTC)},
				{ModTime: time.Date(2025, 10, 15, 12, 29, 0, 0, time.UTC)},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			slices.SortFunc(tt.objs, tt.sort.Compare)
			assert.DeepEqual(t, tt.objs, tt.want)
		})
	}
}
