package curatedpackages

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/templater"
)

const (
	CustomName = "generated-"
	kind       = "Package"
)

type PackageClientOpt func(*PackageClient)

type PackageClient struct {
	bundle         *packagesv1.PackageBundle
	customPackages []string
	kubectl        KubectlRunner
	customConfigs  []string
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

// sourceWithVersions is a wrapper to help get package versions.
//
// This should be pushed upstream to eks-anywhere-packages, then this
// implementation can be removed.
type sourceWithVersions packagesv1.BundlePackageSource

func (s sourceWithVersions) VersionsSlice() []string {
	versions := []string{}
	for _, ver := range packagesv1.BundlePackageSource(s).Versions {
		versions = append(versions, ver.Name)
	}
	return versions
}

// DisplayPackages pretty-prints a table of available packages.
func (pc *PackageClient) DisplayPackages(w io.Writer) error {
	lines := append([][]string{}, packagesHeaderLines...)
	for _, pkg := range pc.bundle.Spec.Packages {
		versions := sourceWithVersions(pkg.Source).VersionsSlice()
		lines = append(lines, []string{pkg.Name, strings.Join(versions, ", ")})
	}

	tw := newCPTabwriter(w, nil)
	defer tw.Flush()
	return tw.writeTable(lines)
}

// packagesHeaderLines pretties-up a table of curated packages info.
var packagesHeaderLines = [][]string{
	{"Package", "Version(s)"},
	{"-------", "----------"},
}

func (pc *PackageClient) GeneratePackages(clusterName string) ([]packagesv1.Package, error) {
	packageMap := pc.packageMap()
	var packages []packagesv1.Package
	for _, p := range pc.customPackages {
		bundlePackage, found := packageMap[strings.ToLower(p)]
		if !found {
			return nil, fmt.Errorf("unknown package %q", p)
		}
		name := CustomName + strings.ToLower(bundlePackage.Name)
		packages = append(packages, convertBundlePackageToPackage(bundlePackage, name, clusterName, pc.bundle.APIVersion, ""))
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

func (pc *PackageClient) InstallPackage(ctx context.Context, bp *packagesv1.BundlePackage, customName string, clusterName string, kubeConfig string) error {
	configString, err := pc.getInstallConfigurations()
	if err != nil {
		return err
	}

	p := convertBundlePackageToPackage(*bp, customName, clusterName, pc.bundle.APIVersion, configString)
	displayPackage := NewDisplayablePackage(&p)
	params := []string{"create", "-f", "-", "--kubeconfig", kubeConfig}
	packageYaml, err := yaml.Marshal(displayPackage)
	if err != nil {
		return err
	}
	stdOut, err := pc.kubectl.ExecuteFromYaml(ctx, packageYaml, params...)
	if err != nil {
		return err
	}
	fmt.Print(&stdOut)
	return nil
}

func (pc *PackageClient) getInstallConfigurations() (string, error) {
	installConfigs, err := ParseConfigurations(pc.customConfigs)
	if err != nil {
		return "", err
	}
	return GenerateAllValidConfigurations(installConfigs)
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

func (pc *PackageClient) DeletePackages(ctx context.Context, packages []string, kubeConfig string, clusterName string) error {
	params := []string{"delete", "packages", "--kubeconfig", kubeConfig, "--namespace", constants.EksaPackagesName + "-" + clusterName}
	params = append(params, packages...)
	stdOut, err := pc.kubectl.ExecuteCommand(ctx, params...)
	if err != nil {
		fmt.Print(&stdOut)
		return err
	}
	fmt.Print(&stdOut)
	return nil
}

func (pc *PackageClient) DescribePackages(ctx context.Context, packages []string, kubeConfig string, clusterName string) error {
	params := []string{"describe", "packages", "--kubeconfig", kubeConfig, "--namespace", constants.EksaPackagesName + "-" + clusterName}
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

func convertBundlePackageToPackage(bp packagesv1.BundlePackage, name string, clusterName string, apiVersion string, config string) packagesv1.Package {
	p := packagesv1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaPackagesName + "-" + clusterName,
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
