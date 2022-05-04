package curatedpackages

import (
	"context"
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	ImageRepositoryName = "eks-anywhere-packages-bundles"
)

type Reader interface {
	ReadBundlesForVersion(eksaVersion string) (*releasev1.Bundles, error)
}

type BundleRegistry interface {
	GetRegistryBaseRef(ctx context.Context) (string, error)
}

type BundleReader struct {
	kubeConfig    string
	source        BundleSource
	kubeVersion   string
	kubectl       KubectlRunner
	bundleManager Manager
	cliVersion    version.Info
	registry      BundleRegistry
}

func NewBundleReader(kubeConfig, kubeVersion string, source BundleSource, k KubectlRunner, bm Manager, cli version.Info, reg BundleRegistry) *BundleReader {
	return &BundleReader{
		kubeConfig:    kubeConfig,
		kubeVersion:   kubeVersion,
		source:        source,
		kubectl:       k,
		bundleManager: bm,
		cliVersion:    cli,
		registry:      reg,
	}
}

func (b *BundleReader) GetLatestBundle(ctx context.Context) (*packagesv1.PackageBundle, error) {
	switch b.source {
	case Cluster:
		return b.getActiveBundleFromCluster(ctx)
	case Registry:
		return b.getLatestBundleFromRegistry(ctx)
	default:
		return nil, fmt.Errorf("unknown source: %q", b.source)
	}
}

func (b *BundleReader) getLatestBundleFromRegistry(ctx context.Context) (*packagesv1.PackageBundle, error) {
	registryBaseRef, err := b.registry.GetRegistryBaseRef(ctx)
	if err != nil {
		return nil, err
	}
	return b.bundleManager.LatestBundle(ctx, registryBaseRef)
}

func (b *BundleReader) getActiveBundleFromCluster(ctx context.Context) (*packagesv1.PackageBundle, error) {
	// Active BundleReader is set at the bundle Controller
	bundleController, err := b.GetActiveController(ctx)
	if err != nil {
		return nil, err
	}
	bundle, err := b.getPackageBundle(ctx, bundleController.Spec.ActiveBundle)
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func (b *BundleReader) getPackageBundle(ctx context.Context, activeBundle string) (*packagesv1.PackageBundle, error) {
	params := []string{"get", "packageBundle", "-o", "json", "--kubeconfig", b.kubeConfig, "--namespace", constants.EksaPackagesName, activeBundle}
	stdOut, err := b.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		return nil, err
	}
	obj := &packagesv1.PackageBundle{}
	if err := json.Unmarshal(stdOut.Bytes(), obj); err != nil {
		return nil, fmt.Errorf("unmarshaling package bundle: %w", err)
	}
	return obj, nil
}

func (b *BundleReader) GetActiveController(ctx context.Context) (*packagesv1.PackageBundleController, error) {
	params := []string{"get", "packageBundleController", "-o", "json", "--kubeconfig", b.kubeConfig, "--namespace", constants.EksaPackagesName, bundle.PackageBundleControllerName}
	stdOut, err := b.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		return nil, err
	}
	obj := &packagesv1.PackageBundleController{}
	if err := json.Unmarshal(stdOut.Bytes(), obj); err != nil {
		return nil, fmt.Errorf("unmarshaling active package bundle controller: %w", err)
	}
	return obj, nil
}

func (b *BundleReader) UpgradeBundle(ctx context.Context, controller *packagesv1.PackageBundleController, newBundle string) error {
	controller.Spec.ActiveBundle = newBundle
	controllerYaml, err := yaml.Marshal(controller)
	if err != nil {
		return err
	}
	params := []string{"create", "-f", "-", "--kubeconfig", b.kubeConfig}
	_, err = b.kubectl.CreateFromYaml(ctx, controllerYaml, params...)
	if err != nil {
		return err
	}
	return nil
}
