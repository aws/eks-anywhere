package collection

// Set is a collection that only contains unique elements.
type Set[T comparable] map[T]struct{}

// NewSet creates an empty Set.
func NewSet[T comparable]() Set[T] {
	return newSet[T](0)
}

// NewSetFrom creates a Set from a list of elements.
func NewSetFrom[T comparable](elements ...T) Set[T] {
	s := NewSet[T]()
	for _, e := range elements {
		s.Add(e)
	}

	return s
}

func newSet[T comparable](size int) Set[T] {
	return make(Set[T], size)
}

// Add stores a new element in the Set if wasn't contained yet.
func (s Set[T]) Add(e T) {
	s[e] = struct{}{}
}

// Delete removes an element from the Set if it existed.
func (s Set[T]) Delete(e T) {
	delete(s, e)
}

// Contains checks if an element is contained in the Set.
func (s Set[T]) Contains(e T) bool {
	_, present := s[e]
	return present
}

// ToSlice generates a new slice with all elements in the Set.
// Order is non deterministic.
func (s Set[T]) ToSlice() []T {
	keys := make([]T, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

// MapSet converts c to a new set. f is used to extract the value for representing each element
// of c.
func MapSet[G any, T comparable](c []G, f func(G) T) Set[T] {
	s := NewSet[T]()
	for _, element := range c {
		s.Add(f(element))
	}

	return s
}
