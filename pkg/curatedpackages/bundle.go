package curatedpackages

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(packagesv1.AddToScheme(scheme))
}

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
	bundleManager Manager
	registry      BundleRegistry
	restClient    rest.Interface
}

func NewBundleReader(kubeConfig string, source BundleSource,
	restClient rest.Interface, bm Manager, reg BundleRegistry,
) *BundleReader {
	return &BundleReader{
		kubeConfig:    kubeConfig,
		source:        source,
		bundleManager: bm,
		registry:      reg,
		restClient:    restClient,
	}
}

func (b *BundleReader) GetLatestBundle(ctx context.Context, kubeVersion string) (*packagesv1.PackageBundle, error) {
	switch b.source {
	case Cluster:
		return b.getActiveBundleFromCluster(ctx)
	case Registry:
		return b.getLatestBundleFromRegistry(ctx, kubeVersion)
	default:
		return nil, fmt.Errorf("unknown source: %q", b.source)
	}
}

func (b *BundleReader) getLatestBundleFromRegistry(ctx context.Context, kubeVersion string) (*packagesv1.PackageBundle, error) {
	registryBaseRef, err := b.registry.GetRegistryBaseRef(ctx)
	if err != nil {
		return nil, err
	}
	return b.bundleManager.LatestBundle(ctx, registryBaseRef, kubeVersion)
}

func (b *BundleReader) getActiveBundleFromCluster(ctx context.Context) (*packagesv1.PackageBundle, error) {
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

func (b *BundleReader) getPackageBundle(ctx context.Context, activeBundle string) (
	*packagesv1.PackageBundle, error,
) {
	pb := &packagesv1.PackageBundle{}
	err := b.restClient.Get().Namespace(constants.EksaPackagesName).
		Resource("packagebundles").Name(activeBundle).Do(ctx).Into(pb)
	if err != nil {
		return nil, fmt.Errorf("getting package bundle: %w", err)
	}

	return pb, nil
}

func (b *BundleReader) GetActiveController(ctx context.Context) (*packagesv1.PackageBundleController, error) {
	// There's some question around how this will handle multiple
	// kubeadmcontrolplanes. This is being tracked in
	// https://github.com/aws/eks-anywhere-packages/issues/425
	kcp := &controlplanev1.KubeadmControlPlane{}
	err := b.restClient.Get().Namespace(constants.EksaSystemNamespace).
		Resource("kubeadmcontrolplane").Do(ctx).Into(kcp)
	if err != nil {
		return nil, fmt.Errorf("getting kubeadm control plane: %w", err)
	}

	pbc := &packagesv1.PackageBundleController{}
	err = b.restClient.Get().Namespace(constants.EksaPackagesName).
		Resource(strings.TrimSpace(kcp.Name)).
		Name(packagesv1.PackageBundleControllerName).Do(ctx).Into(pbc)
	if err != nil {
		return nil, fmt.Errorf("getting package bundle controller: %w", err)
	}

	return pbc, nil
}

func (b *BundleReader) UpgradeBundle(ctx context.Context, controller *packagesv1.PackageBundleController, newBundle string) error {
	controller.Spec.ActiveBundle = newBundle
	err := b.restClient.Put().
		Namespace(constants.EksaPackagesName).
		Resource("packagebundlecontrollers").
		Name(packagesv1.PackageBundleControllerName).
		Body(controller).
		Do(ctx).Error()
	if err != nil {
		return err
	}

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
