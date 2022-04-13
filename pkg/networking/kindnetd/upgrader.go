package kindnetd

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Client interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
}

type Upgrader struct {
	client Client
}

func NewUpgrader(client Client) *Upgrader {
	return &Upgrader{
		client: client,
	}
}

func (u Upgrader) Upgrade(ctx context.Context, cluster *types.Cluster, currentSpec, newSpec *cluster.Spec, namespaces []string) (*types.ChangeDiff, error) {
	diff := kindnetdChangeDiff(currentSpec, newSpec)
	if diff == nil {
		logger.V(1).Info("Nothing to upgrade for Kindnetd")
		return nil, nil
	}

	manifest, err := generateManifest(newSpec)
	if err != nil {
		return nil, err
	}

	if err := u.client.ApplyKubeSpecFromBytes(ctx, cluster, manifest); err != nil {
		return nil, fmt.Errorf("failed applying kindnetd manifest during upgrade: %v", err)
	}

	return types.NewChangeDiff(diff), nil
}

func kindnetdChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	if currentSpec.VersionsBundle.Kindnetd.Version == newSpec.VersionsBundle.Kindnetd.Version {
		return nil
	}

	return &types.ComponentChangeDiff{
		ComponentName: "kindnetd",
		OldVersion:    currentSpec.VersionsBundle.Kindnetd.Version,
		NewVersion:    newSpec.VersionsBundle.Kindnetd.Version,
	}
}
