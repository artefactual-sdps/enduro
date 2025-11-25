package client

import (
	"slices"
	"testing"

	"gotest.tools/v3/assert"
)

func setDefaultTaskBatchSize(t *testing.T, size int) {
	t.Helper()
	prev := defaultTaskBatchSize
	defaultTaskBatchSize = size
	t.Cleanup(func() { defaultTaskBatchSize = prev })
}

func TestChunk(t *testing.T) {
	tests := map[string]struct {
		size     int
		values   []int
		expected [][]int
	}{
		"chunk size 1": {
			size:     1,
			values:   []int{1, 2, 3},
			expected: [][]int{{1}, {2}, {3}},
		},
		"chunk size 2 with exact fit": {
			size:     2,
			values:   []int{1, 2, 3, 4},
			expected: [][]int{{1, 2}, {3, 4}},
		},
		"chunk size 2 with remainder": {
			size:     2,
			values:   []int{1, 2, 3},
			expected: [][]int{{1, 2}, {3}},
		},
		"chunk size larger than input": {
			size:     5,
			values:   []int{1, 2, 3},
			expected: [][]int{{1, 2, 3}},
		},
		"empty input": {
			size:     2,
			values:   []int{},
			expected: [][]int{},
		},
		"chunk size 3 with partial last chunk": {
			size:     3,
			values:   []int{1, 2, 3, 4, 5},
			expected: [][]int{{1, 2, 3}, {4, 5}},
		},
		"chunk size equals input length": {
			size:     3,
			values:   []int{1, 2, 3},
			expected: [][]int{{1, 2, 3}},
		},
		"single element": {
			size:     3,
			values:   []int{42},
			expected: [][]int{{42}},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setDefaultTaskBatchSize(t, tc.size)
			seq := chunk(slices.Values(tc.values))

			got := [][]int{}
			for batch := range seq {
				cp := make([]int, len(batch))
				copy(cp, batch)
				got = append(got, cp)
			}

			assert.DeepEqual(t, got, tc.expected)
		})
	}
}

// TestChunkEarlyTermination ensures chunk stops consuming as soon as the caller
// returns false from the yield callback, so it doesn't read more input than
// needed.
func TestChunkEarlyTermination(t *testing.T) {
	setDefaultTaskBatchSize(t, 2)
	values := []int{1, 2, 3, 4, 5, 6}
	seq := chunk(slices.Values(values))

	// Consume only the first batch, then stop.
	var got [][]int
	for batch := range seq {
		cp := make([]int, len(batch))
		copy(cp, batch)
		got = append(got, cp)
		break // Early termination.
	}

	assert.DeepEqual(t, got, [][]int{{1, 2}})
}
