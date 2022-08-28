package curatedpackages

import (
	"strings"
)

const (
	Download        = "download"
	Import          = "import"
	packageLocation = "783794618700.dkr.ecr.us-west-2.amazonaws.com"
)

type ReaderType string

func NewReader(s string) ReaderType {
	return ReaderType(strings.ToLower(strings.TrimSpace(s)))
}

func (r *ReaderType) GetRegistry(uri string) string {
	switch *r {
	case Import:
		return packageLocation
	default:
		return GetRegistry(uri)
	}
}
