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
func TestSnowKubernetes127UbuntuAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu127()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
	runAWSIamAuthFlow(test)
}

func TestSnowKubernetes126To127AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewSnow(t, framework.WithSnowUbuntu126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube126)),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
	)
}

// Labels
func TestSnowKubernetes127UbuntuLabelsUpgradeFlow(t *testing.T) {
	provider := framework.NewSnow(t,
		framework.WithSnowWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1), api.WithLabel(key1, val1)),
		),
		framework.WithSnowWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1), api.WithLabel(key2, val2)),
		),
		framework.WithSnowUbuntu127(),
	)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(),
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val2)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val1)),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

// OIDC
func TestSnowKubernetes127OIDC(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu127()),
		framework.WithOIDC(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runOIDCFlow(test)
}

// Proxy config
func TestSnowKubernetes127UbuntuProxyConfig(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu127()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube127)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
		// TODO: provide separate Proxy Env Vars for Snow provider. Leaving VSphere for backwards compatibility
		framework.WithProxy(framework.VsphereProxyRequiredEnvVars),
	)
	runProxyConfigFlow(test)
}

// Simpleflow
func TestSnowKubernetes122SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu122()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runSimpleFlow(test)
}

func TestSnowKubernetes123SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu123()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
	)
	runSimpleFlow(test)
}

func TestSnowKubernetes124SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu124()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube124)),
	)
	runSimpleFlow(test)
}

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

// Taints
func TestSnowKubernetes127UbuntuTaintsUpgradeFlow(t *testing.T) {
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
		framework.WithSnowUbuntu127(),
	)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(),
		),
	)

	runTaintsUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker1, api.WithTaint(framework.NoExecuteTaint())),
			api.WithWorkerNodeGroup(worker2, api.WithNoTaints()),
			api.WithControlPlaneTaints([]corev1.Taint{framework.PreferNoScheduleTaint()}),
		),
	)
}

// Upgrade
func TestSnowKubernetes126To127UbuntuMultipleFieldsUpgrade(t *testing.T) {
	provider := framework.NewSnow(t, framework.WithSnowUbuntu126())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube126),
			api.WithControlPlaneCount(3),
			api.WithWorkerNodeCount(1),
		),
	)
	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube127)),
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

func TestSnowKubernetes127UbuntuRemoveWorkerNodeGroups(t *testing.T) {
	provider := framework.NewSnow(t,
		framework.WithSnowWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1)),
		),
		framework.WithSnowWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithSnowUbuntu127(),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube127),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(),
		),
	)

	runSimpleUpgradeFlow(
		test,
		v1alpha1.Kube127,
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

func TestSnowKubernetes122UbuntuTo123Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithUbuntu122(), snow.WithUbuntu123())
}

func TestSnowKubernetes123UbuntuTo124Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithUbuntu123(), snow.WithUbuntu124())
}

func TestSnowKubernetes124UbuntuTo125Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithUbuntu124(), snow.WithUbuntu125())
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

func TestSnowKubernetes122BottlerocketTo123Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocket122(), snow.WithBottlerocket123())
}

func TestSnowKubernetes123BottlerocketTo124Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocket123(), snow.WithBottlerocket124())
}

func TestSnowKubernetes124BottlerocketTo125Upgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocket124(), snow.WithBottlerocket125())
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

func TestSnowKubernetes122To123BottlerocketStaticIPUpgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocketStaticIP122(), snow.WithBottlerocketStaticIP123())
}

func TestSnowKubernetes123To124BottlerocketStaticIPUpgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocketStaticIP123(), snow.WithBottlerocketStaticIP124())
}

func TestSnowKubernetes124To125BottlerocketStaticIPUpgrade(t *testing.T) {
	snow := framework.NewSnow(t)
	test := framework.NewClusterE2ETest(t, snow)

	runSnowUpgradeTest(test, snow, snow.WithBottlerocketStaticIP124(), snow.WithBottlerocketStaticIP125())
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
		snow.WithBottlerocket124(),
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
			snow.WithBottlerocket122(),
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
			snow.WithBottlerocket123(),
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
			snow.WithBottlerocket124(),
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
			snow.WithBottlerocket125(),
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
