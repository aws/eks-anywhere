//go:build e2e
// +build e2e

package vsphere

import (
	"context"
	"testing"

	vsphere2 "github.com/aws/eks-anywhere/internal/pkg/api/vsphere"
	"github.com/aws/eks-anywhere/test/framework/vsphere"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/test/framework"
)

func runAutoImportFlow(test *framework.ClusterE2ETest, provider *vsphere.VSphere) {
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

func deleteTemplates(test *framework.ClusterE2ETest, provider *vsphere.VSphere, machineConfigs map[string]v1alpha1.VSphereMachineConfig) {
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
	provider := vsphere.NewVSphere(t,
		vsphere.WithVSphereFillers(
			vsphere2.WithTemplate(""),
			vsphere2.WithOsFamily(v1alpha1.Ubuntu),
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
	provider := vsphere.NewVSphere(t,
		vsphere.WithVSphereFillers(
			vsphere2.WithTemplate(""),
			vsphere2.WithOsFamily(v1alpha1.Ubuntu),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAutoImportFlow(test, provider)
}

func TestVSphereKubernetes120BottlerocketAutoimport(t *testing.T) {
	provider := vsphere.NewVSphere(t,
		vsphere.WithVSphereFillers(
			vsphere2.WithTemplate(""),
			vsphere2.WithOsFamily(v1alpha1.Bottlerocket),
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
	provider := vsphere.NewVSphere(t,
		vsphere.WithVSphereFillers(
			vsphere2.WithTemplate(""),
			vsphere2.WithOsFamily(v1alpha1.Bottlerocket),
		),
	)
	test := framework.NewClusterE2ETest(
		t,
		provider,
		framework.WithClusterFiller(api.WithKubernetesVersion(v1alpha1.Kube121)),
	)
	runAutoImportFlow(test, provider)
}
