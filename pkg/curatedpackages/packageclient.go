package curatedpackages

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/templater"
)

const (
	minWidth   = 16
	tabWidth   = 8
	padding    = 0
	padChar    = '\t'
	flags      = 0
	CustomName = "my-"
	kind       = "Package"
)

type PackageClientOpt func(*PackageClient)

type PackageClient struct {
	bundle         *packagesv1.PackageBundle
	customPackages []string
	kubectl        KubectlRunner
	customConfigs  []string
	showOptions    bool
}

func NewPackageClient(kubectl KubectlRunner, options ...PackageClientOpt) *PackageClient {
	pc := &PackageClient{
		kubectl: kubectl,
	}
	for _, o := range options {
		o(pc)
	}
	return pc
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
	packageMap := pc.packageMap()
	var packages []packagesv1.Package
	for _, p := range pc.customPackages {
		bundlePackage, found := packageMap[strings.ToLower(p)]
		if !found {
			return nil, fmt.Errorf("unknown package %q", p)
		}
		name := CustomName + strings.ToLower(bundlePackage.Name)
		configString := pc.getGenerateConfigurations(&bundlePackage)
		packages = append(packages, convertBundlePackageToPackage(bundlePackage, name, pc.bundle.APIVersion, configString))
	}
	return packages, nil
}

func (pc *PackageClient) WritePackagesToStdOut(packages []packagesv1.Package) error {
	var output [][]byte
	for _, p := range packages {
		displayPackage := NewDisplayablePackage(&p)
		content, err := yaml.Marshal(displayPackage)
		if err != nil {
			return fmt.Errorf("unable to parse package %s %v", p.Name, err)
		}
		output = append(output, content)
	}
	fmt.Println(string(templater.AppendYamlResources(output...)))
	return nil
}

func (pc *PackageClient) GetPackageFromBundle(packageName string) (*packagesv1.BundlePackage, error) {
	packageMap := pc.packageMap()
	p, ok := packageMap[strings.ToLower(packageName)]
	if !ok {
		return nil, fmt.Errorf("package %s not found", packageName)
	}
	return &p, nil
}

func (pc *PackageClient) packageMap() map[string]packagesv1.BundlePackage {
	pMap := make(map[string]packagesv1.BundlePackage)
	for _, p := range pc.bundle.Spec.Packages {
		pMap[strings.ToLower(p.Name)] = p
	}
	return pMap
}

func (pc *PackageClient) InstallPackage(ctx context.Context, bp *packagesv1.BundlePackage, customName string, kubeConfig string) error {
	configString, err := pc.getInstallConfigurations(bp)
	if err != nil {
		return err
	}

	p := convertBundlePackageToPackage(*bp, customName, pc.bundle.APIVersion, configString)
	displayPackage := NewDisplayablePackage(&p)
	params := []string{"create", "-f", "-", "--kubeconfig", kubeConfig}
	packageYaml, err := yaml.Marshal(displayPackage)
	if err != nil {
		return err
	}
	stdOut, err := pc.kubectl.CreateFromYaml(ctx, packageYaml, params...)
	if err != nil {
		return err
	}
	fmt.Print(&stdOut)
	return nil
}

func (pc *PackageClient) getInstallConfigurations(bp *packagesv1.BundlePackage) (string, error) {
	installConfigs, err := ParseConfigurations(pc.customConfigs)
	if err != nil {
		return "", err
	}

	configs := GetConfigurationsFromBundle(bp)

	err = UpdateConfigurations(configs, installConfigs)
	if err != nil {
		return "", err
	}

	configString := GenerateAllValidConfigurations(configs)
	return configString, nil
}

func (pc *PackageClient) getGenerateConfigurations(bp *packagesv1.BundlePackage) string {
	configs := GetConfigurationsFromBundle(bp)
	configString := GenerateDefaultConfigurations(configs)
	return configString
}

func (pc *PackageClient) ApplyPackages(ctx context.Context, fileName string, kubeConfig string) error {
	params := []string{"apply", "-f", fileName, "--kubeconfig", kubeConfig}
	stdOut, err := pc.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		fmt.Print(&stdOut)
		return err
	}
	fmt.Print(&stdOut)
	return nil
}

func (pc *PackageClient) CreatePackages(ctx context.Context, fileName string, kubeConfig string) error {
	params := []string{"create", "-f", fileName, "--kubeconfig", kubeConfig}
	stdOut, err := pc.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		fmt.Print(&stdOut)
		return err
	}
	fmt.Print(&stdOut)
	return nil
}

func (pc *PackageClient) DeletePackages(ctx context.Context, packages []string, kubeConfig string) error {
	params := []string{"delete", "packages", "--kubeconfig", kubeConfig, "--namespace", constants.EksaPackagesName}
	params = append(params, packages...)
	stdOut, err := pc.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		fmt.Print(&stdOut)
		return err
	}
	fmt.Print(&stdOut)
	return nil
}

func (pc *PackageClient) DescribePackages(ctx context.Context, packages []string, kubeConfig string) error {
	params := []string{"describe", "packages", "--kubeconfig", kubeConfig, "--namespace", constants.EksaPackagesName}
	params = append(params, packages...)
	stdOut, err := pc.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		fmt.Print(&stdOut)
		return fmt.Errorf("kubectl execution failure: \n%v", err)
	}
	if len(stdOut.Bytes()) == 0 {
		return errors.New("no resources found")
	}
	fmt.Print(&stdOut)
	return nil
}

func convertBundlePackageToPackage(bp packagesv1.BundlePackage, name string, apiVersion string, config string) packagesv1.Package {
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
			Config:      config,
		},
	}
	return p
}

func WithBundle(bundle *packagesv1.PackageBundle) func(*PackageClient) {
	return func(config *PackageClient) {
		config.bundle = bundle
	}
}

func WithCustomPackages(customPackages []string) func(*PackageClient) {
	return func(config *PackageClient) {
		config.customPackages = customPackages
	}
}

func WithCustomConfigs(customConfigs []string) func(*PackageClient) {
	return func(config *PackageClient) {
		config.customConfigs = customConfigs
	}
}

func WithShowOptions(showOptions bool) func(client *PackageClient) {
	return func(config *PackageClient) {
		config.showOptions = showOptions
	}
}
