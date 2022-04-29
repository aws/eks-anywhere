package curatedpackages

import (
	"k8s.io/apimachinery/pkg/version"
)

// Discovery
/**
Implements ServerVersionInterface to provide the Kubernetes client version to be used.
*/
type Discovery struct {
	minor string
	major string
}

func NewDiscovery(major string, minor string) *Discovery {
	return &Discovery{
		minor: minor,
		major: major,
	}
}

func (d *Discovery) ServerVersion() (*version.Info, error) {
	v := &version.Info{
		Major: d.major,
		Minor: d.minor,
	}
	return v, nil
}
