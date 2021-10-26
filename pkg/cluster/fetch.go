package cluster

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type BundlesFetch func(ctx context.Context, name, namespace string) (*v1alpha1release.Bundles, error)

func BuildSpecForCluster(ctx context.Context, cluster *v1alpha1.Cluster, fetch BundlesFetch) (*Spec, error) {
	bundles, err := GetBundlesForCluster(ctx, cluster, fetch)
	if err != nil {
		return nil, err
	}

	return BuildSpecFromBundles(cluster, bundles)
}

func GetBundlesForCluster(ctx context.Context, cluster *v1alpha1.Cluster, fetch BundlesFetch) (*v1alpha1release.Bundles, error) {
	bundles, err := fetch(ctx, cluster.Name, cluster.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed fetching Bundles for cluster: %v", err)
	}

	return bundles, nil
}
