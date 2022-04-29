package eksd

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Upgrader struct {
	*Installer
}

func NewUpgrader(client EksdInstallerClient, reader Reader) *Upgrader {
	return &Upgrader{
		NewEksdInstaller(client, reader),
	}
}

func (u *Upgrader) Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	logger.V(1).Info("Checking for EKS-D components upgrade")
	changeDiff := EksdChangeDiff(currentSpec, newSpec)
	if changeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for EKS-D components")
		return nil, nil
	}
	logger.V(1).Info("Starting EKS-D components upgrade")
	if err := u.InstallEksdCRDs(ctx, newSpec, cluster); err != nil {
		return nil, fmt.Errorf("upgrading EKS-D components from version %s to version %s: %v", changeDiff.ComponentReports[0].OldVersion, changeDiff.ComponentReports[0].NewVersion, err)
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
