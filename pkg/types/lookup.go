package types

type Lookup map[string]struct{}

func (l Lookup) IsPresent(v string) bool {
	_, present := l[v]
	return present
}

func SliceToLookup(slice []string) Lookup {
	l := make(map[string]struct{}, len(slice))
	for _, e := range slice {
		l[e] = struct{}{}
	}

	return l
}
