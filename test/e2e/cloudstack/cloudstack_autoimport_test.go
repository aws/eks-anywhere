//go:build e2e
// +build e2e

package cloudstack

import (
	"testing"

	"github.com/aws/eks-anywhere/test/framework/cloudstack"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runAutoImportFlow(test *framework.ClusterE2ETest, provider *cloudstack.CloudStack) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	templates := getMachineConfigs(test)
	test.DeleteCluster()
	//deleteTemplates(test, provider, templates)
}

func getMachineConfigs(test *framework.ClusterE2ETest) map[string]v1alpha1.CloudStackMachineConfig {
	test.T.Log("Getting cloudstack machine configs to extract template")
	machineConfigs := test.GetEksaCloudStackMachineConfigs()
	uniqueMachineConfigs := make(map[string]v1alpha1.CloudStackMachineConfig, len(machineConfigs))
	for _, m := range machineConfigs {
		uniqueMachineConfigs[m.Spec.Template+m.Spec.ResourcePool] = m
	}

	return uniqueMachineConfigs
}

//func deleteTemplates(test *framework.ClusterE2ETest, provider *cloudstack.CloudStack, machineConfigs map[string]v1alpha1.CloudStackMachineConfig) {
//	ctx := context.Background()
//	for _, machineConfig := range machineConfigs {
//		test.T.Logf("Deleting cloudStack template: %s", machineConfig.Spec.Template)
//		err := provider.GovcClient.DeleteTemplate(ctx, machineConfig.Spec.ResourcePool, machineConfig.Spec.Template)
//		if err != nil {
//			test.T.Errorf("Failed deleting template [%s]: %v", machineConfig.Spec.Template, err)
//		}
//	}
//}

func TestCloudStackKubernetes120RedhatAutoimport(t *testing.T) {
	t.Skip("Skipping CloudStack in CI/CD")
	provider := cloudstack.NewCloudStack(t,
		cloudstack.WithCloudStackFillers(
			api.WithTemplate(""),
			cloudstack.WithOsFamily(v1alpha1.Redhat),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAutoImportFlow(test, provider)
}

func TestCloudStackKubernetes121RedhatAutoimport(t *testing.T) {
	t.Skip("Skipping CloudStack in CI/CD")
	provider := cloudstack.NewCloudStack(t,
		cloudstack.WithCloudStackFillers(
			api.WithTemplate(""),
			cloudstack.WithOsFamily(v1alpha1.Redhat),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAutoImportFlow(test, provider)
}
