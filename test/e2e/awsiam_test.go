//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
	"github.com/aws/eks-anywhere/test/framework/cloudstack"
	"github.com/aws/eks-anywhere/test/framework/vsphere"
)

func runAWSIamAuthFlow(test *framework.ClusterE2ETest) {
	test.GenerateClusterConfig()
	test.CreateCluster()
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

func TestVSphereKubernetes120AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu120()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes121AWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithUbuntu121()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes120BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithBottleRocket120()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAWSIamAuthFlow(test)
}

func TestVSphereKubernetes121BottleRocketAWSIamAuth(t *testing.T) {
	test := framework.NewClusterE2ETest(
		t,
		vsphere.NewVSphere(t, vsphere.WithBottleRocket121()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes120AWSIamAuth(t *testing.T) {
	t.Skip("Skipping CloudStack in CI/CD")
	test := framework.NewClusterE2ETest(
		t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat120()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAWSIamAuthFlow(test)
}

func TestCloudStackKubernetes121AWSIamAuth(t *testing.T) {
	t.Skip("Skipping CloudStack in CI/CD")
	test := framework.NewClusterE2ETest(
		t,
		cloudstack.NewCloudStack(t, cloudstack.WithRedhat121()),
		framework.WithAWSIam(),
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAWSIamAuthFlow(test)
}
