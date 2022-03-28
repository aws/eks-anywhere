package curatedpackages

import (
	"fmt"
	"strings"
)

type BundleSource struct {
	source string
}

const (
	Cluster  = "cluster"
	Registry = "registry"
)

func (b BundleSource) String() string {
	return b.source
}

func (b *BundleSource) Set(s string) error {
	switch strings.ToLower(s) {
	case Cluster, Registry:
		b.source = s
	default:
		return fmt.Errorf("unknown source: %q", s)
	}
	return nil
}

func (b BundleSource) Type() string {
	return "BundleSource"
}
