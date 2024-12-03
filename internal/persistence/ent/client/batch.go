package entclient

import "iter"

var defaultBatchSize = 100

// batch groups elements from the provided sequence into batches of size
// defaultBatchSize, yielding a stream of slices, each containing at most
// defaultBatchSize elements, except for the last batch which may contain fewer.
// Note that the slices are reused internally to minimize allocations, so they
// should not be retained or modified after being yielded to avoid unexpected
// behavior.
func batch[T any](seq iter.Seq[T]) iter.Seq[[]T] {
	return func(yield func([]T) bool) {
		b := make([]T, 0, defaultBatchSize)
		for a := range seq {
			b = append(b, a)
			if len(b) == defaultBatchSize {
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
