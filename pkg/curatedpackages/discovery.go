package curatedpackages

import (
	"k8s.io/apimachinery/pkg/version"
)

// Discovery
/**
Implements ServerVersionInterface to provide the Kubernetes client version to be used.
*/
type Discovery struct {
	kubeVersion *KubeVersion
}

type KubeVersion struct {
	major string
	minor string
}

func NewDiscovery(kubeVersion *KubeVersion) *Discovery {
	return &Discovery{
		kubeVersion: kubeVersion,
	}
}

func NewKubeVersion(major, minor string) *KubeVersion {
	return &KubeVersion{
		major: major,
		minor: minor,
	}
}

func (d *Discovery) ServerVersion() (*version.Info, error) {
	v := &version.Info{
		Major: d.kubeVersion.major,
		Minor: d.kubeVersion.minor,
	}
	return v, nil
}
