package semver

import "strconv"

// identifier represents a single identifier in a semver version.
// identifiers are dot separated strings or numbers in the pre release
// and build metadata.
type identifier struct {
	intValue uint64
	strValue string
	isNum    bool
}

func newIdentifier(s string) identifier {
	bi := identifier{}
	if num, err := strconv.ParseUint(s, 10, 64); err == nil {
		bi.intValue = num
		bi.isNum = true
	} else {
		bi.strValue = s
		bi.isNum = false
	}
	return bi
}

// compare compares v and o.
// -1 == v is less than o.
// 0 == v is equal to o.
// 1 == v is greater than o.
// 2 == v is different than o (it is not possible to identify if lower or greater).
// Number is considered lower than string.
func (v identifier) compare(o identifier) int {
	if v.isNum && !o.isNum {
		return -1
	}
	if !v.isNum && o.isNum {
		return 1
	}
	if v.isNum && o.isNum { // both are numbers
		switch {
		case v.intValue < o.intValue:
			return -1
		case v.intValue == o.intValue:
			return 0
		default:
			return 1
		}
	} else { // both are strings
		if v.strValue == o.strValue {
			return 0
		}
		// In order to support random identifiers, like commit hashes,
		// we return 2 when the strings are different to signal the
		// identifiers are different but we can't determine the precedence
		return 2
	}
}

type identifiers []identifier

func newIdentifiers(ids []string) identifiers {
	bis := make(identifiers, 0, len(ids))
	for _, id := range ids {
		bis = append(bis, newIdentifier(id))
	}
	return bis
}

// compare compares 2 identifiers v and o.
// -1 == v is less than o.
// 0 == v is equal to o.
// 1 == v is greater than o.
// If everything else is equal the longer identifier is greater.
func (v identifiers) compare(o identifiers) int {
	i := 0
	for ; i < len(v) && i < len(o); i++ {
		comp := v[i].compare(o[i])
		if comp != 0 {
			return comp
		}
	}

	// if everything is equal until now the longer is greater
	if i == len(v) && i == len(o) {
		return 0
	} else if i == len(v) && i < len(o) {
		return -1
	}

	return 1
}
