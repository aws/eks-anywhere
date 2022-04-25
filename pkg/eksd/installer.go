package eksd

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	filereader "github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests/eksd"
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
	reader  *filereader.Reader
}

func NewEksdInstaller(client EksdInstallerClient) *Installer {
	return &Installer{
		client:  client,
		retrier: retrier.NewWithMaxRetries(maxRetries, backOffPeriod),
		reader:  filereader.NewReader(),
	}
}

func (i *Installer) InstallEksdCRDs(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	eksdCRDs, err := eksd.ReadManifestContent(i.reader, clusterSpec.VersionsBundle.EksD.Components)
	if err != nil {
		return fmt.Errorf("loading manifest for eksd components: %v", err)
	}

	if err = i.retrier.Retry(
		func() error {
			return i.client.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, eksdCRDs, constants.EksaSystemNamespace)
		},
	); err != nil {
		return fmt.Errorf("applying eksd release crd: %v", err)
	}

	return nil
}

func (i *Installer) InstallEksdManifest(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	eksdReleaseManifest, err := eksd.ReadManifestContent(i.reader, clusterSpec.VersionsBundle.EksD.EksDReleaseUrl)
	if err != nil {
		return fmt.Errorf("loading manifest for eksd components: %v", err)
	}

	logger.V(4).Info("Applying eksd manifest to cluster")
	if err = i.retrier.Retry(
		func() error {
			return i.client.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, eksdReleaseManifest, constants.EksaSystemNamespace)
		},
	); err != nil {
		return fmt.Errorf("applying eksd release manifest: %v", err)
	}

	return nil
}
