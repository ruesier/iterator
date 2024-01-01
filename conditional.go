package iterator

func All[E any](iter Iterator[E], predicate func(E) bool) (bool, error) {
	for iter.Next() {
		if !predicate(iter.Get()) {
			return false, nil
		}
	}
	if err := iter.Err(); err != nil {
		return false, err
	}
	return true, nil
}

func Any[E any](iter Iterator[E], predicate func(E) bool) (bool, error) {
	for iter.Next() {
		if predicate(iter.Get()) {
			return true, nil
		}
	}
	return false, iter.Err()
}

type FilterMap[BEFORE any, AFTER any] struct {
	Iterator[BEFORE]

	TestUpdate func(BEFORE) (bool, AFTER)

	next AFTER
}

func (fm FilterMap[BEFORE, AFTER]) Next() bool {
	for fm.Iterator.Next() {
		var shouldReturn bool
		shouldReturn, fm.next = fm.TestUpdate(fm.Iterator.Get())
		if shouldReturn {
			return true
		}
	}
	return false
}

func (fm FilterMap[BEFORE, AFTER]) Get() AFTER {
	return fm.next
}
