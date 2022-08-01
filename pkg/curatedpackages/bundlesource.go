package curatedpackages

import (
	"fmt"
	"strings"
)

/**
BundleSource implements Value interface from cobra to provide validation for customer input
*/

type BundleSource string

const (
	Cluster  = "cluster"
	Registry = "registry"
)

func (b BundleSource) String() string {
	return string(b)
}

func (b *BundleSource) Set(s string) error {
	src := BundleSource(strings.ToLower(strings.TrimSpace(s)))
	switch src {
	case Cluster, Registry:
		*b = src
	default:
		return fmt.Errorf("unknown source: %q", s)
	}
	return nil
}

func (b BundleSource) Type() string {
	return "BundleSource"
}
