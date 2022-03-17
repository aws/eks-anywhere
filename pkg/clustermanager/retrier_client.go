package clustermanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager/internal"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type retrierClient struct {
	*client
	*retrier.Retrier
}

func NewRetrierClient(client *client, retrier *retrier.Retrier) *retrierClient {
	return &retrierClient{
		client:  client,
		Retrier: retrier,
	}
}

func (c *retrierClient) installEksdComponents(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	eksdComponents, err := clusterSpec.ReadEksdManifests(clusterSpec.VersionsBundle.EksD)
	if err != nil {
		return fmt.Errorf("failed loading manifest for eksd components: %v", err)
	}

	if err = c.Retry(
		func() error {
			return c.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, eksdComponents.ReleaseCrdContent, constants.EksaSystemNamespace)
		},
	); err != nil {
		return fmt.Errorf("error applying eksd release crd: %v", err)
	}

	if err = c.Retry(
		func() error {
			return c.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, eksdComponents.ReleaseManifestContent, constants.EksaSystemNamespace)
		},
	); err != nil {
		return fmt.Errorf("error applying eksd release manifest: %v", err)
	}

	return nil
}

func (c *retrierClient) installCustomComponents(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster) error {
	componentsManifest, err := clusterSpec.LoadManifest(clusterSpec.VersionsBundle.Eksa.Components)
	if err != nil {
		return fmt.Errorf("failed loading manifest for eksa components: %v", err)
	}

	err = c.Retry(
		func() error {
			return c.ApplyKubeSpecFromBytes(ctx, cluster, componentsManifest.Content)
		},
	)
	if err != nil {
		return fmt.Errorf("error applying eks-a components spec: %v", err)
	}

	// inject proxy env vars the eksa-controller-manager deployment if proxy is configured
	if clusterSpec.Cluster.Spec.ProxyConfiguration != nil {
		noProxyList := append(clusterSpec.Cluster.Spec.ProxyConfiguration.NoProxy, clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks...)
		noProxyList = append(noProxyList, clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks...)
		envMap := map[string]string{
			"HTTP_PROXY":  clusterSpec.Cluster.Spec.ProxyConfiguration.HttpProxy,
			"HTTPS_PROXY": clusterSpec.Cluster.Spec.ProxyConfiguration.HttpsProxy,
			"NO_PROXY":    strings.Join(noProxyList[:], ","),
		}
		err = c.Retrier.Retry(
			func() error {
				return c.UpdateEnvironmentVariablesInNamespace(ctx, "deployment", "eksa-controller-manager", envMap, cluster, "eksa-system")
			},
		)
		if err != nil {
			return fmt.Errorf("error applying eks-a components spec: %v", err)
		}
	}
	return c.waitForDeployments(ctx, internal.EksaDeployments, cluster)
}
