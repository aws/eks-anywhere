package cluster

import (
	"context"
	"fmt"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const eksdReleaseObjDoesNotExist = "the server doesn't have a resource type \"releases\""

type BundlesFetch func(ctx context.Context, name, namespace string) (*v1alpha1release.Bundles, error)

type GitOpsFetch func(ctx context.Context, name, namespace string) (*v1alpha1.GitOpsConfig, error)

type EksdReleaseFetch func(ctx context.Context, name, namespace string) (*eksdv1alpha1.Release, error)

type OIDCFetch func(ctx context.Context, name, namespace string) (*v1alpha1.OIDCConfig, error)

func BuildSpecForCluster(ctx context.Context, cluster *v1alpha1.Cluster, bundlesFetch BundlesFetch, eksdReleaseFetch EksdReleaseFetch, gitOpsFetch GitOpsFetch, oidcFetch OIDCFetch) (*Spec, error) {
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
	oidcConfig, err := GetOIDCForCluster(ctx, cluster, oidcFetch)
	if err != nil {
		return nil, err
	}
	return BuildSpecFromBundles(cluster, bundles, WithEksdRelease(eksd), WithGitOpsConfig(gitOpsConfig), WithOIDCConfig(oidcConfig))
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
	if !cluster.IsSelfManaged() { // We do not apply the EKS-D objects to the managed clusters, so the EKS-D release manifest will be retrieved from the URL in the bundle
		return nil, nil
	}
	versionsBundle, err := GetVersionsBundle(cluster, bundles)
	if err != nil {
		return nil, fmt.Errorf("failed fetching versions bundle: %v", err)
	}
	eksd, err := fetch(ctx, versionsBundle.EksD.Name, constants.EksaSystemNamespace)
	if err != nil {
		if strings.Contains(err.Error(), eksdReleaseObjDoesNotExist) {
			logger.V(4).Info("EKS-D release objects are not present on the cluster. EKS-D release manifest will be retrieved from the URL in the bundle")
			return nil, nil
		} else {
			return nil, fmt.Errorf("failed fetching EKS-D release for cluster: %v", err)
		}
	}

	return eksd, nil
}

func GetVersionsBundle(clusterConfig *v1alpha1.Cluster, bundles *v1alpha1release.Bundles) (*v1alpha1release.VersionsBundle, error) {
	for _, versionsBundle := range bundles.Spec.VersionsBundles {
		if versionsBundle.KubeVersion == string(clusterConfig.Spec.KubernetesVersion) {
			return &versionsBundle, nil
		}
	}
	return nil, fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", clusterConfig.Spec.KubernetesVersion, bundles.Spec.Number)
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
