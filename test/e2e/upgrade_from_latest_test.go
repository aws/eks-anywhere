// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

func runUpgradeFromLatestReleaseFlow(test *framework.ClusterE2ETest, wantVersion anywherev1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
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
	provider := framework.NewVSphere(t, framework.WithVSphereFillers(
		api.WithTemplateForAllMachines(""), // Use default template from bundle
		api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
	))
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		anywherev1.Kube120,
		provider.WithProviderUpgrade(
			framework.UpdateBottlerocketTemplate120(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes121BottlerocketUpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithVSphereFillers(
		api.WithTemplateForAllMachines(""), // Use default template from bundle
		api.WithOsFamilyForAllMachines(anywherev1.Bottlerocket),
	))
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		anywherev1.Kube121,
		provider.WithProviderUpgrade(
			framework.UpdateBottlerocketTemplate121(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes120UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithVSphereFillers(
		api.WithTemplateForAllMachines(""), // Use default template from bundle
		api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
	))
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube120)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		anywherev1.Kube120,
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate120Var(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes121UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithVSphereFillers(
		api.WithTemplateForAllMachines(""), // Use default template from bundle
		api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
	))
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		anywherev1.Kube121,
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate121Var(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestVSphereKubernetes121UbuntuUpgradeFromLatestMinorReleaseAlwaysNetworkPolicy(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithVSphereFillers(
		api.WithTemplateForAllMachines(""), // Use default template from bundle
		api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
	))
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		anywherev1.Kube121,
		framework.WithClusterFiller(api.WithCiliumPolicyEnforcementMode(anywherev1.CiliumPolicyModeAlways)),
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate121Var(), // Set the template so it doesn't get autoimported
		),
	)
}

func TestDockerKubernetes121UpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		anywherev1.Kube121,
	)
}

func TestVSphereKubernetes121To122UbuntuUpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithVSphereFillers(
		api.WithTemplateForAllMachines(""), // Use default template from bundle
		api.WithOsFamilyForAllMachines(anywherev1.Ubuntu),
	))
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		anywherev1.Kube122,
		provider.WithProviderUpgrade(
			framework.UpdateUbuntuTemplate122Var(), // Set the template so it doesn't get autoimported
		),
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube122)),
		framework.WithEnvVar(features.K8s122SupportEnvVar, "true"),
	)
}

func TestDockerKubernetes121to122UpgradeFromLatestMinorRelease(t *testing.T) {
	provider := framework.NewDocker(t)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(anywherev1.Kube121)),
		framework.WithClusterFiller(api.WithExternalEtcdTopology(1)),
		framework.WithClusterFiller(api.WithControlPlaneCount(1)),
		framework.WithClusterFiller(api.WithWorkerNodeCount(1)),
	)
	runUpgradeFromLatestReleaseFlow(
		test,
		anywherev1.Kube122,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(anywherev1.Kube122)),
		framework.WithEnvVar(features.K8s122SupportEnvVar, "true"),
	)
}
