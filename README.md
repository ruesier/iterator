# iterator
Golang implementation of a generic iterator. Inspiration comes from the `bufio.Scanner` type.
The major advantage of an iterator is in lazy evaluation of the elements.

## The Basics

### Iterator Interface

The core interface that this package defines is `Iterator`. Its three methods are:
- `Next` to prepare the next element.
- `Get` to return the current element.
- `Err` to report any errors that might have occurred.

### Iterators and Slices

This package includes tools for converting between slices and iterators when necessary.

The `Slice` type is a wrapper of a slice that implements `Iterator`.

The `ToSlice` function returns slice representation of an `Iterator`.

Example:
```golang
package main

import "github.com/ruesier/iterator"

func main() {
    data := []int{1, 2, 3, 4, 5}
    sliceIter := &Slice[int]{
        Data: data,
    }

    result, err := iterator.ToSlice(sliceIter)
    if err != nil {
        log.Fatal(err)
    }
    for i, d := range data {
        if d != result[i] {
            log.Fatalf("mismatched result expected %d at index %d, got %d", d, i, result[i])
        }
    }
}
```

### Transformations

The `Filter`, `Map`, and `Limit` types wrap iterators to produce new iterators.

- `Filter` only returns values that pass a provided test function
- `Map` applies an update function to each element
- `Limit` only returns a maximum number of elements

## Async Iteration
