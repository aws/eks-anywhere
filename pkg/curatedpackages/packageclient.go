package curatedpackages

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"
	"sigs.k8s.io/yaml"
	"text/tabwriter"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/validations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	minWidth        = 16
	tabWidth        = 8
	padding         = 0
	padChar         = '\t'
	flags           = 0
	customName      = "my-"
	kind            = "Package"
	filePermission  = 0o644
	dirPermission   = 0o755
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

func (p *PackageClient) DisplayPackages() {
	w := new(tabwriter.Writer)
	defer w.Flush()
	w.Init(os.Stdout, minWidth, tabWidth, padding, padChar, flags)
	fmt.Fprintf(w, "%s\t%s\t \n", "Package", "Version(s)")
	fmt.Fprintf(w, "%s\t%s\t \n", "-------", "----------")
	for _, pkg := range p.bundle.Spec.Packages {
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

func (p *PackageClient) GeneratePackages() ([]packagesv1.Package, error) {
	packageNameToPackage := p.getPackageNameToPackage()
	var packages []packagesv1.Package
	for _, v := range p.packages {
		bundlePackage := packageNameToPackage[strings.ToLower(v)]
		if bundlePackage.Name == "" {
			fmt.Println(fmt.Errorf("unknown package %s", v).Error())
			continue
		}
		name := customName + strings.ToLower(bundlePackage.Name)
		packages = append(packages, convertBundlePackageToPackage(bundlePackage, name, p.bundle.APIVersion))
	}
	return packages, nil
}

func (p *PackageClient) WritePackagesToFile(packages []packagesv1.Package, d string) error {
	directory := filepath.Join(d, packageLocation)
	if !validations.FileExists(directory) {
		if err := os.Mkdir(directory, dirPermission); err != nil {
			return fmt.Errorf("unable to create directory %s", directory)
		}
	}

	for _, pack := range packages {
		displayPackage := NewDisplayPackage(pack)
		content, err := yaml.Marshal(displayPackage)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to parse package %s %v", pack.Name, err).Error())
			continue
		}
		err = writeToFile(directory, pack.Name, content)
		if err != nil {
			fmt.Println(err.Error())
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

func (p *PackageClient) getPackageNameToPackage() map[string]packagesv1.BundlePackage {
	pntop := make(map[string]packagesv1.BundlePackage)
	for _, p := range p.bundle.Spec.Packages {
		pntop[strings.ToLower(p.Name)] = p
	}
	return pntop
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
