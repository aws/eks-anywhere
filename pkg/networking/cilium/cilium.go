package cilium

import (
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	networking "github.com/aws/eks-anywhere/pkg/networking/internal"
)

const namespace = constants.KubeSystemNamespace

type Cilium struct{}

func NewCilium() *Cilium {
	return &Cilium{}
}

func (c *Cilium) GenerateManifest(clusterSpec *cluster.Spec) ([]byte, error) {
	return networking.LoadManifest(clusterSpec, clusterSpec.VersionsBundle.Cilium.Manifest)
}
