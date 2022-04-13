package cilium

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/templater"
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
	ciliumManifest, err := c.templater.GenerateManifest(ctx, clusterSpec)
	if err != nil {
		return nil, err
	}

	if clusterSpec.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode != v1alpha1.CiliumPolicyModeAlways {
		return ciliumManifest, nil
	}

	networkPolicyManifest, err := c.templater.GenerateNetworkPolicyManifest(clusterSpec, providerNamespaces)
	if err != nil {
		return nil, err
	}

	return templater.AppendYamlResources(ciliumManifest, networkPolicyManifest), nil
}
