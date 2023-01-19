//go:build e2e
// +build e2e

package e2e

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
