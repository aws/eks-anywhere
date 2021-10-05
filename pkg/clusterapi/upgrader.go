package clusterapi

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Upgrader struct {
	capiClient capiClient
}

type capiClient interface {
	Upgrade(ctx context.Context, managementCluster *types.Cluster, newSpec *cluster.Spec, changeDiff *CAPIChangeDiff) error
}

func (u *Upgrader) Upgrade(ctx context.Context, managementCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) error {
	changeDiff := u.capiChangeDiff(currentSpec, newSpec)
	if changeDiff == nil {
		return nil
	}

	if err := u.capiClient.Upgrade(ctx, managementCluster, newSpec, changeDiff); err != nil {
		return fmt.Errorf("failed upgrading ClusterAPI from bundles %d to bundles %d: %v", currentSpec.Bundles.Spec.Number, newSpec.Bundles.Spec.Number, err)
	}

	return nil
}

type CAPIChangeDiff struct {
	Core                   *types.ComponentChangeDiff
	ControlPlane           *types.ComponentChangeDiff
	BootstrapProviders     []types.ComponentChangeDiff
	InfrastructureProvider *types.ComponentChangeDiff
}

func (u *Upgrader) capiChangeDiff(currentSpec, newSpec *cluster.Spec) *CAPIChangeDiff {
	// TODO: check version changes for all providers
	return nil
}

func (u *Upgrader) ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ChangeDiff {
	u.capiChangeDiff(currentSpec, newSpec)
	// TODO: convert from capiChangeDiff to generic changeDiff
	return nil
}
