//go:build e2e && (snow || all_providers)
// +build e2e
// +build snow all_providers

package e2e

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

// AWS IAM Auth
func TestSnowKubernetes128UbuntuAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu128()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runAWSIamAuthFlow(test)
}

func TestSnowKubernetes127To128AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewSnow(t, framework.WithSnowUbuntu127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
}

func TestSnowKubernetes127To128UbuntuManagementCPUpgradeAPI(t *testing.T) {
	provider := framework.NewSnow(t, framework.WithSnowUbuntu127())
	test := framework.NewClusterE2ETest(
		t, provider,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
		),
	)
	runUpgradeFlowWithAPI(
		test,
		api.ClusterToConfigFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(3),
		),
		provider.WithUbuntu128(),
	)
}

// Labels
func TestSnowKubernetes128UbuntuLabelsUpgradeFlow(t *testing.T) {
	provider := framework.NewSnow(t,
		framework.WithSnowWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1), api.WithLabel(key1, val1)),
		),
		framework.WithSnowWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1), api.WithLabel(key2, val2)),
		),
		framework.WithSnowUbuntu128(),
	)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(),
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val2)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val1)),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

// OIDC
func TestSnowKubernetes128OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu128()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

// Proxy config
func TestSnowKubernetes128UbuntuProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		// TODO: provide separate Proxy Env Vars for Snow provider. Leaving VSphere for backwards compatibility
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

// Simpleflow
func TestSnowKubernetes125SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu125()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube125)),
	)
	runSimpleFlow(test)
}

func TestSnowKubernetes126SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu126()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runSimpleFlow(test)
}

func TestSnowKubernetes127SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runSimpleFlow(test)
}

func TestSnowKubernetes128SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

// Taints
func TestSnowKubernetes128UbuntuTaintsUpgradeFlow(t *testing.T) {
	provider := framework.NewSnow(t,
		framework.WithSnowWorkerNodeGroup(
			worker0,
			framework.NoScheduleWorkerNodeGroup(worker0, 1),
		),
		framework.WithSnowWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithSnowWorkerNodeGroup(
			worker2,
			framework.PreferNoScheduleWorkerNodeGroup(worker2, 1),
		),
		framework.WithSnowUbuntu128(),
	)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(),
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

// Upgrade
func TestSnowKubernetes127To128UbuntuMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewSnow(t, framework.WithSnowUbuntu127())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(1),
		),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(2),
		),
		provider.WithProviderUpgrade(
			api.WithSnowInstanceTypeForAllMachines("sbe-c.xlarge"),
			api.WithSnowPhysicalNetworkConnectorForAllMachines(v1alpha1.QSFP),
		),
	)
}

func TestSnowKubernetes128UbuntuRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewSnow(t,
		framework.WithSnowWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1)),
		),
		framework.WithSnowWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithSnowUbuntu128(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube128),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube128,
		framework.WithClusterUpgrade(
			api.RemoveWorkerNodeGroup(worker1),
			api.WithWorkerNodeGroup(worker0, api.WithCount(1)),
		),
		provider.WithNewSnowWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(
				worker2,
				api.WithCount(1),
			),
		),
	)
}

func TestSnowKubernetes127UbuntuTo128Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithUbuntu127(), snow.WithUbuntu128())
}

func TestSnowKubernetes125UbuntuTo126Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithUbuntu125(), snow.WithUbuntu126())
}

func TestSnowKubernetes126UbuntuTo127Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithUbuntu126(), snow.WithUbuntu127())
}

func TestSnowKubernetes127BottlerocketTo128Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocket127(), snow.WithBottlerocket128())
}

func TestSnowKubernetes125BottlerocketTo126Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocket125(), snow.WithBottlerocket126())
}

func TestSnowKubernetes126BottlerocketTo127Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocket126(), snow.WithBottlerocket127())
}

func TestSnowKubernetes127To128BottlerocketStaticIPUpgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocketStaticIP127(), snow.WithBottlerocketStaticIP128())
}

func TestSnowKubernetes125To126BottlerocketStaticIPUpgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocketStaticIP125(), snow.WithBottlerocketStaticIP126())
}

func TestSnowKubernetes126To127BottlerocketStaticIPUpgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocketStaticIP126(), snow.WithBottlerocketStaticIP127())
}

// Workload API
func TestSnowMulticlusterWorkloadClusterAPI(t *testing.T) {
	snow := framework.NewSnow(t)
	managementCluster := framework.NewClusterE2ETest(
		t, snow,
	).WithClusterConfig(
		api.ClusterToConfigFiller(
			api.WithControlPlaneCount(1),
			api.WithWorkerNodeCount(1),
			api.WithStackedEtcdTopology(),
		),
		snow.WithBottlerocket126(),
	)
	test := framework.NewMulticlusterE2ETest(t, managementCluster)
	test.WithWorkloadClusters(
		framework.NewClusterE2ETest(
			t, snow, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			snow.WithBottlerocket125(),
		),
		framework.NewClusterE2ETest(
			t, snow, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			snow.WithBottlerocket126(),
		),
		framework.NewClusterE2ETest(
			t, snow, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			snow.WithBottlerocket127(),
		),
		framework.NewClusterE2ETest(
			t, snow, framework.WithClusterName(test.NewWorkloadClusterName()),
		).WithClusterConfig(
			api.ClusterToConfigFiller(
				api.WithManagementCluster(managementCluster.ClusterName),
				api.WithControlPlaneCount(1),
				api.WithWorkerNodeCount(1),
				api.WithStackedEtcdTopology(),
			),
			snow.WithBottlerocket128(),
		),
	)
	test.CreateManagementCluster()
	test.RunConcurrentlyInWorkloadClusters(func(wc *framework.WorkloadCluster) {
		wc.ApplyClusterManifest()
		wc.DeleteClusterWithKubectl()
	})
	test.ManagementCluster.StopIfFailed()
	test.DeleteManagementCluster()
}
