package curatedpackages_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type packageInstallerTest struct {
	*WithT
	ctx                     context.Context
	kubectlRunner           *mocks.MockKubectlRunner
	packageClient           *mocks.MockPackageHandler
	packageControllerClient *mocks.MockPackageController
	spec                    *cluster.Spec
	command                 *curatedpackages.Installer
	packagePath             string
	kubeConfigPath          string
}

func newPackageInstallerTest(t *testing.T) *packageInstallerTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlRunner(ctrl)
	pc := mocks.NewMockPackageHandler(ctrl)
	pcc := mocks.NewMockPackageController(ctrl)
	packagesPath := "/test/package.yaml"
	spec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &anywherev1.Cluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "test-cluster",
				},
			},
		},
		VersionsBundle: &cluster.VersionsBundle{
			VersionsBundle: &v1alpha1.VersionsBundle{
				PackageController: v1alpha1.PackageBundle{
					HelmChart: v1alpha1.Image{
						URI:  "test_registry/test/eks-anywhere-packages:v1",
						Name: "test_chart",
					},
				},
			},
		},
	}
	kubeConfigPath := kubeconfig.FromClusterName(spec.Cluster.Name)
	return &packageInstallerTest{
		WithT:                   NewWithT(t),
		ctx:                     context.Background(),
		kubectlRunner:           k,
		spec:                    spec,
		packagePath:             packagesPath,
		packageClient:           pc,
		packageControllerClient: pcc,
		kubeConfigPath:          kubeConfigPath,
		command:                 curatedpackages.NewInstaller(k, pc, pcc, spec, packagesPath),
	}
}

func TestPackageInstallerSuccess(t *testing.T) {
	tt := newPackageInstallerTest(t)

	tt.packageClient.EXPECT().CreatePackages(tt.ctx, tt.packagePath, tt.kubeConfigPath).Return(nil)
	tt.kubectlRunner.EXPECT().HasResource(tt.ctx, "crd", "certificates.cert-manager.io", tt.kubeConfigPath, "cert-manager").Return(true, nil)
	tt.packageControllerClient.EXPECT().InstallController(tt.ctx).Return(nil)

	err := tt.command.InstallCuratedPackages(tt.ctx)
	tt.Expect(err).To(BeNil())
}

func TestPackageInstallerFailWhenCertManagerFails(t *testing.T) {
	tt := newPackageInstallerTest(t)

	tt.kubectlRunner.EXPECT().HasResource(tt.ctx, "crd", "certificates.cert-manager.io", tt.kubeConfigPath, "cert-manager").Return(false, nil)

	err := tt.command.InstallCuratedPackages(tt.ctx)
	tt.Expect(err).NotTo(BeNil())
}

func TestPackageInstallerFailWhenControllerFails(t *testing.T) {
	tt := newPackageInstallerTest(t)

	tt.kubectlRunner.EXPECT().HasResource(tt.ctx, "crd", "certificates.cert-manager.io", tt.kubeConfigPath, "cert-manager").Return(true, nil)
	tt.packageControllerClient.EXPECT().InstallController(tt.ctx).Return(errors.New("controller installation failed"))

	err := tt.command.InstallCuratedPackages(tt.ctx)
	tt.Expect(err).NotTo(BeNil())
}

func TestPackageInstallerFailWhenPackageFails(t *testing.T) {
	tt := newPackageInstallerTest(t)

	tt.packageClient.EXPECT().CreatePackages(tt.ctx, tt.packagePath, tt.kubeConfigPath).Return(errors.New("path doesn't exist"))
	tt.kubectlRunner.EXPECT().HasResource(tt.ctx, "crd", "certificates.cert-manager.io", tt.kubeConfigPath, "cert-manager").Return(true, nil)
	tt.packageControllerClient.EXPECT().InstallController(tt.ctx).Return(nil)

	err := tt.command.InstallCuratedPackages(tt.ctx)
	tt.Expect(err).NotTo(BeNil())
}
