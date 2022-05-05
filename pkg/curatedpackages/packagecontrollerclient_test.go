package curatedpackages_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere-packages/pkg/bundle"
	"github.com/aws/eks-anywhere/pkg/constants"
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

	tt.chartInstaller.EXPECT().InstallChartFromName(tt.ctx, "oci://"+tt.ociUri, tt.kubeConfig, tt.chartName, tt.chartVersion).Return(nil)

	err := tt.command.InstallController(tt.ctx)
	if err != nil {
		t.Errorf("Install Controller Should succeed when installation passes")
	}
}

func TestInstallControllerFail(t *testing.T) {
	tt := newPackageControllerTest(t)

	tt.chartInstaller.EXPECT().InstallChartFromName(tt.ctx, "oci://"+tt.ociUri, tt.kubeConfig, tt.chartName, tt.chartVersion).Return(errors.New("login failed"))

	err := tt.command.InstallController(tt.ctx)
	if err == nil {
		t.Errorf("Install Controller Should fail when installation fails")
	}
}

func TestGetActiveControllerSuccess(t *testing.T) {
	tt := newPackageControllerTest(t)

	tt.kubectl.EXPECT().GetResource(tt.ctx, "packageBundleController", bundle.PackageBundleControllerName, tt.kubeConfig, constants.EksaPackagesName).Return(true, nil)

	err := tt.command.ValidateControllerDoesNotExist(tt.ctx)
	if err == nil {
		t.Errorf("Get Active Controller should not succeed when controller exists")
	}
}

func TestGetActiveControllerFail(t *testing.T) {
	tt := newPackageControllerTest(t)

	tt.kubectl.EXPECT().GetResource(tt.ctx, "packageBundleController", bundle.PackageBundleControllerName, tt.kubeConfig, constants.EksaPackagesName).Return(false, errors.New("controller doesn't exist"))

	err := tt.command.ValidateControllerDoesNotExist(tt.ctx)
	if err != nil {
		t.Errorf("Get Active Controller should succeed when controller doesn't exist")
	}
}
