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
	Upgrade(ctx context.Context, managementCluster *types.Cluster, newSpec *cluster.Spec, changeReport *CAPIChangeReport) error
}

func (u *Upgrader) Upgrade(ctx context.Context, managementCluster *types.Cluster, currentSpec, newSpec *cluster.Spec) error {
	changeReport := u.capiChangeReport(currentSpec, newSpec)
	if changeReport == nil {
		return nil
	}

	if err := u.capiClient.Upgrade(ctx, managementCluster, newSpec, changeReport); err != nil {
		return fmt.Errorf("failed upgrading ClusterAPI from bundles %d to bundles %d: %v", currentSpec.Bundles.Spec.Number, newSpec.Bundles.Spec.Number, err)
	}

	return nil
}

type CAPIChangeReport struct {
	Core                   *types.ComponentChangeReport
	ControlPlane           *types.ComponentChangeReport
	BootstrapProviders     []types.ComponentChangeReport
	InfrastructureProvider *types.ComponentChangeReport
}

func (u *Upgrader) capiChangeReport(currentSpec, newSpec *cluster.Spec) *CAPIChangeReport {
	// TODO: check version changes for all providers
	return nil
}

func (u *Upgrader) ChangeReport(currentSpec, newSpec *cluster.Spec) *types.ChangeReport {
	u.capiChangeReport(currentSpec, newSpec)
	// TODO: convert from capiChangeReport to generic changeReport
	return nil
}
