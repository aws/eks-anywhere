package curatedpackages_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

type packageTest struct {
	*WithT
	ctx     context.Context
	bundle  *packagesv1.PackageBundle
	command *curatedpackages.PackageClient
}

func newPackageTest(t *testing.T) *packageTest {
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
	}
}

func TestGeneratePackagesSucceed(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"harbor-test"}
	tt.command = curatedpackages.NewPackageClient(tt.bundle, packages...)

	result, err := tt.command.GeneratePackages()
	tt.Expect(err).To(BeNil())
	tt.Expect(result[0].Name).To(BeEquivalentTo(curatedpackages.CustomName + packages[0]))
}

func TestGeneratePackagesFail(t *testing.T) {
	tt := newPackageTest(t)
	packages := []string{"unknown-package"}
	tt.command = curatedpackages.NewPackageClient(tt.bundle, packages...)

	result, err := tt.command.GeneratePackages()
	tt.Expect(err).NotTo(BeNil())
	tt.Expect(result).To(BeNil())
}
