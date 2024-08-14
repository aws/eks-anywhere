package vsphere_test

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/config"
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
