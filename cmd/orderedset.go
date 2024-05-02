package main

// a set ordered by insertion
type orderedSet[T comparable] struct {
	ordered []T
	set     map[T]struct{}
}

func newOrderedSet[T comparable]() orderedSet[T] {
	return orderedSet[T]{
		set: make(map[T]struct{}),
	}
}

func (os *orderedSet[T]) add(vals ...T) {
	for _, val := range vals {
		_, found := os.set[val]
		if found {
			continue
		}

		os.ordered = append(os.ordered, val)
		os.set[val] = struct{}{}
	}
}

func (os orderedSet[T]) getAll() []T {
	return os.ordered
}

func (os orderedSet[T]) len() int {
	return len(os.ordered)
}
