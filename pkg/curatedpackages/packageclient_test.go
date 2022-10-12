package curatedpackages_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
)

type packageTest struct {
	*WithT
	ctx        context.Context
	kubectl    *mocks.MockKubectlRunner
	bundle     *packagesv1.PackageBundle
	command    *curatedpackages.PackageClient
	kubeConfig string
}

func newPackageTest(t *testing.T) *packageTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlRunner(ctrl)
	return &packageTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		bundle: &packagesv1.PackageBundle{
			Spec: packagesv1.PackageBundleSpec{
				Packages: []packagesv1.BundlePackage{
					{
						Name: "harbor-test",
						Source: packagesv1.BundlePackageSource{
							Versions: []packagesv1.SourceVersion{
								{Name: "0.0.1"},
								{Name: "0.0.2"},
							},
						},
					},
					{
						Name: "redis-test",
						Source: packagesv1.BundlePackageSource{
							Versions: []packagesv1.SourceVersion{
								{Name: "0.0.3"},
								{Name: "0.0.4"},
							},
						},
					},
				},
			},
		},
		kubectl:    k,
		kubeConfig: "kubeconfig.kubeconfig",
	}
}

func TestGeneratePackagesSucceed(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	result, err := tt.command.GeneratePackages("billy")

	tt.Expect(err).To(BeNil())
	tt.Expect(result[0].Name).To(Equal(curatedpackages.CustomName + packages[0]))
}

func TestGeneratePackagesFail(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"unknown-package"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	result, err := tt.command.GeneratePackages("billy")
	tt.Expect(err).NotTo(BeNil())
	tt.Expect(result).To(BeNil())
}

func TestGetPackageFromBundleSucceeds(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))
	result, err := tt.command.GetPackageFromBundle(packages[0])

	tt.Expect(err).To(BeNil())
	tt.Expect(result.Name).To(Equal(packages[0]))
}

func TestGetPackageFromBundleFails(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))
	result, err := tt.command.GetPackageFromBundle("nonexisting")

	tt.Expect(err).NotTo(BeNil())
	tt.Expect(result).To(BeNil())
}

func TestInstallPackagesSucceeds(t *testing.T) {
	tt := newPackageTest(t)
	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, gomock.Any(), gomock.Any()).Return(convertJsonToBytes(tt.bundle.Spec.Packages[0]), nil)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	// Suppress output temporarily since it is not needed for testing
	temp := os.Stdout
	os.Stdout = nil // turn it off
	err := tt.command.InstallPackage(tt.ctx, &tt.bundle.Spec.Packages[0], "my-harbor", "billy", "")
	os.Stdout = temp // restore it
	tt.Expect(err).To(BeNil())
}

func TestInstallPackagesFails(t *testing.T) {
	tt := newPackageTest(t)
	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, errors.New("error installing package. Package exists"))
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	err := tt.command.InstallPackage(tt.ctx, &tt.bundle.Spec.Packages[0], "my-harbor", "billy", "")
	tt.Expect(err).To(MatchError(ContainSubstring("error installing package. Package exists")))
}

func TestInstallPackagesFailsWhenInvalidConfigs(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	customConfigs := []string{"test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages), curatedpackages.WithCustomConfigs(customConfigs))

	err := tt.command.InstallPackage(tt.ctx, &tt.bundle.Spec.Packages[0], "my-harbor", "billy", "")
	tt.Expect(err).NotTo(BeNil())
}

func TestApplyPackagesPass(t *testing.T) {
	tt := newPackageTest(t)
	fileName := "test_file.yaml"
	params := []string{"apply", "-f", fileName, "--kubeconfig", tt.kubeConfig}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(convertJsonToBytes(tt.bundle.Spec.Packages[0]), nil)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	err := tt.command.ApplyPackages(tt.ctx, fileName, tt.kubeConfig)
	tt.Expect(err).To(BeNil())
	fmt.Println()
}

func TestApplyPackagesFail(t *testing.T) {
	tt := newPackageTest(t)
	fileName := "non_existing.yaml"
	params := []string{"apply", "-f", fileName, "--kubeconfig", tt.kubeConfig}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, errors.New("file doesn't exist"))
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	err := tt.command.ApplyPackages(tt.ctx, fileName, tt.kubeConfig)
	tt.Expect(err).To(MatchError(ContainSubstring("file doesn't exist")))
}

func TestCreatePackagesPass(t *testing.T) {
	tt := newPackageTest(t)
	fileName := "test_file.yaml"
	params := []string{"create", "-f", fileName, "--kubeconfig", tt.kubeConfig}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(convertJsonToBytes(tt.bundle.Spec.Packages[0]), nil)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	err := tt.command.CreatePackages(tt.ctx, fileName, tt.kubeConfig)
	fmt.Println()
	tt.Expect(err).To(BeNil())
}

func TestCreatePackagesFail(t *testing.T) {
	tt := newPackageTest(t)
	fileName := "non_existing.yaml"
	params := []string{"create", "-f", fileName, "--kubeconfig", tt.kubeConfig}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, errors.New("file doesn't exist"))
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	err := tt.command.CreatePackages(tt.ctx, fileName, tt.kubeConfig)
	tt.Expect(err).To(MatchError(ContainSubstring("file doesn't exist")))
}

func TestDeletePackagesPass(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	args := []string{"harbor-test"}
	params := []string{"delete", "packages", "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName + "-susie"}
	params = append(params, args...)

	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(convertJsonToBytes(tt.bundle.Spec.Packages[0]), nil)

	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))
	err := tt.command.DeletePackages(tt.ctx, args, tt.kubeConfig, "susie")
	fmt.Println()
	tt.Expect(err).To(BeNil())
}

func TestDeletePackagesFail(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	args := []string{"non-working-package"}
	params := []string{"delete", "packages", "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName + "-susie"}
	params = append(params, args...)
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, errors.New("package doesn't exist"))
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	err := tt.command.DeletePackages(tt.ctx, args, tt.kubeConfig, "susie")
	tt.Expect(err).To(MatchError(ContainSubstring("package doesn't exist")))
}

func TestDescribePackagesPass(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	args := []string{"harbor-test"}
	params := []string{"describe", "packages", "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName + "-susie"}
	params = append(params, args...)

	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(convertJsonToBytes(tt.bundle.Spec.Packages[0]), nil)
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	err := tt.command.DescribePackages(tt.ctx, args, tt.kubeConfig, "susie")
	fmt.Println()
	tt.Expect(err).To(BeNil())
}

func TestDescribePackagesFail(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	args := []string{"non-working-package"}
	params := []string{"describe", "packages", "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName + "-susie"}
	params = append(params, args...)

	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, errors.New("package doesn't exist"))
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	err := tt.command.DescribePackages(tt.ctx, args, tt.kubeConfig, "susie")
	tt.Expect(err).To(MatchError(ContainSubstring("package doesn't exist")))
}

func TestDescribePackagesWhenEmptyResources(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	var args []string
	params := []string{"describe", "packages", "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName + "-susie"}
	params = append(params, args...)

	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)
	tt.command = curatedpackages.NewPackageClient(tt.kubectl, curatedpackages.WithBundle(tt.bundle), curatedpackages.WithCustomPackages(packages))

	err := tt.command.DescribePackages(tt.ctx, args, tt.kubeConfig, "susie")
	tt.Expect(err).To(MatchError(ContainSubstring("no resources found")))
}

func TestDisplayPackages(t *testing.T) {
	tt := newPackageTest(t)
	bundle := curatedpackages.WithBundle(tt.bundle)
	pc := curatedpackages.NewPackageClient(nil, bundle)
	buf := &bytes.Buffer{}
	err := pc.DisplayPackages(buf)
	tt.Expect(err).To(BeNil())
	// The expected string needs to have whitespace at the end of the strings,
	// which some editors will remove by default, so it's probably best to use
	// this format, even though it's a little harder to read for humans.
	expected := "Package\t\tVersion(s)\t\n-------\t\t----------\t\nharbor-test\t0.0.1, 0.0.2\t\nredis-test\t0.0.3, 0.0.4\t\n"
	tt.Expect(buf.String()).To(Equal(expected))
}
