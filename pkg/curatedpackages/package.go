package curatedpackages

import (
	"context"
	"fmt"
	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

func DisplayPackages(m map[string][]string) {
	fmt.Println("Package", "Version(s)")
	for key, values := range m {
		fmt.Print(key)
		fmt.Print(values)
		fmt.Println()
	}
}

func GetPackages(ctx context.Context, bundle *api.PackageBundle) (map[string][]string, error) {
	packages := getPackagesFromBundle(ctx, bundle)
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
