//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

func runAWSIamAuthFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateAWSIamAuth()
	test.StopIfFailed()
	test.DeleteCluster()
}

func runUpgradeFlowWithAWSIamAuth(test *framework.ClusterE2ETest, updateVersion v1alpha1.KubernetesVersion, clusterOpts ...framework.ClusterE2ETestOpt) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateAWSIamAuth()
	test.UpgradeCluster(clusterOpts)
	test.ValidateCluster(updateVersion)
	test.ValidateAWSIamAuth()
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestDockerKubernetes120AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes121AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes122AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes123AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithEnvVar(features.K8s123SupportEnvVar, "true"),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes120AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu120()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes121AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu121()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes122AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu122()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes123AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu123()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithEnvVar(features.K8s123SupportEnvVar, "true"),

	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes120BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket120()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes121BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket121()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes122BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket122()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes123BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket123()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube123)),
		framework.WithEnvVar(features.K8s123SupportEnvVar, "true"),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes122To123AWSIamAuthUpgrade(t *testing.T) {
	provider := framework.NewVSphere(t, framework.WithUbuntu122())
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithEnvVar(features.K8s123SupportEnvVar, "true"),
	)
	runUpgradeFlowWithAWSIamAuth(
		test,
		v1alpha1.Kube123,
		framework.WithClusterUpgrade(api.WithKubernetesVersion(v1alpha1.Kube123)),
		provider.WithProviderUpgrade(framework.UpdateUbuntuTemplate123Var()),
		framework.WithEnvVar(features.K8s123SupportEnvVar, "true"),
	)
}