package eksd

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	releavev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
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

// Upgrade checks for EKS-D updates, and if there are updates the EKS-D CRDs in the cluster.
func (u *Upgrader) Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec) error {
	logger.V(1).Info("Checking for EKS-D CRD updates")
	changeDiff := ChangeDiff(currentSpec, newSpec)
	if changeDiff == nil {
		logger.V(1).Info("Nothing to update for EKS-D.")
		return nil
	}
	logger.V(1).Info("Updating EKS-D CRDs")
	if err := u.InstallEksdCRDs(ctx, newSpec, cluster); err != nil {
		return fmt.Errorf("updating EKS-D crds from bundles %d to bundles %d: %v", currentSpec.Bundles.Spec.Number, newSpec.Bundles.Spec.Number, err)
	}
	return nil
}

// ChangeDiff returns the change diff between the current and new EKS-D versions.
func ChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ChangeDiff {
	currentVersionsBundle := currentSpec.RootVersionsBundle()
	newVersionsBundle := newSpec.RootVersionsBundle()
	if currentVersionsBundle.EksD.Name != newVersionsBundle.EksD.Name {
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

func eksdKubernetesVersionTag(eksd releavev1alpha1.EksDRelease) string {
	parts := strings.Split(eksd.Name, "-")
	releaseNumber := strings.Split(eksd.Name, "-")[len(parts)-1]
	return fmt.Sprintf("%s-eks-%s-%s", eksd.KubeVersion, eksd.ReleaseChannel, releaseNumber)
}
