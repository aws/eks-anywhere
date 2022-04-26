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
	lower := strings.ToLower(s)
	switch lower {
	case Cluster, Registry:
		*b = BundleSource(lower)
	default:
		return fmt.Errorf("unknown source: %q", s)
	}
	return nil
}

func (b BundleSource) Type() string {
	return "BundleSource"
}
