package cmd

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/version"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func getImages(spec string) ([]v1alpha1.Image, error) {
	clusterSpec, err := cluster.NewSpecFromClusterConfig(spec, version.Get())
	if err != nil {
		return nil, err
	}
	bundle := clusterSpec.VersionsBundle
	images := append(bundle.Images(), clusterSpec.KubeDistroImages()...)
	return images, nil
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
		WithExecutableBuilder().
		WithManifestReader().
		WithKubectl().
		WithHelm().
		WithCuratedPackagesRegistry(config.registryName, config.kubeVersion, version.Get()).
		Build(ctx)
}

type PackageOpt func(*PackageConfig)

type PackageConfig struct {
	registryName string
	kubeVersion  string
	mountPaths   []string
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
