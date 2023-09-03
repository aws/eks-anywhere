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
		command:                 curatedpackages.NewInstaller(k, pc, pcc, spec, packagesPath, kubeConfigPath),
	}
}

func TestPackageInstallerSuccess(t *testing.T) {
	tt := newPackageInstallerTest(t)

	tt.packageClient.EXPECT().CreatePackages(tt.ctx, tt.packagePath, tt.kubeConfigPath).Return(nil)
	tt.packageControllerClient.EXPECT().Enable(tt.ctx).Return(nil)

	tt.command.InstallCuratedPackages(tt.ctx)
}

func TestPackageInstallerFailWhenControllerFails(t *testing.T) {
	tt := newPackageInstallerTest(t)

	tt.packageControllerClient.EXPECT().Enable(tt.ctx).Return(errors.New("controller installation failed"))

	tt.command.InstallCuratedPackages(tt.ctx)
}

func TestPackageInstallerFailWhenPackageFails(t *testing.T) {
	tt := newPackageInstallerTest(t)

	tt.packageClient.EXPECT().CreatePackages(tt.ctx, tt.packagePath, tt.kubeConfigPath).Return(errors.New("path doesn't exist"))
	tt.packageControllerClient.EXPECT().Enable(tt.ctx).Return(nil)

	tt.command.InstallCuratedPackages(tt.ctx)
}

func TestPackageInstallerDisabled(t *testing.T) {
	tt := newPackageInstallerTest(t)
	tt.spec.Cluster.Spec.Packages = &anywherev1.PackageConfiguration{
		Disable: true,
	}

	tt.command.InstallCuratedPackages(tt.ctx)
}

func TestIsPackageControllerDisabled(t *testing.T) {
	tt := newPackageInstallerTest(t)

	if curatedpackages.IsPackageControllerDisabled(nil) {
		t.Errorf("nil cluster should be enabled")
	}

	if curatedpackages.IsPackageControllerDisabled(tt.spec.Cluster) {
		t.Errorf("nil package controller should be enabled")
	}

	tt.spec.Cluster.Spec.Packages = &anywherev1.PackageConfiguration{
		Disable: false,
	}
	if curatedpackages.IsPackageControllerDisabled(tt.spec.Cluster) {
		t.Errorf("package controller should be enabled")
	}

	tt.spec.Cluster.Spec.Packages = &anywherev1.PackageConfiguration{
		Disable: true,
	}
	if !curatedpackages.IsPackageControllerDisabled(tt.spec.Cluster) {
		t.Errorf("package controller should be disabled")
	}
}
