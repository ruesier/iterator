// Package iterator defines an interface to represent lazy iteration of values.
// Also includes types and functions for interacting with iterators.
package iterator

import "golang.org/x/exp/constraints"

// Iterator is the primary interface this package defines. The basic usage of
// an Iterator looks like this.
//
// ```golang
//
//	for iter.Next() {
//		doSomething(iter.Get())
//	}
//	if err := iter.Err(); err != nil {
//		panic(err)
//	}
//
// ````
type Iterator[E any] interface {
	// Next returns true when there is a value to be accessed with Get.
	// False either means that the Iterator is empty or an error occurred.
	Next() bool
	// Get returns the current value of the Iterator. Get should return the same
	// value until Next is called. Get's behavior is undefined if called before
	// Next or when Next returns false.
	Get() E
	// Err returns any errors occurred. if Next returns true, this should return nil
	Err() error
}

// Filter is an Iterator that wraps another iterator. It only returns the values
// that the Test function returns true for.
type Filter[E any] struct {
	Iterator[E]
	// Test should return true if the provided value should be included.
	Test func(E) bool
}

func (f Filter[E]) Next() bool {
	for f.Iterator.Next() {
		if f.Test(f.Get()) {
			return true
		}
	}
	return false
}

// Map is an Iterator that wraps another iterator. It applies the Convert function
// to each value. This can result in a change in the iterating type.
type Map[BEFORE any, AFTER any] struct {
	Iterator[BEFORE]
	Update func(BEFORE) AFTER
}

func (m Map[BEFORE, AFTER]) Get() AFTER {
	return m.Update(m.Iterator.Get())
}

// Slice is a wrapper of a slice that implements the Iterator interface.
type Slice[E any] struct {
	Slice []E

	started bool
}

func (s *Slice[E]) Next() bool {
	if len(s.Slice) == 0 {
		return false
	}
	if s.started {
		s.Slice = s.Slice[1:]
		return len(s.Slice) > 0
	} else {
		s.started = true
		return true
	}
}

func (s *Slice[E]) Get() E {
	return s.Slice[0]
}

func (s *Slice[E]) Err() error {
	return nil
}

// Range returns an iterator from a start number to an end number, not including the end value.
// Range interprets the number of parameters as follows:
// - Range() => empty Iterator, Next will return false.
// - Range(end) => an Iterator over 0 <= val < end, step = 1.
// - Range(start, end) => an Iterator over start <= val < end, step = 1.
// - Range(start, end, step) => an Iterator over start <= val < end, by the step value.
// any parameters past the first 3 are ignored.
func Range[NUM constraints.Integer | constraints.Float](vals ...NUM) Iterator[NUM] {
	var start, end, step NUM
	switch len(vals) {
	case 0:
		return &Generator[NUM]{
			Generate: func(_ NUM) (NUM, error) {
				return 0, Stop
			},
		}
	case 1:
		start = 0
		end = vals[0]
		step = 1
	case 2:
		start = vals[0]
		end = vals[1]
		step = 1
	default:
		start = vals[0]
		end = vals[1]
		step = vals[2]
	}
	return &Generator[NUM]{
		Generate: func(prev NUM) (NUM, error) {
			next := prev + step
			if next >= end {
				return prev, Stop
			}
			return next, nil
		},
		val: start,
	}
}

// ToSlice reads all the values from an iterator and returns them all. If Iterator
// returns an infinite number of values this will result in an infinite loop.
// Use an `iterator.Limit` to prevent infinite loops.
func ToSlice[E any](iter Iterator[E]) (result []E, err error) {
	for iter.Next() {
		result = append(result, iter.Get())
	}
	return result, iter.Err()
}

type Combine[BEFORE any, AFTER any] struct {
	Iterators []Iterator[BEFORE]
	Join      func(...BEFORE) AFTER
}

func (c Combine[BEFORE, AFTER]) Next() bool {
	for _, iter := range c.Iterators {
		if !iter.Next() {
			return false
		}
	}
	return true
}

func (c Combine[BEFORE, AFTER]) Get() AFTER {
	var gets []BEFORE
	for _, iter := range c.Iterators {
		gets = append(gets, iter.Get())
	}
	return c.Join(gets...)
}

func (c Combine[BEFORE, AFTER]) Err() error {
	for _, iter := range c.Iterators {
		if err := iter.Err(); err != nil {
			return err
		}
	}
	return nil
}

type Error string

func (e Error) Error() string {
	return string(e)
}

const Stop Error = "Stop Iteration"

// Generate is the function used by Generator to create the next value. The parameter
// is the previous value returned by the iterator. The first call to Generate will be passed
// the zero value of E.
// The returns should have the following valid structures
// - return some_value, nil. Continue iterating, some_value will be the next Get Value
// - return zero_of_E, some_error. An error occurred during Generate, Iteration Error.
// - return zero_of_E, Stop. Iteration is finished.
type Generate[E any] func(E) (E, error)

type Generator[E any] struct {
	Generate Generate[E]

	val E
	err error
}

func (g *Generator[E]) Next() bool {
	if g.err != nil {
		return false
	}
	g.val, g.err = g.Generate(g.val)
	return g.err != nil
}

func (g *Generator[E]) Get() E {
	return g.val
}

func (g *Generator[E]) Err() error {
	if g.err == Stop {
		return nil
	}
	return g.err
}

// Clonable types are able to return a copy of themselves. Usually this copy
// has newly allocated space for all the mutable components.
type Clonable[E any] interface {
	Clone() E
}

// CloneNoOp breaks the convention of the Clonable type by having Clone just return
// the same value. This is to allow convenient use of the Echo iterator without
// needing to implement `Clone`. Warning: Reference types will preserve transformations.
type CloneNoOp[E any] struct {
	Wrap E
}

func (c CloneNoOp[E]) Clone() E {
	return c.Wrap
}

// Echo returns a clone of the same data repeatedly.
type Echo[E any] struct {
	Template Clonable[E]
}

func (e Echo[E]) Next() bool {
	return true
}

func (e Echo[E]) Get() E {
	return e.Template.Clone()
}

func (e Echo[E]) Err() error {
	return nil
}

type Limit[E any] struct {
	Iterator[E]
	Max int

	count int
}

func (l *Limit[E]) Next() bool {
	if l.count >= l.Max {
		return false
	}
	if !l.Iterator.Next() {
		// setting count to Max so that future calls to Next will return false
		// without needing to call Next on the underlying iterator.
		l.count = l.Max
		return false
	}
	l.count++
	return true
}

// Fold uses the provided function to update the result with each element from
// the iterator. Returns the final result.
func Fold[ELEMENT any, RESULT any](iter Iterator[ELEMENT], fold func(ELEMENT, RESULT) (RESULT, error), init RESULT) (RESULT, error) {
	result := init
	for iter.Next() {
		var err error
		result, err = fold(iter.Get(), result)
		if err != nil {
			return result, err
		}
	}
	return result, iter.Err()
}

// Reduce applies the provided functions to each element in the iterator along with an acculator of the same type.
func Reduce[ELEMENT any](iter Iterator[ELEMENT], reduce func(ELEMENT, ELEMENT) (ELEMENT, error)) (ELEMENT, error) {
	var current ELEMENT
	if !iter.Next() {
		return iter.Get(), iter.Err()
	}
	current = iter.Get()
	for iter.Next() {
		var err error
		current, err = reduce(current, iter.Get())
		if err != nil {
			return current, err
		}
	}
	return current, iter.Err()
}

type StopWhen[E any] struct {
	Iterator[E]
	When func(E) bool

	next E
	stop bool
}

func (sw *StopWhen[E]) Next() bool {
	if sw.stop || !sw.Iterator.Next() {
		return false
	}
	sw.next = sw.Iterator.Get()
	if sw.When(sw.next) {
		sw.stop = true
		return false
	}
	return true
}

func (sw *StopWhen[E]) Get() E {
	return sw.next
}
