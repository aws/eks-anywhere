//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	prometheusPackageName            = "prometheus"
	prometheusPackagePrefix          = "generated"
	prometheusPackageTargetNamespace = "observability"
)

func runCuratedPackagesPrometheusInstall(test *framework.ClusterE2ETest) {
	packageFullName := prometheusPackagePrefix + "-" + prometheusPackageName
	test.InstallLocalStorageProvisioner()
	test.CreateNamespace(prometheusPackageTargetNamespace)
	test.SetPackageBundleActive()
	test.ValidatePackageBundleControllerRegistry()
	test.InstallCuratedPackage(prometheusPackageName, packageFullName,
		kubeconfig.FromClusterName(test.ClusterName),
		"--set server.persistentVolume.storageClass=local-path")
	test.VerifyPrometheusPackageInstalled(packageFullName, prometheusPackageTargetNamespace)
	test.VerifyPrometheusNodeExporterStates(packageFullName, prometheusPackageTargetNamespace)
	test.VerifyPrometheusPrometheusServerStates(packageFullName, prometheusPackageTargetNamespace, "deployment")
}

func runCuratedPackagesPrometheusUpdate(test *framework.ClusterE2ETest) {
	packageFullName := prometheusPackagePrefix + "-" + prometheusPackageName

	test.InstallLocalStorageProvisioner()
	test.CreateNamespace(prometheusPackageTargetNamespace)
	test.SetPackageBundleActive()
	test.InstallCuratedPackage(prometheusPackageName, packageFullName,
		kubeconfig.FromClusterName(test.ClusterName),
		"--set server.persistentVolume.storageClass=local-path")

	test.ApplyPrometheusPackageServerStatefulSetFile(packageFullName, prometheusPackageTargetNamespace)
	test.VerifyPrometheusPackageInstalled(packageFullName, prometheusPackageTargetNamespace)
	test.VerifyPrometheusPrometheusServerStates(packageFullName, prometheusPackageTargetNamespace, "statefulset")
	test.VerifyPrometheusNodeExporterStates(packageFullName, prometheusPackageTargetNamespace)

	test.ApplyPrometheusPackageServerDeploymentFile(packageFullName, prometheusPackageTargetNamespace)
	test.VerifyPrometheusPackageInstalled(packageFullName, prometheusPackageTargetNamespace)
	test.VerifyPrometheusPrometheusServerStates(packageFullName, prometheusPackageTargetNamespace, "deployment")
	test.VerifyPrometheusNodeExporterStates(packageFullName, prometheusPackageTargetNamespace)
}

// runCuratedPackagesPrometheusCustomScrapeConfigs installs Prometheus with a job from extraScrapeConfigs
// and one from serverFiles."prometheus.yml".scrape_configs, and checks both are active next to the
// default "prometheus" job. This fails on the Prometheus chart shipped before
// the chart fix in aws/eks-anywhere-build-tooling#5791 and passes once a bundle with that fix is active.
func runCuratedPackagesPrometheusCustomScrapeConfigs(test *framework.ClusterE2ETest) {
	packageFullName := prometheusPackagePrefix + "-" + prometheusPackageName
	test.InstallLocalStorageProvisioner()
	test.CreateNamespace(prometheusPackageTargetNamespace)
	test.SetPackageBundleActive()
	test.ValidatePackageBundleControllerRegistry()
	test.InstallPrometheusPackageWithCustomScrapeConfigs(packageFullName, prometheusPackageTargetNamespace)
	test.VerifyPrometheusPackageInstalled(packageFullName, prometheusPackageTargetNamespace)
	test.VerifyPrometheusPrometheusServerStates(packageFullName, prometheusPackageTargetNamespace, "deployment")
	test.VerifyPrometheusScrapeJobsActive(packageFullName, prometheusPackageTargetNamespace,
		framework.PrometheusDefaultScrapeJob,
		framework.PrometheusExtraScrapeConfigsJob,
		framework.PrometheusServerFilesScrapeJob,
	)
}

func runCuratedPackagesPrometheusCustomScrapeConfigsFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackagesPrometheusCustomScrapeConfigs)
}

func runCuratedPackagesPrometheusInstallSimpleFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackagesPrometheusInstall)
}

func runCuratedPackagesPrometheusUpdateFlow(test *framework.ClusterE2ETest) {
	test.WithCluster(runCuratedPackagesPrometheusUpdate)
}

func runCuratedPackagesPrometheusInstallTinkerbellSimpleFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.GenerateHardwareConfig()
	test.CreateCluster(framework.WithControlPlaneWaitTimeout("20m"))
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneNoTaints, framework.ValidateControlPlaneLabels)
	runCuratedPackagesPrometheusInstall(test)
	test.DeleteCluster()
	test.ValidateHardwareDecommissioned()
}
