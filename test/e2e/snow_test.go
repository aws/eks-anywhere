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

// Proxy Config
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

// Simple Flow
func TestSnowKubernetes128SimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
	)
	runSimpleFlow(test)
}

func TestSnowKubernetes128StackedEtcdSimpleFlow(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewSnow(t, framework.WithSnowUbuntu128()),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube128)),
		framework.WithClusterFiller(api.WithStackedEtcdTopology()),
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
