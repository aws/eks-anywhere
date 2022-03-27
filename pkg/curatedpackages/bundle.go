package curatedpackages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere-packages/pkg/testutil"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
)

const (
	RegistryBaseRef             = "public.ecr.aws/q0f6t3x4/eksa-package-bundles"
	PackageBundleControllerName = "eksa-packages-bundle-controller"
)

type BundleSource = string

const (
	Cluster  BundleSource = "cluster"
	Registry              = "registry"
)

func GetLatestBundle(ctx context.Context, kubeConfig string, source BundleSource, kubeVersion string) (*api.PackageBundle, error) {
	switch strings.ToLower(source) {
	case Cluster:
		return getActiveBundleFromCluster(ctx, kubeConfig)
	case Registry:
		return getLatestBundleFromRegistry(ctx, kubeVersion)
	default:
		return nil, errors.New("unknown source")
	}
}

func createBundleManager(kubeVersion string) (manager bundle.Manager) {
	versionSplit := strings.Split(kubeVersion, ".")
	major, minor := versionSplit[0], versionSplit[1]
	log := logr.Discard()
	discovery := testutil.NewFakeDiscovery(major, minor)
	puller := artifacts.NewRegistryPuller()
	return bundle.NewBundleManager(log, discovery, puller)
}

func getLatestBundleFromRegistry(ctx context.Context, kubeVersion string) (*api.PackageBundle, error) {
	bm := createBundleManager(kubeVersion)
	return bm.LatestBundle(ctx, RegistryBaseRef)
}

func getActiveBundleFromCluster(ctx context.Context, kubeConfig string) (*api.PackageBundle, error) {
	params := []executables.KubectlOpt{executables.WithOutput("json"), executables.WithKubeconfig(kubeConfig), executables.WithNamespace(constants.EksaPackagesName), executables.WithArg(PackageBundleControllerName)}
	deps, err := createKubectl(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl

	// Active Bundle is set at the bundle Controller
	bundleController, err := getActiveController(ctx, kubectl, params...)
	if err != nil {
		return nil, err
	}
	params = []executables.KubectlOpt{executables.WithOutput("json"), executables.WithKubeconfig(kubeConfig), executables.WithNamespace(constants.EksaPackagesName), executables.WithArg(bundleController.Spec.ActiveBundle)}
	bundle, err := getPackageBundle(ctx, kubectl, params...)
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func getPackageBundle(ctx context.Context, kubectl *executables.Kubectl, opts ...executables.KubectlOpt) (*api.PackageBundle, error) {
	stdOut, err := kubectl.GetResources(ctx, "packageBundle", opts...)
	if err != nil {
		return nil, err
	}
	obj := &api.PackageBundle{}
	if err = json.NewDecoder(&stdOut).Decode(obj); err != nil {
		return nil, fmt.Errorf("unmarshaling package bundle: %w", err)
	}
	return obj, nil
}

func getActiveController(ctx context.Context, kubectl *executables.Kubectl, opts ...executables.KubectlOpt) (*api.PackageBundleController, error) {
	stdOut, err := kubectl.GetResources(ctx, "packageBundleController", opts...)
	if err != nil {
		return nil, err
	}
	obj := &api.PackageBundleController{}
	if err = json.NewDecoder(&stdOut).Decode(obj); err != nil {
		return nil, fmt.Errorf("unmarshaling active package bundle controller: %w", err)
	}
	return obj, nil
}

func createKubectl(ctx context.Context) (*dependencies.Dependencies, error) {
	return dependencies.NewFactory().
		WithExecutableImage(executables.DefaultEksaImage()).
		WithExecutableBuilder().
		WithKubectl().
		Build(ctx)
}
