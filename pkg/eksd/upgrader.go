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

// UpgraderOpt allows to customize an eksd upgrader on construction.
type UpgraderOpt = InstallerOpt

// NewUpgrader constructs a new eks-d upgrader.
func NewUpgrader(client EksdInstallerClient, reader Reader, opts ...UpgraderOpt) *Upgrader {
	return &Upgrader{
		NewEksdInstaller(client, reader, opts...),
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
	currentVersionsBundle := currentSpec.ControlPlaneVersionsBundle()
	newVersionsBundle := newSpec.ControlPlaneVersionsBundle()
	if currentVersionsBundle.EksD.Name != newVersionsBundle.EksD.Name {
		logger.V(1).Info("EKS-D change diff ", "oldVersion ", currentVersionsBundle.EksD.Name, "newVersion ", newVersionsBundle.EksD.Name)
		return &types.ChangeDiff{
			ComponentReports: []types.ComponentChangeDiff{
				{
					ComponentName: "EKS-D",
					NewVersion:    newVersionsBundle.EksD.Name,
					OldVersion:    currentVersionsBundle.EksD.Name,
				},
			},
		}
	}
	return nil
}
