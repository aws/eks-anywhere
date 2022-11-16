package api

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestNutanixDatacenterConfigFillers(t *testing.T) {
	g := NewWithT(t)
	configFile := "testdata/nutanix/cluster-config.yaml"
	conf, err := newNutanixConfig(configFile)
	assert.NoError(t, err)
	assert.NotNil(t, conf)

	WithNutanixAdditionalTrustBundle("dGVzdEJ1bmRsZQ==")(conf) // "dGVzdEJ1bmRsZQ==" is "testBundle" in base64
	g.Expect(conf.datacenterConfig.Spec.AdditionalTrustBundle).To(Equal("testBundle"))

	WithNutanixEndpoint("prism-test.nutanix.com")(conf)
	g.Expect(conf.datacenterConfig.Spec.Endpoint).To(Equal("prism-test.nutanix.com"))

	WithNutanixPort(8080)(conf)
	g.Expect(conf.datacenterConfig.Spec.Port).To(Equal(8080))
}

func TestNutanixMachineConfigFillers(t *testing.T) {
	g := NewWithT(t)
	configFile := "testdata/nutanix/cluster-config.yaml"
	conf, err := newNutanixConfig(configFile)
	assert.NoError(t, err)
	assert.NotNil(t, conf)

	WithNutanixMachineMemorySize("4Gi")(conf)
	WithNutanixMachineVCPUSocket(2)(conf)
	WithNutanixMachineVCPUsPerSocket(2)(conf)
	WithNutanixMachineSystemDiskSize("20Gi")(conf)
	WithNutanixSubnetName("testSubnet")(conf)
	WithNutanixPrismElementClusterName("testCluster")(conf)
	WithNutanixMachineTemplateImageName("testImage")(conf)
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

func TestAutoFillNutanixProvider(t *testing.T) {
	g := NewWithT(t)
	configFile := "testdata/nutanix/cluster-config.yaml"
	resources, err := AutoFillNutanixProvider(
		configFile,
		WithNutanixEndpoint("prism-test.nutanix.com"),
		WithNutanixSubnetName("testSubnet"),
	)
	g.Expect(err).To(BeNil())
	g.Expect(resources).To(Not(BeNil()))
	expectedResources, err := os.ReadFile("testdata/nutanix/templated-resources.yaml")
	g.Expect(err).To(BeNil())
	g.Expect(resources).To(MatchYAML(expectedResources))
}
