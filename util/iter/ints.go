package itools

import (
	"iter"
)

func Ints32u() iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		var i uint32 = 0
		for {
			if !yield(i) {
				return
			}

			i += 1
		}
	}
}
