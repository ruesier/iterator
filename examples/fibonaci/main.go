package main

import (
	"fmt"

	"github.com/ruesier/iterator"
)

type fib struct {
	next int
	prev int
}

func (f *fib) Next() bool {
	f.next, f.prev = f.next+f.prev, f.next
	return true
}

func (f *fib) Get() int {
	return f.next
}

func (f *fib) Err() error { return nil }

func main() {
	var f iterator.Iterator[int] = &iterator.Limit[int]{
		Iterator: &fib{0, 1},
		Max:      10,
	}

	nums, _ := iterator.ToSlice[int](f)
	fmt.Printf("%v\n", nums)

	f = &fib{0, 1}
	g := &fib{1, 0}
	ratios, _ := iterator.ToSlice[float64](
		&iterator.Limit[float64]{
			Iterator: iterator.Combine[int, float64]{
				Iterators: []iterator.Iterator[int]{f, g},
				Join: func(vals ...int) float64 {
					return float64(vals[1]) / float64(vals[0])
				},
			},
			Max: 20,
		})
	for _, r := range ratios {
		fmt.Printf("%f\n", r)
	}
}
