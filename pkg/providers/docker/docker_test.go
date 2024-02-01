package docker_test

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"path"
	"testing"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	dockerMocks "github.com/aws/eks-anywhere/pkg/providers/docker/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const testdataDir = "testdata"

type dockerTest struct {
	*WithT
	dockerClient *dockerMocks.MockProviderClient
	kubectl      *dockerMocks.MockProviderKubectlClient
	provider     *docker.Provider
}

func newTest(t *testing.T) *dockerTest {
	ctrl := gomock.NewController(t)
	client := dockerMocks.NewMockProviderClient(ctrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(ctrl)
	return &dockerTest{
		WithT:        NewWithT(t),
		dockerClient: client,
		kubectl:      kubectl,
		provider:     docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow),
	}
}

func givenClusterSpec(t *testing.T, fileName string) *cluster.Spec {
	return test.NewFullClusterSpec(t, path.Join(testdataDir, fileName))
}

func TestVersion(t *testing.T) {
	tt := newTest(t)
	dockerProviderVersion := "v0.3.19"
	managementComponents := givenManagementComponents()
	managementComponents.Docker.Version = dockerProviderVersion

	tt.Expect(tt.provider.Version(managementComponents)).To(Equal(dockerProviderVersion))
}

func TestProviderUpdateKubeConfig(t *testing.T) {
	input := []byte(`
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJ
    server: https://172.18.0.3:6443
  name: capi-quickstart`)
	expected := `
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://127.0.0.1:4332
  name: capi-quickstart`

	mockCtrl := gomock.NewController(t)

	type fields struct {
		clusterName string
	}
	type args struct {
		content     *[]byte
		clusterName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test updates for docker config file",
			fields: fields{
				clusterName: "capi-quickstart",
			},
			args: args{
				content:     &input,
				clusterName: "capi-quickstart",
			},
			want: expected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			client.EXPECT().GetDockerLBPort(gomock.Any(), tt.args.clusterName).Return("4332", nil)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)

			if err := p.UpdateKubeConfig(tt.args.content, tt.args.clusterName); (err != nil) != tt.wantErr {
				t.Errorf("UpdateKubeConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if string(*tt.args.content) != tt.want {
				t.Errorf("updateKubeConfigFile() got = %v, want %v", string(*tt.args.content), tt.want)
			}
		})
	}
}

func TestProviderGenerateDeploymentFileSuccessUpdateMachineTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	var cpTaints, wnTaints, wnTaints2 []v1.Taint

	cpTaints = append(cpTaints, v1.Taint{Key: "key1", Value: "val1", Effect: "NoSchedule", TimeAdded: nil})
	cpTaints = append(cpTaints, v1.Taint{Key: "key2", Value: "val2", Effect: "PreferNoSchedule", TimeAdded: nil})
	cpTaints = append(cpTaints, v1.Taint{Key: "key3", Value: "val3", Effect: "NoExecute", TimeAdded: nil})
	wnTaints = append(wnTaints, v1.Taint{Key: "key2", Value: "val2", Effect: "PreferNoSchedule", TimeAdded: nil})
	wnTaints2 = append(wnTaints2, v1.Taint{Key: "wnTaitns2", Value: "true", Effect: "PreferNoSchedule", TimeAdded: nil})

	nodeLabels := map[string]string{"label1": "foo", "label2": "bar"}

	tests := []struct {
		testName    string
		clusterSpec *cluster.Spec
		wantCPFile  string
		wantMDFile  string
	}{
		{
			testName: "valid config",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_expected.yaml",
			wantMDFile: "testdata/valid_deployment_md_expected.yaml",
		},
		{
			testName: "valid config 1.24",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.24"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundles["1.24"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_expected_124onwards.yaml",
			wantMDFile: "testdata/valid_deployment_md_expected_124onwards.yaml",
		},
		{
			testName: "valid config with cp taints",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.Cluster.Spec.ControlPlaneConfiguration.Taints = cpTaints
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_taints_expected.yaml",
			wantMDFile: "testdata/valid_deployment_md_expected.yaml",
		},
		{
			testName: "valid config with md taints",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.Cluster.Spec.ControlPlaneConfiguration.Taints = cpTaints
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), Taints: wnTaints, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_taints_expected.yaml",
			wantMDFile: "testdata/valid_deployment_md_taints_expected.yaml",
		},
		{
			testName: "valid config multiple worker node groups with machine deployment taints",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.Cluster.Spec.ControlPlaneConfiguration.Taints = cpTaints
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), Taints: wnTaints, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}, {Count: ptr.Int(3), Taints: wnTaints2, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-1"}}
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
			}),
			wantCPFile: "testdata/valid_deployment_cp_taints_expected.yaml",
			wantMDFile: "testdata/valid_deployment_multiple_md_taints_expected.yaml",
		},
		{
			testName: "valid config with node labels",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Labels: nodeLabels, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_expected.yaml",
			wantMDFile: "testdata/valid_deployment_node_labels_md_expected.yaml",
		},
		{
			testName: "valid config with cp node labels",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.Cluster.Spec.ControlPlaneConfiguration.Labels = nodeLabels
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_node_labels_cp_expected.yaml",
			wantMDFile: "testdata/valid_deployment_md_expected.yaml",
		},
		{
			testName: "valid config with cidrs and custom resolv.conf",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"10.10.0.0/24", "10.128.0.0/12"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"192.168.0.0/16", "10.10.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.DNS.ResolvConf = &v1alpha1.ResolvConf{Path: "/etc/my-custom-resolv.conf"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
			}),
			wantCPFile: "testdata/valid_deployment_custom_cidrs_cp_expected.yaml",
			wantMDFile: "testdata/valid_deployment_custom_cidrs_md_expected.yaml",
		},
		{
			testName: "with minimal oidc",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}

				s.OIDCConfig = &v1alpha1.OIDCConfig{
					Spec: v1alpha1.OIDCConfigSpec{
						ClientId:  "my-client-id",
						IssuerUrl: "https://mydomain.com/issuer",
					},
				}
			}),
			wantCPFile: "testdata/capd_valid_minimal_oidc_cp_expected.yaml",
			wantMDFile: "testdata/capd_valid_minimal_oidc_md_expected.yaml",
		},
		{
			testName: "with full oidc",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
				s.OIDCConfig = &v1alpha1.OIDCConfig{
					Spec: v1alpha1.OIDCConfigSpec{
						ClientId:     "my-client-id",
						IssuerUrl:    "https://mydomain.com/issuer",
						GroupsClaim:  "claim1",
						GroupsPrefix: "prefix-for-groups",
						RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{
							{
								Claim: "sub",
								Value: "test",
							},
						},
						UsernameClaim:  "username-claim",
						UsernamePrefix: "username-prefix",
					},
				}
			}),
			wantCPFile: "testdata/capd_valid_full_oidc_cp_expected.yaml",
			wantMDFile: "testdata/capd_valid_full_oidc_md_expected.yaml",
		},
		{
			testName: "valid autoscaling config",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "test-cluster"
				s.Cluster.Spec.KubernetesVersion = "1.19"
				s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
				s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.VersionsBundles["1.19"] = versionsBundle
				s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0", AutoScalingConfiguration: &v1alpha1.AutoScalingConfiguration{MinCount: 3, MaxCount: 5}}}
			}),
			wantCPFile: "testdata/valid_deployment_cp_expected.yaml",
			wantMDFile: "testdata/valid_autoscaler_deployment_md_expected.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ctx := context.Background()
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
			cluster := &types.Cluster{
				Name: "test",
			}
			currentSpec := tt.clusterSpec.DeepCopy()
			tt.clusterSpec.Bundles.Spec.Number = 2
			bootstrapCluster := &types.Cluster{
				Name: "bootstrap-test",
			}
			for _, nodeGroup := range tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
				md := &clusterv1.MachineDeployment{
					Spec: clusterv1.MachineDeploymentSpec{
						Template: clusterv1.MachineTemplateSpec{
							Spec: clusterv1.MachineSpec{
								Bootstrap: clusterv1.Bootstrap{
									ConfigRef: &v1.ObjectReference{
										Name: fmt.Sprintf("%s-%s-template-1234567890000", tt.clusterSpec.Cluster.Name, nodeGroup.Name),
									},
								},
							},
						},
					},
				}
				machineDeploymentName := fmt.Sprintf("%s-%s", tt.clusterSpec.Cluster.Name, nodeGroup.Name)
				kubectl.EXPECT().GetMachineDeployment(ctx, machineDeploymentName,
					gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster)),
					gomock.AssignableToTypeOf(executables.WithNamespace(constants.EksaSystemNamespace))).Return(md, nil)
			}
			kubectl.EXPECT().UpdateAnnotation(ctx, "etcdadmcluster", fmt.Sprintf("%s-etcd", tt.clusterSpec.Cluster.Name),
				map[string]string{etcdv1.UpgradeInProgressAnnotation: "true"}, gomock.Any(), gomock.Any())
			cpContent, mdContent, err := p.GenerateCAPISpecForUpgrade(ctx, bootstrapCluster, cluster, currentSpec, tt.clusterSpec)
			if err != nil {
				t.Fatalf("provider.GenerateCAPISpecForUpgrade() error = %v, wantErr nil", err)
			}
			test.AssertContentToFile(t, string(cpContent), tt.wantCPFile)
			test.AssertContentToFile(t, string(mdContent), tt.wantMDFile)
		})
	}
}

func TestProviderGenerateDeploymentFileSuccessUpdateKubeadmConfigTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := test.NewClusterSpec()

	var cpTaints, wnTaints []v1.Taint
	cpTaints = append(cpTaints, v1.Taint{Key: "key1", Value: "val1", Effect: "NoSchedule", TimeAdded: nil})
	cpTaints = append(cpTaints, v1.Taint{Key: "key2", Value: "val2", Effect: "PreferNoSchedule", TimeAdded: nil})
	cpTaints = append(cpTaints, v1.Taint{Key: "key3", Value: "val3", Effect: "NoExecute", TimeAdded: nil})
	wnTaints = append(wnTaints, v1.Taint{Key: "key2", Value: "val2", Effect: "PreferNoSchedule", TimeAdded: nil})

	clusterSpec.Cluster.Name = "test-cluster"
	clusterSpec.Cluster.Spec.KubernetesVersion = "1.19"
	clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
	clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count = 3
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints = cpTaints
	clusterSpec.VersionsBundles["1.19"] = versionsBundle
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}

	p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	cluster := &types.Cluster{
		Name: "test-cluster",
	}
	currentSpec := clusterSpec.DeepCopy()
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Taints = wnTaints
	bootstrapCluster := &types.Cluster{
		Name: "bootstrap-test",
	}

	cp := &controlplanev1.KubeadmControlPlane{
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					Name: "test-cluster-control-plane-template-1234567890000",
				},
			},
		},
	}
	etcdadm := &etcdv1.EtcdadmCluster{
		Spec: etcdv1.EtcdadmClusterSpec{
			InfrastructureTemplate: v1.ObjectReference{
				Name: "test-cluster-etcd-template-1234567890000",
			},
		},
	}

	kubectl.EXPECT().GetKubeadmControlPlane(ctx, cluster, cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(cp, nil)
	kubectl.EXPECT().GetEtcdadmCluster(ctx, cluster, cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(etcdadm, nil)

	cpContent, mdContent, err := p.GenerateCAPISpecForUpgrade(ctx, bootstrapCluster, cluster, currentSpec, clusterSpec)
	if err != nil {
		t.Fatalf("provider.GenerateCAPISpecForUpgrade() error = %v, wantErr nil", err)
	}

	test.AssertContentToFile(t, string(cpContent), "testdata/valid_deployment_cp_taints_expected.yaml")
	test.AssertContentToFile(t, string(mdContent), "testdata/valid_deployment_md_taints_expected.yaml")
}

func TestProviderGenerateDeploymentFileSuccessNotUpdateMachineTemplate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	clusterSpec := test.NewClusterSpec()
	clusterSpec.Cluster.Spec.KubernetesVersion = v1alpha1.Kube119
	clusterSpec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
	clusterSpec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(0), MachineGroupRef: &v1alpha1.Ref{Name: "fluxTestCluster"}, Name: "md-0"}}
	p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	cluster := &types.Cluster{
		Name: "test",
	}
	currentSpec := clusterSpec.DeepCopy()
	bootstrapCluster := &types.Cluster{
		Name: "bootstrap-test",
	}

	cp := &controlplanev1.KubeadmControlPlane{
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					Name: "test-control-plane-template-original",
				},
			},
		},
	}
	md := &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Name: "test-md-0-original",
						},
					},
					InfrastructureRef: v1.ObjectReference{
						Name: "test-md-0-original",
					},
				},
			},
		},
	}
	machineDeploymentName := fmt.Sprintf("%s-%s", clusterSpec.Cluster.Name, clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Name)

	kubectl.EXPECT().GetKubeadmControlPlane(ctx, cluster, cluster.Name, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(cp, nil)
	kubectl.EXPECT().GetMachineDeployment(ctx, machineDeploymentName, gomock.AssignableToTypeOf(executables.WithCluster(bootstrapCluster))).Return(md, nil).Times(2)

	cpContent, mdContent, err := p.GenerateCAPISpecForUpgrade(ctx, bootstrapCluster, cluster, currentSpec, clusterSpec)
	if err != nil {
		t.Fatalf("provider.GenerateCAPISpecForUpgrade() error = %v, wantErr nil", err)
	}

	test.AssertContentToFile(t, string(cpContent), "testdata/no_machinetemplate_update_cp_expected.yaml")
	test.AssertContentToFile(t, string(mdContent), "testdata/no_machinetemplate_update_md_expected.yaml")
}

func TestGetInfrastructureBundleSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	tests := []struct {
		testName             string
		managementComponents *cluster.ManagementComponents
	}{
		{
			testName:             "create overrides layer",
			managementComponents: givenManagementComponents(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			client := dockerMocks.NewMockProviderClient(mockCtrl)
			kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
			p := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)

			infraBundle := p.GetInfrastructureBundle(tt.managementComponents)
			if infraBundle == nil {
				t.Fatalf("provider.GetInfrastructureBundle() should have an infrastructure bundle")
			}
			assert.Equal(t, "infrastructure-docker/v0.3.19/", infraBundle.FolderName, "Incorrect folder name")
			assert.Equal(t, len(infraBundle.Manifests), 3, "Wrong number of files in the infrastructure bundle")

			wantManifests := []releasev1alpha1.Manifest{
				versionsBundle.Docker.Components,
				versionsBundle.Docker.Metadata,
				versionsBundle.Docker.ClusterTemplate,
			}
			assert.ElementsMatch(t, infraBundle.Manifests, wantManifests, "Incorrect manifests")
		})
	}
}

var versionsBundle = &cluster.VersionsBundle{
	KubeDistro: &cluster.KubeDistro{
		Kubernetes: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/kubernetes",
			Tag:        "v1.19.6-eks-1-19-2",
		},
		CoreDNS: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/coredns",
			Tag:        "v1.8.0-eks-1-19-2",
		},
		Etcd: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/etcd-io",
			Tag:        "v3.4.14-eks-1-19-2",
		},
		EtcdVersion: "3.4.14",
		EtcdURL:     "https://distro.eks.amazonaws.com/kubernetes-1-21/releases/4/artifacts/etcd/v3.4.16/etcd-linux-amd64-v3.4.16.tar.gz",
	},
	VersionsBundle: &releasev1alpha1.VersionsBundle{
		EksD: releasev1alpha1.EksDRelease{
			KindNode: releasev1alpha1.Image{
				Description: "kind/node container image",
				Name:        "kind/node",
				URI:         "public.ecr.aws/eks-distro/kubernetes-sigs/kind/node:v1.18.16-eks-1-18-4-216edda697a37f8bf16651af6c23b7e2bb7ef42f-62681885fe3a97ee4f2b110cc277e084e71230fa",
			},
		},
		Docker: releasev1alpha1.DockerBundle{
			Version: "v0.3.19",
			Manager: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/capd-manager:v0.3.15-6bdb9fc78bb926135843c58ec8b77b54d8f2c82c",
			},
			KubeProxy: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/brancz/kube-rbac-proxy:v0.8.0-25df7d96779e2a305a22c6e3f9425c3465a77244",
			},
			Components: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/infrastructure-components-development.yaml",
			},
			ClusterTemplate: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/cluster-template-development.yaml",
			},
			Metadata: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/metadata.yaml",
			},
		},
		Haproxy: releasev1alpha1.HaproxyBundle{
			Image: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/haproxy:v0.11.1-eks-a-v0.0.0-dev-build.1464",
			},
		},
	},
}

var workerVersionsBundle = &cluster.VersionsBundle{
	KubeDistro: &cluster.KubeDistro{
		Kubernetes: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/kubernetes",
			Tag:        "v1.18.4-eks-1-18-3",
		},
		CoreDNS: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/coredns",
			Tag:        "v1.8.0-eks-1-19-2",
		},
		Etcd: cluster.VersionedRepository{
			Repository: "public.ecr.aws/eks-distro/etcd-io",
			Tag:        "v3.4.14-eks-1-19-2",
		},
		EtcdVersion: "3.4.14",
	},
	VersionsBundle: &releasev1alpha1.VersionsBundle{
		EksD: releasev1alpha1.EksDRelease{
			KindNode: releasev1alpha1.Image{
				Description: "kind/node container image",
				Name:        "kind/node",
				URI:         "public.ecr.aws/eks-distro/kubernetes-sigs/kind/node:v1.18.16-eks-1-18-4-216edda697a37f8bf16651af6c23b7e2bb7ef42f-62681885fe3a97ee4f2b110cc277e084e71230fa",
			},
		},
		Docker: releasev1alpha1.DockerBundle{
			Version: "v0.3.19",
			Manager: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/capd-manager:v0.3.15-6bdb9fc78bb926135843c58ec8b77b54d8f2c82c",
			},
			KubeProxy: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/brancz/kube-rbac-proxy:v0.8.0-25df7d96779e2a305a22c6e3f9425c3465a77244",
			},
			Components: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/infrastructure-components-development.yaml",
			},
			ClusterTemplate: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/cluster-template-development.yaml",
			},
			Metadata: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/metadata.yaml",
			},
		},
		Haproxy: releasev1alpha1.HaproxyBundle{
			Image: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/kind/haproxy:v0.11.1-eks-a-v0.0.0-dev-build.1464",
			},
		},
	},
}

func givenManagementComponents() *cluster.ManagementComponents {
	return &cluster.ManagementComponents{
		EksD: releasev1alpha1.EksDRelease{
			KindNode: releasev1alpha1.Image{
				Description: "kind/node container image",
				Name:        "kind/node",
				URI:         "public.ecr.aws/eks-distro/kubernetes-sigs/kind/node:v1.18.16-eks-1-18-4-216edda697a37f8bf16651af6c23b7e2bb7ef42f-62681885fe3a97ee4f2b110cc277e084e71230fa",
			},
		},
		Docker: releasev1alpha1.DockerBundle{
			Version: "v0.3.19",
			Manager: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/kubernetes-sigs/cluster-api/capd-manager:v0.3.15-6bdb9fc78bb926135843c58ec8b77b54d8f2c82c",
			},
			KubeProxy: releasev1alpha1.Image{
				URI: "public.ecr.aws/l0g8r8j6/brancz/kube-rbac-proxy:v0.8.0-25df7d96779e2a305a22c6e3f9425c3465a77244",
			},
			Components: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/infrastructure-components-development.yaml",
			},
			ClusterTemplate: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/cluster-template-development.yaml",
			},
			Metadata: releasev1alpha1.Manifest{
				URI: "embed:///config/clusterctl/overrides/infrastructure-docker/v0.3.19/metadata.yaml",
			},
		},
	}
}

func TestChangeDiffNoChange(t *testing.T) {
	tt := newTest(t)
	assert.Nil(t, tt.provider.ChangeDiff(givenManagementComponents(), givenManagementComponents()))
}

func TestChangeDiffWithChange(t *testing.T) {
	tt := newTest(t)
	currentManagementComponents := givenManagementComponents()
	currentManagementComponents.Docker.Version = "v0.3.18"

	newManagementComponents := givenManagementComponents()
	newManagementComponents.Docker.Version = "v0.3.19"

	wantDiff := &types.ComponentChangeDiff{
		ComponentName: "docker",
		NewVersion:    "v0.3.19",
		OldVersion:    "v0.3.18",
	}

	tt.Expect(tt.provider.ChangeDiff(currentManagementComponents, newManagementComponents)).To(Equal(wantDiff))
}

func TestProviderGenerateCAPISpecForCreateWithPodIAMConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	clusterObj := &types.Cluster{
		Name: "test-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Spec.KubernetesVersion = "1.19"
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 1
		s.VersionsBundles["1.19"] = versionsBundle
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}}}
	})
	clusterSpec.Cluster.Spec.PodIAMConfig = &v1alpha1.PodIAMConfig{ServiceAccountIssuer: "https://test"}

	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, _, err := provider.GenerateCAPISpecForCreate(context.Background(), clusterObj, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/valid_deployment_cp_pod_iam_expected.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithStackedEtcd(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	clusterObj := &types.Cluster{
		Name: "test-cluster",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Spec.KubernetesVersion = "1.19"
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 1
		s.VersionsBundles["1.19"] = versionsBundle
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}}}
	})

	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, _, err := provider.GenerateCAPISpecForCreate(context.Background(), clusterObj, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(cp), "testdata/valid_deployment_cp_stacked_etcd_expected.yaml")
}

func TestProviderGenerateCAPISpecForCreateWithWorkerKubernetesVersion(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	clusterObj := &types.Cluster{
		Name: "test-cluster",
	}
	workerVersion := v1alpha1.KubernetesVersion("1.18")
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Spec.KubernetesVersion = "1.19"
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 1
		s.VersionsBundles["1.19"] = versionsBundle
		s.VersionsBundles["1.18"] = workerVersionsBundle

		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{
			{Name: "md-0", Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, KubernetesVersion: &workerVersion},
			{Name: "md-1", Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}},
		}
	})

	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	_, md, err := provider.GenerateCAPISpecForCreate(context.Background(), clusterObj, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}
	test.AssertContentToFile(t, string(md), "testdata/valid_deployment_md_expected_worker_version.yaml")
}

func TestDockerTemplateBuilderGenerateCAPISpecControlPlane(t *testing.T) {
	type args struct {
		clusterSpec  *cluster.Spec
		buildOptions []providers.BuildMapOption
	}
	tests := []struct {
		name        string
		args        args
		wantContent []byte
		wantErr     error
	}{
		{
			name: "kube 119 test",
			args: args{
				clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
					s.Cluster.Name = "test-cluster"
				}),
				buildOptions: nil,
			},
			wantErr: nil,
		},
		{
			name: "kube version not specified",
			args: args{
				clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
					s.Cluster.Name = "test-cluster"
					s.Cluster.Spec.KubernetesVersion = ""
				}),
				buildOptions: nil,
			},
			wantErr: fmt.Errorf("error building template map for CP "),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			builder := docker.NewDockerTemplateBuilder(time.Now)

			gotContent, err := builder.GenerateCAPISpecControlPlane(tt.args.clusterSpec, tt.args.buildOptions...)
			if err != tt.wantErr && !assert.Contains(t, err.Error(), tt.wantErr.Error()) {
				t.Errorf("Got DockerTemplateBuilder.GenerateCAPISpecControlPlane() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				g.Expect(gotContent).NotTo(BeEmpty())
			}
		})
	}
}

func TestProviderGenerateDeploymentFileForSingleNodeCluster(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	clusterObj := &types.Cluster{Name: "single-node"}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "single-node"
		s.Cluster.Spec.KubernetesVersion = "1.21"
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 1
		s.VersionsBundles["1.21"] = versionsBundle
		s.Cluster.Spec.WorkerNodeGroupConfigurations = nil
	})

	if provider == nil {
		t.Fatalf("provider object is nil")
	}

	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)
	if err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, _, err := provider.GenerateCAPISpecForCreate(context.Background(), clusterObj, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_cluster_docker_cp_single_node.yaml")
}

func TestDockerTemplateBuilderGenerateCAPISpecWorkers(t *testing.T) {
	type args struct {
		clusterSpec *cluster.Spec
	}
	tests := []struct {
		name        string
		args        args
		wantContent []byte
		wantErr     error
	}{
		{
			name: "kube version not specified",
			args: args{
				clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
					s.Cluster.Name = "test-cluster"
					s.Cluster.Spec.KubernetesVersion = ""
				}),
			},
			wantErr: fmt.Errorf("error building template map for MD "),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			builder := docker.NewDockerTemplateBuilder(time.Now)

			gotContent, err := builder.GenerateCAPISpecWorkers(tt.args.clusterSpec, nil, nil)
			if err != tt.wantErr && !assert.Contains(t, err.Error(), tt.wantErr.Error()) {
				t.Errorf("Got DockerTemplateBuilder.GenerateCAPISpecWorkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				g.Expect(gotContent).NotTo(BeEmpty())
			}
		})
	}
}

func TestInvalidDockerTemplateWithControlplaneEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Spec.KubernetesVersion = "1.19"
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.Cluster.Spec.ControlPlaneConfiguration.Endpoint = &v1alpha1.Endpoint{Host: "test-ip"}
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: ptr.Int(3), MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
	})
	wantErr := fmt.Errorf("specifying endpoint host configuration in Cluster is not supported")
	err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec)

	if err == nil || err.Error() != wantErr.Error() {
		t.Fatalf("err %v, wantErr %v", err, wantErr)
	}
}

func TestDockerGenerateDeploymentFileWithMirrorConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	clusterObj := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, "cluster_mirror_config.yaml")

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), clusterObj, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_config_md.yaml")
}

func TestDockerGenerateDeploymentFileWithMirrorAndCertConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	clusterObj := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, "cluster_mirror_with_cert_config.yaml")

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), clusterObj, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_with_cert_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_with_cert_config_md.yaml")
}

func TestDockerGenerateDeploymentFileWithMirrorAndAuthConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	t.Setenv("REGISTRY_USERNAME", "username")
	t.Setenv("REGISTRY_PASSWORD", "password")
	ctx := context.Background()
	client := dockerMocks.NewMockProviderClient(mockCtrl)
	kubectl := dockerMocks.NewMockProviderKubectlClient(mockCtrl)
	provider := docker.NewProvider(&v1alpha1.DockerDatacenterConfig{}, client, kubectl, test.FakeNow)
	clusterObj := &types.Cluster{Name: "test"}
	clusterSpec := givenClusterSpec(t, "cluster_mirror_with_auth_config.yaml")

	if err := provider.SetupAndValidateCreateCluster(ctx, clusterSpec); err != nil {
		t.Fatalf("failed to setup and validate: %v", err)
	}

	cp, md, err := provider.GenerateCAPISpecForCreate(context.Background(), clusterObj, clusterSpec)
	if err != nil {
		t.Fatalf("failed to generate cluster api spec contents: %v", err)
	}

	test.AssertContentToFile(t, string(cp), "testdata/expected_results_mirror_with_auth_config_cp.yaml")
	test.AssertContentToFile(t, string(md), "testdata/expected_results_mirror_with_auth_config_md.yaml")
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

		bldr := docker.NewDockerTemplateBuilder(time.Now)

		data, err := bldr.GenerateCAPISpecControlPlane(clusterSpec)
		g.Expect(err).ToNot(HaveOccurred())

		test.AssertContentToFile(t, string(data), tc.Output)
	}
}

func TestDockerWriteKubeconfig(t *testing.T) {
	for _, tc := range []struct {
		clusterName    string
		kubeconfigPath string
		providerErr    error
		writerErr      error
	}{
		{
			clusterName:    "test",
			kubeconfigPath: "test",
		},
		{
			clusterName:    "test",
			kubeconfigPath: "test",
			writerErr:      fmt.Errorf("failed to write kubeconfig"),
		},
		{
			clusterName:    "test",
			kubeconfigPath: "test",
			providerErr:    fmt.Errorf("failed to get LB port"),
		},
	} {
		g := NewWithT(t)
		ctx := context.Background()
		buf := bytes.NewBuffer([]byte{})
		mockCtrl := gomock.NewController(t)
		dockerClient := dockerMocks.NewMockProviderClient(mockCtrl)
		mocksWriter := dockerMocks.NewMockKubeconfigReader(mockCtrl)
		mocksWriter.EXPECT().GetClusterKubeconfig(ctx, tc.clusterName, tc.kubeconfigPath).Return([]byte{}, tc.writerErr)
		if tc.writerErr == nil {
			dockerClient.EXPECT().GetDockerLBPort(ctx, tc.clusterName).Return("", tc.providerErr)
		}
		writer := docker.NewKubeconfigWriter(dockerClient, mocksWriter)
		err := writer.WriteKubeconfig(ctx, tc.clusterName, tc.kubeconfigPath, buf)
		if tc.writerErr == nil && tc.providerErr == nil {
			g.Expect(err).ToNot(HaveOccurred())
		} else {
			g.Expect(err).To(HaveOccurred())
		}
	}
}
