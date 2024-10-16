package entclient

import (
	"slices"
	"testing"

	"gotest.tools/v3/assert"
)

func TestBatch(t *testing.T) {
	original := defaultBatchSize

	tests := map[string]struct {
		size     int
		values   []int
		expected [][]int
	}{
		"batch size 1": {
			size:     1,
			values:   []int{1, 2, 3},
			expected: [][]int{{1}, {2}, {3}},
		},
		"batch size 2 with exact fit": {
			size:     2,
			values:   []int{1, 2, 3, 4},
			expected: [][]int{{1, 2}, {3, 4}},
		},
		"batch size 2 with remainder": {
			size:     2,
			values:   []int{1, 2, 3},
			expected: [][]int{{1, 2}, {3}},
		},
		"batch size larger than input": {
			size:     5,
			values:   []int{1, 2, 3},
			expected: [][]int{{1, 2, 3}},
		},
		"empty input": {
			size:     2,
			values:   []int{},
			expected: [][]int{},
		},
		"batch size 3 with partial last batch": {
			size:     3,
			values:   []int{1, 2, 3, 4, 5},
			expected: [][]int{{1, 2, 3}, {4, 5}},
		},
		"batch size equals input length": {
			size:     3,
			values:   []int{1, 2, 3},
			expected: [][]int{{1, 2, 3}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			defaultBatchSize = tc.size
			seq := batch(slices.Values(tc.values))

			ret := [][]int{}
			for item := range seq {
				cp := make([]int, len(item))
				copy(cp, item)
				ret = append(ret, cp)
			}

			assert.DeepEqual(t, ret, tc.expected)
		})
	}

	defaultBatchSize = original
}
