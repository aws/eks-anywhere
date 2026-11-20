package vsphere_test

import (
	"regexp"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/providers/common"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

const (
	expectedVSphereUsername = "vsphere_username"
	expectedVSpherePassword = "vsphere_password"
)

func TestVsphereTemplateBuilderGenerateCAPISpecWorkersInvalidSSHKey(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
	firstMachineConfigName := spec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	machineConfig := spec.VSphereMachineConfigs[firstMachineConfigName]
	machineConfig.Spec.Users[0].SshAuthorizedKeys[0] = invalidSSHKey()
	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	_, err := builder.GenerateCAPISpecWorkers(spec, nil, nil)
	g.Expect(err).To(
		MatchError(ContainSubstring("formatting ssh key for vsphere workers template: ssh")),
	)
}

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneInvalidControlPlaneSSHKey(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
	controlPlaneMachineConfigName := spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	machineConfig := spec.VSphereMachineConfigs[controlPlaneMachineConfigName]
	machineConfig.Spec.Users[0].SshAuthorizedKeys[0] = invalidSSHKey()
	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	_, err := builder.GenerateCAPISpecControlPlane(spec, nil, nil)
	g.Expect(err).To(
		MatchError(ContainSubstring("formatting ssh key for vsphere control plane template: ssh")),
	)
}

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneInvalidEtcdSSHKey(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
	etcdMachineConfigName := spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	machineConfig := spec.VSphereMachineConfigs[etcdMachineConfigName]
	machineConfig.Spec.Users[0].SshAuthorizedKeys[0] = invalidSSHKey()
	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	_, err := builder.GenerateCAPISpecControlPlane(spec, nil, nil)
	g.Expect(err).To(
		MatchError(ContainSubstring("formatting ssh key for vsphere etcd template: ssh")),
	)
}

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneValidKubeletConfigWN(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
	spec.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubeletConfiguration = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"maxPods": 20,
		},
	}
	spec.Cluster.Spec.ClusterNetwork.DNS = v1alpha1.DNS{
		ResolvConf: &v1alpha1.ResolvConf{
			Path: "test-path",
		},
	}
	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecWorkers(spec, nil, nil)
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(data), "testdata/expected_kct.yaml")
}

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneValidKubeletConfigCP(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.KubeletConfiguration = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"maxPods": 20,
		},
	}
	spec.Cluster.Spec.ClusterNetwork.DNS = v1alpha1.DNS{
		ResolvConf: &v1alpha1.ResolvConf{
			Path: "test-path",
		},
	}
	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
	})
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(data), "testdata/expected_kcp.yaml")
}

func TestVsphereGenerateCAPISpecControlPlaneValidKubeletConfigCPBR(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_br.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.KubeletConfiguration = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":    "KubeletConfiguration",
			"maxPods": 20,
		},
	}
	spec.Cluster.Spec.ClusterNetwork.DNS = v1alpha1.DNS{
		ResolvConf: &v1alpha1.ResolvConf{
			Path: "test-path",
		},
	}

	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
	})
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(data), "testdata/expected_kcp_br.yaml")
}

func TestTemplateBuilder_CertSANs(t *testing.T) {
	t.Setenv(config.EksavSphereUsernameKey, expectedVSphereUsername)
	t.Setenv(config.EksavSpherePasswordKey, expectedVSpherePassword)

	for _, tc := range []struct {
		Input  string
		Output string
	}{
		{
			Input:  "testdata/cluster_api_server_cert_san_domain_name.yaml",
			Output: "testdata/expected_cluster_api_server_cert_san_domain_name.yaml",
		},
		{
			Input:  "testdata/cluster_api_server_cert_san_ip.yaml",
			Output: "testdata/expected_cluster_api_server_cert_san_ip.yaml",
		},
	} {
		g := NewWithT(t)
		clusterSpec := test.NewFullClusterSpec(t, tc.Input)

		bldr := vsphere.NewVsphereTemplateBuilder(time.Now)

		data, err := bldr.GenerateCAPISpecControlPlane(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())

		test.AssertContentToFile(t, string(data), tc.Output)
	}
}

func invalidSSHKey() string {
	return "ssh-rsa AAAA    B3NzaC1K73CeQ== testemail@test.com"
}

func TestVsphereTemplateBuilderInventoryPaths(t *testing.T) {
	for _, field := range []string{"folder", "resourcePool", "datastore", "datacenter"} {
		for _, value := range []string{"/Datacenter/vm", "/Datacenter/vm's", "'vm'", "", "/Datacenter/vm #1", "path \"quoted\"\\vm: name"} {
			t.Run(field+"/"+value, func(t *testing.T) {
				g := NewWithT(t)
				spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
				for _, machine := range spec.VSphereMachineConfigs {
					switch field {
					case "folder":
						machine.Spec.Folder = value
					case "resourcePool":
						machine.Spec.ResourcePool = value
					case "datastore":
						machine.Spec.Datastore = value
					}
				}
				if field == "datacenter" {
					spec.VSphereDatacenter.Spec.Datacenter = value
				}
				builder := vsphere.NewVsphereTemplateBuilder(time.Now)
				controlPlane, err := builder.GenerateCAPISpecControlPlane(spec)
				g.Expect(err).ToNot(HaveOccurred())
				workers, err := builder.GenerateCAPISpecWorkers(spec, nil, nil)
				g.Expect(err).ToNot(HaveOccurred())
				for name, data := range map[string][]byte{"controlPlane": controlPlane, "workers": workers} {
					t.Run(name, func(t *testing.T) {
						g := NewWithT(t)
						objects, err := test.ParseMultiDocYAML(data)
						g.Expect(err).ToNot(HaveOccurred())
						count := 0
						for _, object := range objects {
							if object["kind"] != "VSphereMachineTemplate" {
								continue
							}
							actual, err := test.GetYAMLPath(object, "spec.template.spec."+field)
							g.Expect(err).ToNot(HaveOccurred())
							g.Expect(actual).To(Equal(value))
							count++
						}
						if name == "controlPlane" {
							g.Expect(count).To(Equal(2))
						} else {
							g.Expect(count).To(Equal(1))
						}
					})
				}
			})
		}
	}
}

func TestVsphereTemplateBuilderFailureDomainInventoryPaths(t *testing.T) {
	for _, value := range []string{"/Datacenter/vm", "/Datacenter/vm's", "'vm'", "", "/Datacenter/vm #1", "path \"quoted\"\\vm: name"} {
		t.Run(value, func(t *testing.T) {
			g := NewWithT(t)
			spec := test.NewFullClusterSpec(t, "testdata/cluster_vsphere_failuredomain.yaml")
			spec.VSphereDatacenter.Spec.Datacenter = value
			spec.VSphereDatacenter.Spec.FailureDomains = []v1alpha1.FailureDomain{{
				Name:           "fd-1",
				ComputeCluster: value,
				Datastore:      value,
				Folder:         value,
				Network:        value,
				ResourcePool:   value,
			}}
			builder := vsphere.NewVsphereTemplateBuilder(time.Now)
			data, err := builder.GenerateVsphereFailureDomainsSpec(spec, map[string]string{"fd-1": "test-test-fd-1"})
			g.Expect(err).ToNot(HaveOccurred())
			objects, err := test.ParseMultiDocYAML(data)
			g.Expect(err).ToNot(HaveOccurred())
			for kind, paths := range map[string][]string{
				"VSphereFailureDomain":  {"spec.region.name", "spec.zone.name", "spec.topology.datacenter", "spec.topology.computeCluster", "spec.topology.datastore"},
				"VSphereDeploymentZone": {"spec.placementConstraint.folder", "spec.placementConstraint.resourcePool"},
			} {
				object, err := test.FindObjectByKind(objects, kind)
				g.Expect(err).ToNot(HaveOccurred())
				for _, path := range paths {
					actual, err := test.GetYAMLPath(object, path)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(actual).To(Equal(value), path)
				}
				if kind == "VSphereFailureDomain" {
					networks, err := test.GetYAMLPath(object, "spec.topology.networks")
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(networks).To(Equal([]interface{}{value}))
				}
			}
		})
	}
}

func TestVsphereGenerateCAPISpecControlPlaneValidKubeletConfigNTPCPBR(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main_br.yaml")
	spec.Cluster.Spec.ControlPlaneConfiguration.KubeletConfiguration = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":    "KubeletConfiguration",
			"maxPods": 20,
		},
	}
	spec.Cluster.Spec.ClusterNetwork.DNS = v1alpha1.DNS{
		ResolvConf: &v1alpha1.ResolvConf{
			Path: "test-path",
		},
	}

	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	spec.Config.VSphereMachineConfigs["test-cp"].Spec.HostOSConfiguration = &v1alpha1.HostOSConfiguration{
		NTPConfiguration: &v1alpha1.NTPConfiguration{
			Servers: []string{"test.ntp"},
		},
	}
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
	})
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(data), "testdata/expected_kcp_br_ntp.yaml")
}

func TestVsphereTemplateBuilderGenerateFailureDomainYaml(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_vsphere_failuredomain.yaml")
	templateNames := map[string]string{
		"fd-1": "test-test-fd-1",
	}
	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateVsphereFailureDomainsSpec(spec, templateNames)
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(data), "testdata/expected_results_failuredomain.yaml")
}

func TestVsphereTemplateBuilderGenerateCAPISpecVCenterTags(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")
	testTags := []string{"urn:vmomi:InventoryServiceTag:test1-uuid:GLOBAL", "urn:vmomi:InventoryServiceTag:test2-uuid:GLOBAL"}
	controlPlaneMachineConfigName := spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	cpMachineConfig := spec.VSphereMachineConfigs[controlPlaneMachineConfigName]
	cpMachineConfig.Spec.TagIDs = testTags
	etcdMachineConfigName := spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
	etcdMachineConfig := spec.VSphereMachineConfigs[etcdMachineConfigName]
	etcdMachineConfig.Spec.TagIDs = testTags
	firstMachineConfigName := spec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	workerMachineConfig := spec.VSphereMachineConfigs[firstMachineConfigName]
	workerMachineConfig.Spec.TagIDs = testTags
	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	cpData, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
	})
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(cpData), "testdata/expected_kcp_vcenter_tags.yaml")
	wData, err := builder.GenerateCAPISpecWorkers(spec, nil, nil)
	g.Expect(err).ToNot(HaveOccurred())
	test.AssertContentToFile(t, string(wData), "testdata/expected_kct_vcenter_tags.yaml")
}

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneWithCustomAuditPolicy(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")

	// Set custom audit policy content
	customAuditPolicy := `apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Metadata
  omitStages:
  - RequestReceived`

	spec.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyContent = customAuditPolicy

	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
	})

	g.Expect(err).ToNot(HaveOccurred())
	str := collapseWhitespace(string(data))
	g.Expect(str).To(ContainSubstring(collapseWhitespace(customAuditPolicy)))

	defaultAuditPolicy, err := common.GetAuditPolicy(spec.Cluster.Spec.KubernetesVersion)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(str).ToNot(ContainSubstring(collapseWhitespace(defaultAuditPolicy)))
}

func collapseWhitespace(s string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
}

func TestVsphereTemplateBuilderGenerateCAPISpecControlPlaneWithDefaultAuditPolicy(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, "testdata/cluster_main.yaml")

	spec.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyContent = ""

	builder := vsphere.NewVsphereTemplateBuilder(time.Now)
	data, err := builder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(spec.Cluster)
	})

	g.Expect(err).ToNot(HaveOccurred())

	defaultAuditPolicy, err := common.GetAuditPolicy(spec.Cluster.Spec.KubernetesVersion)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(collapseWhitespace(string(data))).To(ContainSubstring(collapseWhitespace(defaultAuditPolicy)))
}
