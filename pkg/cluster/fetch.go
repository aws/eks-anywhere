package cluster

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type BundlesFetch func(ctx context.Context, name, namespace string) (*v1alpha1release.Bundles, error)

type GitOpsFetch func(ctx context.Context, name, namespace string) (*v1alpha1.GitOpsConfig, error)

type OIDCFetch func(ctx context.Context, name, namespace string) (*v1alpha1.OIDCConfig, error)

func BuildSpecForCluster(ctx context.Context, cluster *v1alpha1.Cluster, bundlesFetch BundlesFetch, gitOpsFetch GitOpsFetch, oidcFetch OIDCFetch) (*Spec, error) {
	bundles, err := GetBundlesForCluster(ctx, cluster, bundlesFetch)
	if err != nil {
		return nil, err
	}
	gitOpsConfig, err := GetGitOpsForCluster(ctx, cluster, gitOpsFetch)
	if err != nil {
		return nil, err
	}
	oidcConfig, err := GetOIDCForCluster(ctx, cluster, oidcFetch)
	if err != nil {
		return nil, err
	}
	return BuildSpecFromBundles(cluster, bundles, WithGitOpsConfig(gitOpsConfig), WithOIDCConfig(oidcConfig))
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

func GetOIDCForCluster(ctx context.Context, cluster *v1alpha1.Cluster, fetch OIDCFetch) (*v1alpha1.OIDCConfig, error) {
	if fetch == nil || cluster.Spec.IdentityProviderRefs == nil {
		return nil, nil
	}

	for _, identityProvider := range cluster.Spec.IdentityProviderRefs {
		if identityProvider.Kind == v1alpha1.OIDCConfigKind {
			oidc, err := fetch(ctx, identityProvider.Name, cluster.Namespace)
			if err != nil {
				return nil, fmt.Errorf("failed fetching OIDCConfig for cluster: %v", err)
			}
			return oidc, nil
		}
	}
	return nil, nil
}
