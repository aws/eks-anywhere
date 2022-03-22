package curatedpackages

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere-packages/pkg/artifacts"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere-packages/pkg/testutil"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
)

const (
	Cluster         = "cluster"
	Registry        = "registry"
	RegistryBaseRef = "public.ecr.aws/q0f6t3x4/eksa-package-bundles"
)

func GetLatestBundle(ctx context.Context, kubeConfig string, source string, kubeVersion string) (*api.PackageBundle, error) {
	var (
		packageBundle *api.PackageBundle
		err           error
	)

	switch strings.ToLower(source) {
	case Cluster:
		packageBundle, err = getActiveBundleFromCluster(ctx, kubeConfig)
	case Registry:
		packageBundle, err = getLatestBundleFromRegistry(ctx, kubeVersion)
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

func getActiveBundleFromCluster(ctx context.Context, kubeConfig string) (*api.PackageBundle, error) {
	params := []executables.KubectlOpt{executables.WithOutput("json"), executables.WithKubeconfig(kubeConfig), executables.WithNamespace(constants.EksaPackagesName)}
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
	activeBundle := bundleController.Spec.ActiveBundle
	params = append(params, executables.WithArg(activeBundle))
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
	if err = json.Unmarshal(stdOut.Bytes(), obj); err != nil {
		return nil, fmt.Errorf("error parsing packageBundle response: %v", err)
	}
	return obj, nil
}

func getActiveController(ctx context.Context, kubectl *executables.Kubectl, opts ...executables.KubectlOpt) (*api.PackageBundleController, error) {
	stdOut, err := kubectl.GetResources(ctx, "packageBundleController", opts...)
	if err != nil {
		return nil, err
	}
	obj := &api.PackageBundleControllerList{}
	if err = json.Unmarshal(stdOut.Bytes(), obj); err != nil {
		return nil, fmt.Errorf("error parsing packageBundleController response: %v", err)
	}
	activeController, err := getActiveBundleController(obj)
	if err != nil {
		return nil, err
	}
	return activeController, nil
}

func getActiveBundleController(bc *api.PackageBundleControllerList) (*api.PackageBundleController, error) {
	for _, v := range bc.Items {
		if v.Status.State == api.BundleControllerStateActive {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("no Active Bundle Controller Found")
}
