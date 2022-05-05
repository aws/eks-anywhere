package curatedpackages_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
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

	params := []string{"get", "packageBundleController", "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName, bundle.PackageBundleControllerName}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)

	err := tt.command.ValidateControllerExists(tt.ctx)
	if err != nil {
		t.Errorf("Get Active Controller should succeed when controller doesn't exist")
	}
}

func TestGetActiveControllerFail(t *testing.T) {
	tt := newPackageControllerTest(t)
	bundleCtrl := packagesv1.PackageBundleController{
		Spec: packagesv1.PackageBundleControllerSpec{
			ActiveBundle: "v1-21-1000",
		},
	}

	params := []string{"get", "packageBundleController", "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName, bundle.PackageBundleControllerName}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(convertJsonToBytes(bundleCtrl), nil)

	err := tt.command.ValidateControllerExists(tt.ctx)
	if err == nil {
		t.Errorf("Get Active Controller should fail when controller exists")
	}
}

func TestGetActiveControllerPassWhenErr(t *testing.T) {
	tt := newPackageControllerTest(t)

	params := []string{"get", "packageBundleController", "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName, bundle.PackageBundleControllerName}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, errors.New("kubeconfig doesn't exist"))

	err := tt.command.ValidateControllerExists(tt.ctx)
	if err != nil {
		t.Errorf("Get Active Controller should succeed when err encountered")
	}
}
