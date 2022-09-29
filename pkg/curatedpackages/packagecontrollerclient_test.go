package curatedpackages_test

// import (
// 	"context"
// 	_ "embed"
// 	"testing"

// 	"github.com/golang/mock/gomock"
// 	. "github.com/onsi/gomega"

// 	"github.com/aws/eks-anywhere/pkg/curatedpackages"
// 	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
// )

// const (
// 	cronJobName = "cronjob/cron-ecr-renew"
// 	jobName     = "eksa-auth-refresher"
// )

// type packageControllerTest struct {
// 	*WithT
// 	ctx            context.Context
// 	kubectl        *mocks.MockKubectlRunner
// 	chartInstaller *mocks.MockChartInstaller
// 	command        *curatedpackages.PackageControllerClient
// 	clusterName    string
// 	kubeConfig     string
// 	ociUri         string
// 	chartName      string
// 	chartVersion   string
// 	eksaAccessId   string
// 	eksaAccessKey  string
// 	eksaRegion     string
// 	httpProxy      string
// 	httpsProxy     string
// 	noProxy        []string
// }

// func newPackageControllerTest(t *testing.T) *packageControllerTest {
// 	ctrl := gomock.NewController(t)
// 	k := mocks.NewMockKubectlRunner(ctrl)
// 	ci := mocks.NewMockChartInstaller(ctrl)
// 	kubeConfig := "kubeconfig.kubeconfig"
// 	uri := "test/registry_name"
// 	chartName := "test_controller"
// 	chartVersion := "v1.0.0"
// 	eksaAccessId := "test-access-id"
// 	eksaAccessKey := "test-access-key"
// 	eksaRegion := "test-region"
// 	clusterName := "billy"
// 	return &packageControllerTest{
// 		WithT:          NewWithT(t),
// 		ctx:            context.Background(),
// 		kubectl:        k,
// 		chartInstaller: ci,
// 		command: curatedpackages.NewPackageControllerClient(
// 			ci, k, clusterName, kubeConfig, uri, chartName, chartVersion,
// 			curatedpackages.WithEksaSecretAccessKey(eksaAccessKey),
// 			curatedpackages.WithEksaRegion(eksaRegion),
// 			curatedpackages.WithEksaAccessKeyId(eksaAccessId),
// 		),
// 		clusterName:   clusterName,
// 		kubeConfig:    kubeConfig,
// 		ociUri:        uri,
// 		chartName:     chartName,
// 		chartVersion:  chartVersion,
// 		eksaAccessId:  eksaAccessId,
// 		eksaAccessKey: eksaAccessKey,
// 		eksaRegion:    eksaRegion,
// 		httpProxy:     "1.1.1.1",
// 		httpsProxy:    "1.1.1.1",
// 		noProxy:       []string{"1.1.1.1/24"},
// 	}
// }

// func TestInstallControllerSuccess(t *testing.T) {
// 	tt := newPackageControllerTest(t)

// 	registry := curatedpackages.GetRegistry(tt.ociUri)
// 	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
// 	clusterName := fmt.Sprintf("clusterName=%s", "billy")
// 	values := []string{sourceRegistry, clusterName}
// 	params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
// 	dat, err := os.ReadFile("testdata/awssecret_test.yaml")
// 	tt.Expect(err).NotTo(HaveOccurred())
// 	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, dat, params).Return(bytes.Buffer{}, nil)
// 	params = []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName}
// 	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)
// 	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(nil)

// 	err = tt.command.InstallController(tt.ctx)
// 	if err != nil {
// 		t.Errorf("Install Controller Should succeed when installation passes")
// 	}
// }

// func TestInstallControllerWithProxy(t *testing.T) {
// 	tt := newPackageControllerTest(t)
// 	tt.command = curatedpackages.NewPackageControllerClient(
// 		tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.ociUri, tt.chartName, tt.chartVersion,
// 		curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
// 		curatedpackages.WithEksaRegion(tt.eksaRegion),
// 		curatedpackages.WithEksaAccessKeyId(tt.eksaAccessId),
// 		curatedpackages.WithHTTPProxy(tt.httpProxy),
// 		curatedpackages.WithHTTPSProxy(tt.httpsProxy),
// 		curatedpackages.WithNoProxy(tt.noProxy),
// 	)

// 	registry := curatedpackages.GetRegistry(tt.ociUri)
// 	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
// 	clusterName := fmt.Sprintf("clusterName=%s", "billy")
// 	httpProxy := fmt.Sprintf("proxy.HTTP_PROXY=%s", tt.httpProxy)
// 	httpsProxy := fmt.Sprintf("proxy.HTTPS_PROXY=%s", tt.httpsProxy)
// 	noProxy := fmt.Sprintf("proxy.NO_PROXY=%s", strings.Join(tt.noProxy, "\\,"))

// 	values := []string{sourceRegistry, clusterName, httpProxy, httpsProxy, noProxy}
// 	params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
// 	dat, err := os.ReadFile("testdata/awssecret_test.yaml")
// 	tt.Expect(err).NotTo(HaveOccurred())
// 	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, dat, params).Return(bytes.Buffer{}, nil)
// 	params = []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName}
// 	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)
// 	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(nil)

// 	err = tt.command.InstallController(tt.ctx)
// 	if err != nil {
// 		t.Errorf("Install Controller Should succeed when installation passes")
// 	}
// }

// func TestInstallControllerWithEmptyProxy(t *testing.T) {
// 	tt := newPackageControllerTest(t)
// 	tt.command = curatedpackages.NewPackageControllerClient(
// 		tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.ociUri, tt.chartName, tt.chartVersion,
// 		curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
// 		curatedpackages.WithEksaRegion(tt.eksaRegion),
// 		curatedpackages.WithEksaAccessKeyId(tt.eksaAccessId),
// 		curatedpackages.WithHTTPProxy(""),
// 		curatedpackages.WithHTTPSProxy(""),
// 		curatedpackages.WithNoProxy(nil),
// 	)

// 	registry := curatedpackages.GetRegistry(tt.ociUri)
// 	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
// 	clusterName := fmt.Sprintf("clusterName=%s", "billy")

// 	values := []string{sourceRegistry, clusterName}
// 	params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
// 	dat, err := os.ReadFile("testdata/awssecret_test.yaml")
// 	tt.Expect(err).NotTo(HaveOccurred())
// 	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, dat, params).Return(bytes.Buffer{}, nil)
// 	params = []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName}
// 	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)
// 	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(nil)

// 	err = tt.command.InstallController(tt.ctx)
// 	if err != nil {
// 		t.Errorf("Install Controller Should succeed when installation passes")
// 	}
// }

// func TestInstallControllerFail(t *testing.T) {
// 	tt := newPackageControllerTest(t)
// 	registry := curatedpackages.GetRegistry(tt.ociUri)
// 	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
// 	clusterName := fmt.Sprintf("clusterName=%s", "billy")
// 	values := []string{sourceRegistry, clusterName}

// 	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(errors.New("login failed"))

// 	err := tt.command.InstallController(tt.ctx)
// 	if err == nil {
// 		t.Errorf("Install Controller Should fail when installation fails")
// 	}
// }

// func TestInstallControllerSuccessWhenApplySecretFails(t *testing.T) {
// 	tt := newPackageControllerTest(t)

// 	registry := curatedpackages.GetRegistry(tt.ociUri)
// 	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
// 	clusterName := fmt.Sprintf("clusterName=%s", "billy")
// 	values := []string{sourceRegistry, clusterName}
// 	params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
// 	dat, err := os.ReadFile("testdata/awssecret_test.yaml")
// 	tt.Expect(err).To(BeNil())
// 	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, dat, params).Return(bytes.Buffer{}, errors.New("error applying secrets"))
// 	params = []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName}
// 	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, nil)
// 	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(nil)

// 	err = tt.command.InstallController(tt.ctx)
// 	if err != nil {
// 		t.Errorf("Install Controller Should succeed when secret creation fails")
// 	}
// }

// func TestInstallControllerSuccessWhenCronJobFails(t *testing.T) {
// 	tt := newPackageControllerTest(t)

// 	registry := curatedpackages.GetRegistry(tt.ociUri)
// 	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
// 	clusterName := fmt.Sprintf("clusterName=%s", "billy")
// 	values := []string{sourceRegistry, clusterName}
// 	params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
// 	dat, err := os.ReadFile("testdata/awssecret_test.yaml")
// 	tt.Expect(err).To(BeNil())
// 	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, dat, params).Return(bytes.Buffer{}, nil)
// 	params = []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName}
// 	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, errors.New("error creating cron job"))
// 	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(nil)

// 	err = tt.command.InstallController(tt.ctx)
// 	if err != nil {
// 		t.Errorf("Install Controller Should succeed when cron job fails")
// 	}
// }

// func successGetObject(t *testing.T) func(context.Context, string, string, string, string, runtime.Object) error {
// 	return func(_ context.Context, _, _, _, _ string, obj runtime.Object) error {
// 		o := packagesv1.PackageBundleController{
// 			Spec: packagesv1.PackageBundleControllerSpec{
// 				ActiveBundle: "non-empty",
// 			},
// 		}
// 		pbc, ok := obj.(*packagesv1.PackageBundleController)
// 		if !ok {
// 			t.Fatalf("obj is not a package bundle controller")
// 		}
// 		o.DeepCopyInto(pbc)
// 		return nil
// 	}
// }

// func errorGetObject(t *testing.T) func(context.Context, string, string, string, string, runtime.Object) error {
// 	return func(_ context.Context, _, _, _, _ string, obj runtime.Object) error {
// 		return fmt.Errorf("test error")
// 	}
// }

// func TestGetActiveControllerSuccess(t *testing.T) {
// 	tt := newPackageControllerTest(t)

// 	tt.kubectl.EXPECT().
// 		HasResource(tt.ctx, "packageBundleController", tt.clusterName, tt.kubeConfig, packagesv1.PackageNamespace).
// 		Return(true, nil)
// 	tt.kubectl.EXPECT().
// 		GetObject(gomock.Any(), "packageBundleController", tt.clusterName, tt.kubeConfig, packagesv1.PackageNamespace, gomock.Any()).
// 		DoAndReturn(successGetObject(t))

// 	found, err := tt.command.IsInstalled(tt.ctx)
// 	if err != nil || !found {
// 		t.Errorf("Get Active Controller should return true when controller exists")
// 	}
// }

// func TestGetActiveControllerFail(t *testing.T) {
// 	tt := newPackageControllerTest(t)

// 	tt.kubectl.EXPECT().HasResource(tt.ctx, "packageBundleController", tt.clusterName, tt.kubeConfig, constants.EksaPackagesName).Return(true, nil)
// 	tt.kubectl.EXPECT().GetObject(gomock.Any(), "packageBundleController",
// 		tt.clusterName, tt.kubeConfig, constants.EksaPackagesName, gomock.Any()).
// 		DoAndReturn(errorGetObject(t))

// 	found, err := tt.command.IsInstalled(tt.ctx)
// 	if found {
// 		t.Fatalf("expected false, got %t", found)
// 	}
// 	if err == nil {
// 		t.Fatalf("expected error, got %v", err)
// 	}
// }

// func TestDefaultEksaRegionSetWhenNoRegionSpecified(t *testing.T) {
// 	tt := newPackageControllerTest(t)

// 	registry := curatedpackages.GetRegistry(tt.ociUri)
// 	sourceRegistry := fmt.Sprintf("sourceRegistry=%s", registry)
// 	clusterName := fmt.Sprintf("clusterName=%s", "billy")
// 	values := []string{sourceRegistry, clusterName}
// 	params := []string{"create", "-f", "-", "--kubeconfig", tt.kubeConfig}
// 	dat, err := os.ReadFile("testdata/awssecret_defaultregion.yaml")
// 	tt.Expect(err).To(BeNil())
// 	tt.kubectl.EXPECT().ExecuteFromYaml(tt.ctx, dat, params).Return(bytes.Buffer{}, nil)
// 	params = []string{"create", "job", jobName, "--from=" + cronJobName, "--kubeconfig", tt.kubeConfig, "--namespace", constants.EksaPackagesName}
// 	tt.kubectl.EXPECT().ExecuteCommand(tt.ctx, params).Return(bytes.Buffer{}, errors.New("error creating cron job"))
// 	tt.chartInstaller.EXPECT().InstallChart(tt.ctx, tt.chartName, "oci://"+tt.ociUri, tt.chartVersion, tt.kubeConfig, values).Return(nil)

// 	tt.command = curatedpackages.NewPackageControllerClient(
// 		tt.chartInstaller, tt.kubectl, "billy", tt.kubeConfig, tt.ociUri, tt.chartName, tt.chartVersion,
// 		curatedpackages.WithEksaRegion(""),
// 		curatedpackages.WithEksaAccessKeyId(tt.eksaAccessId),
// 		curatedpackages.WithEksaSecretAccessKey(tt.eksaAccessKey),
// 	)
// 	err = tt.command.InstallController(tt.ctx)
// 	if err != nil {
// 		t.Errorf("Install Controller Should succeed when cron job fails")
// 	}
// }
