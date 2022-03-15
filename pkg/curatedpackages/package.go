package curatedpackages

import (
	"context"
	"fmt"
	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
)

func DisplayPackages(ctx context.Context, m map[string][]string) {
	fmt.Println("Package", "Version(s)")
	for key, values := range m {
		fmt.Print(key)
		fmt.Print(values)
		fmt.Println()
	}
}

func GetPackages(ctx context.Context, bundle *api.PackageBundle, kubeConfig string) (map[string][]string, error) {
	packages := getPackagesFromBundle(ctx, bundle)
	packagesInCluster, err := getExistingPackagesFromCluster(ctx, kubeConfig)
	if err != nil {
		return nil, err
	}
	consolidatePackages(ctx, packages, *packagesInCluster)

	return packages, nil
}

func getPackagesFromBundle(ctx context.Context, bundle *api.PackageBundle) map[string][]string {
	packagesInBundle := make(map[string][]string)
	for _, p := range bundle.Spec.Packages {
		packagesInBundle[p.Name] = append(packagesInBundle[p.Name], convertBundleVersionToVersion(p.Source.Versions)...)
	}
	return packagesInBundle
}

func convertBundleVersionToVersion(bundleVersions []api.SourceVersion) []string {
	var versions []string
	for _, v := range bundleVersions {
		versions = append(versions, v.Name)
	}
	return versions
}

func consolidatePackages(ctx context.Context, packages map[string][]string, packagesInCluster api.PackageList) {
	for _, p := range packagesInCluster.Items {
		packages[p.Name] = append(packages[p.Name], p.Spec.PackageVersion)
	}
}

func getExistingPackagesFromCluster(ctx context.Context, kubeConfig string) (*api.PackageList, error) {
	deps, err := createKubectl(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize executables: %v", err)
	}
	kubectl := deps.Kubectl
	params := []executables.KubectlOpt{executables.WithKubeconfig(kubeConfig), executables.WithNamespace(constants.EksaPackagesName)}
	packageList, err := kubectl.GetPackages(ctx, params...)
	if err != nil {
		return nil, err
	}
	return packageList, nil
}
