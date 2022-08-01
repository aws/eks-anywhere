package cluster

import (
	"context"
	"fmt"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type BundlesFetch func(ctx context.Context, name, namespace string) (*v1alpha1release.Bundles, error)

type GitOpsFetch func(ctx context.Context, name, namespace string) (*v1alpha1.GitOpsConfig, error)

type FluxConfigFetch func(ctx context.Context, name, namespace string) (*v1alpha1.FluxConfig, error)

type EksdReleaseFetch func(ctx context.Context, name, namespace string) (*eksdv1alpha1.Release, error)

type OIDCFetch func(ctx context.Context, name, namespace string) (*v1alpha1.OIDCConfig, error)

// BuildSpec constructs a cluster.Spec for an eks-a cluster by retrieving all
// necessary objects using fetch methods
// This is deprecated in favour of BuildSpec
func BuildSpecForCluster(ctx context.Context, cluster *v1alpha1.Cluster, bundlesFetch BundlesFetch, eksdReleaseFetch EksdReleaseFetch, gitOpsFetch GitOpsFetch, fluxConfigFetch FluxConfigFetch, oidcFetch OIDCFetch) (*Spec, error) {
	bundles, err := GetBundlesForCluster(ctx, cluster, bundlesFetch)
	if err != nil {
		return nil, err
	}

	var fluxConfig *v1alpha1.FluxConfig
	var gitOpsConfig *v1alpha1.GitOpsConfig
	if cluster.Spec.GitOpsRef != nil {
		if cluster.Spec.GitOpsRef.Kind == v1alpha1.FluxConfigKind {
			fluxConfig, err = GetFluxConfigForCluster(ctx, cluster, fluxConfigFetch)
			if err != nil {
				return nil, err
			}
		}

		if cluster.Spec.GitOpsRef.Kind == v1alpha1.GitOpsConfigKind {
			gitOpsConfig, err = GetGitOpsForCluster(ctx, cluster, gitOpsFetch)
			if err != nil {
				return nil, err
			}
			fluxConfig = gitOpsConfig.ConvertToFluxConfig()
		}
	}

	eksd, err := GetEksdReleaseForCluster(ctx, cluster, bundles, eksdReleaseFetch)
	if err != nil {
		return nil, err
	}
	oidcConfig, err := GetOIDCForCluster(ctx, cluster, oidcFetch)
	if err != nil {
		return nil, err
	}
	return BuildSpecFromBundles(cluster, bundles, WithEksdRelease(eksd), WithGitOpsConfig(gitOpsConfig), WithFluxConfig(fluxConfig), WithOIDCConfig(oidcConfig))
}

func GetBundlesForCluster(ctx context.Context, cluster *v1alpha1.Cluster, fetch BundlesFetch) (*v1alpha1release.Bundles, error) {
	name, namespace := bundlesNamespacedKey(cluster)
	bundles, err := fetch(ctx, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("fetching Bundles for cluster: %v", err)
	}

	return bundles, nil
}

func bundlesNamespacedKey(cluster *v1alpha1.Cluster) (name, namespace string) {
	if cluster.Spec.BundlesRef != nil {
		name = cluster.Spec.BundlesRef.Name
		namespace = cluster.Spec.BundlesRef.Namespace
	} else {
		// Handles old clusters that don't contain a reference yet to the Bundles
		// For those clusters, the Bundles was created with the same name as the cluster
		// and in the same namespace
		name = cluster.Name
		namespace = cluster.Namespace
	}

	return name, namespace
}

func GetFluxConfigForCluster(ctx context.Context, cluster *v1alpha1.Cluster, fetch FluxConfigFetch) (*v1alpha1.FluxConfig, error) {
	if fetch == nil || cluster.Spec.GitOpsRef == nil {
		return nil, nil
	}
	fluxConfig, err := fetch(ctx, cluster.Spec.GitOpsRef.Name, cluster.Namespace)
	if err != nil {
		return nil, fmt.Errorf("fetching FluxCOnfig for cluster: %v", err)
	}

	return fluxConfig, nil
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
	versionsBundle, err := GetVersionsBundle(cluster, bundles)
	if err != nil {
		return nil, fmt.Errorf("failed fetching versions bundle: %v", err)
	}
	eksd, err := fetch(ctx, versionsBundle.EksD.Name, constants.EksaSystemNamespace)
	if err != nil {
		logger.V(4).Info("EKS-D release objects cannot be retrieved from the cluster. Fetching EKS-D release manifest from the URL in the bundle")
		return nil, nil
	}

	return eksd, nil
}

func GetVersionsBundle(clusterConfig *v1alpha1.Cluster, bundles *v1alpha1release.Bundles) (*v1alpha1release.VersionsBundle, error) {
	return getVersionsBundleForKubernetesVersion(clusterConfig.Spec.KubernetesVersion, bundles)
}

func getVersionsBundleForKubernetesVersion(kubernetesVersion v1alpha1.KubernetesVersion, bundles *v1alpha1release.Bundles) (*v1alpha1release.VersionsBundle, error) {
	for _, versionsBundle := range bundles.Spec.VersionsBundles {
		if versionsBundle.KubeVersion == string(kubernetesVersion) {
			return &versionsBundle, nil
		}
	}
	return nil, fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", kubernetesVersion, bundles.Spec.Number)
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

// BuildSpec constructs a cluster.Spec for an eks-a cluster by retrieving all
// necessary objects from the cluster using a kubernetes client
func BuildSpec(ctx context.Context, client Client, cluster *v1alpha1.Cluster) (*Spec, error) {
	configBuilder := NewDefaultConfigClientBuilder()
	config, err := configBuilder.Build(ctx, client, cluster)
	if err != nil {
		return nil, err
	}

	bundlesName, bundlesNamespace := bundlesNamespacedKey(cluster)
	bundles := &v1alpha1release.Bundles{}
	if err = client.Get(ctx, bundlesName, bundlesNamespace, bundles); err != nil {
		return nil, err
	}

	versionsBundle, err := GetVersionsBundle(cluster, bundles)
	if err != nil {
		return nil, err
	}

	// Ideally we would use the same namespace as the Bundles, but Bundles can be in any namespace and
	// the eksd release is always in eksa-system
	eksdRelease := &eksdv1alpha1.Release{}
	if err = client.Get(ctx, versionsBundle.EksD.Name, constants.EksaSystemNamespace, eksdRelease); err != nil {
		return nil, err
	}

	spec := NewSpec()
	if err := spec.init(config, bundles, versionsBundle, eksdRelease); err != nil {
		return nil, err
	}

	return spec, nil
}
