package eksd

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	maxRetries    = 5
	backOffPeriod = 5 * time.Second
)

type EksdInstallerClient interface {
	ApplyKubeSpecFromBytesWithNamespace(ctx context.Context, cluster *types.Cluster, data []byte, namespace string) error
}

type Installer struct {
	client  EksdInstallerClient
	retrier *retrier.Retrier
}

func NewEksdInstaller(client EksdInstallerClient) *Installer {
	return &Installer{
		client:  client,
		retrier: retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
	}
}

func (e *Installer) InstallEksdCRDs(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	eksdComponents, err := clusterSpec.ReadEksdManifests(clusterSpec.VersionsBundle.EksD)
	if err != nil {
		return fmt.Errorf("loading manifest for eksd components: %v", err)
	}

	if err = e.retrier.Retry(
		func() error {
			return e.client.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, eksdComponents.ReleaseCrdContent, constants.EksaSystemNamespace)
		},
	); err != nil {
		return fmt.Errorf("applying eksd release crd: %v", err)
	}

	return nil
}

func (e *Installer) InstallEksdManifest(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	eksdComponents, err := clusterSpec.ReadEksdManifests(clusterSpec.VersionsBundle.EksD)
	if err != nil {
		return fmt.Errorf("loading manifest for eksd components: %v", err)
	}

	logger.V(4).Info("Applying eksd manifest to cluster")
	if err = e.retrier.Retry(
		func() error {
			return e.client.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, eksdComponents.ReleaseManifestContent, constants.EksaSystemNamespace)
		},
	); err != nil {
		return fmt.Errorf("applying eksd release manifest: %v", err)
	}

	return nil
}
