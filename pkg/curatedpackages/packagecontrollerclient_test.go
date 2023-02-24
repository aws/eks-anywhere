package curatedpackages_test

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	writermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	artifactsv1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const valueFileName = "values.yaml"

//go:embed testdata/expected_all_values.yaml
var expectedAllValues string

type packageControllerTest struct {
	*WithT
	ctx            context.Context
	kubectl        *mocks.MockKubectlRunner
	chartInstaller *mocks.MockChartInstaller
	command        *curatedpackages.PackageControllerClient
	clusterName    string
	kubeConfig     string
	chart          *artifactsv1.Image
	eksaAccessID   string
	eksaAccessKey  string
	eksaRegion     string
	httpProxy      string
	httpsProxy     string
	noProxy        []string
	registryMirror *registrymirror.RegistryMirror
	writer         filewriter.FileWriter
	wantValueFile  string
}

func newPackageControllerTests(t *testing.T) []*packageControllerTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlRunner(ctrl)
	ci := mocks.NewMockChartInstaller(ctrl)
	kubeConfig := "kubeconfig.kubeconfig"
	chart := &artifactsv1.Image{
		Name: "test_controller",
		URI:  "test_registry/eks-anywhere/eks-anywhere-packages:v1",
	}
	chartDev := &artifactsv1.Image{
		Name: "test_controller",
		URI:  "public.ecr.aws/l0g8r8j6/eks-anywhere-packages:v1",
	}
	chartStaging := &artifactsv1.Image{
		Name: "test_controller",
		URI:  "public.ecr.aws/w9m0f3l5/eks-anywhere-packages:v1",
	}
	eksaAccessId := "test-access-id"
	eksaAccessKey := "test-access-key"
	eksaRegion := "test-region"
	clusterName := "billy"
	registryMirror := &registrymirror.RegistryMirror{
		BaseRegistry: "1.2.3.4:443",
		NamespacedRegistryMap: map[string]string{
			constants.DefaultCoreEKSARegistry:             "1.2.3.4:443/public",
			constants.DefaultCuratedPackagesRegistryRegex: "1.2.3.4:443/private",
		},
		Auth:               true,
		CACertContent:      "-----BEGIN CERTIFICATE-----\nabc\nefg\n-----END CERTIFICATE-----\n",
		InsecureSkipVerify: false,
	}
	registryMirrorInsecure := &registrymirror.RegistryMirror{
		BaseRegistry: "1.2.3.4:8443",
		NamespacedRegistryMap: map[string]string{
			constants.DefaultCoreEKSARegistry:             "1.2.3.4:443/public",
			constants.DefaultCuratedPackagesRegistryRegex: "1.2.3.4:443/private",
		},
		Auth:               true,
		CACertContent:      "-----BEGIN CERTIFICATE-----\nabc\nefg\n-----END CERTIFICATE-----\n",
		InsecureSkipVerify: true,
	}
	writer, _ := filewriter.NewWriter(clusterName)
	clusterSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{},
			},
		},
	}
	return []*packageControllerTest{
		{
			WithT:          NewWithT(t),
			ctx:            context.Background(),
			kubectl:        k,
			chartInstaller: ci,
			command: curatedpackages.NewPackageControllerClient(
				ci, k, clusterName, kubeConfig, chart, registryMirror,
				curatedpackages.WithEksaSecretAccessKey(eksaAccessKey),
				curatedpackages.WithEksaRegion(eksaRegion),
				curatedpackages.WithEksaAccessKeyId(eksaAccessId),
				curatedpackages.WithManagementClusterName(clusterName),
				curatedpackages.WithValuesFileWriter(writer),
				curatedpackages.WithClusterSpec(clusterSpec),
			),
			clusterName:    clusterName,
			kubeConfig:     kubeConfig,
			chart:          chart,
			eksaAccessID:   eksaAccessId,
			eksaAccessKey:  eksaAccessKey,
			eksaRegion:     eksaRegion,
			httpProxy:      "1.1.1.1",
			httpsProxy:     "1.1.1.1",
			noProxy:        []string{"1.1.1.1/24"},
			registryMirror: registryMirror,
			writer:         writer,
			wantValueFile:  "testdata/values_test.yaml",
		},
		{
			WithT:          NewWithT(t),
			ctx:            context.Background(),
			kubectl:        k,
			chartInstaller: ci,
			command: curatedpackages.NewPackageControllerClient(
				ci, k, clusterName, kubeConfig, chart,
				nil,
				curatedpackages.WithEksaSecretAccessKey(eksaAccessKey),
				curatedpackages.WithEksaRegion(eksaRegion),
				curatedpackages.WithEksaAccessKeyId(eksaAccessId),
				curatedpackages.WithManagementClusterName(clusterName),
				curatedpackages.WithValuesFileWriter(writer),
			),
			clusterName:    clusterName,
			kubeConfig:     kubeConfig,
			chart:          chart,
			eksaAccessID:   eksaAccessId,
			eksaAccessKey:  eksaAccessKey,
			eksaRegion:     eksaRegion,
			httpProxy:      "1.1.1.1",
			httpsProxy:     "1.1.1.1",
			noProxy:        []string{"1.1.1.1/24"},
			registryMirror: nil,
			writer:         writer,
			wantValueFile:  "testdata/values_empty_registrymirrorsecret.yaml",
		},
		{
			WithT:          NewWithT(t),
			ctx:            context.Background(),
			kubectl:        k,
			chartInstaller: ci,
			command: curatedpackages.NewPackageControllerClient(
				ci, k, clusterName, kubeConfig, chartDev,
				nil,
				curatedpackages.WithEksaSecretAccessKey(eksaAccessKey),
				curatedpackages.WithEksaRegion(eksaRegion),
				curatedpackages.WithEksaAccessKeyId(eksaAccessId),
				curatedpackages.WithManagementClusterName(clusterName),
				curatedpackages.WithValuesFileWriter(writer),
			),
			clusterName:    clusterName,
			kubeConfig:     kubeConfig,
			chart:          chartDev,
			eksaAccessID:   eksaAccessId,
			eksaAccessKey:  eksaAccessKey,
			eksaRegion:     eksaRegion,
			httpProxy:      "1.1.1.1",
			httpsProxy:     "1.1.1.1",
			noProxy:        []string{"1.1.1.1/24"},
			registryMirror: nil,
			writer:         writer,
			wantValueFile:  "testdata/values_empty_registrymirrorsecret.yaml",
		},
		{
			WithT:          NewWithT(t),
			ctx:            context.Background(),
			kubectl:        k,
			chartInstaller: ci,
			command: curatedpackages.NewPackageControllerClient(
				ci, k, clusterName, kubeConfig, chartStaging,
				nil,
				curatedpackages.WithEksaSecretAccessKey(eksaAccessKey),
				curatedpackages.WithEksaRegion(eksaRegion),
				curatedpackages.WithEksaAccessKeyId(eksaAccessId),
				curatedpackages.WithManagementClusterName(clusterName),
				curatedpackages.WithValuesFileWriter(writer),
			),
			clusterName:    clusterName,
			kubeConfig:     kubeConfig,
			chart:          chartStaging,
			eksaAccessID:   eksaAccessId,
			eksaAccessKey:  eksaAccessKey,
			eksaRegion:     eksaRegion,
			httpProxy:      "1.1.1.1",
			httpsProxy:     "1.1.1.1",
			noProxy:        []string{"1.1.1.1/24"},
			registryMirror: nil,
			writer:         writer,
			wantValueFile:  "testdata/values_empty_registrymirrorsecret.yaml",
		},
		{
			WithT:          NewWithT(t),
			ctx:            context.Background(),
			kubectl:        k,
			chartInstaller: ci,
			command: curatedpackages.NewPackageControllerClient(
				ci, k, clusterName, kubeConfig, chart, registryMirrorInsecure,
				curatedpackages.WithManagementClusterName(clusterName),
				curatedpackages.WithValuesFileWriter(writer),
			),
			clusterName:    clusterName,
			kubeConfig:     kubeConfig,
			chart:          chart,
			eksaAccessID:   "",
			eksaAccessKey:  "",
			eksaRegion:     "",
			httpProxy:      "1.1.1.1",
			httpsProxy:     "1.1.1.1",
			noProxy:        []string{"1.1.1.1/24"},
			registryMirror: registryMirrorInsecure,
			writer:         writer,
			wantValueFile:  "testdata/values_empty_awssecret.yaml",
		},
		{
			WithT:          NewWithT(t),
			ctx:            context.Background(),
			kubectl:        k,
			chartInstaller: ci,
			command: curatedpackages.NewPackageControllerClient(
				ci, k, clusterName, kubeConfig, chart, nil,
				curatedpackages.WithManagementClusterName(clusterName),
				curatedpackages.WithValuesFileWriter(writer),
			),
			clusterName:    clusterName,
			kubeConfig:     kubeConfig,
			chart:          chart,
			eksaAccessID:   "",
			eksaAccessKey:  "",
			eksaRegion:     "",
			httpProxy:      "1.1.1.1",
			httpsProxy:     "1.1.1.1",
			noProxy:        []string{"1.1.1.1/24"},
			registryMirror: nil,
			writer:         writer,
			wantValueFile:  "testdata/values_empty.yaml",
		},
	}
}

func TestEnableCuratedPackagesSuccess(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableCuratedPackagesSucceedInWorkloadCluster(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, tt.clusterName, tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
			curatedpackages.WithEksaRegion("us-west-2"),
			curatedpackages.WithEksaAccessKeyId(tt.eksaAccessID),
			curatedpackages.WithManagementClusterName("mgmt"),
			curatedpackages.WithValuesFileWriter(tt.writer),
		)
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		values = append(values, "workloadOnly=true")
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name+"-billy", ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		tt.Expect(err).To(BeNil())
	}
}

func getPBCSuccess(t *testing.T) func(context.Context, string, string, string, string, *packagesv1.PackageBundleController) error {
	return func(_ context.Context, _, _, _, _ string, obj *packagesv1.PackageBundleController) error {
		pbc := &packagesv1.PackageBundleController{
			Spec: packagesv1.PackageBundleControllerSpec{
				ActiveBundle: "test-bundle",
			},
		}
		pbc.DeepCopyInto(obj)
		return nil
	}
}

func getPBCFail(t *testing.T) func(context.Context, string, string, string, string, *packagesv1.PackageBundleController) error {
	return func(_ context.Context, _, _, _, _ string, obj *packagesv1.PackageBundleController) error {
		return fmt.Errorf("test error")
	}
}

func TestEnableCuratedPackagesWithProxy(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
			curatedpackages.WithEksaRegion(tt.eksaRegion),
			curatedpackages.WithEksaAccessKeyId(tt.eksaAccessID),
			curatedpackages.WithHTTPProxy(tt.httpProxy),
			curatedpackages.WithHTTPSProxy(tt.httpsProxy),
			curatedpackages.WithNoProxy(tt.noProxy),
			curatedpackages.WithManagementClusterName(tt.clusterName),
			curatedpackages.WithValuesFileWriter(tt.writer),
		)
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		httpProxy := fmt.Sprintf("proxy.HTTP_PROXY=%s", tt.httpProxy)
		httpsProxy := fmt.Sprintf("proxy.HTTPS_PROXY=%s", tt.httpsProxy)
		noProxy := fmt.Sprintf("proxy.NO_PROXY=%s", strings.Join(tt.noProxy, "\\,"))
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		} else {
			if tt.eksaRegion == "" {
				tt.eksaRegion = "us-west-2"
			}
			defaultImageRegistry = strings.ReplaceAll(defaultImageRegistry, "us-west-2", tt.eksaRegion)
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName, httpProxy, httpsProxy, noProxy}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableCuratedPackagesWithEmptyProxy(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
			curatedpackages.WithEksaRegion(tt.eksaRegion),
			curatedpackages.WithEksaAccessKeyId(tt.eksaAccessID),
			curatedpackages.WithHTTPProxy(""),
			curatedpackages.WithHTTPSProxy(""),
			curatedpackages.WithNoProxy(nil),
			curatedpackages.WithManagementClusterName(tt.clusterName),
			curatedpackages.WithValuesFileWriter(tt.writer),
		)
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		} else {
			if tt.eksaRegion == "" {
				tt.eksaRegion = "us-west-2"
			}
			defaultImageRegistry = strings.ReplaceAll(defaultImageRegistry, "us-west-2", tt.eksaRegion)
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableCuratedPackagesFail(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(errors.New("login failed"))
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err == nil {
			t.Errorf("Install Controller Should fail when installation fails")
		}
	}
}

func TestEnableCuratedPackagesFailNoActiveBundle(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCFail(t)).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	}
}

func TestEnableCuratedPackagesSuccessWhenCronJobFails(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when cron job fails")
		}
	}
}

func TestIsInstalledTrue(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.kubectl.EXPECT().HasResource(tt.ctx, "packageBundleController", tt.clusterName, tt.kubeConfig, constants.EksaPackagesName).Return(false, nil)

		found := tt.command.IsInstalled(tt.ctx)
		if found {
			t.Errorf("expected true, got %t", found)
		}
	}
}

func TestIsInstalledFalse(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {

		tt.kubectl.EXPECT().HasResource(tt.ctx, "packageBundleController", tt.clusterName, tt.kubeConfig, constants.EksaPackagesName).
			Return(false, errors.New("controller doesn't exist"))

		found := tt.command.IsInstalled(tt.ctx)
		if found {
			t.Errorf("expected false, got %t", found)
		}
	}
}

func TestEnableCuratedPackagesActiveBundleCustomTimeout(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
			curatedpackages.WithEksaRegion(tt.eksaRegion),
			curatedpackages.WithEksaAccessKeyId(tt.eksaAccessID),
			curatedpackages.WithActiveBundleTimeout(time.Second),
			curatedpackages.WithManagementClusterName(tt.clusterName),
			curatedpackages.WithValuesFileWriter(tt.writer),
		)
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		} else {
			if tt.eksaRegion == "" {
				tt.eksaRegion = "us-west-2"
			}
			defaultImageRegistry = strings.ReplaceAll(defaultImageRegistry, "us-west-2", tt.eksaRegion)
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableCuratedPackagesActiveBundleWaitLoops(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCLoops(t, 3)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	}
}

func getPBCLoops(t *testing.T, loops int) func(context.Context, string, string, string, string, *packagesv1.PackageBundleController) error {
	return func(_ context.Context, _, _, _, _ string, obj *packagesv1.PackageBundleController) error {
		loops = loops - 1
		if loops > 0 {
			return nil
		}
		pbc := &packagesv1.PackageBundleController{
			Spec: packagesv1.PackageBundleControllerSpec{
				ActiveBundle: "test-bundle",
			},
		}
		pbc.DeepCopyInto(obj)
		return nil
	}
}

func TestEnableCuratedPackagesActiveBundleTimesOut(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
			curatedpackages.WithEksaRegion(tt.eksaRegion),
			curatedpackages.WithEksaAccessKeyId(tt.eksaAccessID),
			curatedpackages.WithActiveBundleTimeout(time.Millisecond),
			curatedpackages.WithManagementClusterName(tt.clusterName),
			curatedpackages.WithValuesFileWriter(tt.writer),
		)
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		} else {
			if tt.eksaRegion == "" {
				tt.eksaRegion = "us-west-2"
			}
			defaultImageRegistry = strings.ReplaceAll(defaultImageRegistry, "us-west-2", tt.eksaRegion)
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCDelay(t, time.Second)).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		expectedErr := fmt.Errorf("timed out finding an active package bundle / eksa-packages-billy namespace for the current cluster: %v", context.DeadlineExceeded)
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	}
}

func TestEnableCuratedPackagesActiveBundleNamespaceTimesOut(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
			curatedpackages.WithEksaRegion(tt.eksaRegion),
			curatedpackages.WithEksaAccessKeyId(tt.eksaAccessID),
			curatedpackages.WithActiveBundleTimeout(time.Millisecond),
			curatedpackages.WithManagementClusterName(tt.clusterName),
			curatedpackages.WithValuesFileWriter(tt.writer),
		)
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries()
		sourceRegistry = fmt.Sprintf("sourceRegistry=%s", sourceRegistry)
		defaultRegistry = fmt.Sprintf("defaultRegistry=%s", defaultRegistry)
		defaultImageRegistry = fmt.Sprintf("defaultImageRegistry=%s", defaultImageRegistry)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		} else {
			if tt.eksaRegion == "" {
				tt.eksaRegion = "us-west-2"
			}
			defaultImageRegistry = strings.ReplaceAll(defaultImageRegistry, "us-west-2", tt.eksaRegion)
		}
		values := []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
		if (tt.eksaAccessID == "" || tt.eksaAccessKey == "") && tt.registryMirror == nil {
			values = append(values, "cronjob.suspend=true")
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", valueFilePath, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return false, nil }).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		expectedErr := fmt.Errorf("timed out finding an active package bundle / eksa-packages-billy namespace for the current cluster: %v", context.DeadlineExceeded)
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	}
}

func getPBCDelay(t *testing.T, delay time.Duration) func(context.Context, string, string, string, string, *packagesv1.PackageBundleController) error {
	return func(_ context.Context, _, _, _, _ string, obj *packagesv1.PackageBundleController) error {
		time.Sleep(delay)
		return fmt.Errorf("test error")
	}
}

func TestCreateHelmOverrideValuesYaml(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		}
		filePath, content, err := tt.command.CreateHelmOverrideValuesYaml()
		tt.Expect(err).To(BeNil())
		tt.Expect(filePath).To(Equal(filepath.Join(tt.clusterName, filewriter.DefaultTmpFolder, "values.yaml")))
		test.AssertContentToFile(t, string(content), tt.wantValueFile)
	}
}

func TestCreateHelmOverrideValuesYamlFail(t *testing.T) {
	_ = os.Unsetenv("REGISTRY_USERNAME")
	_ = os.Unsetenv("REGISTRY_PASSWORD")
	for _, tt := range newPackageControllerTests(t) {
		filePath, content, err := tt.command.CreateHelmOverrideValuesYaml()
		if tt.registryMirror != nil {
			tt.Expect(err).NotTo(BeNil())
			tt.Expect(filePath).To(Equal(""))
		} else {
			tt.Expect(err).To(BeNil())
			tt.Expect(filePath).To(Equal(filepath.Join(tt.clusterName, filewriter.DefaultTmpFolder, "values.yaml")))
			test.AssertContentToFile(t, string(content), tt.wantValueFile)
		}
	}
}

func TestCreateHelmOverrideValuesYamlFailWithNoWriter(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
			curatedpackages.WithEksaRegion(tt.eksaRegion),
			curatedpackages.WithEksaAccessKeyId(tt.eksaAccessID),
			curatedpackages.WithActiveBundleTimeout(time.Second),
			curatedpackages.WithManagementClusterName(tt.clusterName),
		)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		}

		err := tt.command.EnableCuratedPackages(tt.ctx)
		expectedErr := fmt.Errorf("valuesFileWriter is nil")
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	}
}

func TestCreateHelmOverrideValuesYamlFailWithWriteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	writer := writermocks.NewMockFileWriter(ctrl)
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithValuesFileWriter(writer),
		)
		if tt.registryMirror != nil {
			t.Setenv("REGISTRY_USERNAME", "username")
			t.Setenv("REGISTRY_PASSWORD", "password")
		}
		writer.EXPECT().Write(gomock.Any(), gomock.Any()).Return("", errors.New("writer errors out"))

		filePath, content, err := tt.command.CreateHelmOverrideValuesYaml()
		tt.Expect(filePath).To(Equal(""))
		tt.Expect(content).NotTo(BeNil())
		tt.Expect(err).NotTo(BeNil())
	}
}

func TestGetPackageControllerConfigurationNil(t *testing.T) {
	g := NewWithT(t)
	sut := curatedpackages.NewPackageControllerClient(nil, nil, "billy", "", nil, nil)
	result, err := sut.GetPackageControllerConfiguration()
	g.Expect(result).To(Equal(""))
	g.Expect(err).To(BeNil())
}

func TestGetPackageControllerConfigurationAll(t *testing.T) {
	clusterSpec := v1alpha1.ClusterSpec{
		Packages: &v1alpha1.PackageConfiguration{
			Disable: false,
			Controller: &v1alpha1.PackageControllerConfiguration{
				Repository:      "my-repo",
				Digest:          "my-digest",
				DisableWebhooks: true,
				Tag:             "my-tag",
				Env:             []string{"A=B"},
				Resources: v1alpha1.PackageControllerResources{
					Limits: v1alpha1.ImageResource{
						CPU:    "my-cpu",
						Memory: "my-memory",
					},
					Requests: v1alpha1.ImageResource{
						CPU:    "my-requests-cpu",
						Memory: "my-requests-memory",
					},
				},
			},
			CronJob: &v1alpha1.PackageControllerCronJob{
				Repository: "my-cronjob-repo",
				Digest:     "my-cronjob-digest",
				Tag:        "my-cronjob-tag",
				Disable:    true,
			},
		},
	}
	cluster := cluster.Spec{Config: &cluster.Config{Cluster: &v1alpha1.Cluster{Spec: clusterSpec}}}
	g := NewWithT(t)
	sut := curatedpackages.NewPackageControllerClient(nil, nil, "billy", "", nil, nil, curatedpackages.WithClusterSpec(&cluster))
	result, err := sut.GetPackageControllerConfiguration()
	g.Expect(result).To(Equal(expectedAllValues))
	g.Expect(err).To(BeNil())
}

func TestGetPackageControllerConfigurationNothing(t *testing.T) {
	clusterSpec := v1alpha1.ClusterSpec{
		Packages: &v1alpha1.PackageConfiguration{
			Disable: true,
		},
	}
	g := NewWithT(t)
	cluster := cluster.Spec{Config: &cluster.Config{Cluster: &v1alpha1.Cluster{Spec: clusterSpec}}}
	sut := curatedpackages.NewPackageControllerClient(nil, nil, "billy", "", nil, nil, curatedpackages.WithClusterSpec(&cluster))
	result, err := sut.GetPackageControllerConfiguration()
	g.Expect(result).To(Equal(""))
	g.Expect(err).To(BeNil())
}

func TestGetPackageControllerConfigurationError(t *testing.T) {
	clusterSpec := v1alpha1.ClusterSpec{
		Packages: &v1alpha1.PackageConfiguration{
			Disable: false,
			Controller: &v1alpha1.PackageControllerConfiguration{
				Env: []string{"AB"},
			},
		},
	}
	g := NewWithT(t)
	cluster := cluster.Spec{Config: &cluster.Config{Cluster: &v1alpha1.Cluster{Spec: clusterSpec}}}
	sut := curatedpackages.NewPackageControllerClient(nil, nil, "billy", "", nil, nil, curatedpackages.WithClusterSpec(&cluster))
	_, err := sut.GetPackageControllerConfiguration()
	g.Expect(err).NotTo(BeNil())
	g.Expect(err.Error()).To(Equal("invalid environment in specification <AB>"))
}
