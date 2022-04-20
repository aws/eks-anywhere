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

func (u *Upgrader) Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	logger.V(1).Info("Checking for EKS-A components upgrade")
	if !newSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Skipping EKS-A components upgrade, not a self-managed cluster")
		return nil, nil
	}
	changeDiff := EksaChangeDiff(currentSpec, newSpec)
	if changeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for controller and CRDs")
		return nil, nil
	}
	logger.V(1).Info("Starting EKS-A components upgrade")
	oldVersion := currentSpec.VersionsBundle.Eksa.Version
	newVersion := newSpec.VersionsBundle.Eksa.Version
	if err := u.retrier.installCustomComponents(ctx, newSpec, cluster); err != nil {
		return nil, fmt.Errorf("failed upgrading EKS-A components from version %v to version %v: %v", oldVersion, newVersion, err)
	}

	return changeDiff, nil
}

func EksaChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ChangeDiff {
	if currentSpec.VersionsBundle.Eksa.Version != newSpec.VersionsBundle.Eksa.Version {
		return &types.ChangeDiff{
			ComponentReports: []types.ComponentChangeDiff{
				{
					ComponentName: "EKS-A",
					NewVersion:    newSpec.VersionsBundle.Eksa.Version,
					OldVersion:    currentSpec.VersionsBundle.Eksa.Version,
				},
			},
		}
	}
	return nil
}
