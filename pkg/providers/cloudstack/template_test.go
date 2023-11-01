package cloudstack_test

import (
	"path"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
)

const (
	testClusterConfigMainFilename        = "cluster_main.yaml"
	testClusterConfigMainWithAZsFilename = "cluster_main_with_availability_zones.yaml"
	testDataDir                          = "testdata"
)

func TestTemplateBuilderGenerateCAPISpecControlPlaneNilDatacenter(t *testing.T) {
	g := NewWithT(t)
	templateBuilder := cloudstack.NewTemplateBuilder(time.Now)
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	clusterSpec.CloudStackDatacenter = nil
	_, err := templateBuilder.GenerateCAPISpecControlPlane(clusterSpec)
	g.Expect(err).To(MatchError(ContainSubstring("provided clusterSpec CloudStackDatacenter is nil. Unable to generate CAPI spec control plane")))
}

func TestTemplateBuilderGenerateCAPISpecControlPlaneMissingNames(t *testing.T) {
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))

	tests := []struct {
		name        string
		buildOption func(values map[string]interface{})
		wantErr     string
	}{
		{
			name: "missing control plane template name",
			buildOption: func(values map[string]interface{}) {
				values["etcdTemplateName"] = clusterapi.EtcdMachineTemplateName(clusterSpec.Cluster)
			},
			wantErr: "unable to determine control plane template name",
		},
		{
			name: "missing etcd machine template name",
			buildOption: func(values map[string]interface{}) {
				values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(clusterSpec.Cluster)
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			templateBuilder := cloudstack.NewTemplateBuilder(time.Now)

			_, err := templateBuilder.GenerateCAPISpecControlPlane(clusterSpec, tt.buildOption)
			g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
		})
	}
}

func TestTemplateBuilderGenerateCAPISpecControlPlaneInvalidSSHKey(t *testing.T) {
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))

	controlPlaneMachineConfigName := clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
	etcdMachineConfigName := clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name

	tests := []struct {
		name              string
		machineConfigName string
		wantErr           string
	}{
		{
			name:              "invalid controlplane ssh key",
			machineConfigName: controlPlaneMachineConfigName,
			wantErr:           "formatting ssh key for cloudstack control plane template: ssh",
		},
		{
			name:              "invalid etcd ssh key",
			machineConfigName: etcdMachineConfigName,
			wantErr:           "formatting ssh key for cloudstack etcd template: ssh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			spec := clusterSpec.DeepCopy()
			templateBuilder := cloudstack.NewTemplateBuilder(time.Now)
			machineConfig := spec.CloudStackMachineConfigs[tt.machineConfigName]
			machineConfig.Spec.Users[0].SshAuthorizedKeys[0] = "ssh-rsa AAAA    B3NzaC1K73CeQ== testemail@test.com"
			_, err := templateBuilder.GenerateCAPISpecControlPlane(spec, func(values map[string]interface{}) {
				values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(clusterSpec.Cluster)
				values["etcdTemplateName"] = clusterapi.EtcdMachineTemplateName(clusterSpec.Cluster)
			})
			g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
		})
	}
}

func TestTemplateBuilderGenerateCAPISpecControlPlaneInvalidEndpoint(t *testing.T) {
	g := NewWithT(t)
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "1.1.1.1::"
	templateBuilder := cloudstack.NewTemplateBuilder(time.Now)

	_, err := templateBuilder.GenerateCAPISpecControlPlane(clusterSpec, func(values map[string]interface{}) {
		values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(clusterSpec.Cluster)
		values["etcdTemplateName"] = clusterapi.EtcdMachineTemplateName(clusterSpec.Cluster)
	})

	g.Expect(err).To(MatchError(ContainSubstring("error building template map from CP host 1.1.1.1:: is invalid: address 1.1.1.1::: too many colons in address")))
}

func TestTemplateBuilderGenerateCAPISpecWorkersInvalidSSHKey(t *testing.T) {
	g := NewWithT(t)
	templateBuilder := cloudstack.NewTemplateBuilder(time.Now)
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	firstMachineConfigName := clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
	machineConfig := clusterSpec.CloudStackMachineConfigs[firstMachineConfigName]
	machineConfig.Spec.Users[0].SshAuthorizedKeys[0] = "ssh-rsa AAAA    B3NzaC1K73CeQ== testemail@test.com"
	machineTemplateNames, kubeadmConfigTemplateNames := clusterapi.InitialTemplateNamesForWorkers(clusterSpec)
	_, err := templateBuilder.GenerateCAPISpecWorkers(clusterSpec, machineTemplateNames, kubeadmConfigTemplateNames)
	g.Expect(err).To(
		MatchError(ContainSubstring("formatting ssh key for cloudstack worker template: ssh")),
	)
}

func TestTemplateBuilderGenerateCAPISpecWorkersInvalidEndpoint(t *testing.T) {
	g := NewWithT(t)
	templateBuilder := cloudstack.NewTemplateBuilder(time.Now)
	clusterSpec := test.NewFullClusterSpec(t, path.Join(testDataDir, testClusterConfigMainFilename))
	clusterSpec.Cluster.Spec.ProxyConfiguration = &v1alpha1.ProxyConfiguration{}
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "1.1.1.1::"
	machineTemplateNames, kubeadmConfigTemplateNames := clusterapi.InitialTemplateNamesForWorkers(clusterSpec)
	_, err := templateBuilder.GenerateCAPISpecWorkers(clusterSpec, machineTemplateNames, kubeadmConfigTemplateNames)
	g.Expect(err).To(MatchError(ContainSubstring("building template map for MD host 1.1.1.1:: is invalid: address 1.1.1.1::: too many colons in address")))
}

func TestTemplateBuilder_CertSANs(t *testing.T) {
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

		bldr := cloudstack.NewTemplateBuilder(time.Now)

		data, err := bldr.GenerateCAPISpecControlPlane(clusterSpec, func(values map[string]interface{}) {
			values["controlPlaneTemplateName"] = clusterapi.ControlPlaneMachineTemplateName(clusterSpec.Cluster)
		})
		g.Expect(err).ToNot(HaveOccurred())

		test.AssertContentToFile(t, string(data), tc.Output)
	}
}
