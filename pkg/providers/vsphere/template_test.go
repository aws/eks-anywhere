package vsphere_test

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
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

func TestTemplateBuilder_CertSANs(t *testing.T) {
	os.Unsetenv(config.EksavSphereUsernameKey)
	os.Unsetenv(config.EksavSpherePasswordKey)

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
