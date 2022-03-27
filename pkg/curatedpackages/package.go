package curatedpackages

import (
	"encoding/json"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
	"text/tabwriter"

	api "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

const (
	minWidth   = 16
	tabWidth   = 8
	padding    = 0
	padChar    = '\t'
	flags      = 0
	customName = "my"
	kind       = "Package"
	permission = 0644
)

func DisplayPackages(packages []api.BundlePackage) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, minWidth, tabWidth, padding, padChar, flags)
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
		packages = append(packages, convertBundlePackageToPackage(bundlePackage, bundle.APIVersion))
	}
	return packages, nil
}

func WritePackagesToFile(packages []api.Package, dir string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to get the current working directory: %v", err)
	}
	for _, p := range packages {
		e, err := json.Marshal(p)
		if err != nil {
			err = fmt.Errorf("unable to parse package %s %v", p.Name, err)
			fmt.Println(err.Error())
			continue
		}
		writeToFile(dir, currentDir, p.Name, e)
	}
	return nil
}

func writeToFile(dir string, currentDir string, packageName string, content []byte) {
	pathSep := string(os.PathSeparator)
	dir = strings.TrimSuffix(dir, pathSep)
	dir = strings.TrimPrefix(dir, pathSep)
	directory := currentDir + pathSep + dir
	os.Mkdir(directory, permission)
	file := directory + pathSep + packageName + ".yaml"
	err := os.WriteFile(file, content, permission)
	if err != nil {
		err = fmt.Errorf("%v", err)
		fmt.Println(err.Error())
	}
}

func getPackageNameToPackage(packages []api.BundlePackage) map[string]api.BundlePackage {
	pntop := make(map[string]api.BundlePackage)
	for _, p := range packages {
		pntop[strings.ToLower(p.Name)] = p
	}
	return pntop
}

func convertBundlePackageToPackage(bp api.BundlePackage, apiVersion string) api.Package {
	versionToUse := bp.Source.Versions[0]
	p := api.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name: customName + bp.Name,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       kind,
			APIVersion: apiVersion,
		},
		Spec: api.PackageSpec{
			PackageName:     bp.Name,
			PackageVersion:  versionToUse.Name,
			TargetNamespace: constants.EksaPackagesName,
		},
		Status: api.PackageStatus{
			Source: api.PackageOCISource{
				Registry:   bp.Source.Registry,
				Repository: bp.Source.Repository,
				Digest:     versionToUse.Digest,
			},
		},
	}
	return p
}
