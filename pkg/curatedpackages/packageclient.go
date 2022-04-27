package curatedpackages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/validations"
)

const (
	minWidth        = 16
	tabWidth        = 8
	padding         = 0
	padChar         = '\t'
	flags           = 0
	CustomName      = "my-"
	kind            = "Package"
	filePermission  = 0o600
	dirPermission   = 0o700
	packageLocation = "curated-packages"
)

type PackageClient struct {
	bundle   *packagesv1.PackageBundle
	packages []string
}

func NewPackageClient(bundle *packagesv1.PackageBundle, packages ...string) *PackageClient {
	return &PackageClient{
		bundle:   bundle,
		packages: packages,
	}
}

func (pc *PackageClient) DisplayPackages() {
	w := new(tabwriter.Writer)
	defer w.Flush()
	w.Init(os.Stdout, minWidth, tabWidth, padding, padChar, flags)
	fmt.Fprintf(w, "%s\t%s\t \n", "Package", "Version(s)")
	fmt.Fprintf(w, "%s\t%s\t \n", "-------", "----------")
	for _, pkg := range pc.bundle.Spec.Packages {
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

func (pc *PackageClient) GeneratePackages() ([]packagesv1.Package, error) {
	packageMap := pc.getPackageNameToPackage()
	var packages []packagesv1.Package
	for _, p := range pc.packages {
		bundlePackage, found := packageMap[strings.ToLower(p)]
		if !found {
			fmt.Fprintf(os.Stderr, "unknown package %q\n", p)
			continue
		}
		name := CustomName + strings.ToLower(bundlePackage.Name)
		packages = append(packages, convertBundlePackageToPackage(bundlePackage, name, pc.bundle.APIVersion))
	}
	return packages, nil
}

func (pc *PackageClient) WritePackagesToFile(packages []packagesv1.Package, d string) error {
	directory := filepath.Join(d, packageLocation)
	if !validations.FileExists(directory) {
		if err := os.MkdirAll(directory, dirPermission); err != nil {
			return fmt.Errorf("unable to create directory %q: %w", directory, err)
		}
	}

	for _, p := range packages {
		displayPackage := NewDisplayablePackage(p)
		content, err := yaml.Marshal(displayPackage)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to parse package %s %v", p.Name, err).Error())
			continue
		}
		err = writeToFile(directory, p.Name, content)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue // or return an error to abort.
		}
	}
	return nil
}

func writeToFile(dir string, packageName string, content []byte) error {
	file := filepath.Join(dir, packageName) + ".yaml"
	if err := os.WriteFile(file, content, filePermission); err != nil {
		return fmt.Errorf("unable to write to the file: %s %v", file, err)
	}
	return nil
}

func (pc *PackageClient) getPackageNameToPackage() map[string]packagesv1.BundlePackage {
	pMap := make(map[string]packagesv1.BundlePackage)
	for _, p := range pc.bundle.Spec.Packages {
		pMap[strings.ToLower(p.Name)] = p
	}
	return pMap
}

func convertBundlePackageToPackage(bp packagesv1.BundlePackage, name string, apiVersion string) packagesv1.Package {
	p := packagesv1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaPackagesName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: apiVersion,
		},
		Spec: packagesv1.PackageSpec{
			PackageName: bp.Name,
		},
	}
	return p
}
