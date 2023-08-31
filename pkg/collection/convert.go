package collection

// ToMap is a utility function that converts the slice s to a map using key to retrieve map
// keys.
func ToMap[K comparable, V any](s []V, key func(V) K) map[K]V {
	m := make(map[K]V)
	for _, elem := range s {
		m[key(elem)] = elem
	}
	return m
}

// ToSlice is a utility function that converts the map m to a slice of m's values. The returned
// slices order is undefined.
func ToSlice[K comparable, V any](m map[K]V) []V {
	s := make([]V, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}
