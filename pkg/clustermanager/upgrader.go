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

func (u *Upgrader) EksaUpgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
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

func (u *Upgrader) EksdUpgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	logger.V(1).Info("Checking for EKS-D components upgrade")
	if !newSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Skipping EKS-D components upgrade, not a self-managed cluster")
		return nil, nil
	}
	changeDiff := EksdChangeDiff(currentSpec, newSpec)
	if changeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for EKS-D components")
		return nil, nil
	}
	logger.V(1).Info("Starting EKS-D components upgrade")
	oldVersion := currentSpec.VersionsBundle.EksD.Name
	newVersion := newSpec.VersionsBundle.EksD.Name
	if err := u.retrier.installEksdComponents(ctx, newSpec, cluster); err != nil {
		return nil, fmt.Errorf("failed upgrading EKS-D components from version %v to version %v: %v", oldVersion, newVersion, err)
	}
	return changeDiff, nil
}

func EksdChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ChangeDiff {
	if currentSpec.VersionsBundle.EksD.Name != newSpec.VersionsBundle.EksD.Name {
		logger.V(1).Info("EKS-D change diff ", "oldVersion ", currentSpec.VersionsBundle.EksD.Name, "newVersion ", newSpec.VersionsBundle.EksD.Name)
		return &types.ChangeDiff{
			ComponentReports: []types.ComponentChangeDiff{
				{
					ComponentName: "EKS-D",
					NewVersion:    newSpec.VersionsBundle.EksD.Name,
					OldVersion:    currentSpec.VersionsBundle.EksD.Name,
				},
			},
		}
	}
	return nil
}
