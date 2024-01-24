package eksd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
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

// Upgrade checks whether upgrading EKS-D components is necessary and, if so, installs the new EKS-D CRDs.
func (u *Upgrader) Upgrade(ctx context.Context, cluster *types.Cluster, currentManagementComponentsVersionsBundle *releasev1alpha1.VersionsBundle, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	logger.V(1).Info("Checking for EKS-D components upgrade")

	changeDiff := ChangeDiff(currentManagementComponentsVersionsBundle, newSpec.FirstVersionsBundle())
	if changeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for EKS-D components")
		return nil, nil
	}
	logger.V(1).Info("Starting EKS-D components upgrade")
	if err := u.InstallEksdCRDs(ctx, newSpec, cluster); err != nil {
		return nil, fmt.Errorf("upgrading EKS-D components to EKS-A version %s: %v", *newSpec.Cluster.Spec.EksaVersion, err)
	}
	return changeDiff, nil
}

// ChangeDiff generates a version change diff for the EksD component.
func ChangeDiff(currentVersionsBundle, newVersionsBundle *releasev1alpha1.VersionsBundle) *types.ChangeDiff {
	if currentVersionsBundle.EksD.Name != newVersionsBundle.EksD.Name {
		logger.V(1).Info("EKS-D change diff ", "oldVersion ", currentVersionsBundle.EksD.Name, "newVersion ", newVersionsBundle.EksD.Name)
		return &types.ChangeDiff{
			ComponentReports: []types.ComponentChangeDiff{
				{
					ComponentName: "kubernetes",
					NewVersion:    eksdKubernetesVersionTag(newVersionsBundle.EksD),
					OldVersion:    eksdKubernetesVersionTag(currentVersionsBundle.EksD),
				},
			},
		}
	}
	return nil
}

func eksdKubernetesVersionTag(eksd releasev1alpha1.EksDRelease) string {
	releaseNumber := strings.Split(eksd.Name, "eks-")[1]
	return fmt.Sprintf("%s-eks-%s-%s", eksd.KubeVersion, eksd.ReleaseChannel, releaseNumber)
}
