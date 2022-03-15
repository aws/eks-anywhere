package curatedpackages

import (
	"context"
	"fmt"
	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere-packages/pkg/testutil"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/go-logr/logr"
	"strings"
)

const (
	FromCluster     = "cluster"
	FromRegistry    = "registry"
	RegistryBaseRef = "public.ecr.aws/q0f6t3x4/eksa-package-bundles"
)

func GetLatestBundle(ctx context.Context, kubeConfig string, location string, kubeVersion string) (*api.PackageBundle, error) {
	var (
		packageBundle *api.PackageBundle
		err           error
	)

	switch strings.ToLower(location) {
	case FromCluster:
		packageBundle, err = getLatestBundleFromCluster(ctx, kubeConfig)
	case FromRegistry:
		packageBundle, err = getLatestBundleFromRegistry(ctx, kubeVersion)
	default:
		return nil, fmt.Errorf("unsupported location, Please use the options: %v, or %v", FromCluster, FromRegistry)
	}

	if err != nil {
		return nil, err
	}
	return packageBundle, nil
}

func createBundleManager(kubeVersion string) (manager bundle.Manager) {
	versionSplit := strings.Split(kubeVersion, ".")
	major, minor := versionSplit[0], versionSplit[1]
	log := logr.Discard()
	discovery := testutil.NewFakeDiscovery(major, minor)
	puller := artifacts.NewRegistryPuller()
	bm := bundle.NewBundleManager(log, discovery, puller)
	return bm
}

func getLatestBundleFromRegistry(ctx context.Context, kubeVersion string) (*api.PackageBundle, error) {
	bm := createBundleManager(kubeVersion)
	bundle, err := bm.LatestBundle(ctx, RegistryBaseRef)
	return bundle, err
}

func getLatestBundleFromCluster(ctx context.Context, kubeConfig string) (*api.PackageBundle, error) {
	params := []executables.KubectlOpt{executables.WithKubeconfig(kubeConfig), executables.WithNamespace(constants.EksaPackagesName)}
	deps, err := createKubectl(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl
	bundleList, err := kubectl.GetPackageBundles(ctx, params...)
	if err != nil {
		return nil, err
	}
	allBundles := bundleList.Items
	return &allBundles[0], nil
}
