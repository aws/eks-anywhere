package curatedpackages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/semver"
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
	clusterName   string
	kubectl       KubectlRunner
	bundleManager Manager
	registry      BundleRegistry
}

func NewBundleReader(kubeConfig string, clusterName string, k KubectlRunner, bm Manager, reg BundleRegistry) *BundleReader {
	return &BundleReader{
		kubeConfig:    kubeConfig,
		clusterName:   clusterName,
		kubectl:       k,
		bundleManager: bm,
		registry:      reg,
	}
}

func (b *BundleReader) GetLatestBundle(ctx context.Context, kubeVersion string) (*packagesv1.PackageBundle, error) {
	if len(kubeVersion) > 0 {
		return b.getLatestBundleFromRegistry(ctx, kubeVersion)
	}
	return b.getActiveBundleFromCluster(ctx)
}

func (b *BundleReader) getLatestBundleFromRegistry(ctx context.Context, kubeVersion string) (*packagesv1.PackageBundle, error) {
	registryBaseRef, err := b.registry.GetRegistryBaseRef(ctx)
	if err != nil {
		return nil, err
	}

	kubeSemVer, err := semver.New(kubeVersion + ".0")
	if err != nil {
		return nil, err
	}

	return b.bundleManager.LatestBundle(ctx, registryBaseRef, fmt.Sprintf("%d", kubeSemVer.Major), fmt.Sprintf("%d", kubeSemVer.Minor), "")
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

func (b *BundleReader) getPackageBundle(ctx context.Context, bundleName string) (*packagesv1.PackageBundle, error) {
	params := []string{"get", "packageBundle", "-o", "json", "--kubeconfig", b.kubeConfig, "--namespace", constants.EksaPackagesName, bundleName}
	if bundleName == "" {
		return nil, fmt.Errorf("no bundle name specified")
	}
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
	params := []string{"get", "packageBundleController", "-o", "json", "--kubeconfig", b.kubeConfig, "--namespace", constants.EksaPackagesName, b.clusterName}
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

func (b *BundleReader) UpgradeBundle(ctx context.Context, controller *packagesv1.PackageBundleController, newBundleVersion string) error {
	controller.Spec.ActiveBundle = newBundleVersion
	controllerYaml, err := yaml.Marshal(controller)
	if err != nil {
		return err
	}
	params := []string{"apply", "-f", "-", "--kubeconfig", b.kubeConfig}
	stdOut, err := b.kubectl.ExecuteFromYaml(ctx, controllerYaml, params...)
	if err != nil {
		return err
	}
	fmt.Print(&stdOut)
	return nil
}

func GetPackageBundleRef(vb releasev1.VersionsBundle) (string, error) {
	packageController := vb.PackageController
	// Use package controller registry to fetch packageBundles.
	// Format of controller image is: <uri>/<env_type>/<repository_name>
	controllerImage := strings.Split(packageController.Controller.Image(), "/")
	major, minor, err := parseKubeVersion(vb.KubeVersion)
	if err != nil {
		logger.MarkFail("unable to parse kubeversion", "error", err)
		return "", fmt.Errorf("unable to parse kubeversion %s %v", vb.KubeVersion, err)
	}
	latestBundle := fmt.Sprintf("v%s-%s-%s", major, minor, "latest")
	registryBaseRef := fmt.Sprintf("%s/%s/%s:%s", controllerImage[0], controllerImage[1], "eks-anywhere-packages-bundles", latestBundle)
	return registryBaseRef, nil
}
