package v1alpha1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestValidateClusterName(t *testing.T) {
	tests := []struct {
		clusterName, name string
		wantErr           error
	}{
		{
			name:        "FailureSpecialChars",
			clusterName: "test-cluster@123_",
			wantErr:     errors.New("test-cluster@123_ is not a valid cluster name, cluster names must start with lowercase/uppercase letters and can include numbers and dashes. For instance 'testCluster-123' is a valid name but '123testCluster' is not. "),
		},
		{
			name:        "FailureDotChars",
			clusterName: "test-cluster1.20",
			wantErr:     errors.New("test-cluster1.20 is not a valid cluster name, cluster names must start with lowercase/uppercase letters and can include numbers and dashes. For instance 'testCluster-123' is a valid name but '123testCluster' is not. "),
		},
		{
			name:        "FailureFirstCharNumeric",
			clusterName: "123test-Cluster",
			wantErr:     errors.New("123test-Cluster is not a valid cluster name, cluster names must start with lowercase/uppercase letters and can include numbers and dashes. For instance 'testCluster-123' is a valid name but '123testCluster' is not. "),
		},
		{
			name:        "SuccessUpperCaseChars",
			clusterName: "test-Cluster",
			wantErr:     nil,
		},
		{
			name:        "SuccessLowerCase",
			clusterName: "test-cluster",
			wantErr:     nil,
		},
		{
			name:        "SuccessLowerCaseDashNumeric",
			clusterName: "test-cluster123",
			wantErr:     nil,
		},
		{
			name:        "SuccessLowerCaseNumeric",
			clusterName: "test123cluster",
			wantErr:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got := ValidateClusterName(tc.clusterName)
			if !reflect.DeepEqual(tc.wantErr, got) {
				t.Errorf("%v got = %v, want %v", tc.name, got, tc.wantErr)
			}
		})
	}
}

func TestClusterNameLength(t *testing.T) {
	tests := []struct {
		clusterName, name string
		wantErr           error
	}{
		{
			name:        "SuccessClusterNameLength",
			clusterName: "qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnm",
			wantErr:     nil,
		},
		{
			name:        "FailureClusterNameLength",
			clusterName: "qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnm12345",
			wantErr:     errors.New("number of characters in qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnm12345 should be less than 81"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got := ValidateClusterNameLength(tc.clusterName)
			if !reflect.DeepEqual(tc.wantErr, got) {
				t.Errorf("%v got = %v, want %v", tc.name, got, tc.wantErr)
			}
		})
	}
}

func TestValidateExternalEtcdSupport(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name        string
		wantCluster *Cluster
		wantErr     bool
	}{
		{
			name: "tinkerbell config without external etcd",
			wantCluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: TinkerbellDatacenterKind,
						Name: "eksa-unit-test",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "tinkerbell config with external etcd",
			wantCluster: &Cluster{
				Spec: ClusterSpec{
					ExternalEtcdConfiguration: &ExternalEtcdConfiguration{Count: 1},
					DatacenterRef: Ref{
						Kind: TinkerbellDatacenterKind,
						Name: "eksa-unit-test",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got := validateExternalEtcdSupport(tc.wantCluster)
			if tc.wantErr {
				g.Expect(got).To(MatchError(ContainSubstring("external etcd configuration is unsupported")))
			} else {
				g.Expect(got).To(Succeed())
			}
		})
	}
}

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
						Name:  "md-0",
						Count: ptr.Int(3),
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					ClusterNetwork: ClusterNetwork{
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
			testName: "namespace mismatch between cluster and datacenter",
			fileName: "cluster_1_20_namespace_mismatch_between_cluster_and_datacenter.yaml",
			wantErr:  true,
		},
		{
			testName: "namespace mismatch between cluster and machineconfigs",
			fileName: "cluster_1_20_namespace_mismatch_between_cluster_and_machineconfigs",
			wantErr:  true,
		},
		{
			testName: "valid 1.20 with non eksa resources",
			fileName: "testdata/cluster_1_20_with_non_eksa_resources.yaml",
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
			testName: "with valid ip proxy configuration",
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
			testName: "with valid domain name proxy configuration",
			fileName: "testdata/cluster_valid_domainname_proxy.yaml",
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
					ProxyConfiguration: &ProxyConfiguration{
						HttpProxy:  "http://google.com:1",
						HttpsProxy: "google.com:1",
						NoProxy:    []string{"localhost"},
					},
				},
			},
			wantErr: false,
		},
		{
			testName: "valid different tainted machine configs",
			fileName: "testdata/cluster_valid_taints_multiple_worker_node_groups.yaml",
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
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{
						{
							Name:  "md-0",
							Count: ptr.Int(3),
							MachineGroupRef: &Ref{
								Kind: VSphereMachineConfigKind,
								Name: "eksa-unit-test-2",
							},
							Taints: []v1.Taint{
								{
									Key:    "key1",
									Value:  "val1",
									Effect: v1.TaintEffectNoSchedule,
								},
							},
						},
						{
							Name:  "md-1",
							Count: ptr.Int(3),
							MachineGroupRef: &Ref{
								Kind: VSphereMachineConfigKind,
								Name: "eksa-unit-test-2",
							},
							Taints: []v1.Taint{
								{
									Key:    "key1",
									Value:  "val1",
									Effect: v1.TaintEffectPreferNoSchedule,
								},
							},
						},
					},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
			testName: "valid tainted workload cluster machine configs",
			fileName: "testdata/cluster_valid_taints_workload_cluster.yaml",
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
					ManagementCluster: ManagementCluster{
						Name: "mgmt",
					},
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
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{
						{
							Name:  "md-0",
							Count: ptr.Int(3),
							MachineGroupRef: &Ref{
								Kind: VSphereMachineConfigKind,
								Name: "eksa-unit-test-2",
							},
							Taints: []v1.Taint{
								{
									Key:    "key1",
									Value:  "val1",
									Effect: v1.TaintEffectNoSchedule,
								},
							},
						},
						{
							Name:  "md-1",
							Count: ptr.Int(3),
							MachineGroupRef: &Ref{
								Kind: VSphereMachineConfigKind,
								Name: "eksa-unit-test-2",
							},
							Taints: []v1.Taint{
								{
									Key:    "key1",
									Value:  "val1",
									Effect: v1.TaintEffectNoExecute,
								},
							},
						},
					},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
			testName:    "with invalid worker node group taints",
			fileName:    "testdata/cluster_invalid_taints.yaml",
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
			testName:    "invalid cluster name",
			fileName:    "testdata/cluster_invalid_cluster_name.yaml",
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
		{
			testName: "tinkerbell without worker nodes",
			fileName: "testdata/tinkerbell_cluster_without_worker_nodes.yaml",
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "single-node",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube123,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 1,
						Endpoint: &Endpoint{
							Host: "10.80.8.90",
						},
						MachineGroupRef: &Ref{
							Kind: TinkerbellMachineConfigKind,
							Name: "single-node-cp",
						},
					},
					WorkerNodeGroupConfigurations: nil,
					DatacenterRef: Ref{
						Kind: TinkerbellDatacenterKind,
						Name: "single-node",
					},
					ClusterNetwork: ClusterNetwork{
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
					ManagementCluster: ManagementCluster{
						Name: "single-node",
					},
				},
			},
			wantErr: false,
		},
		{
			testName:    "tinkerbell 1.21 cluster without worker nodes",
			fileName:    "testdata/tinkerbell_121cluster_without_worker_nodes.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "nontinkerbell datacenter without worker nodes",
			fileName:    "testdata/vsphere_cluster_without_worker_nodes.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "without worker nodes but has control plane taints",
			fileName:    "testdata/cluster_without_worker_nodes_has_cp_taints.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "Invalid registry",
			fileName:    "testdata/invalid_registry.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName: "valid package config",
			fileName: "testdata/cluster_package_configuration.yaml",
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
					Packages: &PackageConfiguration{
						Disable: true,
					},
				},
			},
			wantErr: false,
		},
		{
			testName:    "invalid controller package configuration",
			fileName:    "testdata/cluster_package_configuration_invalid.yaml",
			wantCluster: nil,
			wantErr:     true,
		},
		{
			testName:    "invalid package cronjob",
			fileName:    "testdata/cluster_package_cronjob_invalid.yaml",
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Name:  "md-0",
						Count: ptr.Int(3),
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
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Name:  "md-0",
						Count: ptr.Int(3),
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					ClusterNetwork: ClusterNetwork{
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
						Count: ptr.Int(3),
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
			matchError: fmt.Errorf("converting YAML to JSON: yaml: line 12: did not find expected key"),
		},
		{
			name: "Invalid key",
			args: args{
				fileName:      "testdata/invalid_key.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr:    true,
			matchError: fmt.Errorf("unmarshaling JSON: while decoding JSON: json: unknown field \"registryMirro rConfiguration\""),
		},
		{
			name: "Invalid yaml",
			args: args{
				fileName:      "testdata/invalid_format.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr:    true,
			matchError: fmt.Errorf("unable to parse testdata/invalid_format.yaml file: error converting YAML to JSON: yaml: did not find expected node content"),
		},
		{
			name: "Invalid spec field",
			args: args{
				fileName:      "testdata/invalid_spec_field.yaml",
				clusterConfig: &Cluster{},
			},
			wantErr:    true,
			matchError: fmt.Errorf("unmarshaling JSON: while decoding JSON: json: unknown field \"invalidField\""),
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
						Count: ptr.Int(3),
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

func Test_ParseClusterConfigFromContent(t *testing.T) {
	tests := []struct {
		name          string
		fileName      string
		clusterConfig KindAccessor
		expectedError error
	}{
		{
			name:          "Good cluster config parse",
			fileName:      "testdata/cluster_vsphere.yaml",
			clusterConfig: &Cluster{},
			expectedError: nil,
		},
		{
			name:          "non-splitable manifest",
			fileName:      "testdata/invalid_manifest.yaml",
			clusterConfig: &Cluster{},
			expectedError: errors.New("invalid Yaml document separator: \\nkey: value\\ninvalid_separator\\n"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content, err := ioutil.ReadFile(test.fileName)
			require.NoError(t, err)

			err = ParseClusterConfigFromContent(content, test.clusterConfig)

			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
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

func TestCluster_AddRemoveManagedByCLIAnnotation(t *testing.T) {
	g := NewWithT(t)
	c := &Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster_test",
		},
	}
	c.AddManagedByCLIAnnotation()
	val, ok := c.Annotations[ManagedByCLIAnnotation]

	g.Expect(ok).To(BeTrue())
	g.Expect(val).To(ContainSubstring("true"))

	c.ClearManagedByCLIAnnotation()
	_, ok = c.Annotations[ManagedByCLIAnnotation]
	g.Expect(ok).To(BeFalse())
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
		name              string
		datacenterRefKind string
		want              bool
		prev              *Endpoint
		new               *Endpoint
	}{
		{
			name:              "previous and new == nil",
			datacenterRefKind: VSphereDatacenterKind,
			want:              true,
			prev:              nil,
			new:               nil,
		},
		{
			name:              "previous == nil",
			datacenterRefKind: VSphereDatacenterKind,
			want:              false,
			prev:              nil,
			new:               &Endpoint{},
		},
		{
			name:              "new == nil",
			datacenterRefKind: VSphereDatacenterKind,
			want:              false,
			prev:              &Endpoint{},
			new:               nil,
		},
		{
			name:              "previous == new",
			datacenterRefKind: VSphereDatacenterKind,
			want:              true,
			prev:              &Endpoint{Host: "host"},
			new:               &Endpoint{Host: "host"},
		},
		{
			name:              "previous != new",
			datacenterRefKind: VSphereDatacenterKind,
			want:              false,
			prev:              &Endpoint{Host: "host"},
			new:               &Endpoint{Host: "new-host"},
		},
		{
			name:              "same host, no port",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              true,
			prev:              &Endpoint{Host: "host"},
			new:               &Endpoint{Host: "host"},
		},
		{
			name:              "same host, same default port",
			want:              true,
			datacenterRefKind: CloudStackDatacenterKind,
			prev:              &Endpoint{Host: "host:6443"},
			new:               &Endpoint{Host: "host:6443"},
		},
		{
			name:              "same host, same custom port",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              true,
			prev:              &Endpoint{Host: "host:6442"},
			new:               &Endpoint{Host: "host:6442"},
		},
		{
			name:              "different host, no port",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              false,
			prev:              &Endpoint{Host: "host"},
			new:               &Endpoint{Host: "new-host"},
		},
		{
			name:              "different host, different port",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              false,
			prev:              &Endpoint{Host: "host:6443"},
			new:               &Endpoint{Host: "new-host:6442"},
		},
		{
			name:              "same host, old custom port, new no port",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              false,
			prev:              &Endpoint{Host: "host:6442"},
			new:               &Endpoint{Host: "host"},
		},
		{
			name:              "same host, old default port, new no port",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              true,
			prev:              &Endpoint{Host: "host:6443"},
			new:               &Endpoint{Host: "host"},
		},
		{
			name:              "same host, old no port, new custom port",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              false,
			prev:              &Endpoint{Host: "host"},
			new:               &Endpoint{Host: "host:6442"},
		},
		{
			name:              "same host, old no port, new default port",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              true,
			prev:              &Endpoint{Host: "host"},
			new:               &Endpoint{Host: "host:6443"},
		},
		{
			name:              "same host, old default port, new no port",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              true,
			prev:              &Endpoint{Host: "host:6443"},
			new:               &Endpoint{Host: "host"},
		},
		{
			name:              "invalid host",
			datacenterRefKind: CloudStackDatacenterKind,
			want:              false,
			prev:              &Endpoint{Host: "host:6443"},
			new:               &Endpoint{Host: "host::"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != tt.prev.Equal(tt.new, tt.datacenterRefKind) {
				t.Errorf("Endpoint %+v should be equals to  %+v", tt.prev, tt.new)
			}
		})
	}
}

func TestCloudStackEndPointEquals(t *testing.T) {
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
			name: "same host, no port",
			want: true,
			prev: &Endpoint{Host: "host"},
			new:  &Endpoint{Host: "host"},
		},
		{
			name: "same host, same default port",
			want: true,
			prev: &Endpoint{Host: "host:6443"},
			new:  &Endpoint{Host: "host:6443"},
		},
		{
			name: "same host, same custom port",
			want: true,
			prev: &Endpoint{Host: "host:6442"},
			new:  &Endpoint{Host: "host:6442"},
		},
		{
			name: "different host, no port",
			want: false,
			prev: &Endpoint{Host: "host"},
			new:  &Endpoint{Host: "new-host"},
		},
		{
			name: "different host, different port",
			want: false,
			prev: &Endpoint{Host: "host:6443"},
			new:  &Endpoint{Host: "new-host:6442"},
		},
		{
			name: "same host, old custom port, new no port",
			want: false,
			prev: &Endpoint{Host: "host:6442"},
			new:  &Endpoint{Host: "host"},
		},
		{
			name: "same host, old default port, new no port",
			want: true,
			prev: &Endpoint{Host: "host:6443"},
			new:  &Endpoint{Host: "host"},
		},
		{
			name: "same host, old no port, new custom port",
			want: false,
			prev: &Endpoint{Host: "host"},
			new:  &Endpoint{Host: "host:6442"},
		},
		{
			name: "same host, old no port, new default port",
			want: true,
			prev: &Endpoint{Host: "host"},
			new:  &Endpoint{Host: "host:6443"},
		},
		{
			name: "same host, old default port, new no port",
			want: true,
			prev: &Endpoint{Host: "host:6443"},
			new:  &Endpoint{Host: "host"},
		},
		{
			name: "invalid host",
			want: false,
			prev: &Endpoint{Host: "host:6443"},
			new:  &Endpoint{Host: "host::"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != tt.prev.CloudStackEqual(tt.new) {
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
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
				Pods: Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
						"1.2.3.4/6",
					},
				},
			},
			new: &ClusterNetwork{
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
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
			name: "previous == new, same cni, older format",
			want: true,
			prev: &ClusterNetwork{
				CNI: Cilium,
			},
			new: &ClusterNetwork{
				CNI: Cilium,
			},
		},
		{
			name: "previous != new, diff CNI, older format",
			want: false,
			prev: &ClusterNetwork{
				CNI: Kindnetd,
			},
			new: &ClusterNetwork{
				CNI: Cilium,
			},
		},
		{
			name: "previous == new, same cni, diff format",
			want: true,
			prev: &ClusterNetwork{
				CNI: Cilium,
			},
			new: &ClusterNetwork{
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
			},
		},
		{
			name: "previous != new, same cni, diff format, diff cilium policy  mode",
			want: false,
			prev: &ClusterNetwork{
				CNI: Cilium,
			},
			new: &ClusterNetwork{
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{PolicyEnforcementMode: "always"}},
			},
		},
		{
			name: "previous != new, different cni, different format",
			want: false,
			prev: &ClusterNetwork{
				CNI: Kindnetd,
			},
			new: &ClusterNetwork{
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
			},
		},
		{
			name: "previous != new, new cniConfig format, diff cni",
			want: false,
			prev: &ClusterNetwork{
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
			},
			new: &ClusterNetwork{
				CNIConfig: &CNIConfig{Kindnetd: &KindnetdConfig{}},
			},
		},
		{
			name: "previous == new, new cniConfig format, same cni",
			want: true,
			prev: &ClusterNetwork{
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
			},
			new: &ClusterNetwork{
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
			},
		},
		{
			name: "previous != new, new cniConfig format, same cilium cni, diff configuration",
			want: false,
			prev: &ClusterNetwork{
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{PolicyEnforcementMode: "always"}},
			},
			new: &ClusterNetwork{
				CNIConfig: &CNIConfig{Cilium: &CiliumConfig{PolicyEnforcementMode: "default"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != tt.new.Equal(tt.prev) {
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

func TestValidateNetworking(t *testing.T) {
	nodeCidrMaskSize := new(int)
	*nodeCidrMaskSize = 28
	tests := []struct {
		name    string
		wantErr error
		cluster *Cluster
	}{
		{
			name:    "both formats used",
			wantErr: fmt.Errorf("invalid format for cni plugin: both old and new formats used, use only the CNIConfig field"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/8",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						CNI:       Cilium,
						CNIConfig: &CNIConfig{Cilium: &CiliumConfig{}},
					},
				},
			},
		},
		{
			name:    "deprecated CNI field",
			wantErr: nil,
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/8",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "no CNI plugin input",
			wantErr: fmt.Errorf("cni not specified"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/8",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						CNI:       "",
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "node cidr mask size valid",
			wantErr: nil,
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/24",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						Nodes: &Nodes{
							CIDRMaskSize: nodeCidrMaskSize,
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "node cidr mask size invalid",
			wantErr: fmt.Errorf("the size of pod subnet with mask 30 is smaller than or equal to the size of node subnet with mask 28"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/30",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						Nodes: &Nodes{
							CIDRMaskSize: nodeCidrMaskSize,
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "node cidr mask size invalid diff",
			wantErr: fmt.Errorf("pod subnet mask (6) and node-mask (28) difference is greater than 16"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/6",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						Nodes: &Nodes{
							CIDRMaskSize: nodeCidrMaskSize,
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "both pods CIDR block and service CIDR block do not conflict with control plane endpoint",
			wantErr: nil,
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: SnowDatacenterKind,
					},
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Endpoint: &Endpoint{
							Host: "192.168.1.10",
						},
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"10.1.0.0/16",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"10.96.0.0/12",
							},
						},
						Nodes: &Nodes{
							CIDRMaskSize: nodeCidrMaskSize,
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "invalid pods CIDR block",
			wantErr: fmt.Errorf("invalid CIDR block format for Pods: {[1.2.3]}. Please specify a valid CIDR block for pod subnet"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						Nodes: &Nodes{
							CIDRMaskSize: nodeCidrMaskSize,
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "pods CIDR block conflicts with control plane endpoint",
			wantErr: fmt.Errorf("control plane endpoint 192.168.1.10 conflicts with pods CIDR block 192.168.1.0/24"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: SnowDatacenterKind,
					},
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Endpoint: &Endpoint{
							Host: "192.168.1.10",
						},
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"192.168.1.0/24",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/6",
							},
						},
						Nodes: &Nodes{
							CIDRMaskSize: nodeCidrMaskSize,
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "invalid services CIDR block",
			wantErr: fmt.Errorf("invalid CIDR block for Services: {[1.2.3]}. Please specify a valid CIDR block for service subnet"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/6",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3",
							},
						},
						Nodes: &Nodes{
							CIDRMaskSize: nodeCidrMaskSize,
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "services CIDR block conflicts with control plane endpoint",
			wantErr: fmt.Errorf("control plane endpoint 192.168.1.10 conflicts with services CIDR block 192.168.1.0/24"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: SnowDatacenterKind,
					},
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Endpoint: &Endpoint{
							Host: "192.168.1.10",
						},
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/6",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"192.168.1.0/24",
							},
						},
						Nodes: &Nodes{
							CIDRMaskSize: nodeCidrMaskSize,
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "control plane endpoint is invalid",
			wantErr: fmt.Errorf("control plane endpoint 192.168.1 is invalid"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: SnowDatacenterKind,
					},
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Endpoint: &Endpoint{
							Host: "192.168.1",
						},
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/6",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"10.1.0.0/24",
							},
						},
						Nodes: &Nodes{
							CIDRMaskSize: nodeCidrMaskSize,
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "vsphere cluster uses cilium CNI",
			wantErr: nil,
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/8",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						CNI:       Cilium,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "vsphere cluster uses kindnetd CNI",
			wantErr: fmt.Errorf("kindnetd is only supported on Docker provider for development and testing. For all other providers please use Cilium CNI"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/8",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						CNI:       Kindnetd,
						CNIConfig: nil,
					},
				},
			},
		},
		{
			name:    "vsphere cluster uses kindnetd CNIConfig",
			wantErr: fmt.Errorf("kindnetd is only supported on Docker provider for development and testing. For all other providers please use Cilium CNI"),
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/8",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						CNI:       "",
						CNIConfig: &CNIConfig{Kindnetd: &KindnetdConfig{}},
					},
				},
			},
		},
		{
			name:    "docker cluster uses kindnetd CNIConfig",
			wantErr: nil,
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: DockerDatacenterKind,
					},
					ClusterNetwork: ClusterNetwork{
						Pods: Pods{
							CidrBlocks: []string{
								"1.2.3.4/8",
							},
						},
						Services: Services{
							CidrBlocks: []string{
								"1.2.3.4/7",
							},
						},
						CNI:       "",
						CNIConfig: &CNIConfig{Kindnetd: &KindnetdConfig{}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateNetworking(tt.cluster)
			if !reflect.DeepEqual(tt.wantErr, got) {
				t.Errorf("%v got = %v, want %v", tt.name, got, tt.wantErr)
			}
		})
	}
}

func TestValidateCNIConfig(t *testing.T) {
	tests := []struct {
		name           string
		wantErr        error
		clusterNetwork *ClusterNetwork
	}{
		{
			name:    "CNI plugin not specified",
			wantErr: fmt.Errorf("validating cniConfig: no cni plugin specified"),
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{},
			},
		},
		{
			name:    "multiple CNI plugins specified",
			wantErr: fmt.Errorf("validating cniConfig: cannot specify more than one cni plugins"),
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{
					Cilium:   &CiliumConfig{},
					Kindnetd: &KindnetdConfig{},
				},
			},
		},
		{
			name:    "invalid cilium policy enforcement mode",
			wantErr: fmt.Errorf("validating cniConfig: cilium policyEnforcementMode \"invalid\" not supported"),
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{
					Cilium: &CiliumConfig{
						PolicyEnforcementMode: "invalid",
					},
				},
			},
		},
		{
			name:    "invalid cilium policy enforcement mode and > 1 plugins",
			wantErr: fmt.Errorf("validating cniConfig: [cilium policyEnforcementMode \"invalid\" not supported, cannot specify more than one cni plugins]"),
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{
					Cilium: &CiliumConfig{
						PolicyEnforcementMode: "invalid",
					},
					Kindnetd: &KindnetdConfig{},
				},
			},
		},
		{
			name:    "valid cilium policy enforcement mode",
			wantErr: nil,
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{
					Cilium: &CiliumConfig{
						PolicyEnforcementMode: "default",
					},
				},
			},
		},
		{
			name:    "valid cilium policy enforcement mode",
			wantErr: nil,
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{
					Cilium: &CiliumConfig{
						PolicyEnforcementMode: "always",
					},
				},
			},
		},
		{
			name:    "valid cilium policy enforcement mode",
			wantErr: nil,
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{
					Cilium: &CiliumConfig{
						PolicyEnforcementMode: "never",
					},
				},
			},
		},
		{
			name:    "CiliumSkipUpgradeWithoutOtherFields",
			wantErr: nil,
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{
					Cilium: &CiliumConfig{
						SkipUpgrade: ptr.Bool(true),
					},
				},
			},
		},
		{
			name: "CiliumSkipUpgradeWithOtherFields",
			wantErr: fmt.Errorf("validating cniConfig: when using skipUpgrades for cilium all " +
				"other fields must be empty"),
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{
					Cilium: &CiliumConfig{
						SkipUpgrade:           ptr.Bool(true),
						PolicyEnforcementMode: "never",
					},
				},
			},
		},
		{
			name: "CiliumSkipUpgradeExplicitFalseWithOtherFields",
			clusterNetwork: &ClusterNetwork{
				CNIConfig: &CNIConfig{
					Cilium: &CiliumConfig{
						SkipUpgrade:           ptr.Bool(false),
						PolicyEnforcementMode: "never",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateCNIConfig(tt.clusterNetwork.CNIConfig)
			if !reflect.DeepEqual(tt.wantErr, got) {
				t.Errorf("%v got = %v, want %v", tt.name, got, tt.wantErr)
			}
		})
	}
}

func TestValidateMirrorConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
		cluster *Cluster
	}{
		{
			name:    "registry mirror not specified",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					RegistryMirrorConfiguration: nil,
				},
			},
		},
		{
			name:    "endpoint not specified",
			wantErr: "no value set for RegistryMirrorConfiguration.Endpoint",
			cluster: &Cluster{
				Spec: ClusterSpec{
					RegistryMirrorConfiguration: &RegistryMirrorConfiguration{
						Endpoint: "",
					},
				},
			},
		},
		{
			name:    "invalid port",
			wantErr: "registry mirror port 65536 is invalid",
			cluster: &Cluster{
				Spec: ClusterSpec{
					RegistryMirrorConfiguration: &RegistryMirrorConfiguration{
						Endpoint: "1.2.3.4",
						Port:     "65536",
					},
				},
			},
		},
		{
			name:    "multiple mappings for curated packages",
			wantErr: "only one registry mirror for curated packages is suppported",
			cluster: &Cluster{
				Spec: ClusterSpec{
					RegistryMirrorConfiguration: &RegistryMirrorConfiguration{
						Endpoint: "1.2.3.4",
						Port:     "30003",
						OCINamespaces: []OCINamespace{
							{
								Registry:  "783794618700.dkr.ecr.us-west-2.amazonaws.com",
								Namespace: "",
							},
							{
								Registry:  "783794618700.dkr.ecr.us-east-1.amazonaws.com",
								Namespace: "",
							},
						},
					},
				},
			},
		},
		{
			name:    "empty registry in OCINamespace",
			wantErr: "registry can't be set to empty in OCINamespaces",
			cluster: &Cluster{
				Spec: ClusterSpec{
					RegistryMirrorConfiguration: &RegistryMirrorConfiguration{
						Endpoint: "1.2.3.4",
						Port:     "30003",
						OCINamespaces: []OCINamespace{
							{
								Registry:  "",
								Namespace: "",
							},
						},
					},
				},
			},
		},
		{
			name:    "insecureSkipVerify on snow provider",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					RegistryMirrorConfiguration: &RegistryMirrorConfiguration{
						Endpoint:           "1.2.3.4",
						Port:               "443",
						InsecureSkipVerify: true,
					},
					DatacenterRef: Ref{
						Kind: SnowDatacenterKind,
					},
				},
			},
		},
		{
			name:    "insecureSkipVerify on nutanix provider",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					RegistryMirrorConfiguration: &RegistryMirrorConfiguration{
						Endpoint:           "1.2.3.4",
						Port:               "443",
						InsecureSkipVerify: true,
					},
					DatacenterRef: Ref{
						Kind: NutanixDatacenterKind,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := validateMirrorConfig(tt.cluster)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestValidateAutoscalingConfig(t *testing.T) {
	tests := []struct {
		name                         string
		wantErr                      string
		workerNodeGroupConfiguration *WorkerNodeGroupConfiguration
	}{
		{
			name:                         "autoscaling config nil",
			wantErr:                      "",
			workerNodeGroupConfiguration: nil,
		},
		{
			name:    "autoscaling config valid",
			wantErr: "",
			workerNodeGroupConfiguration: &WorkerNodeGroupConfiguration{
				Count: ptr.Int(2),
				AutoScalingConfiguration: &AutoScalingConfiguration{
					MinCount: 1,
					MaxCount: 2,
				},
			},
		},
		{
			name:    "negative min count",
			wantErr: "min count must be non negative",
			workerNodeGroupConfiguration: &WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
				AutoScalingConfiguration: &AutoScalingConfiguration{
					MinCount: -1,
				},
			},
		},
		{
			name:    "min count > max count",
			wantErr: "min count must be no greater than max count",
			workerNodeGroupConfiguration: &WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
				AutoScalingConfiguration: &AutoScalingConfiguration{
					MinCount: 2,
					MaxCount: 1,
				},
			},
		},
		{
			name:    "count < min count",
			wantErr: "min count must be less than or equal to count",
			workerNodeGroupConfiguration: &WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
				AutoScalingConfiguration: &AutoScalingConfiguration{
					MinCount: 2,
					MaxCount: 3,
				},
			},
		},
		{
			name:    "count > max count",
			wantErr: "max count must be greater than or equal to count",
			workerNodeGroupConfiguration: &WorkerNodeGroupConfiguration{
				Count: ptr.Int(4),
				AutoScalingConfiguration: &AutoScalingConfiguration{
					MinCount: 2,
					MaxCount: 3,
				},
			},
		},
		{
			name:    "count < 0 with nil autoscaling",
			wantErr: "worker node count must be zero or greater if autoscaling is not enabled",
			workerNodeGroupConfiguration: &WorkerNodeGroupConfiguration{
				Count: ptr.Int(-1),
			},
		},
		{
			name:    "nil autoscaling",
			wantErr: "",
			workerNodeGroupConfiguration: &WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := validateAutoscalingConfig(tt.workerNodeGroupConfiguration)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestClusterRegistryAuth(t *testing.T) {
	tests := []struct {
		name    string
		cluster *Cluster
		want    bool
	}{
		{
			name: "with registry mirror auth",
			cluster: &Cluster{
				Spec: ClusterSpec{
					RegistryMirrorConfiguration: &RegistryMirrorConfiguration{
						Endpoint:     "1.2.3.4",
						Port:         "443",
						Authenticate: true,
					},
				},
			},
			want: true,
		},
		{
			name: "without registry mirror auth",
			cluster: &Cluster{
				Spec: ClusterSpec{
					RegistryMirrorConfiguration: &RegistryMirrorConfiguration{
						Endpoint: "1.2.3.4",
						Port:     "443",
					},
				},
			},
			want: false,
		},
		{
			name:    "without registry mirror",
			cluster: &Cluster{},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster.RegistryAuth()).To(Equal(tt.want))
		})
	}
}

func TestClusterProxyConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		cluster *Cluster
		want    map[string]string
	}{
		{
			name: "with proxy configuration",
			cluster: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Endpoint: &Endpoint{
							Host: "1.2.3.4",
						},
					},
					ProxyConfiguration: &ProxyConfiguration{
						HttpProxy:  "test-http",
						HttpsProxy: "test-https",
						NoProxy:    []string{"test-noproxy-1", "test-noproxy-2", "test-noproxy-3"},
					},
				},
			},
			want: map[string]string{
				"HTTP_PROXY":  "test-http",
				"HTTPS_PROXY": "test-https",
				"NO_PROXY":    "test-noproxy-1,test-noproxy-2,test-noproxy-3,1.2.3.4",
			},
		},
		{
			name:    "without proxy configuration",
			cluster: &Cluster{},
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster.ProxyConfiguration()).To(Equal(tt.want))
		})
	}
}

func TestValidateControlPlaneEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
		cluster *Cluster
	}{
		{
			name:    "docker provider - control plane endpoint is not set",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					DatacenterRef: Ref{
						Kind: DockerDatacenterKind,
					},
					ControlPlaneConfiguration: ControlPlaneConfiguration{},
				},
			},
		},
		{
			name:    "control plane ip is not set",
			wantErr: "cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty",
			cluster: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Endpoint: &Endpoint{
							Host: "",
						},
					},
				},
			},
		},
		{
			name:    "control plane ip is set",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := validateControlPlaneEndpoint(tt.cluster)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestValidateCPUpgradeRolloutStrategy(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
		cluster *Cluster
	}{
		{
			name:    "rolling upgrade strategy invalid",
			wantErr: "ControlPlaneConfiguration: only 'RollingUpdate' and 'InPlace' are supported for upgrade rollout strategy type",
			cluster: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						UpgradeRolloutStrategy: &ControlPlaneUpgradeRolloutStrategy{Type: "NotRollingUpdate"},
					},
				},
			},
		},
		{
			name:    "rolling upgrade knobs not specified",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						UpgradeRolloutStrategy: &ControlPlaneUpgradeRolloutStrategy{Type: "RollingUpdate"},
					},
				},
			},
		},
		{
			name:    "rolling upgrade invalid",
			wantErr: "maxSurge for control plane must be 0 or 1",
			cluster: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						UpgradeRolloutStrategy: &ControlPlaneUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: ControlPlaneRollingUpdateParams{MaxSurge: 2}},
					},
				},
			},
		},
		{
			name:    "rolling upgrade knobs valid value 0",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						UpgradeRolloutStrategy: &ControlPlaneUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: ControlPlaneRollingUpdateParams{MaxSurge: 0}},
					},
				},
			},
		},
		{
			name:    "rolling upgrade knobs valid value 1",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						UpgradeRolloutStrategy: &ControlPlaneUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: ControlPlaneRollingUpdateParams{MaxSurge: 1}},
					},
				},
			},
		},
		{
			name:    "rolling upgrade knobs invalid value -1",
			wantErr: "ControlPlaneConfiguration: maxSurge for control plane cannot be a negative value",
			cluster: &Cluster{
				Spec: ClusterSpec{
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						UpgradeRolloutStrategy: &ControlPlaneUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: ControlPlaneRollingUpdateParams{MaxSurge: -1}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := validateCPUpgradeRolloutStrategy(tt.cluster)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestValidateMDUpgradeRolloutStrategy(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
		cluster *Cluster
	}{
		{
			name:    "rolling upgrade strategy invalid",
			wantErr: "WorkerNodeGroupConfiguration: only 'RollingUpdate' and 'InPlace' are supported for upgrade rollout strategy type",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "NotRollingUpdate"},
					}},
				},
			},
		},
		{
			name:    "rolling upgrade knobs not specified",
			wantErr: "WorkerNodeGroupConfiguration: maxSurge and maxUnavailable not specified or are 0. maxSurge and maxUnavailable cannot both be 0",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate"},
					}},
				},
			},
		},
		{
			name:    "rolling upgrade invalid",
			wantErr: "WorkerNodeGroupConfiguration: maxSurge and maxUnavailable not specified or are 0. maxSurge and maxUnavailable cannot both be 0",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: 0, MaxUnavailable: 0}},
					}},
				},
			},
		},
		{
			name:    "rolling upgrade knobs valid value 0,1",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: 0, MaxUnavailable: 1}},
					}},
				},
			},
		},
		{
			name:    "rolling upgrade knobs valid value 1,0",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: 1, MaxUnavailable: 0}},
					}},
				},
			},
		},
		{
			name:    "rolling upgrade knobs valid value 5,0",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: 5, MaxUnavailable: 0}},
					}},
				},
			},
		},
		{
			name:    "rolling upgrade knobs valid value 3,1",
			wantErr: "",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: 3, MaxUnavailable: 1}},
					}},
				},
			},
		},
		{
			name:    "rolling upgrade knobs invalid values 3,-1",
			wantErr: "WorkerNodeGroupConfiguration: maxSurge and maxUnavailable values cannot be negative",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: 3, MaxUnavailable: -1}},
					}},
				},
			},
		},
		{
			name:    "rolling upgrade knobs invalid values -3,1",
			wantErr: "WorkerNodeGroupConfiguration: maxSurge and maxUnavailable values cannot be negative",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: -3, MaxUnavailable: 1}},
					}},
				},
			},
		},
		{
			name:    "rolling upgrade knobs invalid values -3,-1",
			wantErr: "WorkerNodeGroupConfiguration: maxSurge and maxUnavailable values cannot be negative",
			cluster: &Cluster{
				Spec: ClusterSpec{
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: -3, MaxUnavailable: -1}},
					}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := validateMDUpgradeRolloutStrategy(&tt.cluster.Spec.WorkerNodeGroupConfigurations[0])
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestValidateEksaVersion(t *testing.T) {
	tests := []struct {
		name       string
		wantErr    string
		version    string
		bundlesRef *BundlesRef
	}{
		{
			name:       "both bundlesref and version",
			wantErr:    "cannot pass both bundlesRef and eksaVersion",
			version:    "v0.0.0",
			bundlesRef: &BundlesRef{},
		},
		{
			name:       "eksaversion success",
			wantErr:    "",
			version:    "v0.0.0",
			bundlesRef: nil,
		},
		{
			name:       "eksaversion fail",
			wantErr:    "eksaVersion is not a valid semver",
			version:    "invalid",
			bundlesRef: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			config := &Cluster{
				Spec: ClusterSpec{
					EksaVersion: (*EksaVersion)(&tt.version),
					BundlesRef:  tt.bundlesRef,
				},
			}
			err := validateEksaVersion(config)
			if tt.wantErr == "" {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestGetClusterDefaultKubernetesVersion(t *testing.T) {
	g := NewWithT(t)
	g.Expect(GetClusterDefaultKubernetesVersion()).To(Equal(Kube128))
}

func TestClusterWorkerNodeConfigCount(t *testing.T) {
	tests := []struct {
		name    string
		cluster *Cluster
		want    []WorkerNodeGroupConfiguration
	}{
		{
			name: "with worker node config count",
			cluster: &Cluster{
				Spec: ClusterSpec{},
			},
			want: []WorkerNodeGroupConfiguration{
				{
					Name:                     "",
					Count:                    ptr.Int(5),
					AutoScalingConfiguration: nil,
					MachineGroupRef:          nil,
					Taints:                   nil,
					Labels:                   nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := NewClusterGenerate("test-cluster", WorkerNodeConfigCount(5))
			g := NewWithT(t)
			g.Expect(cg.Spec.WorkerNodeGroupConfigurations).To(Equal(tt.want))
		})
	}
}

func TestClusterCPUpgradeRolloutStrategyNil(t *testing.T) {
	tests := []struct {
		name    string
		cluster *Cluster
		want    ControlPlaneConfiguration
	}{
		{
			name: "with control plane rollout upgrade strategy nil",
			cluster: &Cluster{
				Spec: ClusterSpec{},
			},
			want: ControlPlaneConfiguration{
				Endpoint:               nil,
				Count:                  1,
				MachineGroupRef:        nil,
				Taints:                 nil,
				Labels:                 nil,
				UpgradeRolloutStrategy: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := NewClusterGenerate("test-cluster", ControlPlaneConfigCount(1))
			g := NewWithT(t)
			g.Expect(cg.Spec.ControlPlaneConfiguration).To(Equal(tt.want))
		})
	}
}

func TestClusterCPUpgradeRolloutStrategyNotNil(t *testing.T) {
	tests := []struct {
		name    string
		cluster *Cluster
		want    ControlPlaneConfiguration
	}{
		{
			name: "with control plane rollout upgrade strategy non-nil",
			cluster: &Cluster{
				Spec: ClusterSpec{},
			},
			want: ControlPlaneConfiguration{
				Endpoint:               nil,
				Count:                  1,
				MachineGroupRef:        nil,
				Taints:                 nil,
				Labels:                 nil,
				UpgradeRolloutStrategy: &ControlPlaneUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: ControlPlaneRollingUpdateParams{MaxSurge: 5}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := NewClusterGenerate("test-cluster", ControlPlaneConfigCount(1), WithCPUpgradeRolloutStrategy(5, 2))
			g := NewWithT(t)
			g.Expect(cg.Spec.ControlPlaneConfiguration).To(Equal(tt.want))
		})
	}
}

func TestClusterMDUpgradeRolloutStrategyNil(t *testing.T) {
	tests := []struct {
		name    string
		cluster *Cluster
		want    []WorkerNodeGroupConfiguration
	}{
		{
			name: "with md rollout upgrade strategy knobs not specified",
			cluster: &Cluster{
				Spec: ClusterSpec{},
			},
			want: []WorkerNodeGroupConfiguration{
				{
					Count:           ptr.Int(1),
					MachineGroupRef: nil,
					Taints:          nil,
					Labels:          nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := NewClusterGenerate("test-cluster", WorkerNodeConfigCount(1))
			g := NewWithT(t)
			g.Expect(cg.Spec.WorkerNodeGroupConfigurations).To(Equal(tt.want))
		})
	}
}

func TestClusterMDUpgradeRolloutStrategyNotNil(t *testing.T) {
	tests := []struct {
		name    string
		cluster *Cluster
		want    []WorkerNodeGroupConfiguration
	}{
		{
			name: "with md rollout upgrade strategy",
			cluster: &Cluster{
				Spec: ClusterSpec{},
			},
			want: []WorkerNodeGroupConfiguration{
				{
					Count:                  ptr.Int(1),
					MachineGroupRef:        nil,
					Taints:                 nil,
					Labels:                 nil,
					UpgradeRolloutStrategy: &WorkerNodesUpgradeRolloutStrategy{Type: "RollingUpdate", RollingUpdate: WorkerNodesRollingUpdateParams{MaxSurge: 5, MaxUnavailable: 2}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := NewClusterGenerate("test-cluster", WorkerNodeConfigCount(1), WithWorkerMachineUpgradeRolloutStrategy(5, 2))
			g := NewWithT(t)
			g.Expect(cg.Spec.WorkerNodeGroupConfigurations).To(Equal(tt.want))
		})
	}
}
