package iterator

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestChain(t *testing.T) {
	iter := &Limit[string]{
		Iterator: Map[int, string]{
			Iterator: Filter[int]{
				Iterator: &Generator[int]{
					Generate: func(i int) (bool, int, error) {
						return true, i + 1, nil
					},
				},
				Test: func(i int) bool {
					return i%2 == 0
				},
			},
			Update: func(i int) string {
				return strconv.Itoa(i)
			},
		},
		Max: 5,
	}
	want := []string{"2", "4", "6", "8", "10"}
	got, err := ToSlice(iter)
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
			&Slice[int]{Data: []int{1, 2, 3, 4, 5}},
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
