package curatedpackages_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	artifactsv1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed testdata/packagebundlectrl_test.yaml
var packageBundleControllerTest string

//go:embed testdata/awssecret_test.yaml
var awsSecretTest []byte

//go:embed testdata/awssecret_test_empty.yaml
var awsSecretTestEmpty []byte

//go:embed testdata/awssecret_defaultregion.yaml
var awsSecretDefaultRegion []byte

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
	eksaAccessId := "test-access-id"
	eksaAccessKey := "test-access-key"
	eksaRegion := "test-region"
	clusterName := "billy"
	registyMirror := &registrymirror.RegistryMirror{
		BaseRegistry: "1.2.3.4:443",
		NamespacedRegistryMap: map[string]string{
			constants.DefaultCoreEKSARegistry:             "1.2.3.4:443/public",
			constants.DefaultCuratedPackagesRegistryRegex: "1.2.3.4:443/private",
		},
	}
	return []*packageControllerTest{
		{
			WithT:          NewWithT(t),
			ctx:            context.Background(),
			kubectl:        k,
			chartInstaller: ci,
			command: curatedpackages.NewPackageControllerClient(
				ci, k, clusterName, kubeConfig, chart, registyMirror,
				curatedpackages.WithEksaSecretAccessKey(eksaAccessKey),
				curatedpackages.WithEksaRegion(eksaRegion),
				curatedpackages.WithEksaAccessKeyId(eksaAccessId),
				curatedpackages.WithManagementClusterName(clusterName),
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
			registryMirror: registyMirror,
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
		},
	}
}

func TestEnableCuratedPackagesSuccess(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {

		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTest, params).Return(bytes.Buffer{}, nil)
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableCuratedPackagesNoCronjob(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, tt.clusterName, tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithEksaSecretAccessKey(""),
			curatedpackages.WithEksaRegion(tt.eksaRegion),
			curatedpackages.WithEksaAccessKeyId(""),
			curatedpackages.WithManagementClusterName(tt.clusterName),
		)
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTestEmpty, params).Return(bytes.Buffer{}, fmt.Errorf("boom"))
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName, "cronjob.suspend=true"}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName, "cronjob.suspend=true"}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName, "cronjob.suspend=true"}
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCSuccess(t)).
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
			curatedpackages.WithManagementClusterName("mgmt"),
		)

		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, []byte(packageBundleControllerTest), params).Return(bytes.Buffer{}, nil)

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
		)
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		httpProxy := fmt.Sprintf("proxy.HTTP_PROXY=%s", tt.httpProxy)
		httpsProxy := fmt.Sprintf("proxy.HTTPS_PROXY=%s", tt.httpsProxy)
		noProxy := fmt.Sprintf("proxy.NO_PROXY=%s", strings.Join(tt.noProxy, "\\,"))
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName, httpProxy, httpsProxy, noProxy}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName, httpProxy, httpsProxy, noProxy}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName, httpProxy, httpsProxy, noProxy}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTest, params).Return(bytes.Buffer{}, nil)
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCSuccess(t)).
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
		)
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTest, params).Return(bytes.Buffer{}, nil)
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableCuratedPackagesFail(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(errors.New("login failed"))
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
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
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTest, params).Return(bytes.Buffer{}, nil)
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCFail(t)).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	}
}

func TestEnableCuratedPackagesSuccessWhenApplySecretFails(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTest, params).Return(bytes.Buffer{}, errors.New("failed applying secrets"))
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when secret creation fails")
		}
	}
}

func TestPbcCreationSuccess(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {

		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, []byte(packageBundleControllerTest), params).Return(bytes.Buffer{}, nil)

		err := tt.command.InstallPBCResources(tt.ctx)
		tt.Expect(err).To(BeNil())
	}
}

func TestPbcCreationFail(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, []byte(packageBundleControllerTest), params).Return(bytes.Buffer{}, errors.New("creating pbc"))

		err := tt.command.InstallPBCResources(tt.ctx)
		tt.Expect(err).NotTo(BeNil())
	}
}

func TestEnableCuratedPackagesSuccessWhenCronJobFails(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTest, params).Return(bytes.Buffer{}, nil)
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCSuccess(t)).
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

func TestDefaultEksaRegionSetWhenNoRegionSpecified(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretDefaultRegion, params).Return(bytes.Buffer{}, nil)
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()

		tt.command = curatedpackages.NewPackageControllerClient(
			tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.chart,
			tt.registryMirror,
			curatedpackages.WithEksaRegion(""),
			curatedpackages.WithEksaAccessKeyId(tt.eksaAccessID),
			curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
			curatedpackages.WithManagementClusterName(tt.clusterName),
		)
		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when cron job fails")
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
		)
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTest, params).Return(bytes.Buffer{}, nil)
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCSuccess(t)).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		if err != nil {
			t.Errorf("Install Controller Should succeed when installation passes")
		}
	}
}

func TestEnableCuratedPackagesActiveBundleWaitLoops(t *testing.T) {
	for _, tt := range newPackageControllerTests(t) {
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTest, params).Return(bytes.Buffer{}, nil)
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCLoops(t, 3)).
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
		)
		var values []string
		clusterName := fmt.Sprintf("clusterName=%s", "billy")
		ociURI := fmt.Sprintf("%s%s", "oci://", tt.registryMirror.ReplaceRegistry(tt.chart.Image()))
		if tt.registryMirror != nil {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			defaultRegistry := fmt.Sprintf("defaultRegistry=%s/eks-anywhere", tt.registryMirror.CoreEKSAMirror())
			gatedOCINamespace := tt.registryMirror.CuratedPackagesMirror()
			if gatedOCINamespace == "" {
				values = []string{sourceRegistry, defaultRegistry, clusterName}
			} else {
				defaultImageRegistry := fmt.Sprintf("defaultImageRegistry=%s", gatedOCINamespace)
				values = []string{sourceRegistry, defaultRegistry, defaultImageRegistry, clusterName}
			}
		} else {
			sourceRegistry := fmt.Sprintf("sourceRegistry=%s", curatedpackages.GetRegistry(tt.chart.Image()))
			values = []string{sourceRegistry, clusterName}
		}
		params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
		tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, awsSecretTest, params).Return(bytes.Buffer{}, nil)
		tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chart.Name, ociURI, tt.chart.Tag(), tt.kubeConfig, "", values).Return(nil)
		any := gomock.Any()
		tt.kubectl.EXPECT().
			GetObject(any, any, any, any, any, any).
			DoAndReturn(getPBCDelay(t, time.Second)).
			AnyTimes()

		err := tt.command.EnableCuratedPackages(tt.ctx)
		expectedErr := fmt.Errorf("timed out finding an active package bundle for the current cluster: %v", context.DeadlineExceeded)
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
