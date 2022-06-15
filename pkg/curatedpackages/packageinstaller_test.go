package curatedpackages_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	ctx            context.Context
	chartInstaller *mocks.MockChartInstaller
	kubectlRunner  *mocks.MockKubectlRunner
	spec           *cluster.Spec
	command        *curatedpackages.Installer
	packagePath    string
}

func newPackageInstallerTest(t *testing.T) *packageInstallerTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlRunner(ctrl)
	c := mocks.NewMockChartInstaller(ctrl)
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
	return &packageInstallerTest{
		WithT:          NewWithT(t),
		ctx:            context.Background(),
		chartInstaller: c,
		kubectlRunner:  k,
		spec:           spec,
		packagePath:    packagesPath,
		command:        curatedpackages.NewInstaller(c, k, spec, packagesPath),
	}
}

func TestPackageInstallerSuccess(t *testing.T) {
	tt := newPackageInstallerTest(t)

	kubeConfigPath := kubeconfig.FromClusterName(tt.spec.Cluster.Name)
	helmChart := tt.spec.VersionsBundle.PackageController.HelmChart
	params := []string{"create", "-f", tt.packagePath, "--kubeconfig", kubeConfigPath}
	registry := curatedpackages.GetRegistry(helmChart.URI)
	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
	values := []string{sourceRegistry}
	tt.kubectlRunner.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)
	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, helmChart.Name, "oci://"+helmChart.Image(), helmChart.Tag(), kubeConfigPath, values).Return(nil)

	err := tt.command.InstallCuratedPackages(tt.ctx)
	tt.Expect(err).To(BeNil())
}

func TestPackageInstallerFailWhenControllerFails(t *testing.T) {
	tt := newPackageInstallerTest(t)

	kubeConfigPath := kubeconfig.FromClusterName(tt.spec.Cluster.Name)
	helmChart := tt.spec.VersionsBundle.PackageController.HelmChart
	registry := curatedpackages.GetRegistry(helmChart.URI)
	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
	values := []string{sourceRegistry}
	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, helmChart.Name, "oci://"+helmChart.Image(), helmChart.Tag(), kubeConfigPath, values).Return(errors.New("controller installation failed"))

	err := tt.command.InstallCuratedPackages(tt.ctx)
	tt.Expect(err).NotTo(BeNil())
}

func TestPackageInstallerFailWhenPackageFails(t *testing.T) {
	tt := newPackageInstallerTest(t)

	kubeConfigPath := kubeconfig.FromClusterName(tt.spec.Cluster.Name)
	helmChart := tt.spec.VersionsBundle.PackageController.HelmChart
	params := []string{"create", "-f", tt.packagePath, "--kubeconfig", kubeConfigPath}
	registry := curatedpackages.GetRegistry(helmChart.URI)
	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
	values := []string{sourceRegistry}
	tt.kubectlRunner.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, errors.New("path doesn't exist"))
	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, helmChart.Name, "oci://"+helmChart.Image(), helmChart.Tag(), kubeConfigPath, values).Return(nil)

	err := tt.command.InstallCuratedPackages(tt.ctx)
	tt.Expect(err).NotTo(BeNil())
}
