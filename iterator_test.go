package iterator

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestChain(t *testing.T) {
	gen := &Generator[int]{
		Generate: func(i int) (int, error) {
			return i + 1, nil
		},
	}
	filter := Filter[int]{
		Iterator: gen,
		Test: func(i int) bool {
			return i%2 == 0
		},
	}
	mapped := Map[int, string]{
		Iterator: filter,
		Update: func(i int) string {
			return strconv.Itoa(i)
		},
	}
	limited := &Limit[string]{
		Iterator: mapped,
		Max:      5,
	}
	want := []string{"2", "4", "6", "8", "10"}
	got, err := ToSlice(limited)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("mismatch result (-got, +want): %s", diff)
	}
}

func TestCombine(t *testing.T) {
	iter := Combine[int, int]{
		Iterators: []Iterator[int]{
			&Slice[int]{Slice: []int{1, 2, 3, 4, 5}},
			Echo[int]{CloneNoOp[int]{2}},
		},
		Join: func(values ...int) int {
			var sum int
			for _, v := range values {
				sum += v
			}
			return sum
		},
	}
	want := []int{3, 4, 5, 6, 7}
	got, err := ToSlice[int](iter)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("mismatch result (-got, +want): %s", diff)
	}
}
