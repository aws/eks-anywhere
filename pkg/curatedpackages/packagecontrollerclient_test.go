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

	"github.com/go-logr/logr/testr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
	chartManager   *mocks.MockChartManager
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
	cm := mocks.NewMockChartManager(ctrl)
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
	eksaAwsConfigFile := "test-aws-config-file"
	eksaRegion := "test-region"
	clusterName := "billy"
	registryMirror := &registrymirror.RegistryMirror{
		BaseRegistry: "1.2.3.4:443",
		NamespacedRegistryMap: map[string]string{
			constants.DefaultCoreEKSARegistry:        "1.2.3.4:443/public",
			constants.DefaultCuratedPackagesRegistry: "1.2.3.4:443/private",
		},
		Auth:               true,
		CACertContent:      "-----BEGIN CERTIFICATE-----\nabc\nefg\n-----END CERTIFICATE-----\n",
		InsecureSkipVerify: false,
	}
	registryMirrorInsecure := &registrymirror.RegistryMirror{
		BaseRegistry: "1.2.3.4:8443",
		NamespacedRegistryMap: map[string]string{
			constants.DefaultCoreEKSARegistry:        "1.2.3.4:443/public",
			constants.DefaultCuratedPackagesRegistry: "1.2.3.4:443/private",
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
			WithT:        NewWithT(t),
			ctx:          context.Background(),
			kubectl:      k,
			chartManager: cm,
			command: curatedpackages.NewPackageControllerClient(
				cm, k, clusterName, kubeConfig, chart, registryMirror,
				curatedpackages.WithEksaSecretAccessKey(eksaAccessKey),
				curatedpackages.WithEksaRegion(eksaRegion),
				curatedpackages.WithEksaAccessKeyId(eksaAccessId),
				curatedpackages.WithEksaAwsConfig(eksaAwsConfigFile),
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
			WithT:        NewWithT(t),
			ctx:          context.Background(),
			kubectl:      k,
			chartManager: cm,
			command: curatedpackages.NewPackageControllerClient(
				cm, k, clusterName, kubeConfig, chart,
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
			WithT:        NewWithT(t),
			ctx:          context.Background(),
			kubectl:      k,
			chartManager: cm,
			command: curatedpackages.NewPackageControllerClient(
				cm, k, clusterName, kubeConfig, chartDev,
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
			WithT:        NewWithT(t),
			ctx:          context.Background(),
			kubectl:      k,
			chartManager: cm,
			command: curatedpackages.NewPackageControllerClient(
				cm, k, clusterName, kubeConfig, chartStaging,
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
			WithT:        NewWithT(t),
			ctx:          context.Background(),
			kubectl:      k,
			chartManager: cm,
			command: curatedpackages.NewPackageControllerClient(
				cm, k, clusterName, kubeConfig, chart, registryMirrorInsecure,
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
			WithT:        NewWithT(t),
			ctx:          context.Background(),
			kubectl:      k,
			chartManager: cm,
			command: curatedpackages.NewPackageControllerClient(
				cm, k, clusterName, kubeConfig, chart, nil,
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

func TestEnableSuccess(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableSucceedInWorkloadCluster(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartManager, tt.kubectl, tt.clusterName, tt.kubeConfig, tt.chart,
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
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		values = append(values, "managementClusterName=mgmt")
		values = append(values, "workloadPackageOnly=true")
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name+"-billy", ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, true, gomock.InAnyOrder(values)).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
		tt.Expect(err).To(BeNil())
	}
}

func TestEnableSucceedInWorkloadClusterWhenPackageBundleControllerNotExist(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartManager, tt.kubectl, tt.clusterName, tt.kubeConfig, tt.chart,
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
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		values = append(values, "managementClusterName=mgmt")
		values = append(values, "workloadPackageOnly=true")
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name+"-billy", ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, true, gomock.InAnyOrder(values)).Return(nil)
		gomock.InOrder(
			tt.kubectl.EXPECT().
				GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(getPBCNotFound(t)),
			tt.kubectl.EXPECT().
				GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(getPBCSuccess(t)))
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
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

func getPBCNotFound(t *testing.T) func(context.Context, string, string, string, string, *packagesv1.PackageBundleController) error {
	return func(_ context.Context, _, _, _, _ string, obj *packagesv1.PackageBundleController) error {
		return apierrors.NewNotFound(schema.GroupResource{
			Group:    "test group",
			Resource: "test resource",
		}, "test")
	}
}

func getPBCFail(t *testing.T) func(context.Context, string, string, string, string, *packagesv1.PackageBundleController) error {
	return func(_ context.Context, _, _, _, _ string, obj *packagesv1.PackageBundleController) error {
		return fmt.Errorf("test error")
	}
}

func TestEnableWithProxy(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartManager, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
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
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableWithEmptyProxy(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartManager, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
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
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableFail(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(errors.New("login failed"))
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
		if err == nil {
			t.Errorf("Install Controller Should fail when installation fails")
		}
	}
}

func TestEnableFailNoActiveBundle(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCFail(t)).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	}
}

func TestEnableSuccessWhenCronJobFails(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
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

func TestEnableActiveBundleCustomTimeout(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartManager, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
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
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableActiveBundleWaitLoops(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCLoops(t, 3)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
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

func TestEnableActiveBundleTimesOut(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartManager, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
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
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCDelay(t, time.Second)).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
		expectedErr := fmt.Errorf("timed out finding an active package bundle / eksa-packages-billy namespace for the current cluster: %v", context.DeadlineExceeded)
		if err.Error() != expectedErr.Error() {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	}
}

func TestEnableActiveBundleNamespaceTimesOut(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartManager, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
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
		sourceRegistry, defaultRegistry, defaultImageRegistry := tt.command.GetCuratedPackagesRegistries(context.Background())
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
		tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, false, values).Return(nil)
		tt.kubectl.EXPECT().
			GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()
		tt.kubectl.EXPECT().
			HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return false, nil }).
			AnyTimes()

		err := tt.command.Enable(tt.ctx)
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
			tt.chartManager, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
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

		err := tt.command.Enable(tt.ctx)
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
			tt.chartManager, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
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

func TestGetCuratedPackagesRegistriesDefaultRegion(t *testing.T) {
	clusterSpec := v1alpha1.ClusterSpec{
		Packages: &v1alpha1.PackageConfiguration{
			Disable: true,
		},
	}
	chart := &artifactsv1.Image{
		Name: "test_controller",
		URI:  "test_registry/eks-anywhere/eks-anywhere-packages:v1",
	}
	g := NewWithT(t)
	cluster := cluster.Spec{Config: &cluster.Config{Cluster: &v1alpha1.Cluster{Spec: clusterSpec}}}
	sut := curatedpackages.NewPackageControllerClient(nil, nil, "billy", "", chart, nil, curatedpackages.WithClusterSpec(&cluster))
	_, _, img := sut.GetCuratedPackagesRegistries(context.Background())
	g.Expect(img).To(Equal("783794618700.dkr.ecr.us-west-2.amazonaws.com"))
}

func TestGetCuratedPackagesRegistriesCustomRegion(t *testing.T) {
	clusterSpec := v1alpha1.ClusterSpec{
		Packages: &v1alpha1.PackageConfiguration{
			Disable: true,
		},
	}
	chart := &artifactsv1.Image{
		Name: "test_controller",
		URI:  "test_registry/eks-anywhere/eks-anywhere-packages:v1",
	}
	g := NewWithT(t)
	cluster := cluster.Spec{Config: &cluster.Config{Cluster: &v1alpha1.Cluster{Spec: clusterSpec}}}
	sut := curatedpackages.NewPackageControllerClient(nil, nil, "billy", "", chart, nil, curatedpackages.WithClusterSpec(&cluster), curatedpackages.WithEksaRegion("test"))
	_, _, img := sut.GetCuratedPackagesRegistries(context.Background())
	g.Expect(img).To(Equal("783794618700.dkr.ecr.test.amazonaws.com"))
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

func TestReconcileDeleteGoldenPath(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	log := testr.New(t)

	cluster := &v1alpha1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "billy"}}
	kubeconfig := "test.kubeconfig"
	nsName := constants.EksaPackagesName + "-" + cluster.Name
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
	client := fake.NewClientBuilder().WithRuntimeObjects(ns).Build()
	ctrl := gomock.NewController(t)
	chartManager := mocks.NewMockChartManager(ctrl)
	chartManager.EXPECT().Delete(ctx, kubeconfig, "eks-anywhere-packages-"+cluster.Name, constants.EksaPackagesName)

	sut := curatedpackages.NewPackageControllerClient(chartManager, nil, "billy", kubeconfig, nil, nil)

	err := sut.ReconcileDelete(ctx, log, client, cluster)
	g.Expect(err).To(BeNil())
}

func TestReconcileDeleteNamespaceErrorHandling(s *testing.T) {
	s.Run("ignores not found errors", func(t *testing.T) {
		g := NewWithT(t)
		ctx := context.Background()
		log := testr.New(t)
		cluster := &v1alpha1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "billy"}}
		kubeconfig := "test.kubeconfig"
		ctrl := gomock.NewController(t)
		client := mocks.NewMockKubeDeleter(ctrl)
		ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "eksa-packages-" + cluster.Name}}
		notFoundErr := apierrors.NewNotFound(schema.GroupResource{}, "NOT FOUND: test error")
		client.EXPECT().Delete(ctx, ns).Return(notFoundErr)
		chartManager := mocks.NewMockChartManager(ctrl)
		chartManager.EXPECT().Delete(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		sut := curatedpackages.NewPackageControllerClient(chartManager, nil, "billy", kubeconfig, nil, nil)

		err := sut.ReconcileDelete(ctx, log, client, cluster)
		g.Expect(err).ShouldNot(HaveOccurred())
	})

	s.Run("aborts on errors other than not found", func(t *testing.T) {
		g := NewWithT(t)
		ctx := context.Background()
		log := testr.New(t)
		cluster := &v1alpha1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "billy"}}
		kubeconfig := "test.kubeconfig"
		testErr := fmt.Errorf("test error")
		ctrl := gomock.NewController(t)
		client := mocks.NewMockKubeDeleter(ctrl)
		client.EXPECT().Delete(ctx, gomock.Any()).Return(testErr)
		chartManager := mocks.NewMockChartManager(ctrl)

		sut := curatedpackages.NewPackageControllerClient(chartManager, nil, "billy", kubeconfig, nil, nil)

		err := sut.ReconcileDelete(ctx, log, client, cluster)
		g.Expect(err).Should(HaveOccurred())
	})
}

func TestReconcileDeleteHelmErrorsHandling(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	log := testr.New(t)

	cluster := &v1alpha1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: "billy"}}
	kubeconfig := "test.kubeconfig"
	nsName := constants.EksaPackagesName + "-" + cluster.Name
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
	client := fake.NewClientBuilder().WithRuntimeObjects(ns).Build()
	ctrl := gomock.NewController(t)
	chartManager := mocks.NewMockChartManager(ctrl)
	// TODO this namespace should no longer be empty, following PR 5081
	testErr := fmt.Errorf("test error")
	chartManager.EXPECT().
		Delete(ctx, kubeconfig, "eks-anywhere-packages-"+cluster.Name, constants.EksaPackagesName).
		Return(testErr)

	sut := curatedpackages.NewPackageControllerClient(chartManager, nil, "billy", kubeconfig, nil, nil)

	err := sut.ReconcileDelete(ctx, log, client, cluster)
	g.Expect(err).Should(HaveOccurred())
	g.Expect(err.Error()).Should(Equal("test error"))
}

func TestEnableFullLifecyclePath(t *testing.T) {
	log := testr.New(t)
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlRunner(ctrl)
	cm := mocks.NewMockChartManager(ctrl)
	kubeConfig := "kubeconfig.kubeconfig"
	chart := &artifactsv1.Image{
		Name: "test_controller",
		URI:  "test_registry/eks-anywhere/eks-anywhere-packages:v1",
	}
	clusterName := "billy"
	writer, _ := filewriter.NewWriter(clusterName)

	tt := packageControllerTest{
		WithT:          NewWithT(t),
		ctx:            context.Background(),
		kubectl:        k,
		chartManager:   cm,
		command:        curatedpackages.NewPackageControllerClientFullLifecycle(log, cm, k, nil),
		clusterName:    clusterName,
		kubeConfig:     kubeConfig,
		chart:          chart,
		registryMirror: nil,
		writer:         writer,
		wantValueFile:  "testdata/values_empty.yaml",
	}

	valueFilePath := filepath.Join("billy", filewriter.DefaultTmpFolder, valueFileName)
	ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
	// GetCuratedPackagesRegistries can't be used here, as when initialized
	// via full cluster lifecycle the package controller client hasn't yet
	// determined its chart.
	values := []string{
		"clusterName=" + clusterName,
		"managementClusterName=mgmt",
		"workloadPackageOnly=true",
		"sourceRegistry=public.ecr.aws/eks-anywhere",
		"defaultRegistry=public.ecr.aws/eks-anywhere",
		"defaultImageRegistry=783794618700.dkr.ecr.us-west-2.amazonaws.com",
		"cronjob.suspend=true",
	}

	tt.chartManager.EXPECT().InstallChart(tt.ctx, tt.chart.Name+"-"+clusterName, ociURI, tt.chart.Tag(), tt.kubeConfig, constants.EksaPackagesName, valueFilePath, true, gomock.InAnyOrder(values)).Return(nil)
	tt.kubectl.EXPECT().
		GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(getPBCSuccess(t)).
		AnyTimes()
	tt.kubectl.EXPECT().
		HasResource(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_, _, _, _, _ interface{}) (bool, error) { return true, nil }).
		AnyTimes()
	chartImage := &artifactsv1.Image{
		Name: "test_controller",
		URI:  "test_registry/eks-anywhere/eks-anywhere-packages:v1",
	}

	err := tt.command.EnableFullLifecycle(tt.ctx, log, clusterName, kubeConfig, chartImage, tt.registryMirror,
		curatedpackages.WithEksaRegion("us-west-2"),
		curatedpackages.WithManagementClusterName("mgmt"))
	if err != nil {
		t.Errorf("Install Controller Should succeed when installation passes")
	}
}

type stubRegistryAccessTester struct{}

func (s *stubRegistryAccessTester) Test(ctx context.Context, accessKey, secret, registry, region, awsConfig string) error {
	return nil
}

func TestGetCuratedPackagesRegistries(s *testing.T) {
	s.Run("substitutes a region if set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		k := mocks.NewMockKubectlRunner(ctrl)
		cm := mocks.NewMockChartManager(ctrl)
		kubeConfig := "kubeconfig.kubeconfig"
		chart := &artifactsv1.Image{
			Name: "test_controller",
			URI:  "test_registry/eks-anywhere/eks-anywhere-packages:v1",
		}
		// eksaRegion := "test-region"
		clusterName := "billy"
		writer, _ := filewriter.NewWriter(clusterName)
		client := curatedpackages.NewPackageControllerClient(
			cm, k, clusterName, kubeConfig, chart, nil,
			curatedpackages.WithManagementClusterName(clusterName),
			curatedpackages.WithValuesFileWriter(writer),
			curatedpackages.WithEksaRegion("testing"),
		)

		expected := "783794618700.dkr.ecr.testing.amazonaws.com"
		_, _, got := client.GetCuratedPackagesRegistries(context.Background())

		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	s.Run("won't substitute a blank region", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		k := mocks.NewMockKubectlRunner(ctrl)
		cm := mocks.NewMockChartManager(ctrl)
		kubeConfig := "kubeconfig.kubeconfig"
		chart := &artifactsv1.Image{
			Name: "test_controller",
			URI:  "test_registry/eks-anywhere/eks-anywhere-packages:v1",
		}
		// eksaRegion := "test-region"
		clusterName := "billy"
		writer, _ := filewriter.NewWriter(clusterName)
		client := curatedpackages.NewPackageControllerClient(
			cm, k, clusterName, kubeConfig, chart, nil,
			curatedpackages.WithManagementClusterName(clusterName),
			curatedpackages.WithValuesFileWriter(writer),
		)

		expected := "783794618700.dkr.ecr.us-west-2.amazonaws.com"
		_, _, got := client.GetCuratedPackagesRegistries(context.Background())

		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	s.Run("get regional registries", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		k := mocks.NewMockKubectlRunner(ctrl)
		cm := mocks.NewMockChartManager(ctrl)
		kubeConfig := "kubeconfig.kubeconfig"
		chart := &artifactsv1.Image{
			Name: "test_controller",
			URI:  "test_registry/eks-anywhere/eks-anywhere-packages:v1",
		}
		// eksaRegion := "test-region"
		clusterName := "billy"
		writer, _ := filewriter.NewWriter(clusterName)
		client := curatedpackages.NewPackageControllerClient(
			cm, k, clusterName, kubeConfig, chart, nil,
			curatedpackages.WithManagementClusterName(clusterName),
			curatedpackages.WithValuesFileWriter(writer),
			curatedpackages.WithRegistryAccessTester(&stubRegistryAccessTester{}),
		)

		expected := "346438352937.dkr.ecr.us-west-2.amazonaws.com"
		_, actualDefaultRegistry, actualImageRegistry := client.GetCuratedPackagesRegistries(context.Background())

		if actualDefaultRegistry != expected {
			t.Errorf("expected %q, got %q", expected, actualDefaultRegistry)
		}
		if actualImageRegistry != expected {
			t.Errorf("expected %q, got %q", expected, actualImageRegistry)
		}
	})
}

func TestReconcile(s *testing.T) {
	s.Run("golden path", func(t *testing.T) {
		ctx := context.Background()
		log := testr.New(t)
		cluster := newReconcileTestCluster()
		ctrl := gomock.NewController(t)
		k := mocks.NewMockKubectlRunner(ctrl)
		cm := mocks.NewMockChartManager(ctrl)
		bundles := createBundle(cluster)
		bundles.Spec.VersionsBundles[0].KubeVersion = string(cluster.Spec.KubernetesVersion)
		bundles.ObjectMeta.Name = cluster.Spec.BundlesRef.Name
		bundles.ObjectMeta.Namespace = cluster.Spec.BundlesRef.Namespace
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: constants.EksaSystemNamespace,
				Name:      cluster.Name + "-kubeconfig",
			},
		}
		eksaRelease := createEKSARelease(cluster, bundles)
		cluster.Spec.BundlesRef = nil
		objs := []runtime.Object{cluster, bundles, secret, eksaRelease}
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		cm.EXPECT().InstallChart(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		pcc := curatedpackages.NewPackageControllerClientFullLifecycle(log, cm, k, nil)
		err := pcc.Reconcile(ctx, log, fakeClient, cluster)
		if err != nil {
			t.Errorf("expected nil error, got %s", err)
		}
	})

	s.Run("errors when bundles aren't found", func(t *testing.T) {
		ctx := context.Background()
		log := testr.New(t)
		cluster := newReconcileTestCluster()
		ctrl := gomock.NewController(t)
		k := mocks.NewMockKubectlRunner(ctrl)
		cm := mocks.NewMockChartManager(ctrl)

		bundles := createBundle(cluster)
		eksaRelease := createEKSARelease(cluster, bundles)
		objs := []runtime.Object{cluster, eksaRelease}
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

		pcc := curatedpackages.NewPackageControllerClientFullLifecycle(log, cm, k, nil)
		err := pcc.Reconcile(ctx, log, fakeClient, cluster)
		if err == nil || !apierrors.IsNotFound(err) {
			t.Errorf("expected not found err getting cluster resource, got %s", err)
		}
	})

	s.Run("errors when eksarelease isn't found", func(t *testing.T) {
		ctx := context.Background()
		log := testr.New(t)
		cluster := newReconcileTestCluster()
		ctrl := gomock.NewController(t)
		k := mocks.NewMockKubectlRunner(ctrl)
		cm := mocks.NewMockChartManager(ctrl)

		objs := []runtime.Object{cluster}
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

		pcc := curatedpackages.NewPackageControllerClientFullLifecycle(log, cm, k, nil)
		err := pcc.Reconcile(ctx, log, fakeClient, cluster)
		if err == nil || !apierrors.IsNotFound(err) {
			t.Errorf("expected not found err getting cluster resource, got %s", err)
		}
	})

	s.Run("errors when a matching k8s bundle version isn't found", func(t *testing.T) {
		ctx := context.Background()
		log := testr.New(t)
		cluster := newReconcileTestCluster()
		cluster.Spec.KubernetesVersion = "non-existent"
		ctrl := gomock.NewController(t)
		k := mocks.NewMockKubectlRunner(ctrl)
		cm := mocks.NewMockChartManager(ctrl)
		bundles := createBundle(cluster)
		bundles.ObjectMeta.Name = cluster.Spec.BundlesRef.Name
		bundles.ObjectMeta.Namespace = cluster.Spec.BundlesRef.Namespace
		eksaRelease := createEKSARelease(cluster, bundles)
		objs := []runtime.Object{cluster, bundles, eksaRelease}
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

		pcc := curatedpackages.NewPackageControllerClientFullLifecycle(log, cm, k, nil)
		err := pcc.Reconcile(ctx, log, fakeClient, cluster)
		if err == nil || !strings.Contains(err.Error(), "kubernetes version non-existent") {
			t.Errorf("expected \"kubernetes version non-existent\" error, got %s", err)
		}
	})

	s.Run("errors when helm fails", func(t *testing.T) {
		ctx := context.Background()
		log := testr.New(t)
		cluster := newReconcileTestCluster()
		ctrl := gomock.NewController(t)
		k := mocks.NewMockKubectlRunner(ctrl)
		cm := mocks.NewMockChartManager(ctrl)
		bundles := createBundle(cluster)
		bundles.Spec.VersionsBundles[0].KubeVersion = string(cluster.Spec.KubernetesVersion)
		bundles.ObjectMeta.Name = cluster.Spec.BundlesRef.Name
		bundles.ObjectMeta.Namespace = cluster.Spec.BundlesRef.Namespace
		eksaRelease := createEKSARelease(cluster, bundles)
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: constants.EksaSystemNamespace,
				Name:      cluster.Name + "-kubeconfig",
			},
		}
		objs := []runtime.Object{cluster, bundles, secret, eksaRelease}
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		cm.EXPECT().InstallChart(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test error"))

		pcc := curatedpackages.NewPackageControllerClientFullLifecycle(log, cm, k, nil)
		err := pcc.Reconcile(ctx, log, fakeClient, cluster)
		if err == nil || !strings.Contains(err.Error(), "packages client error: test error") {
			t.Errorf("expected packages client error, got %s", err)
		}
	})
}

func newReconcileTestCluster() *anywherev1.Cluster {
	version := test.DevEksaVersion()
	return &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-workload-cluster",
			Namespace: "my-namespace",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "v1.25",
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "my-bundles-ref",
				Namespace: "my-namespace",
			},
			ManagementCluster: anywherev1.ManagementCluster{
				Name: "my-management-cluster",
			},
			EksaVersion: &version,
		},
	}
}

func createBundle(cluster *anywherev1.Cluster) *artifactsv1.Bundles {
	return &artifactsv1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: "default",
		},
		Spec: artifactsv1.BundlesSpec{
			VersionsBundles: []artifactsv1.VersionsBundle{
				{
					KubeVersion: "1.20",
					EksD: artifactsv1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "testdata/release.yaml",
						KubeVersion:    "1.20",
					},
					CertManager:                artifactsv1.CertManagerBundle{},
					ClusterAPI:                 artifactsv1.CoreClusterAPI{},
					Bootstrap:                  artifactsv1.KubeadmBootstrapBundle{},
					ControlPlane:               artifactsv1.KubeadmControlPlaneBundle{},
					VSphere:                    artifactsv1.VSphereBundle{},
					Docker:                     artifactsv1.DockerBundle{},
					Eksa:                       artifactsv1.EksaBundle{},
					Cilium:                     artifactsv1.CiliumBundle{},
					Kindnetd:                   artifactsv1.KindnetdBundle{},
					Flux:                       artifactsv1.FluxBundle{},
					BottleRocketHostContainers: artifactsv1.BottlerocketHostContainersBundle{},
					ExternalEtcdBootstrap:      artifactsv1.EtcdadmBootstrapBundle{},
					ExternalEtcdController:     artifactsv1.EtcdadmControllerBundle{},
					Tinkerbell:                 artifactsv1.TinkerbellBundle{},
				},
			},
		},
	}
}

func createEKSARelease(cluster *anywherev1.Cluster, bundle *artifactsv1.Bundles) *artifactsv1.EKSARelease {
	return &artifactsv1.EKSARelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-v0-0-0-dev",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: artifactsv1.EKSAReleaseSpec{
			BundlesRef: artifactsv1.BundlesRef{
				Name:      bundle.Name,
				Namespace: bundle.Namespace,
			},
		},
	}
}
