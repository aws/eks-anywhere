package curatedpackages_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/templater"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
)

//go:embed config/awssecret.yaml
var awsSecretYaml string

const (
	eksaDefaultRegion = "us-west-2"
	cronJobName       = "cronjob/ecr-refresher"
	jobName           = "eksa-auth-refresher"
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
	eksaAccessId   string
	eksaAccessKey  string
	eksaRegion     string
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
		eksaAccessId:   "test-access-id",
		eksaAccessKey:  "test-access-key",
		eksaRegion:     "test-region",
	}
}

func TestInstallControllerSuccess(t *testing.T) {
	tt := newPackageControllerTest(t)

	registry := curatedpackages.GetRegistry(tt.ociUri)
	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
	values := []string{sourceRegistry}
	params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
	templateValues := map[string]string{
		"eksaAccessKeyId":     "",
		"eksaSecretAccessKey": "",
		"eksaRegion":          eksaDefaultRegion,
	}
	result, err := templater.Execute(awsSecretYaml, templateValues)
	tt.kubectl.EXPECT().CreateFromYaml(tt.ctx, result, params).Return(bytes.Buffer{}, nil)
	params = []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)
	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(nil)

	err = tt.command.InstallController(tt.ctx)
	if err != nil {
		t.Errorf("Install Controller Should succeed when installation passes")
	}
}

func TestInstallControllerFail(t *testing.T) {
	tt := newPackageControllerTest(t)
	registry := curatedpackages.GetRegistry(tt.ociUri)
	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
	values := []string{sourceRegistry}

	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(errors.New("login failed"))

	err := tt.command.InstallController(tt.ctx)
	if err == nil {
		t.Errorf("Install Controller Should fail when installation fails")
	}
}

func TestInstallControllerSuccessWhenApplySecretFails(t *testing.T) {
	tt := newPackageControllerTest(t)

	registry := curatedpackages.GetRegistry(tt.ociUri)
	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
	values := []string{sourceRegistry}
	params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
	templateValues := map[string]string{
		"eksaAccessKeyId":     "",
		"eksaSecretAccessKey": "",
		"eksaRegion":          eksaDefaultRegion,
	}
	result, err := templater.Execute(awsSecretYaml, templateValues)
	tt.kubectl.EXPECT().CreateFromYaml(tt.ctx, result, params).Return(bytes.Buffer{}, errors.New("error applying secrets"))
	params = []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)
	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(nil)

	err = tt.command.InstallController(tt.ctx)
	if err != nil {
		t.Errorf("Install Controller Should succeed when secret creation fails")
	}
}

func TestInstallControllerSuccessWhenCronJobFails(t *testing.T) {
	tt := newPackageControllerTest(t)

	registry := curatedpackages.GetRegistry(tt.ociUri)
	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
	values := []string{sourceRegistry}
	params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
	templateValues := map[string]string{
		"eksaAccessKeyId":     "",
		"eksaSecretAccessKey": "",
		"eksaRegion":          eksaDefaultRegion,
	}
	result, err := templater.Execute(awsSecretYaml, templateValues)
	tt.kubectl.EXPECT().CreateFromYaml(tt.ctx, result, params).Return(bytes.Buffer{}, nil)
	params = []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName}
	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, errors.New("error creating cron job"))
	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(nil)

	err = tt.command.InstallController(tt.ctx)
	if err != nil {
		t.Errorf("Install Controller Should succeed when cron job fails")
	}
}

func TestGetActiveControllerSuccess(t *testing.T) {
	tt := newPackageControllerTest(t)

	tt.kubectl.EXPECT().GetResource(tt.ctx, "packageBundleController", packagesv1.PackageBundleControllerName, tt.kubeConfig, constants.EksaPackagesName).Return(true, nil)

	err := tt.command.ValidateControllerDoesNotExist(tt.ctx)
	if err == nil {
		t.Errorf("Get Active Controller should not succeed when controller exists")
	}
}

func TestGetActiveControllerFail(t *testing.T) {
	tt := newPackageControllerTest(t)

	tt.kubectl.EXPECT().GetResource(tt.ctx, "packageBundleController", packagesv1.PackageBundleControllerName, tt.kubeConfig, constants.EksaPackagesName).Return(false, errors.New("controller doesn't exist"))

	err := tt.command.ValidateControllerDoesNotExist(tt.ctx)
	if err != nil {
		t.Errorf("Get Active Controller should succeed when controller doesn't exist")
	}
}

func TestApplySecret(t *testing.T) {

}
