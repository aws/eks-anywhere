package kindnetd

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/types"
)

// Upgrader allows to upgrade a kindnetd installation in a EKS-A cluster.
type Upgrader struct {
	client Client
	reader manifests.FileReader
}

// NewUpgrader constructs a new Upgrader.
func NewUpgrader(client Client, reader manifests.FileReader) *Upgrader {
	return &Upgrader{
		client: client,
		reader: reader,
	}
}

// Upgrade configures a kindnetd installation to match the desired state in the cluster Spec.
func (u Upgrader) Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec, namespaces []string) (*types.ChangeDiff, error) {
	diff := kindnetdChangeDiff(currentSpec, newSpec)
	if diff == nil {
		logger.V(1).Info("Nothing to upgrade for Kindnetd")
		return nil, nil
	}

	manifest, err := generateManifest(u.reader, newSpec)
	if err != nil {
		return nil, err
	}

	if err := u.client.ApplyKubeSpecFromBytes(ctx, cluster, manifest); err != nil {
		return nil, fmt.Errorf("failed applying kindnetd manifest during upgrade: %v", err)
	}

	return types.NewChangeDiff(diff), nil
}

func kindnetdChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	currentVersionsBundle := currentSpec.RootVersionsBundle()
	newVersionsBundle := newSpec.RootVersionsBundle()
	if currentVersionsBundle.Kindnetd.Version == newVersionsBundle.Kindnetd.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: "kindnetd",
		OldVersion:    currentVersionsBundle.Kindnetd.Version,
		NewVersion:    newVersionsBundle.Kindnetd.Version,
	}
}

// RunPostControlPlaneUpgradeSetup satisfies the clustermanager.Networking interface.
// It is a noop for kindnetd.
func (u Upgrader) RunPostControlPlaneUpgradeSetup(_ context.Context, _ *types.Cluster) error {
	return nil
}
