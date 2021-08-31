package networking

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Cilium struct{}

func NewCilium() *Cilium {
	return &Cilium{}
}

func (c *Cilium) GenerateManifest(clusterSpec *cluster.Spec) ([]byte, error) {
	return loadManifest(clusterSpec, clusterSpec.VersionsBundle.Cilium.Manifest)
}

func loadManifest(clusterSpec *cluster.Spec, manifest v1alpha1.Manifest) ([]byte, error) {
	m, err := clusterSpec.LoadManifest(manifest)
	if err != nil {
		return nil, fmt.Errorf("can't load networking manifest [%s]: %v", manifest.URI, err)
	}

	return m.Content, nil
}
