package cmd

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// getImages returns all the images in the Bundle for the cluster kubernetes versions.
// This is deprecated. It builds a file reader in line, prefer using the dependency factory.
func getImages(clusterSpecPath, bundlesOverride string) ([]v1alpha1.Image, error) {
	var specOpts []cluster.FileSpecBuilderOpt
	if bundlesOverride != "" {
		specOpts = append(specOpts, cluster.WithOverrideBundlesManifest(bundlesOverride))
	}
	cliVersion := version.Get()
	spec, err := readClusterSpec(clusterSpecPath, cliVersion, specOpts...)
	if err != nil {
		return nil, err
	}

	kubeVersions := spec.Cluster.KubernetesVersions()
	kubeVersionsFilter := make([]string, 0, len(kubeVersions))
	for _, version := range kubeVersions {
		kubeVersionsFilter = append(kubeVersionsFilter, string(version))
	}

	return bundles.ReadImages(
		files.NewReader(files.WithEKSAUserAgent("cli", cliVersion.GitVersion)),
		spec.Bundles,
		kubeVersionsFilter...,
	)
}

// getKubeconfigPath returns an EKS-A kubeconfig path. The return van be overriden using override
// to give preference to a user specified kubeconfig.
func getKubeconfigPath(clusterName, override string) string {
	if override == "" {
		return kubeconfig.FromClusterName(clusterName)
	}
	return override
}

func NewDependenciesForPackages(ctx context.Context, opts ...PackageOpt) (*dependencies.Dependencies, error) {
	config := New(opts...)
	return dependencies.NewFactory().
		WithExecutableMountDirs(config.mountPaths...).
		WithCustomBundles(config.bundlesOverride).
		WithExecutableBuilder().
		WithManifestReader().
		WithKubectl().
		WithHelm(helm.WithInsecure()).
		WithCuratedPackagesRegistry(config.registryName, config.kubeVersion, version.Get()).
		WithPackageControllerClient(config.spec, config.kubeConfig).
		WithLogger().
		Build(ctx)
}

type PackageOpt func(*PackageConfig)

type PackageConfig struct {
	registryName    string
	kubeVersion     string
	kubeConfig      string
	mountPaths      []string
	spec            *cluster.Spec
	bundlesOverride string
}

func New(options ...PackageOpt) *PackageConfig {
	pc := &PackageConfig{}
	for _, o := range options {
		o(pc)
	}
	return pc
}

func WithRegistryName(registryName string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.registryName = registryName
	}
}

func WithKubeVersion(kubeVersion string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.kubeVersion = kubeVersion
	}
}

func WithMountPaths(mountPaths ...string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.mountPaths = mountPaths
	}
}

func WithClusterSpec(spec *cluster.Spec) func(config *PackageConfig) {
	return func(config *PackageConfig) {
		config.spec = spec
	}
}

func WithKubeConfig(kubeConfig string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.kubeConfig = kubeConfig
	}
}

// WithBundlesOverride sets bundlesOverride in the config with incoming value.
func WithBundlesOverride(bundlesOverride string) func(*PackageConfig) {
	return func(config *PackageConfig) {
		config.bundlesOverride = bundlesOverride
	}
}
