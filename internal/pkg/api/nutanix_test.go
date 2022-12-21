package api

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestNutanixDatacenterConfigFillers(t *testing.T) {
	g := NewWithT(t)
	conf := nutanixConfig()

	WithNutanixAdditionalTrustBundle("dGVzdEJ1bmRsZQ==")(conf) // "dGVzdEJ1bmRsZQ==" is "testBundle" in base64
	g.Expect(conf.datacenterConfig.Spec.AdditionalTrustBundle).To(Equal("testBundle"))

	WithNutanixEndpoint("prism-test.nutanix.com")(conf)
	g.Expect(conf.datacenterConfig.Spec.Endpoint).To(Equal("prism-test.nutanix.com"))

	WithNutanixPort(8080)(conf)
	g.Expect(conf.datacenterConfig.Spec.Port).To(Equal(8080))

	WithNutanixInsecure(true)(conf)
	g.Expect(conf.datacenterConfig.Spec.Insecure).To(Equal(true))
}

func TestNutanixMachineConfigFillers(t *testing.T) {
	g := NewWithT(t)
	conf := nutanixConfig()

	WithNutanixMachineMemorySize("4Gi")(conf)
	WithNutanixMachineVCPUSocket(2)(conf)
	WithNutanixMachineVCPUsPerSocket(2)(conf)
	WithNutanixMachineSystemDiskSize("20Gi")(conf)
	WithNutanixSubnetName("testSubnet")(conf)
	WithNutanixPrismElementClusterName("testCluster")(conf)
	WithNutanixMachineTemplateImageName("testImage")(conf)
	WithOsFamilyForAllNutanixMachines("ubuntu")(conf)
	WithNutanixSSHAuthorizedKey("testKey")(conf)

	for _, machineConfig := range conf.machineConfigs {
		g.Expect(machineConfig.Spec.MemorySize).To(Equal(resource.MustParse("4Gi")))
		g.Expect(machineConfig.Spec.VCPUSockets).To(Equal(int32(2)))
		g.Expect(machineConfig.Spec.VCPUsPerSocket).To(Equal(int32(2)))
		g.Expect(machineConfig.Spec.SystemDiskSize).To(Equal(resource.MustParse("20Gi")))
		g.Expect(machineConfig.Spec.Subnet.Type).To(Equal(anywherev1.NutanixIdentifierName))
		g.Expect(*machineConfig.Spec.Subnet.Name).To(Equal("testSubnet"))
		g.Expect(machineConfig.Spec.Cluster.Type).To(Equal(anywherev1.NutanixIdentifierName))
		g.Expect(*machineConfig.Spec.Cluster.Name).To(Equal("testCluster"))
		g.Expect(machineConfig.Spec.Image.Type).To(Equal(anywherev1.NutanixIdentifierName))
		g.Expect(*machineConfig.Spec.Image.Name).To(Equal("testImage"))
		g.Expect(machineConfig.Spec.Users[0].SshAuthorizedKeys[0]).To(Equal("testKey"))
		g.Expect(machineConfig.Spec.OSFamily).To(Equal(anywherev1.Ubuntu))
	}

	WithNutanixSubnetUUID("90ad37a4-6dc0-4ae7-bcb3-a121dfb3fffa")(conf)
	WithNutanixPrismElementClusterUUID("90ad37a4-6dc0-4ae7-bcb3-a121dfb3fffb")(conf)
	WithNutanixMachineTemplateImageUUID("90ad37a4-6dc0-4ae7-bcb3-a121dfb3fffc")(conf)
	for _, machineConfig := range conf.machineConfigs {
		g.Expect(machineConfig.Spec.Subnet.Type).To(Equal(anywherev1.NutanixIdentifierUUID))
		g.Expect(*machineConfig.Spec.Subnet.UUID).To(Equal("90ad37a4-6dc0-4ae7-bcb3-a121dfb3fffa"))
		g.Expect(machineConfig.Spec.Cluster.Type).To(Equal(anywherev1.NutanixIdentifierUUID))
		g.Expect(*machineConfig.Spec.Cluster.UUID).To(Equal("90ad37a4-6dc0-4ae7-bcb3-a121dfb3fffb"))
		g.Expect(machineConfig.Spec.Image.Type).To(Equal(anywherev1.NutanixIdentifierUUID))
		g.Expect(*machineConfig.Spec.Image.UUID).To(Equal("90ad37a4-6dc0-4ae7-bcb3-a121dfb3fffc"))
	}
}

func nutanixConfig() *NutanixConfig {
	return &NutanixConfig{
		datacenterConfig: &anywherev1.NutanixDatacenterConfig{},
		machineConfigs:   map[string]*anywherev1.NutanixMachineConfig{},
	}
}
