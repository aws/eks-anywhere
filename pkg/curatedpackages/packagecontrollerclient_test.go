package curatedpackages_test

import (
	"bytes"
	"context"
	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/constants"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
)

type packageControllerTest struct {
	*WithT
	ctx            context.Context
	kubectl        *mocks.MockKubectlRunner
	chartInstaller *mocks.MockChartInstaller
	command        *curatedpackages.PackageControllerClient
	kubeConfig     string
	ociUri         string
	chartName      string
	chartVersion   string
}

func newPackageControllerTest(t *testing.T) *packageControllerTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlRunner(ctrl)
	ci := mocks.NewMockChartInstaller(ctrl)
	kubeConfig := "kubeconfig.kubeconfig"
	uri := "test/registry_name"
	chartName := "test_controller"
	chartVersion := "v1.0.0"
	return &packageControllerTest{
		WithT:          NewWithT(t),
		ctx:            context.Background(),
		kubectl:        k,
		chartInstaller: ci,
		command:        curatedpackages.NewPackageControllerClient(ci, k, kubeConfig, uri, chartName, chartVersion),
		kubeConfig:     kubeConfig,
		ociUri:         uri,
		chartName:      chartName,
		chartVersion:   chartVersion,
	}
}

func TestInstallControllerSuccess(t *testing.T) {
	tt := newPackageControllerTest(t)

	tt.chartInstaller.EXPECT().InstallChartFromName(tt.ctx, "oci://"+tt.ociUri, tt.kubeConfig, tt.chartName, tt.chartVersion)

	tt.command.InstallController(tt.ctx)
}

func TestGetActiveControllerSuccess(t *testing.T) {
	tt := newPackageControllerTest(t)

	params := []string{"get", "packageBundleController", "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName, bundle.PackageBundleControllerName}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)

	tt.command.GetActiveController(tt.ctx)
}
