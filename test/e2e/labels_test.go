//go:build e2e
// +build e2e

package e2e

import (
	"github.com/aws/eks-anywhere/pkg/constants"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

const (
	key1   = framework.LabelPrefix + "/" + "key1"
	key2   = framework.LabelPrefix + "/" + "key2"
	cpKey1 = framework.LabelPrefix + "/" + "cp-key1"
	val1   = "val1"
	val2   = "val2"
	cpVal1 = "cp-val1"
)

func runLabelsUpgradeFlow(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeLabels)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneLabels)
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeLabels)
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneLabels)
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestDockerKubernetes123Labels(t *testing.T) {
	provider := framework.NewDocker(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
			api.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
			api.WithWorkerNodeGroup(worker1, api.WithCount(1)),
			api.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestVSphereKubernetes123Labels(t *testing.T) {
	provider := ubuntu123ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestVSphereKubernetes123LabelsBottlerocket(t *testing.T) {
	provider := bottlerocket123ProviderWithLabels(t)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithExternalEtcdTopology(1),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(), // This gives us a blank slate
		),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val1)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val2)),
			api.WithWorkerNodeGroup(worker2),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
	)
}

func TestCloudStackKubernetes121LabelsRedhat(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewCloudStack(t,
			framework.WithCloudStackRedhat121(),
		),
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube121),
			api.WithControlPlaneLabel(framework.FailureDomainLabel, framework.FailureDomainPlaceholder),
			api.WithWorkerNodeGroup(constants.DefaultWorkerNodeGroupName,
				api.WithCount(1),
				api.WithLabel(framework.FailureDomainLabel, framework.FailureDomainPlaceholder),
			),
		),
	)
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateControlPlaneNodes(framework.ValidateControlPlaneFailureDomainLabels)
	test.ValidateWorkerNodes(framework.ValidateWorkerNodeFailureDomainLabels)
	test.ValidateCluster(v1alpha1.Kube121)
	test.DeleteCluster()
}

func TestSnowKubernetes123LabelsUbuntu(t *testing.T) {
	provider := framework.NewSnow(t,
		framework.WithSnowWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(1), api.WithLabel(key1, val1)),
		),
		framework.WithSnowWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1), api.WithLabel(key2, val2)),
		),
		framework.WithSnowUbuntu123(),
	)

	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(
			api.WithKubernetesVersion(v1alpha1.Kube123),
			api.WithControlPlaneCount(1),
			api.RemoveAllWorkerNodeGroups(),
		),
		framework.WithEnvVar("SNOW_PROVIDER", "true"),
		framework.WithEnvVar("FULL_LIFECYCLE_API", "true"),
	)

	runLabelsUpgradeFlow(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(
			api.WithWorkerNodeGroup(worker0, api.WithLabel(key1, val2)),
			api.WithWorkerNodeGroup(worker1, api.WithLabel(key2, val1)),
			api.WithControlPlaneLabel(cpKey1, cpVal1),
		),
		framework.WithEnvVar("SNOW_PROVIDER", "true"),
		framework.WithEnvVar("FULL_LIFECYCLE_API", "true"),
	)
}

func ubuntu123ProviderWithLabels(t *testing.T) *framework.VSphere {
	return framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithUbuntu123(),
	)
}

func bottlerocket123ProviderWithLabels(t *testing.T) *framework.VSphere {
	return framework.NewVSphere(t,
		framework.WithVSphereWorkerNodeGroup(
			worker0,
			framework.WithWorkerNodeGroup(worker0, api.WithCount(2),
				api.WithLabel(key1, val2)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker1,
			framework.WithWorkerNodeGroup(worker1, api.WithCount(1)),
		),
		framework.WithVSphereWorkerNodeGroup(
			worker2,
			framework.WithWorkerNodeGroup(worker2, api.WithCount(1),
				api.WithLabel(key2, val2)),
		),
		framework.WithBottleRocket123(),
	)
}
