package tinkerbell

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/types"
)

func (p *Provider) SetupAndValidateDeleteCluster(ctx context.Context, cluster *types.Cluster, _ *cluster.Spec) error {
	// noop
	return nil
}

func (p *Provider) PostClusterDeleteValidate(ctx context.Context, managementCluster *types.Cluster) error {
	if err := p.stackInstaller.UninstallLocal(ctx); err != nil {
		return err
	}

	return nil
}
