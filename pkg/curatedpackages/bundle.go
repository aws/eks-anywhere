package curatedpackages

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
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
		packageBundle, err = getLatestBundleFromCluster(ctx, kubeConfig)
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

func getLatestBundleFromCluster(ctx context.Context, kubeConfig string) (*api.PackageBundle, error) {
	params := []executables.KubectlOpt{executables.WithOutput("json"), executables.WithKubeconfig(kubeConfig), executables.WithNamespace(constants.EksaPackagesName)}
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
	sortBundlesNewestFirst(allBundles)
	return &allBundles[0], nil
}

func sortBundlesNewestFirst(bundles []api.PackageBundle) {
	sortFn := func(i, j int) bool {
		older, err := isBundleOlderThan(bundles[i].Name, bundles[j].Name)
		if err != nil {
			return true
		}
		return !older
	}
	sort.Slice(bundles, sortFn)
}

func isBundleOlderThan(current, candidate string) (bool, error) {
	if current == "" {
		return true, nil
	}

	curK8sVer, err := kubeVersion(current)
	if err != nil {
		return false, fmt.Errorf("parsing current kube version: %s", err)
	}

	newK8sVer, err := kubeVersion(candidate)
	if err != nil {
		return false, fmt.Errorf("parsing candidate kube version: %s", err)
	}

	if curK8sVer < newK8sVer {
		return true, nil
	}

	curBuildNum, err := buildNumber(current)
	if err != nil {
		return false, fmt.Errorf("parsing current build number: %s", err)
	}

	newBuildNum, err := buildNumber(candidate)
	if err != nil {
		return false, fmt.Errorf("parsing candidate build number: %s", err)
	}

	return curBuildNum < newBuildNum, nil
}

var kubeVersionRe = regexp.MustCompile(`^(v[^-]+)-.*$`)

func kubeVersion(name string) (string, error) {
	matches := kubeVersionRe.FindStringSubmatch(name)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("no kubernetes version found in %q", name)
}

func buildNumber(name string) (int, error) {
	matches := bundleNameRe.FindStringSubmatch(name)
	if len(matches) > 1 {
		buildNumber, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, fmt.Errorf("parsing build number: %s", err)
		}

		return buildNumber, nil
	}

	return 0, fmt.Errorf("no build number found in %q", name)
}

var bundleNameRe = regexp.MustCompile(`^.*-(\d+)$`)
