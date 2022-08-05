package cilium

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const namespace = constants.KubeSystemNamespace

type Cilium struct {
	*Upgrader
}

func NewCilium(client Client, helm Helm) *Cilium {
	return &Cilium{
		Upgrader: NewUpgrader(client, helm),
	}
}

func (c *Cilium) GenerateManifest(ctx context.Context, clusterSpec *cluster.Spec, providerNamespaces []string) ([]byte, error) {
	return c.templater.GenerateManifest(ctx, clusterSpec, WithPolicyAllowedNamespaces(providerNamespaces))
}
