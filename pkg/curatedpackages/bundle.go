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
)

const RegistryBaseRef = "public.ecr.aws/q0f6t3x4/eksa-package-bundles"

func GetLatestBundle(ctx context.Context, kubeConfig string) (*api.PackageBundle, error) {
	packageBundle, err := GetLatestBundleFromCluster(ctx, kubeConfig)
	if err != nil {
		packageBundle, err = GetLatestBundleFromRegistry(ctx)
	}

	if err != nil {
		return nil, err
	}
	return packageBundle, nil
}

func createBundleManager() (manager bundle.Manager) {
	log := logr.Discard()
	discovery := testutil.NewFakeDiscoveryWithDefaults() // TODO: Implement get kube version
	puller := artifacts.NewRegistryPuller()
	bm := bundle.NewBundleManager(log, discovery, puller)
	return bm
}

func GetLatestBundleFromRegistry(ctx context.Context) (*api.PackageBundle, error) {
	bm := createBundleManager()
	bundle, err := bm.LatestBundle(ctx, RegistryBaseRef)
	return bundle, err
}

func GetLatestBundleFromCluster(ctx context.Context, kubeConfig string) (*api.PackageBundle, error) {
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
