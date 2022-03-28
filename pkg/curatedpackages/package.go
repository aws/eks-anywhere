package curatedpackages

import (
	"fmt"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"text/tabwriter"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	MinWidth        = 16
	TabWidth        = 8
	Padding         = 0
	PadChar         = '\t'
	flags           = 0
	CustomName      = "my-"
	Kind            = "Package"
	FilePermission  = 0644
	DirPermission   = 0755
	PackageLocation = "curated-packages"
)

func DisplayPackages(packages []api.BundlePackage) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, MinWidth, TabWidth, Padding, PadChar, flags)
	defer w.Flush()
	fmt.Fprintf(w, "\n %s\t%s\t", "Package", "Version(s)")
	fmt.Fprintf(w, "\n %s\t%s\t", "----", "----")
	for _, pkg := range packages {
		versions := convertBundleVersionToPackageVersion(pkg.Source.Versions)
		fmt.Fprintf(w, "\n %s\t%s\t", pkg.Name, strings.Join(versions, ","))
	}
}

func convertBundleVersionToPackageVersion(bundleVersions []api.SourceVersion) []string {
	var versions []string
	for _, v := range bundleVersions {
		versions = append(versions, v.Name)
	}
	return versions
}

func GeneratePackages(bundle *api.PackageBundle, args []string) ([]api.Package, error) {
	packagesInBundle := bundle.Spec.Packages
	packageNameToPackage := getPackageNameToPackage(packagesInBundle)
	var packages []api.Package
	for _, v := range args {
		bundlePackage := packageNameToPackage[strings.ToLower(v)]
		if bundlePackage == nil {
			fmt.Println(fmt.Errorf("unknown package %s", v).Error())
			continue
		}
		packages = append(packages, convertBundlePackageToPackage(bundlePackage, bundle.APIVersion))
	}
	return packages, nil
}

func WritePackagesToFile(packages []api.Package, d string) error {
	directory := filepath.Join(d, PackageLocation)
	if err := os.Mkdir(directory, DirPermission); err != nil {
		return fmt.Errorf("unable to create directory %s", directory)
	}

	for _, p := range packages {
		e, err := yaml.Marshal(p)
		if err != nil {
			fmt.Println(fmt.Errorf("unable to parse package %s %v", p.Name, err).Error())
			continue
		}
		writeToFile(directory, p.Name, e)
	}
	return nil
}

func writeToFile(dir string, packageName string, content []byte) {
	file := filepath.Join(dir, packageName) + ".yaml"
	err := os.WriteFile(file, content, FilePermission)
	if err != nil {
		err = fmt.Errorf("%v", err)
		fmt.Println(err.Error())
	}
}

func getPackageNameToPackage(packages []api.BundlePackage) map[string]*api.BundlePackage {
	pntop := make(map[string]*api.BundlePackage)
	for _, p := range packages {
		pntop[strings.ToLower(p.Name)] = &p
	}
	return pntop
}

func convertBundlePackageToPackage(bp *api.BundlePackage, apiVersion string) api.Package {
	versionToUse := bp.Source.Versions[0]
	p := api.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name: CustomName + strings.ToLower(bp.Name),
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       Kind,
			APIVersion: apiVersion,
		},
		Spec: api.PackageSpec{
			PackageName:     bp.Name,
			PackageVersion:  versionToUse.Name,
			TargetNamespace: constants.EksaPackagesName,
		},
	}
	return p
}
