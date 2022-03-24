package curatedpackages

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

const (
	MinWidth = 16
	TabWidth = 8
	Padding  = 0
	PadChar  = '\t'
	flags    = 0
)

func DisplayPackages(packages []api.BundlePackage) {
	m := packagesToString(packages)

	// initialize tabwriter
	w := new(tabwriter.Writer)

	w.Init(os.Stdout, MinWidth, TabWidth, Padding, PadChar, flags)
	defer w.Flush()

	fmt.Fprintf(w, "\n %s\t%s\t", "Package", "Version(s)")
	fmt.Fprintf(w, "\n %s\t%s\t", "----", "----")
	for key, values := range m {
		fmt.Fprintf(w, "\n %s\t%s\t", key, strings.Join(values, ","))
	}
}

func GetPackages(bundle *api.PackageBundle) []api.BundlePackage {
	packagesInBundle := bundle.Spec.Packages
	return packagesInBundle
}

func packagesToString(packages []api.BundlePackage) map[string][]string {
	ptopv := make(map[string][]string)
	for _, p := range packages {
		ptopv[p.Name] = append(ptopv[p.Name], convertBundleVersionToPackageVersion(p.Source.Versions)...)
	}
	return ptopv
}

func convertBundleVersionToPackageVersion(bundleVersions []api.SourceVersion) []string {
	var versions []string
	for _, v := range bundleVersions {
		versions = append(versions, v.Name)
	}
	return versions
}
