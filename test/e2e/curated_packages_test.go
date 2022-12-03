//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	EksaPackagesNamespace              = "eksa-packages"
	EksaPackageControllerHelmChartName = "eks-anywhere-packages"
	EksaPackagesSourceRegistry         = "public.ecr.aws/l0g8r8j6"
	EksaPackageControllerHelmURI       = "oci://" + EksaPackagesSourceRegistry + "/eks-anywhere-packages"
	EksaPackageControllerHelmVersion   = "0.2.20-eks-a-v0.0.0-dev-build.4894"
	EksaPackageBundleURI               = "oci://" + EksaPackagesSourceRegistry + "/eks-anywhere-packages-bundles"
)

var EksaPackageControllerHelmValues = []string{"sourceRegistry=public.ecr.aws/l0g8r8j6"}

type resourcePredicate func(string, error) bool

func NoErrorPredicate(_ string, err error) bool {
	return err == nil
}

// TODO turn them into generics using comparable once 1.18 is allowed.
func StringMatchPredicate(s string) resourcePredicate {
	return func(in string, err error) bool {
		return err == nil && strings.Compare(s, in) == 0
	}
}

func IntEqualPredicate(i int) resourcePredicate {
	return func(in string, err error) bool {
		inInt, err := strconv.Atoi(in)
		return err == nil && inInt == i
	}
}

func WaitForResource(
	test *framework.ClusterE2ETest,
	ctx context.Context,
	resource string,
	namespace string,
	jsonpath string,
	timeout time.Duration,
	predicates ...resourcePredicate,
) error {
	end := time.Now().Add(timeout)
	if !strings.HasPrefix(jsonpath, "jsonpath") {
		jsonpath = fmt.Sprintf("jsonpath='%s'", jsonpath)
	}
	command := []string{}
	for time.Now().Before(end) {
		out, err := test.KubectlClient.Execute(ctx, "get", "-n", namespace,
			"--kubeconfig="+kubeconfig.FromClusterName(test.ClusterName),
			"-o", jsonpath, resource)
		outStr := out.String()
		trimmed := strings.Trim(outStr, "'")
		allPredicates := true
		for _, f := range predicates {
			allPredicates = allPredicates && f(trimmed, err)
		}
		if allPredicates {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf(
		"timed out waiting for resource: %s [namespace: %s, jsonpath: %s, timeout: %s]",
		command,
		namespace,
		jsonpath,
		timeout,
	)
}

func WaitForDaemonset(
	test *framework.ClusterE2ETest,
	ctx context.Context,
	daemonsetName string,
	namespace string,
	numberOfNodes int,
	timeout time.Duration,
) error {
	return WaitForResource(
		test,
		ctx,
		"daemonset/"+daemonsetName,
		namespace,
		"{.status.numberAvailable}",
		timeout,
		NoErrorPredicate,
		IntEqualPredicate(numberOfNodes),
	)
}

// Hackish way to get the latest bundle. This assumes no bundle is created outside of the normal PBC bundle fetch timer.
// This should be modified to get the bundle from the previous build step and use that only.
func WaitForLatestBundleToBeAvailable(
	test *framework.ClusterE2ETest,
	ctx context.Context,
	timeout time.Duration,
) error {
	return WaitForResource(test, ctx, "packagebundle", "eksa-packages", "{.items[0]}", timeout, NoErrorPredicate)
}

func WaitForPackageToBeInstalled(
	test *framework.ClusterE2ETest,
	ctx context.Context,
	packageName string,
	timeout time.Duration,
) error {
	//--for=jsonpath isn't supported in v1.22. Update once it's supported
	//_, err = test.KubectlClient.Execute(
	//    ctx, "wait", "--timeout", "1m",
	//    "--for", "jsonpath='{.status.state}'=installed",
	//    "package", packagePrefix, "--kubeconfig", kubeconfig,
	//    "-n", "eksa-packages",
	//)
	return WaitForResource(
		test,
		ctx,
		"package/"+packageName,
		EksaPackagesNamespace,
		"{.status.state}",
		timeout,
		NoErrorPredicate,
		StringMatchPredicate("installed"),
	)
}

func UpgradePackages(test *framework.ClusterE2ETest, bundleVersion string) {
	test.RunEKSA([]string{"upgrade", "packages", "--bundle-version=" + bundleVersion})
}

func GetLatestBundleFromCluster(test *framework.ClusterE2ETest) (string, error) {
	bundleBytes, err := test.KubectlClient.ExecuteCommand(
		context.Background(),
		"get",
		"packagebundle",
		"-n", "eksa-packages",
		"--kubeconfig="+kubeconfig.FromClusterName(test.ClusterName),
		"-o", "jsonpath='{.items[0].metadata.name}'",
	)
	if err != nil {
		return "", err
	}
	bundle := bundleBytes.String()
	return strings.Trim(bundle, "'"), nil
}

// packageBundleURI uses a KubernetesVersion argument to complete a package
// bundle URI by adding the approprate tag.
func packageBundleURI(version v1alpha1.KubernetesVersion) string {
	tag := "v" + strings.Replace(string(version), ".", "-", 1) + "-latest"
	return fmt.Sprintf("%s:%s", EksaPackageBundleURI, tag)
}

func withMgmtCluster(cluster *framework.ClusterE2ETest) *types.Cluster {
	return &types.Cluster{
		Name:               cluster.ClusterName,
		KubeconfigFile:     filepath.Join(cluster.ClusterName, fmt.Sprintf("%s-eks-a-cluster.kubeconfig", cluster.ClusterName)),
		ExistingManagement: true,
	}
}

func setupSimpleMultiCluster(t *testing.T, provider framework.Provider, kubeVersion v1alpha1.KubernetesVersion) *framework.MulticlusterE2ETest {
	test := framework.NewMulticlusterE2ETest(
		t,
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(kubeVersion),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
		framework.NewClusterE2ETest(
			t,
			provider,
			framework.WithClusterFiller(
				api.WithKubernetesVersion(kubeVersion),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
		),
	)
	return test
}

func runCuratedPackageInstall(test *framework.ClusterE2ETest) {
	test.SetPackageBundleActive()
	packageName := "hello-eks-anywhere"
	packagePrefix := "test"
	packageFile := test.BuildPackageConfigFile(packageName, packagePrefix, EksaPackagesNamespace)
	test.InstallCuratedPackageFile(packageFile, kubeconfig.FromClusterName(test.ClusterName))
	test.VerifyHelloPackageInstalled(packagePrefix+"-"+packageName, withMgmtCluster(test))
}

func runCuratedPackageRemoteClusterInstallSimpleFlow(test *framework.MulticlusterE2ETest) {
	test.CreateManagementCluster()
	test.RunInWorkloadClusters(func(e *framework.WorkloadCluster) {
		e.GenerateClusterConfig()
		e.CreateCluster()
		e.VerifyPackageControllerNotInstalled()
		test.ManagementCluster.SetPackageBundleActive()
		packageName := "hello-eks-anywhere"
		packagePrefix := "test"
		packageFile := e.BuildPackageConfigFile(packageName, packagePrefix, EksaPackagesNamespace)
		test.ManagementCluster.InstallCuratedPackageFile(packageFile, kubeconfig.FromClusterName(test.ManagementCluster.ClusterName))
		e.VerifyHelloPackageInstalled(packagePrefix+"-"+packageName, withMgmtCluster(test.ManagementCluster))
		e.DeleteCluster()
	})
	time.Sleep(5 * time.Minute)
	test.DeleteManagementCluster()
}

func runCuratedPackageInstallTinkerbellSingleNodeFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.PowerOnHardware()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneNoTaints, framework.ValidateControlPlaneLabels)
	runCuratedPackageInstall(test)
	test.DeleteCluster()
	test.PowerOffHardware()
	test.ValidateHardwareDecommissioned()
}

func runCuratedPackageInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackageInstall)
}

// There are many tests here, each covers a different combination described in
// the matrix found in
// https://github.com/aws/eks-anywhere-packages/issues/96. They're each named
// according to the columns of that matrix, that is,
// "TestCPackages<Provider><OS><K8s ver>SimpleFlow". Better organization,
// whether via test suites, testing tables, or other functionality is welcome,
// but this is a simple solution for now, without having to make any major
// decisions about test packages or methodologies, right now.

func TestCPackagesDockerUbuntuKubernetes121SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesDockerUbuntuKubernetes122SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesDockerUbuntuKubernetes123SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesDockerUbuntuKubernetes124SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes121SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes121BottleRocketSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes122SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes122BottleRocketSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes123SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes123BottleRocketSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes124SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithUbuntu124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes124BottleRocketSimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(t,
		framework.NewVSphere(t, framework.WithBottleRocket124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes121UbuntuWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu121())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube121)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes121BottleRocketWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket121())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube121)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes122UbuntuWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu122())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube122)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes122BottleRocketWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket122())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube122)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes123UbuntuWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu123())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube123)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes123BottleRocketWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket123())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube123)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes124UbuntuWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithUbuntu124())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube124)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesVSphereKubernetes124BottleRocketWorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewVSphere(t, framework.WithBottleRocket124())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube124)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesTinkerbellUbuntuKubernetes122SingleNodeFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu122Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube122),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)

	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestCPackagesTinkerbellUbuntuKubernetes123SingleNodeFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithUbuntu123Tinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube123),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestCPackagesTinkerbellBottleRocketKubernetes122SingleNodeFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube122),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)

	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestCPackagesTinkerbellBottleRocketKubernetes123SingleNodeFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewTinkerbell(t, framework.WithBottleRocketTinkerbell()),
		framework.WithClusterSingleNode(v1alpha1.Kube123),
		framework.WithControlPlaneHardware(1),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)

	runCuratedPackageInstallTinkerbellSingleNodeFlow(test)
}

func TestCPackagesCloudStackRedhatKubernetes121SimpleFlow(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t, framework.WithCloudStackRedhat121()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			"my-packages-test", EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesCloudStackRedhatKubernetes121WorkloadCluster(t *testing.T) {
	framework.CheckCuratedPackagesCredentials(t)
	provider := framework.NewCloudStack(t, framework.WithCloudStackRedhat121())
	test := setupSimpleMultiCluster(t, provider, v1alpha1.Kube121)
	runCuratedPackageRemoteClusterInstallSimpleFlow(test)
}

func TestCPackagesNutanixKubernetes121SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu121Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube121),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesNutanixKubernetes122SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu122Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube122),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesNutanixKubernetes123SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu123Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube123),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}

func TestCPackagesNutanixKubernetes124SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewNutanix(t, framework.WithUbuntu124Nutanix()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
		framework.WithPackageConfig(t, packageBundleURI(v1alpha1.Kube124),
			EksaPackageControllerHelmChartName, EksaPackageControllerHelmURI,
			EksaPackageControllerHelmVersion, EksaPackageControllerHelmValues),
	)
	runCuratedPackageInstallSimpleFlow(test)
}
