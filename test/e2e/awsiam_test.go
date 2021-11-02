// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runAWSIamAuthFlow(test *framework.E2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	test.ValidateAWSIamAuth()
	test.StopIfFailed()
	test.DeleteCluster()
}

func TestDockerKubernetes120AWSIamAuth(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAWSIamAuthFlow(test)
}

func TestDockerKubernetes121AWSIamAuth(t *testing.T) {
	test := framework.NewE2ETest(t,
		framework.NewDocker(t),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes120AWSIamAuth(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu120()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes121AWSIamAuth(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithUbuntu121()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes120BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket120()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes121BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewE2ETest(
		t,
		framework.NewVSphere(t, framework.WithBottleRocket121()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAWSIamAuthFlow(test)
}
