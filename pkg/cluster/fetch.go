package cluster

import (
	"context"
	"fmt"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	v1alpha1release "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// ManagementComponents bundles the resource definitions of all EKS-A management components.
type ManagementComponents struct {
	EksD                   v1alpha1release.EksDRelease
	CertManager            v1alpha1release.CertManagerBundle
	ClusterAPI             v1alpha1release.CoreClusterAPI
	Bootstrap              v1alpha1release.KubeadmBootstrapBundle
	ControlPlane           v1alpha1release.KubeadmControlPlaneBundle
	VSphere                v1alpha1release.VSphereBundle
	CloudStack             v1alpha1release.CloudStackBundle
	Docker                 v1alpha1release.DockerBundle
	Eksa                   v1alpha1release.EksaBundle
	Flux                   v1alpha1release.FluxBundle
	ExternalEtcdBootstrap  v1alpha1release.EtcdadmBootstrapBundle
	ExternalEtcdController v1alpha1release.EtcdadmControllerBundle
	Tinkerbell             v1alpha1release.TinkerbellBundle
	Snow                   v1alpha1release.SnowBundle
	Nutanix                v1alpha1release.NutanixBundle
}

// ManagementComponentsFromBundles returns ManagementComponents built from a VersionsBundle.
//
// For decoupled component upgrades, the management components can be upgraded to the new EKS-A version
// separately from the Cluster. So, here we have the management components bundles for that new version,
// but, there are still multiple Kubernetes versions to choose from within the bundle to get the
// components information. However, because management component images are the same for every Kubernetes
// version within the same bundle manifest, it's OK to use the first bundle. If there are is differences between
// the management components on this first versions bundle, and the new cluster specs first versions bundle,
// that indicates an upgrade is required. In the future, we might change the bundles API to remove the assumption
// and make this explicit. When that happens, this method will need to change.
func ManagementComponentsFromBundles(bundles *v1alpha1release.Bundles) *ManagementComponents {
	return newManagementComponents(&bundles.Spec.VersionsBundles[0])
}

// newManagementComponents returns a ManagementComponents object built from a VersionsBundle.
func newManagementComponents(vb *v1alpha1release.VersionsBundle) *ManagementComponents {
	return &ManagementComponents{
		EksD:                   vb.EksD,
		CertManager:            vb.CertManager,
		ClusterAPI:             vb.ClusterAPI,
		Bootstrap:              vb.Bootstrap,
		ControlPlane:           vb.ControlPlane,
		VSphere:                vb.VSphere,
		CloudStack:             vb.CloudStack,
		Docker:                 vb.Docker,
		Eksa:                   vb.Eksa,
		Flux:                   vb.Flux,
		ExternalEtcdBootstrap:  vb.ExternalEtcdBootstrap,
		ExternalEtcdController: vb.ExternalEtcdController,
		Tinkerbell:             vb.Tinkerbell,
		Snow:                   vb.Snow,
		Nutanix:                vb.Nutanix,
	}
}

func bundlesNamespacedKey(cluster *v1alpha1.Cluster, release *v1alpha1release.EKSARelease) (name, namespace string) {
	if release != nil && release.Spec.BundlesRef.Name != "" {
		name = release.Spec.BundlesRef.Name
		namespace = release.Spec.BundlesRef.Namespace
	} else if cluster.Spec.BundlesRef != nil {
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

// GetManagementComponents returns the first VersionsBundle from the Bundles object for cluster's the management components version.
func GetManagementComponents(ctx context.Context, client Client, cluster *v1alpha1.Cluster) (*ManagementComponents, error) {
	managementComponentsVersion := cluster.ManagementComponentsVersion()
	if managementComponentsVersion == "" {
		if cluster.Spec.EksaVersion == nil {
			return nil, fmt.Errorf("either management components version or cluster's EksaVersion need to be set")
		}
		managementComponentsVersion = string(*cluster.Spec.EksaVersion)
	}

	eksaReleaseName := v1alpha1release.GenerateEKSAReleaseName(managementComponentsVersion)
	eksaRelease := &v1alpha1release.EKSARelease{}
	err := client.Get(ctx, eksaReleaseName, constants.EksaSystemNamespace, eksaRelease)
	if err != nil {
		return nil, err
	}

	bundles := &v1alpha1release.Bundles{}
	err = client.Get(ctx, eksaRelease.Spec.BundlesRef.Name, eksaRelease.Spec.BundlesRef.Namespace, bundles)
	if err != nil {
		return nil, err
	}

	return ManagementComponentsFromBundles(bundles), nil
}

// GetVersionsBundle gets the VersionsBundle that corresponds to KubernetesVersion.
func GetVersionsBundle(version v1alpha1.KubernetesVersion, bundles *v1alpha1release.Bundles) (*v1alpha1release.VersionsBundle, error) {
	return getVersionsBundleForKubernetesVersion(version, bundles)
}

func getVersionsBundleForKubernetesVersion(kubernetesVersion v1alpha1.KubernetesVersion, bundles *v1alpha1release.Bundles) (*v1alpha1release.VersionsBundle, error) {
	for _, versionsBundle := range bundles.Spec.VersionsBundles {
		if versionsBundle.KubeVersion == string(kubernetesVersion) {
			return &versionsBundle, nil
		}
	}
	return nil, fmt.Errorf("kubernetes version %s is not supported by bundles manifest %d", kubernetesVersion, bundles.Spec.Number)
}

// BuildSpec constructs a cluster.Spec for an eks-a cluster by retrieving all
// necessary objects from the cluster using a kubernetes client.
func BuildSpec(ctx context.Context, client Client, cluster *v1alpha1.Cluster) (*Spec, error) {
	configBuilder := NewDefaultConfigClientBuilder()
	config, err := configBuilder.Build(ctx, client, cluster)
	if err != nil {
		return nil, err
	}

	return BuildSpecFromConfig(ctx, client, config)
}

// BuildSpecFromConfig constructs a cluster.Spec for an eks-a cluster config by retrieving all dependencies objects from the cluster using a kubernetes client.
func BuildSpecFromConfig(ctx context.Context, client Client, config *Config) (*Spec, error) {
	eksaRelease, err := eksaReleaseForCluster(ctx, client, config.Cluster)
	if err != nil {
		return nil, err
	}

	bundles, err := bundlesForEksaRelease(ctx, client, config.Cluster, eksaRelease)
	if err != nil {
		return nil, err
	}

	eksdReleases, err := fetchAllEksdReleases(ctx, client, config.Cluster, bundles)
	if err != nil {
		return nil, err
	}

	return NewSpec(config, bundles, eksdReleases, eksaRelease)
}

func fetchAllEksdReleases(ctx context.Context, client Client, cluster *v1alpha1.Cluster, bundles *v1alpha1release.Bundles) ([]eksdv1alpha1.Release, error) {
	versions := cluster.KubernetesVersions()
	m := make([]eksdv1alpha1.Release, 0, len(versions))
	for _, version := range versions {
		eksd, err := getEksdRelease(ctx, client, version, bundles)
		if err != nil {
			return nil, err
		}
		m = append(m, *eksd)
	}

	return m, nil
}

func getEksdRelease(ctx context.Context, client Client, version v1alpha1.KubernetesVersion, bundles *v1alpha1release.Bundles) (*eksdv1alpha1.Release, error) {
	versionsBundle, err := GetVersionsBundle(version, bundles)
	if err != nil {
		return nil, err
	}

	// Ideally we would use the same namespace as the Bundles, but Bundles can be in any namespace and
	// the eksd release is always in eksa-system
	eksdRelease := &eksdv1alpha1.Release{}
	if err = client.Get(ctx, versionsBundle.EksD.Name, constants.EksaSystemNamespace, eksdRelease); err != nil {
		return nil, err
	}

	return eksdRelease, nil
}

// BundlesForCluster returns a bundles resource for the cluster.
func BundlesForCluster(ctx context.Context, client Client, cluster *v1alpha1.Cluster) (*v1alpha1release.Bundles, error) {
	release, err := eksaReleaseForCluster(ctx, client, cluster)
	if err != nil {
		return nil, err
	}

	return bundlesForEksaRelease(ctx, client, cluster, release)
}

func eksaReleaseForCluster(ctx context.Context, client Client, cluster *v1alpha1.Cluster) (*v1alpha1release.EKSARelease, error) {
	eksaRelease := &v1alpha1release.EKSARelease{}
	if cluster.Spec.BundlesRef == nil {
		if cluster.Spec.EksaVersion == nil {
			return nil, fmt.Errorf("either cluster's EksaVersion or BundlesRef need to be set")
		}
		version := string(*cluster.Spec.EksaVersion)
		eksaReleaseName := v1alpha1release.GenerateEKSAReleaseName(version)
		if err := client.Get(ctx, eksaReleaseName, constants.EksaSystemNamespace, eksaRelease); err != nil {
			return nil, fmt.Errorf("error getting EKSARelease %s", eksaReleaseName)
		}
	}

	return eksaRelease, nil
}

func bundlesForEksaRelease(ctx context.Context, client Client, cluster *v1alpha1.Cluster, eksaRelease *v1alpha1release.EKSARelease) (*v1alpha1release.Bundles, error) {
	bundlesName, bundlesNamespace := bundlesNamespacedKey(cluster, eksaRelease)
	bundles := &v1alpha1release.Bundles{}
	if err := client.Get(ctx, bundlesName, bundlesNamespace, bundles); err != nil {
		return nil, err
	}

	return bundles, nil
}
