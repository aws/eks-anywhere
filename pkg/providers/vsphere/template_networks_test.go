package vsphere_test

import (
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

// TestVsphereTemplateBuilderGenerateCAPISpecWorkersWithNetworks tests the complete YAML output
// using golden file comparison for exact structure validation.
func TestVsphereTemplateBuilderGenerateCAPISpecWorkersWithNetworks(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_with_networks.yaml")

	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecWorkers(spec, nil, nil)
	g.Expect(err).ToNot(HaveOccurred())

	// Use AssertContentToFile for exact YAML structure validation
	test.AssertContentToFile(t, string(data), "testdata/expected_results_networks_md.yaml")
}

// TestNetworksFieldTableDriven comprehensively tests all network configuration scenarios
// Test case 1: WorkerNode with 2 networks config
// Test case 2: WorkerNode with 1 networks config
// Test case 3: WorkerNode with networks field empty
// Test case 4: WorkerNode without networks field.
func TestNetworksFieldConfigurationCase(t *testing.T) {
	testCases := []struct {
		name             string
		networks         []string
		expectedNetworks []string
	}{
		{
			name:             "Multiple networks",
			networks:         []string{"/SDDC-Datacenter/network/net1", "/SDDC-Datacenter/network/net2"},
			expectedNetworks: []string{"/SDDC-Datacenter/network/net1", "/SDDC-Datacenter/network/net2"},
		},
		{
			name:             "Single network",
			networks:         []string{"/SDDC-Datacenter/network/single"},
			expectedNetworks: []string{"/SDDC-Datacenter/network/single"},
		},
		{
			name:             "Empty networks - uses datacenter default",
			networks:         []string{},
			expectedNetworks: []string{"/SDDC-Datacenter/network/sddc-cgw-network-1"},
		},
		{
			name:             "Nil networks - uses datacenter default",
			networks:         nil,
			expectedNetworks: []string{"/SDDC-Datacenter/network/sddc-cgw-network-1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")

			// Configure networks for the worker machine config
			firstMachineConfigName := spec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
			machineConfig := spec.VSphereMachineConfigs[firstMachineConfigName]
			machineConfig.Spec.Networks = tc.networks

			builder := vsphere.NewVsphereTemplateBuilder(time.Now)
			data, err := builder.GenerateCAPISpecWorkers(spec, nil, nil)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(data).ToNot(BeEmpty())

			// Validate the exact network structure in YAML
			validateNetworkDevicesInYAML(t, data, tc.expectedNetworks)
		})
	}
}

// validateNetworkDevicesInYAML parses the generated YAML and validates the network devices structure.
func validateNetworkDevicesInYAML(t *testing.T, yamlData []byte, expectedNetworks []string) {
	g := NewWithT(t)

	// Split YAML documents
	docs := strings.Split(string(yamlData), "---")

	// Find the VSphereMachineTemplate document
	var machineTemplate map[string]any
	for _, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var obj map[string]any
		err := yaml.Unmarshal([]byte(doc), &obj)
		g.Expect(err).ToNot(HaveOccurred())

		if kind, ok := obj["kind"].(string); ok && kind == "VSphereMachineTemplate" {
			machineTemplate = obj
			break
		}
	}

	g.Expect(machineTemplate).ToNot(BeNil(), "VSphereMachineTemplate not found in generated YAML")

	// Navigate to spec.template.spec.network.devices
	spec, ok := machineTemplate["spec"].(map[string]any)
	g.Expect(ok).To(BeTrue(), "spec field not found in VSphereMachineTemplate")

	template, ok := spec["template"].(map[string]any)
	g.Expect(ok).To(BeTrue(), "template field not found in spec")

	templateSpec, ok := template["spec"].(map[string]any)
	g.Expect(ok).To(BeTrue(), "spec field not found in template")

	network, ok := templateSpec["network"].(map[string]any)
	g.Expect(ok).To(BeTrue(), "network field not found in template spec")

	devices, ok := network["devices"].([]any)
	g.Expect(ok).To(BeTrue(), "devices field not found in network")

	// Validate the number of devices matches expected networks
	g.Expect(devices).To(HaveLen(len(expectedNetworks)),
		"Expected %d network devices, got %d", len(expectedNetworks), len(devices))

	// Validate each device
	for i, expectedNetwork := range expectedNetworks {
		device, ok := devices[i].(map[string]any)
		g.Expect(ok).To(BeTrue(), "Device %d is not a map", i)

		// Validate dhcp4 is true
		dhcp4, ok := device["dhcp4"].(bool)
		g.Expect(ok).To(BeTrue(), "dhcp4 field not found or not boolean in device %d", i)
		g.Expect(dhcp4).To(BeTrue(), "dhcp4 should be true for device %d", i)

		// Validate networkName matches expected
		networkName, ok := device["networkName"].(string)
		g.Expect(ok).To(BeTrue(), "networkName field not found or not string in device %d", i)
		g.Expect(networkName).To(Equal(expectedNetwork),
			"Expected networkName %s, got %s for device %d", expectedNetwork, networkName, i)
	}
}
