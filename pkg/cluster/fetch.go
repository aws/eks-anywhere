package cluster

import (
	"context"
	"fmt"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type BundlesFetch func(ctx context.Context, name, namespace string) (*v1alpha1release.Bundles, error)

type GitOpsFetch func(ctx context.Context, name, namespace string) (*v1alpha1.GitOpsConfig, error)

type EksdReleaseFetch func(ctx context.Context, name, namespace string) (*eksdv1alpha1.Release, error)

func BuildSpecForCluster(ctx context.Context, cluster *v1alpha1.Cluster, bundlesFetch BundlesFetch, eksdReleaseFetch EksdReleaseFetch, gitOpsFetch GitOpsFetch) (*Spec, error) {
	bundles, err := GetBundlesForCluster(ctx, cluster, bundlesFetch)
	if err != nil {
		return nil, err
	}
	gitOpsConfig, err := GetGitOpsForCluster(ctx, cluster, gitOpsFetch)
	if err != nil {
		return nil, err
	}
	eksd, err := GetEksdReleaseForCluster(ctx, cluster, bundles, eksdReleaseFetch)
	if err != nil {
		return nil, err
	}
	return BuildSpecFromBundles(cluster, bundles, eksd, WithGitOpsConfig(gitOpsConfig))
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

func GetEksdReleaseForCluster(ctx context.Context, cluster *v1alpha1.Cluster, bundles *v1alpha1release.Bundles, fetch EksdReleaseFetch) (*eksdv1alpha1.Release, error) {
	versionsBundle, err := getVersionsBundle(cluster, bundles)
	if err != nil {
		return nil, fmt.Errorf("failed fetching versions bundle: %v", err)
	}
	eksd, err := fetch(ctx, versionsBundle.EksD.Name, constants.EksaSystemNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed fetching EKS-D release for cluster: %v", err)
	}

	return eksd, nil
}

func getVersionsBundle(clusterConfig *v1alpha1.Cluster, bundles *v1alpha1release.Bundles) (*v1alpha1release.VersionsBundle, error) {
	for _, versionsBundle := range bundles.Spec.VersionsBundles {
		if versionsBundle.KubeVersion == string(clusterConfig.Spec.KubernetesVersion) {
			return &versionsBundle, nil
		}
	}
	return nil, fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", clusterConfig.Spec.KubernetesVersion, bundles.Spec.Number)
}
