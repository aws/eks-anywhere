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

func runCuratedPackageInstall(test *framework.ClusterE2ETest) {
	test.SetPackageBundleActive()
	test.GenerateSupportBundleOnCleanupIfTestFailed()
	err := WaitForPackageToBeInstalled(test, context.Background(), "eks-anywhere-packages", 3*time.Minute)
	if err != nil {
		test.T.Fatalf("packages controller not in installed state: %s", err)
	}
	err = WaitForPackageToBeInstalled(test, context.Background(), "eks-anywhere-packages-crds", 3*time.Minute)
	if err != nil {
		test.T.Fatalf("packages controller crds not in installed state: %s", err)
	}
	packageName := "hello-eks-anywhere"
	packagePrefix := "test"
	packageFile := test.BuildPackageConfigFile(packageName, packagePrefix, EksaPackagesNamespace)
	test.InstallCuratedPackageFile(packageFile, kubeconfig.FromClusterName(test.ClusterName))
	test.VerifyHelloPackageInstalled(packagePrefix+"-"+packageName, withMgmtCluster(test))
}

func runCuratedPackageInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackageInstall)
}

func runDisabledCuratedPackage(test *framework.ClusterE2ETest) {
	test.ValidatingNoPackageController()
}

func runDisabledCuratedPackageInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runDisabledCuratedPackage)
}

func runCuratedPackageRemoteClusterInstallSimpleFlow(test *framework.MulticlusterE2ETest) {
	test.CreateManagementClusterWithConfig()
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

type resourcePredicate func(string, error) bool

func NoErrorPredicate(_ string, err error) bool {
	return err == nil
}

// TODO turn them into generics using comparable once 1.18 is allowed
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
		fmt.Sprintf("%s-%s", EksaPackagesNamespace, test.ClusterName),
		"{.status.state}",
		timeout,
		NoErrorPredicate,
		StringMatchPredicate("installed"),
	)
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
	}
}

func SetupSimpleMultiCluster(t *testing.T, provider framework.Provider, kubeVersion v1alpha1.KubernetesVersion) *framework.MulticlusterE2ETest {
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
