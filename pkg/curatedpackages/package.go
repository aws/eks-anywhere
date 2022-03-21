package curatedpackages

import (
	"fmt"
	"os"
	"strings"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"text/tabwriter"
)

const (
	MinWidth = 16
	TabWidth = 8
	Padding  = 0
	PadChar  = '\t'
	flags    = 0
)

func DisplayPackages(m map[string][]string) {
	// initialize tabwriter
	w := new(tabwriter.Writer)

	w.Init(os.Stdout, MinWidth, TabWidth, Padding, PadChar, flags)
	defer w.Flush()

	fmt.Fprintf(w, "\n %s\t%s\t", "Package", "Version(s)")
	fmt.Fprintf(w, "\n %s\t%s\t", "----", "----")
	fmt.Println("Package", "Version(s)")
	for key, values := range m {
		fmt.Fprintf(w, "\n %s\t%s\t", key, strings.Join(values, ","))
	}
}

func GetPackages(bundle *api.PackageBundle) map[string][]string {
	packages := getPackagesFromBundle(bundle)
	return packages
}

func getPackagesFromBundle(bundle *api.PackageBundle) map[string][]string {
	packagesInBundle := make(map[string][]string)
	for _, p := range bundle.Spec.Packages {
		packagesInBundle[p.Name] = append(packagesInBundle[p.Name], convertBundleVersionToPackageVersion(p.Source.Versions)...)
	}
	return packagesInBundle
}

func convertBundleVersionToPackageVersion(bundleVersions []api.SourceVersion) []string {
	var versions []string
	for _, v := range bundleVersions {
		versions = append(versions, v.Name)
	}
	return versions
}
