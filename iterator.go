// Package iterator defines an interface to represent lazy iteration of values.
// Also includes types and functions for interacting with iterators.
package iterator

// Iterator is the primary interface this package defines. The basic usage of
// an Iterator looks like this.
//
// ```golang
//
//		for iter.Next() {
//			doSomething(iter.Get())
//	 }
//	 if err := iter.Err(); err != nil {
//			panic(err)
//		}
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
	Convert func(BEFORE) AFTER
}

func (m Map[BEFORE, AFTER]) GET() AFTER {
	return m.Convert(m.Iterator.Get())
}

// Slice is a wrapper of a slice that implements the Iterator interface.
type Slice[E any] struct {
	Data []E

	started bool
}

func (s *Slice[E]) Next() bool {
	if len(s.Data) == 0 {
		return false
	}
	if s.started {
		s.Data = s.Data[1:]
		return len(s.Data) > 0
	} else {
		s.started = true
		return true
	}
}

func (s *Slice[E]) Get() E {
	return s.Data[0]
}

func (s *Slice[E]) Err() error {
	return nil
}

// ToSlice reads all the values from an iterator and returns them all. If Iterator
// returns an infinite number of values use GetN instead.
func ToSlice[E any](iter Iterator[E]) (result []E, err error) {
	for iter.Next() {
		result = append(result, iter.Get())
	}
	return result, iter.Err()
}

// GetN returns the next N elements of the iterator. Result may have less than n
// elements if iter finishes before returning n elements.
func GetN[E any](iter Iterator[E], n int) (result []E, err error) {
	for i := 0; i < n && iter.Next(); i++ {
		result = append(result, iter.Get())
	}
	return result, iter.Err()
}

type Combine[BEFORE any, AFTER any] struct {
	Iters []Iterator[BEFORE]
	Join  func(...BEFORE) AFTER
}

func (c Combine[BEFORE, AFTER]) Next() bool {
	for _, iter := range c.Iters {
		if !iter.Next() {
			return false
		}
	}
	return true
}

func (c Combine[BEFORE, AFTER]) Get() AFTER {
	var gets []BEFORE
	for _, iter := range c.Iters {
		gets = append(gets, iter.Get())
	}
	return c.Join(gets...)
}

func (c Combine[BEFORE, AFTER]) Err() error {
	for _, iter := range c.Iters {
		if err := iter.Err(); err != nil {
			return err
		}
	}
	return nil
}
