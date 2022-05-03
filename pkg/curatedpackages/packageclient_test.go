package curatedpackages_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
)

type packageTest struct {
	*WithT
	ctx     context.Context
	kubectl *mocks.MockKubectlRunner
	bundle  *packagesv1.PackageBundle
	command *curatedpackages.PackageClient
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
					},
					{
						Name: "redis-test",
					},
				},
			},
		},
		kubectl: k,
	}
}

func TestGeneratePackagesSucceed(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.bundle, tt.kubectl, packages...)

	result, err := tt.command.GeneratePackages()
	tt.Expect(err).To(BeNil())
	tt.Expect(result[0].Name).To(BeEquivalentTo(curatedpackages.CustomName + packages[0]))
}

func TestGeneratePackagesFail(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"unknown-package"}
	tt.command = curatedpackages.NewPackageClient(tt.bundle, tt.kubectl, packages...)

	result, err := tt.command.GeneratePackages()
	tt.Expect(err).NotTo(BeNil())
	tt.Expect(result).To(BeNil())
}

func TestGetPackageFromBundleSucceeds(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.bundle, tt.kubectl, packages...)
	result, err := tt.command.GetPackageFromBundle(packages[0])

	tt.Expect(err).To(BeNil())
	tt.Expect(result.Name).To(BeEquivalentTo(packages[0]))
}

func TestGetPackageFromBundleFails(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.bundle, tt.kubectl, packages...)
	result, err := tt.command.GetPackageFromBundle("nonexisting")

	tt.Expect(err).NotTo(BeNil())
	tt.Expect(result).To(BeNil())
}

func TestInstallPackagesSucceeds(t *testing.T) {
	tt := newPackageTest(t)
	tt.kubectl.EXPECT().CreateFromYaml(tt.ctx, gomock.Any(), gomock.Any()).Return(convertJsonToBytes(tt.bundle.Spec.Packages[0]), nil)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.bundle, tt.kubectl, packages...)

	err := tt.command.InstallPackage(tt.ctx, &tt.bundle.Spec.Packages[0], "my-harbor", "")
	tt.Expect(err).To(BeNil())
}

func TestInstallPackagesFails(t *testing.T) {
	tt := newPackageTest(t)
	tt.kubectl.EXPECT().CreateFromYaml(tt.ctx, gomock.Any(), gomock.Any()).Return(bytes.Buffer{}, errors.New("error installing package. Package exists"))
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.bundle, tt.kubectl, packages...)

	err := tt.command.InstallPackage(tt.ctx, &tt.bundle.Spec.Packages[0], "my-harbor", "")
	tt.Expect(err).To(MatchError(ContainSubstring("error installing package. Package exists")))
}
