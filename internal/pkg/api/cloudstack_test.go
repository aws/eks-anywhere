package api

import (
	_ "embed"
	"testing"

	. "github.com/onsi/gomega"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

const clusterConfigFile = "testdata/cloudstack-config.yaml"

func TestCloudStackMachineConfigFillers(t *testing.T) {
	testOffering := "test-compute-offering"
	testTemplate := "test-template"
	testSshKey := "test-ssh-key"
	g := NewWithT(t)
	config, err := cluster.ParseConfigFromFile(clusterConfigFile)
	if err != nil {
		g.Fail("failed to parse cluster from file")
	}

	cloudStackConfig := CloudStackConfig{
		machineConfigs: config.CloudStackMachineConfigs,
	}
	WithCloudStackComputeOfferingForAllMachines(testOffering)(cloudStackConfig)
	WithCloudStackTemplateForAllMachines(testTemplate)(cloudStackConfig)
	WithCloudStackSSHAuthorizedKey(testSshKey)(cloudStackConfig)

	for _, machineConfig := range cloudStackConfig.machineConfigs {
		g.Expect(machineConfig.Spec.ComputeOffering.Name).To(Equal(testOffering))
		g.Expect(machineConfig.Spec.Template.Name).To(Equal(testTemplate))
		g.Expect(machineConfig.Spec.Users[0].SshAuthorizedKeys[0]).To(Equal(testSshKey))
	}
}

func TestCloudStackDatacenterConfigFillers(t *testing.T) {
	testAz := anywherev1.CloudStackAvailabilityZone{
		Name:           "testAz",
		CredentialsRef: "testCreds",
		Zone: anywherev1.CloudStackZone{
			Name: "zone1",
			Network: anywherev1.CloudStackResourceIdentifier{
				Name: "SharedNet1",
			},
		},
		Domain:                "testDomain",
		Account:               "testAccount",
		ManagementApiEndpoint: "testApiEndpoint",
	}
	g := NewWithT(t)
	config, err := cluster.ParseConfigFromFile(clusterConfigFile)
	if err != nil {
		g.Fail("failed to parse cluster from file")
	}

	cloudStackConfig := CloudStackConfig{
		datacenterConfig: config.CloudStackDatacenter,
	}
	WithFirstCloudStackAz(testAz)(cloudStackConfig)
	g.Expect(len(cloudStackConfig.datacenterConfig.Spec.AvailabilityZones)).To(Equal(1))
	g.Expect(cloudStackConfig.datacenterConfig.Spec.AvailabilityZones[0]).To(Equal(testAz))

	testAz2 := *testAz.DeepCopy()
	testAz2.Name = "testAz2"
	WithCloudStackAz(testAz2)(cloudStackConfig)
	g.Expect(len(cloudStackConfig.datacenterConfig.Spec.AvailabilityZones)).To(Equal(2))
	g.Expect(cloudStackConfig.datacenterConfig.Spec.AvailabilityZones[0]).To(Equal(testAz))
	g.Expect(cloudStackConfig.datacenterConfig.Spec.AvailabilityZones[1]).To(Equal(testAz2))

	WithoutCloudStackAz(testAz)(cloudStackConfig)
	g.Expect(len(cloudStackConfig.datacenterConfig.Spec.AvailabilityZones)).To(Equal(1))
	g.Expect(cloudStackConfig.datacenterConfig.Spec.AvailabilityZones[0]).To(Equal(testAz2))
}
