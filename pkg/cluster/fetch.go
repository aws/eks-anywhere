package cluster

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type BundlesFetch func(ctx context.Context, name, namespace string) (*v1alpha1release.Bundles, error)

type GitOpsFetch func(ctx context.Context, name, namespace string) (*v1alpha1.GitOpsConfig, error)

func BuildSpecForCluster(ctx context.Context, cluster *v1alpha1.Cluster, bundlesFetch BundlesFetch, gitOpsFetch GitOpsFetch) (*Spec, error) {
	bundles, err := GetBundlesForCluster(ctx, cluster, bundlesFetch)
	if err != nil {
		return nil, err
	}
	gitOpsConfig, err := GetGitOpsForCluster(ctx, cluster, gitOpsFetch)
	if err != nil {
		return nil, err
	}
	return BuildSpecFromBundles(cluster, bundles, WithGitOpsConfig(gitOpsConfig))
}

func GetBundlesForCluster(ctx context.Context, cluster *v1alpha1.Cluster, fetch BundlesFetch) (*v1alpha1release.Bundles, error) {
	bundles, err := fetch(ctx, cluster.Name, cluster.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed fetching Bundles for cluster: %v", err)
	}

	return bundles, nil
}

func GetGitOpsForCluster(ctx context.Context, cluster *v1alpha1.Cluster, fetch GitOpsFetch) (*v1alpha1.GitOpsConfig, error) {
	if fetch == nil || cluster.Spec.GitOpsRef == nil {
		return nil, nil
	}
	gitops, err := fetch(ctx, cluster.Spec.GitOpsRef.Name, cluster.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed fetching GitOpsConfig for cluster: %v", err)
	}

	return gitops, nil
}
