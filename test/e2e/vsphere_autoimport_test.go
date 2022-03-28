// +build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/test/framework"
)

func runAutoImportFlow(test *framework.ClusterE2ETest, provider *framework.VSphere) {
	test.GenerateClusterConfig()
	test.CreateCluster()
	templates := getMachineConfigs(test)
	test.DeleteCluster()
	deleteTemplates(test, provider, templates)
}

func getMachineConfigs(test *framework.ClusterE2ETest) map[string]v1alpha1.VSphereMachineConfig {
	test.T.Log("Getting vsphere machine configs to extract template and resource pool")
	machineConfigs := test.GetEksaVSphereMachineConfigs()
	uniqueMachineConfigs := make(map[string]v1alpha1.VSphereMachineConfig, len(machineConfigs))
	for _, m := range machineConfigs {
		uniqueMachineConfigs[m.Spec.Template+m.Spec.ResourcePool] = m
	}

	return uniqueMachineConfigs
}

func deleteTemplates(test *framework.ClusterE2ETest, provider *framework.VSphere, machineConfigs map[string]v1alpha1.VSphereMachineConfig) {
	ctx := context.Background()
	for _, machineConfig := range machineConfigs {
		test.T.Logf("Deleting vSphere template: %s", machineConfig.Spec.Template)
		err := provider.GovcClient.DeleteTemplate(ctx, machineConfig.Spec.ResourcePool, machineConfig.Spec.Template)
		if err != nil {
			test.T.Errorf("Failed deleting template [%s]: %v", machineConfig.Spec.Template, err)
		}
	}
}

func TestVSphereKubernetes120UbuntuAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes121UbuntuAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes122UbuntuAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Ubuntu),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithEnvVar(features.K8s122SupportEnvVar, "true"),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes120BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube120)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes121BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes122BottlerocketAutoimport(t *testing.T) {
	provider := framework.NewVSphere(t,
		framework.WithVSphereFillers(
			api.WithTemplateForAllMachines(""),
			api.WithOsFamilyForAllMachines(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube122)),
		framework.WithEnvVar(features.K8s122SupportEnvVar, "true"),
	)
	runAutoImportFlow(test, provider)
}
