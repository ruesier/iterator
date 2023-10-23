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
	f := &fib{0, 1}
	nums, _ := iterator.GetN[int](f, 10)
	fmt.Printf("%v\n", nums)

	f = &fib{0, 1}
	g := &fib{1, 0}
	ratios, _ := iterator.GetN[float64](iterator.Combine[int, float64]{
		Iters: []iterator.Iterator[int]{f, g},
		Join: func(vals ...int) float64 {
			return float64(vals[1]) / float64(vals[0])
		},
	}, 20)
	for _, r := range ratios {
		fmt.Printf("%f\n", r)
	}
}
