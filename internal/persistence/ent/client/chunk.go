package client

import "iter"

var defaultTaskBatchSize = 100

// chunk groups elements from the provided sequence into batches of size
// defaultTaskBatchSize, yielding a stream of slices, each containing at most
// defaultTaskBatchSize elements, except for the last batch which may contain
// fewer. Note that the slices are reused internally to minimize allocations, so
// they should not be retained or modified after being yielded to avoid
// unexpected behavior.
func chunk[T any](seq iter.Seq[T]) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		b := make([]T, 0, defaultTaskBatchSize)
		for a := range seq {
			b = append(b, a)
			if len(b) == defaultTaskBatchSize {
				if !yield(b) {
					return
				}
				b = b[:0]
			}
		}
		if len(b) != 0 {
			if !yield(b) {
				return
			}
		}
	}
}
