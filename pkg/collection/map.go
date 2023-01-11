package collection

// MapDiff returns the difference between the two provided maps - the items in s1 that aren't in s2.
func MapDiff(s1, s2 map[string]string) map[string]string {
	diff := map[string]string{}
	for key, val := range s1 {
		v, ok := s2[key]
		if !ok || v != val {
			diff[key] = val
		}
	}
	return diff
}
