package itools

import (
	"iter"
)

func Map[T, U any](
	original iter.Seq[T],
	mapper func(T) U,
) iter.Seq[U] {
	return func(yield func(U) bool) {
		for t := range original {
			var mapd U = mapper(t)
			if !yield(mapd) {
				return
			}
		}
	}
}
