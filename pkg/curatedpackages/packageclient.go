package curatedpackages

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

const (
	minWidth = 16
	tabWidth = 8
	padding  = 0
	padChar  = '\t'
	flags    = 0
)

type PackageClient struct {
	packages []packagesv1.BundlePackage
}

func NewPackageClient(packages []packagesv1.BundlePackage) *PackageClient {
	return &PackageClient{
		packages: packages,
	}
}

func (p *PackageClient) DisplayPackages() {
	w := new(tabwriter.Writer)
	defer w.Flush()
	w.Init(os.Stdout, minWidth, tabWidth, padding, padChar, flags)
	fmt.Fprintf(w, "%s\t%s\t \n", "PackageClient", "Version(s)")
	fmt.Fprintf(w, "%s\t%s\t \n", "-------", "----------")
	for _, pkg := range p.packages {
		versions := convertBundleVersionToPackageVersion(pkg.Source.Versions)
		fmt.Fprintf(w, "%s\t%s\t \n", pkg.Name, strings.Join(versions, ","))
	}
}

func convertBundleVersionToPackageVersion(bundleVersions []packagesv1.SourceVersion) []string {
	var versions []string
	for _, v := range bundleVersions {
		versions = append(versions, v.Name)
	}
	return versions
}
