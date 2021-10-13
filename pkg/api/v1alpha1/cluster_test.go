package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetAndValidateClusterConfig(t *testing.T) {
	tests := []struct {
		testName    string
		fileName    string
		wantCluster *Cluster
		wantErr     bool
	}{
		{
			testName:    "file doesn't exist",
			fileName:    "testdata/fake_file.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "not parseable file",
			fileName:    "testdata/not_parseable_cluster.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName: "valid 1.18",
			fileName: "testdata/cluster_1_18.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube118,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube120,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid different machine configs",
			fileName: "testdata/cluster_different_machine_configs.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test-2",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "with valid GitOps",
			fileName: "testdata/cluster_1_19_gitops.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					GitOpsRef: &Ref{
						Kind: "GitOpsConfig",
						Name: "test-gitops",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "with GitOps branch valid",
			fileName: "testdata/cluster_1_19_gitops_branch.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Cluster",
					APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					GitOpsRef: &Ref{
						Kind: "GitOpsConfig",
						Name: "test-gitops",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "with valid proxy configuration",
			fileName: "testdata/cluster_valid_proxy.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
					ProxyConfiguration: &ProxyConfiguration{
						HttpProxy:  "http://0.0.0.0:1",
						HttpsProxy: "0.0.0.0:1",
						NoProxy:    []string{"localhost"},
					},
				},
			},
			wantErr: false,
		},
		{
			testName:    "with no worker node groups",
			fileName:    "testdata/cluster_invalid_no_worker_node_groups.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "with multiple worker node groups",
			fileName:    "testdata/cluster_invalid_multiple_worker_node_groups.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "with GitOps branch invalid",
			fileName:    "testdata/cluster_1_19_gitops_invalid_branch.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "with gitops invalid repo name",
			fileName:    "testdata/cluster_1_19_gitops_invalid_repo.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "Empty Git Provider",
			fileName:    "testdata/cluster_invalid_gitops_empty_gitprovider.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "Invalid Git Provider",
			fileName:    "testdata/cluster_invalid_gitops_invalid_gitprovider.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "Empty Git Repository",
			fileName:    "testdata/cluster_invalid_gitops_empty_gitrepo.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "Git Repository not set",
			fileName:    "testdata/cluster_invalid_gitops_unset_gitrepo.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "invalid kind",
			fileName:    "testdata/cluster_invalid_kinds.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "even controlPlaneReplicas",
			fileName:    "testdata/cluster_even_control_plane_replicas.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "even unstacked etcd replicas",
			fileName:    "testdata/unstacked_etcd_even_replicas.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "empty identity providers",
			fileName:    "testdata/cluster_invalid_empty_identity_providers.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "extra identity providers",
			fileName:    "testdata/cluster_invalid_extra_identity_providers.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "non oidc identity provider",
			fileName:    "testdata/cluster_invalid_non_oidc_identity_providers.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "with invalid proxy ip configuration",
			fileName:    "testdata/cluster_invalid_proxy_ip.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "with invalid proxy port configuration",
			fileName:    "testdata/cluster_invalid_proxy_port.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "with invalid proxy missing http proxy",
			fileName:    "testdata/cluster_invalid_missing_http_proxy.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "with invalid proxy missing https proxy",
			fileName:    "testdata/cluster_invalid_missing_https_proxy.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "with empty CNI",
			fileName:    "testdata/cluster_empty_cni.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "with not supported CNI",
			fileName:    "testdata/cluster_not_supported_cni.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetAndValidateClusterConfig(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetClusterConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCluster) {
				t.Fatalf("GetClusterConfig() = %#v, want %#v", got, tt.wantCluster)
			}
		})
	}
}

func TestGetClusterConfig(t *testing.T) {
	tests := []struct {
		testName    string
		fileName    string
		wantCluster *Cluster
		wantErr     bool
	}{
		{
			testName: "valid 1.18",
			fileName: "testdata/cluster_1_18.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube118,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.19",
			fileName: "testdata/cluster_1_19.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid with extra delimiters",
			fileName: "testdata/cluster_extra_delimiters.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid 1.20",
			fileName: "testdata/cluster_1_20.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube120,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got, err := GetClusterConfig(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetClusterConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.wantCluster) {
				t.Fatalf("GetClusterConfig() = %#v, want %#v", got, tt.wantCluster)
			}
		})
	}
}

func TestParseClusterConfig(t *testing.T) {
	type args struct {
		fileName      string
		clusterConfig KindAccessor
	}
	tests := []struct {
		name        string
		args        args
		matchError  error
		wantErr     bool
		wantCluster *Cluster
	}{
		{
			name: "Good cluster config parse",
			args: args{
				fileName:      "testdata/cluster_vsphere.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr: false,
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
		},
		{
			name: "Invalid data type",
			args: args{
				fileName:      "testdata/not_parseable_cluster.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr:    true,
			matchError: fmt.Errorf("cannot unmarshal string into Go struct field WorkerNodeGroupConfiguration.spec.workerNodeGroupConfigurations.count of type int"),
		},
		{
			name: "Incorrect indentation",
			args: args{
				fileName:      "testdata/incorrect_indentation.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr:    true,
			matchError: fmt.Errorf("error converting YAML to JSON: yaml: line 12: did not find expected key"),
		},
		{
			name: "Invalid key",
			args: args{
				fileName:      "testdata/invalid_key.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr:    true,
			matchError: fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"registryMirro rConfiguration\""),
		},
		{
			name: "Invalid yaml",
			args: args{
				fileName:      "testdata/invalid_format.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr:    true,
			matchError: fmt.Errorf("error converting YAML to JSON: yaml: did not find expected node content"),
		},
		{
			name: "Invalid spec field",
			args: args{
				fileName:      "testdata/invalid_spec_field.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr:    true,
			matchError: fmt.Errorf("error unmarshaling JSON: while decoding JSON: json: unknown field \"invalidField\""),
		},
		{
			name: "Cluster definition at the end",
			args: args{
				fileName:      "testdata/cluster_definition_at_the_end.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr: false,
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ParseClusterConfig(tt.args.fileName, tt.args.clusterConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseClusterConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.args.clusterConfig, tt.wantCluster) {
				t.Fatalf("GetClusterConfig() = %#v, want %#v", tt.args.clusterConfig, tt.wantCluster)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.matchError.Error()) {
				t.Errorf("ParseClusterConfig() error = %v, wantErr %v err %v", err, tt.wantErr, tt.matchError)
			}
		})
	}
}

func TestCluster_PauseReconcile(t *testing.T) {
	tests := []struct {
		name  string
		want  string
		pause bool
	}{
		{
			name:  "pause should set pause annotation",
			want:  "true",
			pause: true,
		},
		{
			name:  "pause should set pause annotation",
			pause: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "cluster_test",
					Annotations: map[string]string{},
				},
			}
			if tt.pause {
				c.PauseReconcile()
				val, ok := c.Annotations["anywhere.eks.amazonaws.com/paused"]
				if ok && val != tt.want {
					t.Errorf("expected value on annotation is not set got=%s, want=%s", val, tt.want)
				}
				if !ok {
					t.Errorf("pause annotation is not set")
				}
			}
			if !tt.pause {
				if _, ok := c.Annotations["anywhere.eks.amazonaws.com/paused"]; ok {
					t.Errorf("pause annotation is shouldn't be set")
				}
			}
		})
	}
}

func TestCluster_IsReconcilePaused(t *testing.T) {
	tests := []struct {
		name  string
		want  bool
		pause bool
	}{
		{
			name:  "reconcile is paused",
			want:  true,
			pause: true,
		},
		{
			name:  "reconcile is not paused",
			pause: false,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "cluster_test",
					Annotations: map[string]string{},
				},
			}
			if tt.pause {
				c.PauseReconcile()
			}
			if got := c.IsReconcilePaused(); got != tt.want {
				t.Errorf("IsReconcilePaused() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGitOpsEquals(t *testing.T) {
	tests := []struct {
		name string
		want bool
		prev *GitOpsConfigSpec
		new  *GitOpsConfigSpec
	}{
		{
			name: "previous and new == nil",
			want: true,
			prev: nil,
			new:  nil,
		},
		{
			name: "previous == nil",
			want: false,
			prev: nil,
			new:  &GitOpsConfigSpec{},
		},
		{
			name: "previous == nil",
			want: false,
			prev: &GitOpsConfigSpec{},
			new:  nil,
		},
		{
			name: "previous == new",
			want: true,
			prev: &GitOpsConfigSpec{
				Flux: Flux{
					Github: Github{
						Owner:               "owner",
						Repository:          "repo",
						FluxSystemNamespace: "namespace",
						Branch:              "main",
						ClusterConfigPath:   "path/test",
						Personal:            false,
					},
				},
			},
			new: &GitOpsConfigSpec{
				Flux: Flux{
					Github: Github{
						Owner:               "owner",
						Repository:          "repo",
						FluxSystemNamespace: "namespace",
						Branch:              "main",
						ClusterConfigPath:   "path/test",
						Personal:            false,
					},
				},
			},
		},
		{
			name: "previous != new",
			want: false,
			prev: &GitOpsConfigSpec{Flux: Flux{
				Github: Github{
					Owner:               "owner",
					Repository:          "repo",
					FluxSystemNamespace: "namespace",
					Branch:              "main",
					ClusterConfigPath:   "path/test",
					Personal:            false,
				},
			}},
			new: &GitOpsConfigSpec{
				Flux: Flux{
					Github: Github{
						Owner:               "owner",
						Repository:          "new-repo",
						FluxSystemNamespace: "namespace",
						Branch:              "main",
						ClusterConfigPath:   "path/test",
						Personal:            false,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != tt.prev.Equal(tt.new) {
				t.Errorf("GitOps %+v should be equals to  %+v", tt.prev, tt.new)
			}
		})
	}
}

func TestEndPointEquals(t *testing.T) {
	tests := []struct {
		name string
		want bool
		prev *Endpoint
		new  *Endpoint
	}{
		{
			name: "previous and new == nil",
			want: true,
			prev: nil,
			new:  nil,
		},
		{
			name: "previous == nil",
			want: false,
			prev: nil,
			new:  &Endpoint{},
		},
		{
			name: "previous == nil",
			want: false,
			prev: &Endpoint{},
			new:  nil,
		},
		{
			name: "previous == new",
			want: true,
			prev: &Endpoint{Host: "host"},
			new:  &Endpoint{Host: "host"},
		},
		{
			name: "previous != new",
			want: false,
			prev: &Endpoint{Host: "host"},
			new:  &Endpoint{Host: "new-host"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != tt.prev.Equal(tt.new) {
				t.Errorf("Endpoint %+v should be equals to  %+v", tt.prev, tt.new)
			}
		})
	}
}

func TestProxyConfigurationEquals(t *testing.T) {
	tests := []struct {
		name string
		want bool
		prev *ProxyConfiguration
		new  *ProxyConfiguration
	}{
		{
			name: "previous and new == nil",
			want: true,
			prev: nil,
			new:  nil,
		},
		{
			name: "previous == nil",
			want: false,
			prev: nil,
			new:  &ProxyConfiguration{},
		},
		{
			name: "previous == nil",
			want: false,
			prev: &ProxyConfiguration{},
			new:  nil,
		},
		{
			name: "previous == new, all exists",
			want: true,
			prev: &ProxyConfiguration{
				HttpProxy:  "httpproxy",
				HttpsProxy: "httpsproxy",
				NoProxy: []string{
					"noproxy1",
					"noproxy2",
				},
			},
			new: &ProxyConfiguration{
				HttpProxy:  "httpproxy",
				HttpsProxy: "httpsproxy",
				NoProxy: []string{
					"noproxy1",
					"noproxy2",
				},
			},
		},
		{
			name: "previous == new, only httpproxy",
			want: true,
			prev: &ProxyConfiguration{HttpProxy: "httpproxy"},
			new:  &ProxyConfiguration{HttpProxy: "httpproxy"},
		},
		{
			name: "previous == new, only noproxy, order diff",
			want: true,
			prev: &ProxyConfiguration{NoProxy: []string{
				"noproxy1",
				"noproxy2",
				"noproxy3",
			}},
			new: &ProxyConfiguration{NoProxy: []string{
				"noproxy2",
				"noproxy3",
				"noproxy1",
			}},
		},
		{
			name: "previous != new, httpsproxy diff",
			want: false,
			prev: &ProxyConfiguration{HttpsProxy: "httpsproxy1"},
			new:  &ProxyConfiguration{HttpsProxy: "httpsproxy2"},
		},
		{
			name: "previous != new, noproxy diff val",
			want: false,
			prev: &ProxyConfiguration{
				HttpProxy: "",
				NoProxy: []string{
					"noproxy1",
					"noproxy2",
				},
			},
			new: &ProxyConfiguration{
				HttpProxy: "",
				NoProxy: []string{
					"noproxy2",
					"noproxy3",
				},
			},
		},
		{
			name: "previous != new, noproxy diff one empty",
			want: false,
			prev: &ProxyConfiguration{
				NoProxy: []string{
					"noproxy1",
					"noproxy2",
				},
			},
			new: &ProxyConfiguration{},
		},
		{
			name: "previous != new, noproxy diff length",
			want: false,
			prev: &ProxyConfiguration{
				NoProxy: []string{
					"noproxy1",
					"noproxy2",
				},
			},
			new: &ProxyConfiguration{
				NoProxy: []string{
					"noproxy1",
					"noproxy2",
					"noproxy3",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != tt.prev.Equal(tt.new) {
				t.Errorf("ProxyConfiguration %+v should be equals to  %+v", tt.prev, tt.new)
			}
		})
	}
}

func TestClusterNetworkEquals(t *testing.T) {
	tests := []struct {
		name string
		want bool
		prev *ClusterNetwork
		new  *ClusterNetwork
	}{
		{
			name: "previous and new == nil",
			want: true,
			prev: nil,
			new:  nil,
		},
		{
			name: "previous == nil",
			want: false,
			prev: nil,
			new:  &ClusterNetwork{},
		},
		{
			name: "previous == nil",
			want: false,
			prev: &ClusterNetwork{},
			new:  nil,
		},
		{
			name: "previous == new, all exists",
			want: true,
			prev: &ClusterNetwork{
				CNI: Cilium,
				Pods: Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
						"1.2.3.4/6",
					},
				},
				Services: Services{
					CidrBlocks: []string{
						"1.2.3.4/7",
						"1.2.3.4/8",
					},
				},
			},
			new: &ClusterNetwork{
				CNI: Cilium,
				Pods: Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
						"1.2.3.4/6",
					},
				},
				Services: Services{
					CidrBlocks: []string{
						"1.2.3.4/7",
						"1.2.3.4/8",
					},
				},
			},
		},
		{
			name: "previous == new, pods empty",
			want: true,
			prev: &ClusterNetwork{
				Services: Services{
					CidrBlocks: []string{},
				},
			},
			new: &ClusterNetwork{
				Pods:     Pods{},
				Services: Services{},
			},
		},
		{
			name: "previous == new, order diff",
			want: true,
			prev: &ClusterNetwork{
				CNI: Cilium,
				Pods: Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
						"1.2.3.4/6",
					},
				},
				Services: Services{
					CidrBlocks: []string{
						"1.2.3.4/7",
						"1.2.3.4/8",
					},
				},
			},
			new: &ClusterNetwork{
				CNI: Cilium,
				Pods: Pods{
					CidrBlocks: []string{
						"1.2.3.4/6",
						"1.2.3.4/5",
					},
				},
				Services: Services{
					CidrBlocks: []string{
						"1.2.3.4/8",
						"1.2.3.4/7",
					},
				},
			},
		},
		{
			name: "previous != new, pods diff",
			want: false,
			prev: &ClusterNetwork{
				CNI: Cilium,
				Pods: Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
						"1.2.3.4/6",
					},
				},
			},
			new: &ClusterNetwork{
				CNI: Cilium,
				Pods: Pods{
					CidrBlocks: []string{
						"1.2.3.4/6",
					},
				},
			},
		},
		{
			name: "previous != new, services diff, one empty",
			want: false,
			prev: &ClusterNetwork{},
			new: &ClusterNetwork{
				Services: Services{
					CidrBlocks: []string{
						"1.2.3.4/7",
						"1.2.3.4/8",
					},
				},
			},
		},
		{
			name: "previous != new, services diff, CidrBlocks empty",
			want: false,
			prev: &ClusterNetwork{
				Services: Services{
					CidrBlocks: []string{},
				},
			},
			new: &ClusterNetwork{
				Services: Services{
					CidrBlocks: []string{
						"1.2.3.4/7",
						"1.2.3.4/8",
					},
				},
			},
		},
		{
			name: "previous != new, diff CNI",
			want: false,
			prev: &ClusterNetwork{
				CNI: CiliumEnterprise,
			},
			new: &ClusterNetwork{
				CNI: Cilium,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != tt.prev.Equal(tt.new) {
				t.Errorf("ClusterNetwork %+v should be equals to  %+v", tt.prev, tt.new)
			}
		})
	}
}

func TestRefEquals(t *testing.T) {
	tests := []struct {
		name string
		want bool
		prev *Ref
		new  *Ref
	}{
		{
			name: "previous and new == nil",
			want: true,
			prev: nil,
			new:  nil,
		},
		{
			name: "previous == nil",
			want: false,
			prev: nil,
			new:  &Ref{},
		},
		{
			name: "previous == nil",
			want: false,
			prev: &Ref{},
			new:  nil,
		},
		{
			name: "previous == new",
			want: true,
			prev: &Ref{Kind: "kind", Name: "name"},
			new:  &Ref{Kind: "kind", Name: "name"},
		},
		{
			name: "previous != new, val diff",
			want: false,
			prev: &Ref{Kind: "kind1", Name: "name1"},
			new:  &Ref{Kind: "kind2", Name: "name2"},
		},
		{
			name: "previous != new, one missing kind",
			want: false,
			prev: &Ref{Name: "name"},
			new:  &Ref{Kind: "kind", Name: "name"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != tt.prev.Equal(tt.new) {
				t.Errorf("Ref %+v should be equals to  %+v", tt.prev, tt.new)
			}
		})
	}
}
