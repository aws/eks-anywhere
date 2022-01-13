// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runUpgradeFromLatestReleaseFlow(test *framework.ClusterE2ETest, wantVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	latestRelease, err := framework.GetLatestMinorReleaseFromTestBranch()
	if err != nil {
		test.T.Fatal(err)
	}
	test.GenerateClusterConfigForVersion(latestRelease.Version, framework.ExecuteWithEksaRelease(latestRelease))
	test.CreateCluster(framework.ExecuteWithEksaRelease(latestRelease))
	// Adding this manual wait because old versions of the cli don't wait long enough
	// after creation, which makes the upgrade preflight validations fail
	test.WaitForControlPlaneReady()
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(wantVersion)
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestVSphereKubernetes120BottlerocketUpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		v1alpha1.Kube120,
	)
}

func TestVSphereKubernetes121BottlerocketUpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithBottleRocket121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		v1alpha1.Kube121,
	)
}

func TestVSphereKubernetes120UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu120())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		v1alpha1.Kube120,
	)
}

func TestVSphereKubernetes121UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu121())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		v1alpha1.Kube121,
	)
}

func TestDockerKubernetes121UpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
		framework.WithClusterFiller(api.WithControlPlaneCount(3)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		v1alpha1.Kube121,
	)
}
