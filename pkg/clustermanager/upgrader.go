package clustermanager

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Upgrader struct {
	retrier *retrierClient
}

func NewUpgrader(retrier *retrierClient) *Upgrader {
	return &Upgrader{
		retrier: retrier,
	}
}

func (u *Upgrader) Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) error {
	logger.V(1).Info("Checking for EKS-A components upgrade")
	if !newSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Skipping EKS-A components upgrade, not a self-managed cluster")
		return nil
	}
	if !u.isUpgradeNeeded(currentSpec, newSpec) {
		logger.V(1).Info("Nothing to upgrade for controller and CRDs")
		return nil
	}
	logger.V(1).Info("Starting EKS-A components upgrade")
	oldVersion := currentSpec.VersionsBundle.Eksa.Version
	newVersion := newSpec.VersionsBundle.Eksa.Version
	if err := u.retrier.installCustomComponents(ctx, newSpec, cluster); err != nil {
		return fmt.Errorf("failed upgrading EKS-A components from version %v to version %v: %v", oldVersion, newVersion, err)
	}

	return nil
}

func (u *Upgrader) isUpgradeNeeded(currentSpec, newSpec *cluster.Spec) bool {
	return currentSpec.VersionsBundle.Eksa.Version != newSpec.VersionsBundle.Eksa.Version
}
