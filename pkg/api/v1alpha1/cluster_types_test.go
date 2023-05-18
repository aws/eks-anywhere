package v1alpha1_test

import (
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestClusterMachineConfigRefs(t *testing.T) {
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			KubernetesVersion: v1alpha1.Kube119,
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Count: 3,
				Endpoint: &v1alpha1.Endpoint{
					Host: "test-ip",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			},
			WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(3),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: v1alpha1.VSphereMachineConfigKind,
						Name: "eksa-unit-test-1",
					},
				},
				{
					Count: ptr.Int(3),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: v1alpha1.VSphereMachineConfigKind,
						Name: "eksa-unit-test-2",
					},
				},
				{
					Count: ptr.Int(5),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: v1alpha1.VSphereMachineConfigKind,
						Name: "eksa-unit-test", // This tests duplicates
					},
				},
			},
			ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{
				MachineGroupRef: &v1alpha1.Ref{
					Kind: v1alpha1.VSphereMachineConfigKind,
					Name: "eksa-unit-test-etcd",
				},
			},
			DatacenterRef: v1alpha1.Ref{
				Kind: v1alpha1.VSphereDatacenterKind,
				Name: "eksa-unit-test",
			},
		},
	}

	want := []v1alpha1.Ref{
		{
			Kind: v1alpha1.VSphereMachineConfigKind,
			Name: "eksa-unit-test",
		},
		{
			Kind: v1alpha1.VSphereMachineConfigKind,
			Name: "eksa-unit-test-1",
		},
		{
			Kind: v1alpha1.VSphereMachineConfigKind,
			Name: "eksa-unit-test-2",
		},
		{
			Kind: v1alpha1.VSphereMachineConfigKind,
			Name: "eksa-unit-test-etcd",
		},
	}

	got := cluster.MachineConfigRefs()

	if !v1alpha1.RefSliceEqual(got, want) {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}

func TestClusterIsSelfManaged(t *testing.T) {
	testCases := []struct {
		testName string
		cluster  *v1alpha1.Cluster
		want     bool
	}{
		{
			testName: "empty name",
			cluster:  &v1alpha1.Cluster{},
			want:     true,
		},
		{
			testName: "name same as self",
			cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-1",
				},
				Spec: v1alpha1.ClusterSpec{
					ManagementCluster: v1alpha1.ManagementCluster{
						Name: "cluster-1",
					},
				},
			},
			want: true,
		},
		{
			testName: "name different tha self",
			cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-2",
				},
				Spec: v1alpha1.ClusterSpec{
					ManagementCluster: v1alpha1.ManagementCluster{
						Name: "cluster-1",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster.IsSelfManaged()).To(Equal(tt.want))
		})
	}
}

func TestClusterSetManagedBy(t *testing.T) {
	c := &v1alpha1.Cluster{}
	managementClusterName := "managament-cluster"
	c.SetManagedBy(managementClusterName)

	g := NewWithT(t)
	g.Expect(c.IsSelfManaged()).To(BeFalse())
	g.Expect(c.ManagedBy()).To(Equal(managementClusterName))
}

func TestClusterSetSelfManaged(t *testing.T) {
	c := &v1alpha1.Cluster{}
	c.SetSelfManaged()

	g := NewWithT(t)
	g.Expect(c.IsSelfManaged()).To(BeTrue())
}

func TestClusterManagementClusterEqual(t *testing.T) {
	testCases := []struct {
		testName                                 string
		cluster1SelfManaged, cluster2SelfManaged bool
		want                                     bool
	}{
		{
			testName:            "both self managed",
			cluster1SelfManaged: true,
			cluster2SelfManaged: true,
			want:                true,
		},
		{
			testName:            "both managed",
			cluster1SelfManaged: false,
			cluster2SelfManaged: false,
			want:                true,
		},
		{
			testName:            "one managed, one self managed",
			cluster1SelfManaged: false,
			cluster2SelfManaged: true,
			want:                false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{}
			setSelfManaged(cluster1, tt.cluster1SelfManaged)
			cluster2 := &v1alpha1.Cluster{}
			setSelfManaged(cluster2, tt.cluster2SelfManaged)

			g := NewWithT(t)
			g.Expect(cluster1.ManagementClusterEqual(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterResolvConfEqual(t *testing.T) {
	testCases := []struct {
		testName                               string
		cluster1ResolvConf, cluster2ResolvConf string
		want                                   bool
	}{
		{
			testName:           "both empty",
			cluster1ResolvConf: "",
			cluster2ResolvConf: "",
			want:               true,
		},
		{
			testName:           "both defined",
			cluster1ResolvConf: "my-file.conf",
			cluster2ResolvConf: "my-file.conf",
			want:               true,
		},
		{
			testName:           "one empty, one defined",
			cluster1ResolvConf: "",
			cluster2ResolvConf: "my-file.conf",
			want:               false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ClusterNetwork: v1alpha1.ClusterNetwork{
						DNS: v1alpha1.DNS{
							ResolvConf: &v1alpha1.ResolvConf{
								Path: tt.cluster1ResolvConf,
							},
						},
					},
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ClusterNetwork: v1alpha1.ClusterNetwork{
						DNS: v1alpha1.DNS{
							ResolvConf: &v1alpha1.ResolvConf{
								Path: tt.cluster2ResolvConf,
							},
						},
					},
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Spec.ClusterNetwork.DNS.ResolvConf.Equal(cluster2.Spec.ClusterNetwork.DNS.ResolvConf)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualKubernetesVersion(t *testing.T) {
	testCases := []struct {
		testName                         string
		cluster1Version, cluster2Version v1alpha1.KubernetesVersion
		want                             bool
	}{
		{
			testName:        "both empty",
			cluster1Version: "",
			cluster2Version: "",
			want:            true,
		},
		{
			testName:        "one empty, one exists",
			cluster1Version: "",
			cluster2Version: v1alpha1.Kube118,
			want:            false,
		},
		{
			testName:        "both exists, diff",
			cluster1Version: v1alpha1.Kube118,
			cluster2Version: v1alpha1.Kube119,
			want:            false,
		},
		{
			testName:        "both exists, same",
			cluster1Version: v1alpha1.Kube118,
			cluster2Version: v1alpha1.Kube118,
			want:            true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: tt.cluster1Version,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					KubernetesVersion: tt.cluster2Version,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualWorkerNodeGroupConfigurations(t *testing.T) {
	var emptyTaints []corev1.Taint
	taint1 := corev1.Taint{Key: "key1"}
	taint2 := corev1.Taint{Key: "key2"}
	taints1 := []corev1.Taint{taint1, taint2}
	taints1DiffOrder := []corev1.Taint{taint2, taint1}
	taints2 := []corev1.Taint{taint1}

	testCases := []struct {
		testName                   string
		cluster1Wngs, cluster2Wngs []v1alpha1.WorkerNodeGroupConfiguration
		want                       bool
	}{
		{
			testName:     "both empty",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{},
			want:         true,
		},
		{
			testName: "one empty, one exists",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
				},
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k",
						Name: "n",
					},
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k",
						Name: "n",
					},
				},
			},
			want: true,
		},
		{
			testName: "both exist, order diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k1",
						Name: "n1",
					},
				},
				{
					Count: ptr.Int(2),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k2",
						Name: "n2",
					},
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(2),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k2",
						Name: "n2",
					},
				},
				{
					Count: ptr.Int(1),
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k1",
						Name: "n1",
					},
				},
			},
			want: true,
		},
		{
			testName: "both exist, count diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(2),
				},
			},
			want: false,
		},
		{
			testName: "both exist, autoscaling config diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					AutoScalingConfiguration: &v1alpha1.AutoScalingConfiguration{
						MinCount: 1,
						MaxCount: 3,
					},
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					AutoScalingConfiguration: nil,
				},
			},
			want: false,
		},
		{
			testName: "both exist, autoscaling config min diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					AutoScalingConfiguration: &v1alpha1.AutoScalingConfiguration{
						MinCount: 1,
						MaxCount: 3,
					},
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					AutoScalingConfiguration: &v1alpha1.AutoScalingConfiguration{
						MinCount: 2,
						MaxCount: 3,
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, autoscaling config max diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					AutoScalingConfiguration: &v1alpha1.AutoScalingConfiguration{
						MinCount: 1,
						MaxCount: 2,
					},
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					AutoScalingConfiguration: &v1alpha1.AutoScalingConfiguration{
						MinCount: 1,
						MaxCount: 3,
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, ref diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k1",
						Name: "n1",
					},
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					MachineGroupRef: &v1alpha1.Ref{
						Kind: "k2",
						Name: "n2",
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, same taints",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: taints1,
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: taints1,
				},
			},
			want: true,
		},
		{
			testName: "both exist, same taints in different order",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: taints1,
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: taints1DiffOrder,
				},
			},
			want: true,
		},
		{
			testName: "both exist, taints diff",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: taints1,
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: taints2,
				},
			},
			want: false,
		},
		{
			testName: "both exist, one with no taints",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: taints1,
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{},
			},
			want: false,
		},
		{
			testName: "both exist, one with empty taints",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: taints1,
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: emptyTaints,
				},
			},
			want: false,
		},
		{
			testName: "both exist, both with empty taints",
			cluster1Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: emptyTaints,
				},
			},
			cluster2Wngs: []v1alpha1.WorkerNodeGroupConfiguration{
				{
					Taints: emptyTaints,
				},
			},
			want: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					WorkerNodeGroupConfigurations: tt.cluster1Wngs,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					WorkerNodeGroupConfigurations: tt.cluster2Wngs,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualDatacenterRef(t *testing.T) {
	testCases := []struct {
		testName                                     string
		cluster1DatacenterRef, cluster2DatacenterRef v1alpha1.Ref
		want                                         bool
	}{
		{
			testName: "both empty",
			want:     true,
		},
		{
			testName: "one empty, one exists",
			cluster1DatacenterRef: v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			want: false,
		},
		{
			testName: "both exist, diff",
			cluster1DatacenterRef: v1alpha1.Ref{
				Kind: "k1",
				Name: "n1",
			},
			cluster2DatacenterRef: v1alpha1.Ref{
				Kind: "k2",
				Name: "n2",
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1DatacenterRef: v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			cluster2DatacenterRef: v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			want: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					DatacenterRef: tt.cluster1DatacenterRef,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					DatacenterRef: tt.cluster2DatacenterRef,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualIdentityProviderRefs(t *testing.T) {
	testCases := []struct {
		testName                 string
		cluster1Ipr, cluster2Ipr []v1alpha1.Ref
		want                     bool
	}{
		{
			testName:    "both empty",
			cluster1Ipr: []v1alpha1.Ref{},
			cluster2Ipr: []v1alpha1.Ref{},
			want:        true,
		},
		{
			testName: "one empty, one exists",
			cluster1Ipr: []v1alpha1.Ref{
				{
					Kind: "k",
					Name: "n",
				},
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1Ipr: []v1alpha1.Ref{
				{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2Ipr: []v1alpha1.Ref{
				{
					Kind: "k",
					Name: "n",
				},
			},
			want: true,
		},
		{
			testName: "both exist, order diff",
			cluster1Ipr: []v1alpha1.Ref{
				{
					Kind: "k1",
					Name: "n1",
				},
				{
					Kind: "k2",
					Name: "n2",
				},
			},
			cluster2Ipr: []v1alpha1.Ref{
				{
					Kind: "k2",
					Name: "n2",
				},
				{
					Kind: "k1",
					Name: "n1",
				},
			},
			want: true,
		},
		{
			testName: "both exist, count diff",
			cluster1Ipr: []v1alpha1.Ref{
				{
					Kind: "k1",
					Name: "n1",
				},
			},
			cluster2Ipr: []v1alpha1.Ref{
				{
					Kind: "k2",
					Name: "n2",
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					IdentityProviderRefs: tt.cluster1Ipr,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					IdentityProviderRefs: tt.cluster2Ipr,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualGitOpsRef(t *testing.T) {
	testCases := []struct {
		testName                             string
		cluster1GitOpsRef, cluster2GitOpsRef *v1alpha1.Ref
		want                                 bool
	}{
		{
			testName:          "both nil",
			cluster1GitOpsRef: nil,
			cluster2GitOpsRef: nil,
			want:              true,
		},
		{
			testName: "one nil, one exists",
			cluster1GitOpsRef: &v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			cluster2GitOpsRef: nil,
			want:              false,
		},
		{
			testName: "both exist, diff",
			cluster1GitOpsRef: &v1alpha1.Ref{
				Kind: "k1",
				Name: "n1",
			},
			cluster2GitOpsRef: &v1alpha1.Ref{
				Kind: "k2",
				Name: "n2",
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1GitOpsRef: &v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			cluster2GitOpsRef: &v1alpha1.Ref{
				Kind: "k",
				Name: "n",
			},
			want: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					GitOpsRef: tt.cluster1GitOpsRef,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					GitOpsRef: tt.cluster2GitOpsRef,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualClusterNetwork(t *testing.T) {
	testCases := []struct {
		testName                                       string
		cluster1ClusterNetwork, cluster2ClusterNetwork v1alpha1.ClusterNetwork
		want                                           bool
	}{
		{
			testName:               "both nil",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{},
			want:                   true,
		},
		{
			testName: "one empty, one exists",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, diff",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/6",
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, same",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{
				Pods: v1alpha1.Pods{
					CidrBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			want: true,
		},
		{
			testName: "same cni plugin (cilium), diff format",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				CNIConfig: &v1alpha1.CNIConfig{Cilium: &v1alpha1.CiliumConfig{}},
			},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{
				CNI: v1alpha1.Cilium,
			},
			want: true,
		},
		{
			testName: "different cni plugin (cilium), diff format",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				CNIConfig: &v1alpha1.CNIConfig{Kindnetd: &v1alpha1.KindnetdConfig{}},
			},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{
				CNIConfig: &v1alpha1.CNIConfig{Cilium: &v1alpha1.CiliumConfig{}},
			},
			want: false,
		},
		{
			testName: "same cni plugin (cilium), diff cilium configuration",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				CNIConfig: &v1alpha1.CNIConfig{Cilium: &v1alpha1.CiliumConfig{PolicyEnforcementMode: "always"}},
			},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{
				CNIConfig: &v1alpha1.CNIConfig{Cilium: &v1alpha1.CiliumConfig{PolicyEnforcementMode: "default"}},
			},
			want: false,
		},
		{
			testName: "diff Nodes",
			cluster1ClusterNetwork: v1alpha1.ClusterNetwork{
				Nodes: &v1alpha1.Nodes{},
			},
			cluster2ClusterNetwork: v1alpha1.ClusterNetwork{
				Nodes: &v1alpha1.Nodes{
					CIDRMaskSize: ptr.Int(3),
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ClusterNetwork: tt.cluster1ClusterNetwork,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ClusterNetwork: tt.cluster2ClusterNetwork,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualExternalEtcdConfiguration(t *testing.T) {
	testCases := []struct {
		testName                   string
		cluster1Etcd, cluster2Etcd *v1alpha1.ExternalEtcdConfiguration
		want                       bool
	}{
		{
			testName:     "both nil",
			cluster1Etcd: nil,
			cluster2Etcd: nil,
			want:         true,
		},
		{
			testName: "one nil, one exists",
			cluster1Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2Etcd: nil,
			want:         false,
		},
		{
			testName: "both exist, same",
			cluster1Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			want: true,
		},
		{
			testName: "both exist, count diff",
			cluster1Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 2,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			want: false,
		},
		{
			testName: "both exist, ref diff",
			cluster1Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k1",
					Name: "n1",
				},
			},
			cluster2Etcd: &v1alpha1.ExternalEtcdConfiguration{
				Count: 1,
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k2",
					Name: "n2",
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ExternalEtcdConfiguration: tt.cluster1Etcd,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ExternalEtcdConfiguration: tt.cluster2Etcd,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualProxyConfiguration(t *testing.T) {
	testCases := []struct {
		testName                     string
		cluster1Proxy, cluster2Proxy *v1alpha1.ProxyConfiguration
		want                         bool
	}{
		{
			testName:      "both nil",
			cluster1Proxy: nil,
			cluster2Proxy: nil,
			want:          true,
		},
		{
			testName: "one nil, one exists",
			cluster1Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.4",
			},
			cluster2Proxy: nil,
			want:          false,
		},
		{
			testName: "both exist, same",
			cluster1Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.4",
			},
			cluster2Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.4",
			},
			want: true,
		},
		{
			testName: "both exist, diff",
			cluster1Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.4",
			},
			cluster2Proxy: &v1alpha1.ProxyConfiguration{
				HttpProxy: "1.2.3.5",
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ProxyConfiguration: tt.cluster1Proxy,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ProxyConfiguration: tt.cluster2Proxy,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualRegistryMirrorConfiguration(t *testing.T) {
	testCases := []struct {
		testName                   string
		cluster1Regi, cluster2Regi *v1alpha1.RegistryMirrorConfiguration
		want                       bool
	}{
		{
			testName:     "both nil",
			cluster1Regi: nil,
			cluster2Regi: nil,
			want:         true,
		},
		{
			testName: "one nil, one exists",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			cluster2Regi: nil,
			want:         false,
		},
		{
			testName: "both exist, same",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			want: true,
		},
		{
			testName: "both exist, endpoint diff",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.5",
				CACertContent: "ca",
			},
			want: false,
		},
		{
			testName: "both exist, namespaces diff (one nil, one exists)",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{},
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, namespaces diff (registry)",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "1.2.3.4",
						Namespace: "eks-anywhere",
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, namespaces diff (namespace)",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					RegistryMirrorConfiguration: tt.cluster1Regi,
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					RegistryMirrorConfiguration: tt.cluster2Regi,
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualPackageConfigurationNetwork(t *testing.T) {
	testCases := []struct {
		testName           string
		disable1, disable2 bool
		want               bool
	}{
		{
			testName: "equal",
			disable1: true,
			disable2: true,
			want:     true,
		},
		{
			testName: "not equal",
			disable1: true,
			disable2: false,
			want:     false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					Packages: &v1alpha1.PackageConfiguration{
						Disable: tt.disable1,
					},
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					Packages: &v1alpha1.PackageConfiguration{
						Disable: tt.disable2,
					},
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualManagement(t *testing.T) {
	testCases := []struct {
		testName                               string
		cluster1Management, cluster2Management string
		want                                   bool
	}{
		{
			testName:           "both empty",
			cluster1Management: "",
			cluster2Management: "",
			want:               true,
		},
		{
			testName:           "one empty, one equal to self",
			cluster1Management: "",
			cluster2Management: "cluster-1",
			want:               true,
		},
		{
			testName:           "both equal to self",
			cluster1Management: "cluster-1",
			cluster2Management: "cluster-1",
			want:               true,
		},
		{
			testName:           "one empty, one not equal to self",
			cluster1Management: "",
			cluster2Management: "cluster-2",
			want:               false,
		},
		{
			testName:           "one equal to self, one not equal to self",
			cluster1Management: "cluster-1",
			cluster2Management: "cluster-2",
			want:               false,
		},
		{
			testName:           "both not equal to self and different",
			cluster1Management: "cluster-2",
			cluster2Management: "cluster-3",
			want:               false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-1",
				},
				Spec: v1alpha1.ClusterSpec{
					ManagementCluster: v1alpha1.ManagementCluster{
						Name: tt.cluster1Management,
					},
				},
			}
			cluster2 := &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-1",
				},
				Spec: v1alpha1.ClusterSpec{
					ManagementCluster: v1alpha1.ManagementCluster{
						Name: tt.cluster2Management,
					},
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestClusterEqualDifferentBundlesRef(t *testing.T) {
	cluster1 := &v1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-1",
		},
		Spec: v1alpha1.ClusterSpec{
			BundlesRef: &v1alpha1.BundlesRef{
				APIVersion: "v1",
				Name:       "bundles-1",
				Namespace:  "eksa-system",
			},
		},
	}

	cluster2 := cluster1.DeepCopy()
	cluster2.Spec.BundlesRef.Name = "bundles-2"

	g := NewWithT(t)
	g.Expect(cluster1.Equal(cluster2)).To(BeFalse())
}

func TestControlPlaneConfigurationEqual(t *testing.T) {
	var emptyTaints []corev1.Taint
	taint1 := corev1.Taint{Key: "key1"}
	taint2 := corev1.Taint{Key: "key2"}
	taints1 := []corev1.Taint{taint1, taint2}
	taints1DiffOrder := []corev1.Taint{taint2, taint1}
	taints2 := []corev1.Taint{taint1}

	testCases := []struct {
		testName                           string
		cluster1CPConfig, cluster2CPConfig *v1alpha1.ControlPlaneConfiguration
		want                               bool
	}{
		{
			testName: "both exist, same",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			want: true,
		},
		{
			testName: "one nil, one exists",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2CPConfig: nil,
			want:             false,
		},
		{
			testName: "one nil, one exists",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
				MachineGroupRef: &v1alpha1.Ref{
					Kind: "k",
					Name: "n",
				},
			},
			cluster2CPConfig: nil,
			want:             false,
		},
		{
			testName: "count exists, diff",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 2,
			},
			want: false,
		},
		{
			testName: "one count empty",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Count: 1,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{},
			want:             false,
		},
		{
			testName: "both taints equal",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: taints1,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: taints1,
			},
			want: true,
		},
		{
			testName: "taints in different orders",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: taints1,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: taints1DiffOrder,
			},
			want: true,
		},
		{
			testName: "different taints",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: taints1,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: taints2,
			},
			want: false,
		},
		{
			testName: "One taints set empty",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: taints1,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: emptyTaints,
			},
			want: false,
		},
		{
			testName: "one taints set not present",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: taints1,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{},
			want:             false,
		},
		{
			testName: "both taints set empty",
			cluster1CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: emptyTaints,
			},
			cluster2CPConfig: &v1alpha1.ControlPlaneConfiguration{
				Taints: emptyTaints,
			},
			want: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster1CPConfig.Equal(tt.cluster2CPConfig)).To(Equal(tt.want))
		})
	}
}

func TestControlPlaneConfigurationEndpointEqual(t *testing.T) {
	testCases := []struct {
		testName, cluster1CPHost, cluster2CPHost, clusterDatacenterKind string
		want                                                            bool
	}{
		{
			testName:              "one default port, one no port",
			cluster1CPHost:        "1.2.3.4",
			clusterDatacenterKind: v1alpha1.VSphereDatacenterKind,
			cluster2CPHost:        "1.2.3.5",
			want:                  false,
		},
		{
			testName:              "one default port, one no port",
			cluster1CPHost:        "1.2.3.4",
			clusterDatacenterKind: v1alpha1.VSphereDatacenterKind,
			cluster2CPHost:        "",
			want:                  false,
		},
		{
			testName:              "one default port, one no port",
			cluster1CPHost:        "",
			clusterDatacenterKind: v1alpha1.VSphereDatacenterKind,
			cluster2CPHost:        "",
			want:                  true,
		},
		{
			testName:              "one default port, one no port",
			cluster1CPHost:        "1.1.1.1:6443",
			clusterDatacenterKind: v1alpha1.CloudStackDatacenterKind,
			cluster2CPHost:        "1.1.1.1",
			want:                  true,
		},
		{
			testName:              "one default port, one no port",
			cluster1CPHost:        "1.1.1.1",
			clusterDatacenterKind: v1alpha1.CloudStackDatacenterKind,
			cluster2CPHost:        "1.1.1.1:6443",
			want:                  true,
		},
		{
			testName:              "one default port, one no port",
			cluster1CPHost:        "1.1.1.1",
			clusterDatacenterKind: v1alpha1.CloudStackDatacenterKind,
			cluster2CPHost:        "1.1.1.2",
			want:                  false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			cluster1 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					DatacenterRef: v1alpha1.Ref{
						Kind: tt.clusterDatacenterKind,
					},
					ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
						Endpoint: &v1alpha1.Endpoint{
							Host: tt.cluster1CPHost,
						},
					},
				},
			}
			cluster2 := &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					DatacenterRef: v1alpha1.Ref{
						Kind: tt.clusterDatacenterKind,
					},
					ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
						Endpoint: &v1alpha1.Endpoint{
							Host: tt.cluster2CPHost,
						},
					},
				},
			}

			g := NewWithT(t)
			g.Expect(cluster1.Equal(cluster2)).To(Equal(tt.want))
		})
	}
}

func TestRegistryMirrorConfigurationEqual(t *testing.T) {
	testCases := []struct {
		testName                   string
		cluster1Regi, cluster2Regi *v1alpha1.RegistryMirrorConfiguration
		want                       bool
	}{
		{
			testName:     "both nil",
			cluster1Regi: nil,
			cluster2Regi: nil,
			want:         true,
		},
		{
			testName: "one nil, one exists",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
			},
			cluster2Regi: nil,
			want:         false,
		},
		{
			testName: "both exist, same",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "ca",
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			want: true,
		},
		{
			testName: "both exist, endpoint diff",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				CACertContent: "",
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint: "1.2.3.5",
			},
			want: false,
		},
		{
			testName: "both exist, ca diff",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				CACertContent: "ca1",
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				CACertContent: "ca2",
			},
			want: false,
		},
		{
			testName: "both exist, namespaces diff (one nil, one exists)",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{},
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, namespaces diff (registry)",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "",
						Namespace: "eks-anywhere",
					},
				},
			},
			want: false,
		},
		{
			testName: "both exist, namespaces diff (namespace)",
			cluster1Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "eks-anywhere",
					},
				},
			},
			cluster2Regi: &v1alpha1.RegistryMirrorConfiguration{
				OCINamespaces: []v1alpha1.OCINamespace{
					{
						Registry:  "public.ecr.aws",
						Namespace: "",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster1Regi.Equal(tt.cluster2Regi)).To(Equal(tt.want))
		})
	}
}

func TestPodIAMServiceAccountIssuerHasNotChanged(t *testing.T) {
	testCases := []struct {
		testName                                   string
		cluster1PodIAMConfig, cluster2PodIAMConfig *v1alpha1.PodIAMConfig
		want                                       bool
	}{
		{
			testName:             "both nil",
			cluster1PodIAMConfig: nil,
			cluster2PodIAMConfig: nil,
			want:                 true,
		},
		{
			testName: "one nil, one exists",
			cluster1PodIAMConfig: &v1alpha1.PodIAMConfig{
				ServiceAccountIssuer: "https://test",
			},
			cluster2PodIAMConfig: nil,
			want:                 false,
		},
		{
			testName: "both exist, same",
			cluster1PodIAMConfig: &v1alpha1.PodIAMConfig{
				ServiceAccountIssuer: "https://test",
			},
			cluster2PodIAMConfig: &v1alpha1.PodIAMConfig{
				ServiceAccountIssuer: "https://test",
			},
			want: true,
		},
		{
			testName: "both exist, service account issuer different",
			cluster1PodIAMConfig: &v1alpha1.PodIAMConfig{
				ServiceAccountIssuer: "https://test1",
			},
			cluster2PodIAMConfig: &v1alpha1.PodIAMConfig{
				ServiceAccountIssuer: "https://test2",
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster1PodIAMConfig.Equal(tt.cluster2PodIAMConfig)).To(Equal(tt.want))
		})
	}
}

func TestBundlesRefEqual(t *testing.T) {
	testCases := []struct {
		testName                 string
		bundlesRef1, bundlesRef2 *v1alpha1.BundlesRef
		want                     bool
	}{
		{
			testName:    "both nil",
			bundlesRef1: nil,
			bundlesRef2: nil,
			want:        true,
		},
		{
			testName:    "1 nil, 2 not nil",
			bundlesRef1: nil,
			bundlesRef2: &v1alpha1.BundlesRef{},
			want:        false,
		},
		{
			testName:    "1 not nil, 2 nil",
			bundlesRef1: &v1alpha1.BundlesRef{},
			bundlesRef2: nil,
			want:        false,
		},
		{
			testName: "diff APIVersion",
			bundlesRef1: &v1alpha1.BundlesRef{
				APIVersion: "v1",
				Name:       "bundles-1",
				Namespace:  "eksa-system",
			},
			bundlesRef2: &v1alpha1.BundlesRef{
				APIVersion: "v2",
				Name:       "bundles-1",
				Namespace:  "eksa-system",
			},
			want: false,
		},
		{
			testName: "diff Name",
			bundlesRef1: &v1alpha1.BundlesRef{
				APIVersion: "v1",
				Name:       "bundles-1",
				Namespace:  "eksa-system",
			},
			bundlesRef2: &v1alpha1.BundlesRef{
				APIVersion: "v1",
				Name:       "bundles-2",
				Namespace:  "eksa-system",
			},
			want: false,
		},
		{
			testName: "diff Namespace",
			bundlesRef1: &v1alpha1.BundlesRef{
				APIVersion: "v1",
				Name:       "bundles-1",
				Namespace:  "eksa-system",
			},
			bundlesRef2: &v1alpha1.BundlesRef{
				APIVersion: "v1",
				Name:       "bundles-1",
				Namespace:  "default",
			},
			want: false,
		},
		{
			testName: "everything different",
			bundlesRef1: &v1alpha1.BundlesRef{
				APIVersion: "v1",
				Name:       "bundles-1",
				Namespace:  "eksa-system",
			},
			bundlesRef2: &v1alpha1.BundlesRef{
				APIVersion: "v2",
				Name:       "bundles-2",
				Namespace:  "default",
			},
			want: false,
		},
		{
			testName: "equal",
			bundlesRef1: &v1alpha1.BundlesRef{
				APIVersion: "v1",
				Name:       "bundles-1",
				Namespace:  "eksa-system",
			},
			bundlesRef2: &v1alpha1.BundlesRef{
				APIVersion: "v1",
				Name:       "bundles-1",
				Namespace:  "eksa-system",
			},
			want: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.bundlesRef1.Equal(tt.bundlesRef2)).To(Equal(tt.want))
		})
	}
}

func setSelfManaged(c *v1alpha1.Cluster, s bool) {
	if s {
		c.SetSelfManaged()
	} else {
		c.SetManagedBy("management-cluster")
	}
}

func TestNodes_Equal(t *testing.T) {
	tests := []struct {
		name           string
		nodes1, nodes2 *v1alpha1.Nodes
		want           bool
	}{
		{
			name:   "one nil",
			nodes1: nil,
			nodes2: &v1alpha1.Nodes{},
			want:   false,
		},
		{
			name:   "other nil",
			nodes1: &v1alpha1.Nodes{},
			nodes2: nil,
			want:   false,
		},
		{
			name:   "both nil",
			nodes1: nil,
			nodes2: nil,
			want:   true,
		},
		{
			name:   "one nil CIDRMasK",
			nodes1: &v1alpha1.Nodes{},
			nodes2: &v1alpha1.Nodes{
				CIDRMaskSize: ptr.Int(2),
			},
			want: false,
		},
		{
			name:   "both nil CIDRMasK",
			nodes1: &v1alpha1.Nodes{},
			nodes2: &v1alpha1.Nodes{},
			want:   true,
		},
		{
			name: "different not nil CIDRMasK",
			nodes1: &v1alpha1.Nodes{
				CIDRMaskSize: ptr.Int(3),
			},
			nodes2: &v1alpha1.Nodes{
				CIDRMaskSize: ptr.Int(2),
			},
			want: false,
		},
		{
			name: "equal not nil CIDRMasK",
			nodes1: &v1alpha1.Nodes{
				CIDRMaskSize: ptr.Int(2),
			},
			nodes2: &v1alpha1.Nodes{
				CIDRMaskSize: ptr.Int(2),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.nodes1.Equal(tt.nodes2)).To(Equal(tt.want))
		})
	}
}

func TestClusterHasAWSIamConfig(t *testing.T) {
	tests := []struct {
		name    string
		cluster *v1alpha1.Cluster
		want    bool
	}{
		{
			name: "has AWSIamConfig",
			cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-cluster",
					Namespace: "eksa-system",
				},
				Spec: v1alpha1.ClusterSpec{
					IdentityProviderRefs: []v1alpha1.Ref{
						{
							Name: "aws-config",
							Kind: "AWSIamConfig",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "no AWSIamConfig",
			cluster: &v1alpha1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-cluster",
					Namespace: "eksa-system",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster.HasAWSIamConfig()).To(Equal(tt.want))
		})
	}
}

func TestPackageConfiguration_Equal(t *testing.T) {
	same := &v1alpha1.PackageConfiguration{Disable: false}
	tests := []struct {
		name     string
		pcn, pco *v1alpha1.PackageConfiguration
		want     bool
	}{
		{
			name: "one nil",
			pcn:  &v1alpha1.PackageConfiguration{},
			pco:  nil,
			want: false,
		},
		{
			name: "other nil",
			pcn:  nil,
			pco:  &v1alpha1.PackageConfiguration{},
			want: false,
		},
		{
			name: "both nil",
			pcn:  nil,
			pco:  nil,
			want: true,
		},
		{
			name: "equal",
			pcn:  &v1alpha1.PackageConfiguration{Disable: true},
			pco:  &v1alpha1.PackageConfiguration{Disable: true},
			want: true,
		},
		{
			name: "not equal",
			pcn:  &v1alpha1.PackageConfiguration{Disable: true},
			pco:  &v1alpha1.PackageConfiguration{Disable: false},
			want: false,
		},
		{
			name: "not equal controller",
			pcn: &v1alpha1.PackageConfiguration{
				Disable: true,
				Controller: &v1alpha1.PackageControllerConfiguration{
					Tag: "v1",
				},
			},
			pco: &v1alpha1.PackageConfiguration{
				Disable: false,
				Controller: &v1alpha1.PackageControllerConfiguration{
					Tag: "v2",
				},
			},
			want: false,
		},
		{
			name: "equal controller",
			pcn: &v1alpha1.PackageConfiguration{
				Controller: &v1alpha1.PackageControllerConfiguration{
					Tag: "v1",
				},
			},
			pco: &v1alpha1.PackageConfiguration{
				Controller: &v1alpha1.PackageControllerConfiguration{
					Tag: "v1",
				},
			},
			want: true,
		},
		{
			name: "not equal cronjob",
			pcn: &v1alpha1.PackageConfiguration{
				Disable: true,
				CronJob: &v1alpha1.PackageControllerCronJob{
					Tag: "v1",
				},
			},
			pco: &v1alpha1.PackageConfiguration{
				Disable: false,
				CronJob: &v1alpha1.PackageControllerCronJob{
					Tag: "v2",
				},
			},
			want: false,
		},
		{
			name: "equal cronjob",
			pcn: &v1alpha1.PackageConfiguration{
				CronJob: &v1alpha1.PackageControllerCronJob{
					Tag: "v1",
				},
			},
			pco: &v1alpha1.PackageConfiguration{
				CronJob: &v1alpha1.PackageControllerCronJob{
					Tag: "v1",
				},
			},
			want: true,
		},
		{
			name: "same",
			pcn:  same,
			pco:  same,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.pcn.Equal(tt.pco)).To(Equal(tt.want))
		})
	}
}

func TestPackageControllerConfiguration_Equal(t *testing.T) {
	same := &v1alpha1.PackageControllerConfiguration{Tag: "v1"}
	tests := []struct {
		name     string
		pcn, pco *v1alpha1.PackageControllerConfiguration
		want     bool
	}{
		{
			name: "one nil",
			pcn:  &v1alpha1.PackageControllerConfiguration{},
			pco:  nil,
			want: false,
		},
		{
			name: "other nil",
			pcn:  nil,
			pco:  &v1alpha1.PackageControllerConfiguration{},
			want: false,
		},
		{
			name: "both nil",
			pcn:  nil,
			pco:  nil,
			want: true,
		},
		{
			name: "equal Repository",
			pcn:  &v1alpha1.PackageControllerConfiguration{Repository: "a"},
			pco:  &v1alpha1.PackageControllerConfiguration{Repository: "a"},
			want: true,
		},
		{
			name: "not equal Repository",
			pcn:  &v1alpha1.PackageControllerConfiguration{Repository: "a"},
			pco:  &v1alpha1.PackageControllerConfiguration{Repository: "b"},
			want: false,
		},
		{
			name: "equal Tag",
			pcn:  &v1alpha1.PackageControllerConfiguration{Tag: "v1"},
			pco:  &v1alpha1.PackageControllerConfiguration{Tag: "v1"},
			want: true,
		},
		{
			name: "not equal Tag",
			pcn:  &v1alpha1.PackageControllerConfiguration{Tag: "v1"},
			pco:  &v1alpha1.PackageControllerConfiguration{Tag: "v2"},
			want: false,
		},
		{
			name: "equal Digest",
			pcn:  &v1alpha1.PackageControllerConfiguration{Digest: "a"},
			pco:  &v1alpha1.PackageControllerConfiguration{Digest: "a"},
			want: true,
		},
		{
			name: "not equal Digest",
			pcn:  &v1alpha1.PackageControllerConfiguration{Digest: "a"},
			pco:  &v1alpha1.PackageControllerConfiguration{Digest: "b"},
			want: false,
		},
		{
			name: "equal DisableWebhooks",
			pcn:  &v1alpha1.PackageControllerConfiguration{DisableWebhooks: true},
			pco:  &v1alpha1.PackageControllerConfiguration{DisableWebhooks: true},
			want: true,
		},
		{
			name: "not equal DisableWebhooks",
			pcn:  &v1alpha1.PackageControllerConfiguration{DisableWebhooks: true},
			pco:  &v1alpha1.PackageControllerConfiguration{},
			want: false,
		},
		{
			name: "equal Env",
			pcn:  &v1alpha1.PackageControllerConfiguration{Env: []string{"a"}},
			pco:  &v1alpha1.PackageControllerConfiguration{Env: []string{"a"}},
			want: true,
		},
		{
			name: "not equal Env",
			pcn:  &v1alpha1.PackageControllerConfiguration{Env: []string{"a"}},
			pco:  &v1alpha1.PackageControllerConfiguration{Env: []string{"b"}},
			want: false,
		},
		{
			name: "not equal Resources",
			pcn: &v1alpha1.PackageControllerConfiguration{
				Resources: v1alpha1.PackageControllerResources{
					Requests: v1alpha1.ImageResource{
						CPU: "1",
					},
				},
			},
			pco: &v1alpha1.PackageControllerConfiguration{
				Resources: v1alpha1.PackageControllerResources{
					Requests: v1alpha1.ImageResource{
						CPU: "2",
					},
				},
			},
			want: false,
		},
		{
			name: "same",
			pcn:  same,
			pco:  same,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.pcn.Equal(tt.pco)).To(Equal(tt.want))
		})
	}
}

func TestPackageControllerResources_Equal(t *testing.T) {
	same := &v1alpha1.PackageControllerResources{
		Limits: v1alpha1.ImageResource{
			CPU: "3",
		},
	}
	tests := []struct {
		name     string
		pcn, pco *v1alpha1.PackageControllerResources
		want     bool
	}{
		{
			name: "one nil",
			pcn:  &v1alpha1.PackageControllerResources{},
			pco:  nil,
			want: false,
		},
		{
			name: "other nil",
			pcn:  nil,
			pco:  &v1alpha1.PackageControllerResources{},
			want: false,
		},
		{
			name: "both nil",
			pcn:  nil,
			pco:  nil,
			want: true,
		},
		{
			name: "equal Requests",
			pcn: &v1alpha1.PackageControllerResources{
				Requests: v1alpha1.ImageResource{
					CPU: "1",
				},
			},
			pco: &v1alpha1.PackageControllerResources{
				Requests: v1alpha1.ImageResource{
					CPU: "1",
				},
			},
			want: true,
		},
		{
			name: "not equal Requests",
			pcn: &v1alpha1.PackageControllerResources{
				Requests: v1alpha1.ImageResource{
					CPU: "1",
				},
			},
			pco: &v1alpha1.PackageControllerResources{
				Requests: v1alpha1.ImageResource{
					CPU: "2",
				},
			},
			want: false,
		},
		{
			name: "equal Limits",
			pcn: &v1alpha1.PackageControllerResources{
				Limits: v1alpha1.ImageResource{
					CPU: "1",
				},
			},
			pco: &v1alpha1.PackageControllerResources{
				Limits: v1alpha1.ImageResource{
					CPU: "1",
				},
			},
			want: true,
		},
		{
			name: "not equal Limits",
			pcn: &v1alpha1.PackageControllerResources{
				Limits: v1alpha1.ImageResource{
					CPU: "1",
				},
			},
			pco: &v1alpha1.PackageControllerResources{
				Limits: v1alpha1.ImageResource{
					CPU: "2",
				},
			},
			want: false,
		},
		{
			name: "same",
			pcn:  same,
			pco:  same,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.pcn.Equal(tt.pco)).To(Equal(tt.want))
		})
	}
}

func TestImageResource_Equal(t *testing.T) {
	same := &v1alpha1.ImageResource{
		CPU: "3",
	}
	tests := []struct {
		name     string
		pcn, pco *v1alpha1.ImageResource
		want     bool
	}{
		{
			name: "one nil",
			pcn:  &v1alpha1.ImageResource{},
			pco:  nil,
			want: false,
		},
		{
			name: "other nil",
			pcn:  nil,
			pco:  &v1alpha1.ImageResource{},
			want: false,
		},
		{
			name: "both nil",
			pcn:  nil,
			pco:  nil,
			want: true,
		},
		{
			name: "equal CPU",
			pcn: &v1alpha1.ImageResource{
				CPU: "1",
			},
			pco: &v1alpha1.ImageResource{
				CPU: "1",
			},
			want: true,
		},
		{
			name: "not equal CPU",
			pcn: &v1alpha1.ImageResource{
				CPU: "1",
			},
			pco: &v1alpha1.ImageResource{
				CPU: "2",
			},
			want: false,
		},
		{
			name: "equal Memory",
			pcn: &v1alpha1.ImageResource{
				Memory: "1",
			},
			pco: &v1alpha1.ImageResource{
				Memory: "1",
			},
			want: true,
		},
		{
			name: "not equal Memory",
			pcn: &v1alpha1.ImageResource{
				Memory: "1",
			},
			pco: &v1alpha1.ImageResource{
				Memory: "2",
			},
			want: false,
		},
		{
			name: "same",
			pcn:  same,
			pco:  same,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.pcn.Equal(tt.pco)).To(Equal(tt.want))
		})
	}
}

func TestPackageControllerCronJob_Equal(t *testing.T) {
	same := &v1alpha1.PackageControllerCronJob{
		Repository: "3",
	}
	tests := []struct {
		name     string
		pcn, pco *v1alpha1.PackageControllerCronJob
		want     bool
	}{
		{
			name: "one nil",
			pcn:  &v1alpha1.PackageControllerCronJob{},
			pco:  nil,
			want: false,
		},
		{
			name: "other nil",
			pcn:  nil,
			pco:  &v1alpha1.PackageControllerCronJob{},
			want: false,
		},
		{
			name: "both nil",
			pcn:  nil,
			pco:  nil,
			want: true,
		},
		{
			name: "equal Repository",
			pcn: &v1alpha1.PackageControllerCronJob{
				Repository: "1",
			},
			pco: &v1alpha1.PackageControllerCronJob{
				Repository: "1",
			},
			want: true,
		},
		{
			name: "not equal Repository",
			pcn: &v1alpha1.PackageControllerCronJob{
				Repository: "1",
			},
			pco: &v1alpha1.PackageControllerCronJob{
				Repository: "2",
			},
			want: false,
		},
		{
			name: "equal Tag",
			pcn: &v1alpha1.PackageControllerCronJob{
				Tag: "1",
			},
			pco: &v1alpha1.PackageControllerCronJob{
				Tag: "1",
			},
			want: true,
		},
		{
			name: "not equal Tag",
			pcn: &v1alpha1.PackageControllerCronJob{
				Tag: "1",
			},
			pco: &v1alpha1.PackageControllerCronJob{
				Tag: "2",
			},
			want: false,
		},
		{
			name: "equal Digest",
			pcn: &v1alpha1.PackageControllerCronJob{
				Digest: "1",
			},
			pco: &v1alpha1.PackageControllerCronJob{
				Digest: "1",
			},
			want: true,
		},
		{
			name: "not equal Digest",
			pcn: &v1alpha1.PackageControllerCronJob{
				Digest: "1",
			},
			pco: &v1alpha1.PackageControllerCronJob{
				Digest: "2",
			},
			want: false,
		},
		{
			name: "equal Disable",
			pcn: &v1alpha1.PackageControllerCronJob{
				Disable: true,
			},
			pco: &v1alpha1.PackageControllerCronJob{
				Disable: true,
			},
			want: true,
		},
		{
			name: "not equal Disable",
			pcn: &v1alpha1.PackageControllerCronJob{
				Disable: true,
			},
			pco: &v1alpha1.PackageControllerCronJob{
				Disable: false,
			},
			want: false,
		},
		{
			name: "same",
			pcn:  same,
			pco:  same,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.pcn.Equal(tt.pco)).To(Equal(tt.want))
		})
	}
}

func TestCiliumConfigEquality(t *testing.T) {
	tests := []struct {
		Name  string
		A     *v1alpha1.CiliumConfig
		B     *v1alpha1.CiliumConfig
		Equal bool
	}{
		{
			Name:  "Nils",
			A:     nil,
			B:     nil,
			Equal: true,
		},
		{
			Name:  "NilA",
			A:     nil,
			B:     &v1alpha1.CiliumConfig{},
			Equal: false,
		},
		{
			Name:  "NilB",
			A:     &v1alpha1.CiliumConfig{},
			B:     nil,
			Equal: false,
		},
		{
			Name:  "ZeroValues",
			A:     &v1alpha1.CiliumConfig{},
			B:     &v1alpha1.CiliumConfig{},
			Equal: true,
		},
		{
			Name: "EqualPolicyEnforcement",
			A: &v1alpha1.CiliumConfig{
				PolicyEnforcementMode: "always",
			},
			B: &v1alpha1.CiliumConfig{
				PolicyEnforcementMode: "always",
			},
			Equal: true,
		},
		{
			Name: "DiffPolicyEnforcement",
			A: &v1alpha1.CiliumConfig{
				PolicyEnforcementMode: "always",
			},
			B: &v1alpha1.CiliumConfig{
				PolicyEnforcementMode: "default",
			},
			Equal: false,
		},
		{
			Name: "NilSkipUpgradeAFalse",
			A: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(false),
			},
			B:     &v1alpha1.CiliumConfig{},
			Equal: true,
		},
		{
			Name: "NilSkipUpgradeBFalse",
			A:    &v1alpha1.CiliumConfig{},
			B: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(false),
			},
			Equal: true,
		},
		{
			Name: "SkipUpgradeBothFalse",
			A: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(false),
			},
			B: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(false),
			},
			Equal: true,
		},
		{
			Name: "NilSkipUpgradeATrue",
			A: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
			B:     &v1alpha1.CiliumConfig{},
			Equal: false,
		},
		{
			Name: "NilSkipUpgradeBTrue",
			A:    &v1alpha1.CiliumConfig{},
			B: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
			Equal: false,
		},
		{
			Name: "SkipUpgradeBothTrue",
			A: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
			B: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
			Equal: true,
		},
		{
			Name: "SkipUpgradeAFalseBTrue",
			A: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(false),
			},
			B: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
			Equal: false,
		},
		{
			Name: "SkipUpgradeATrueBFalse",
			A: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(true),
			},
			B: &v1alpha1.CiliumConfig{
				SkipUpgrade: ptr.Bool(false),
			},
			Equal: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tc.A.Equal(tc.B)).To(Equal(tc.Equal))
		})
	}
}

func TestKubeVersionToValidSemver(t *testing.T) {
	type args struct {
		kubeVersion v1alpha1.KubernetesVersion
	}
	tests := []struct {
		name    string
		args    args
		want    *semver.Version
		wantErr error
	}{
		{
			name: "convert kube 1.22",
			args: args{
				kubeVersion: v1alpha1.Kube122,
			},
			want: &semver.Version{
				Major: 1,
				Minor: 22,
				Patch: 0,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v1alpha1.KubeVersionToSemver(tt.args.kubeVersion)
			if err != tt.wantErr {
				t.Errorf("KubeVersionToSemver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KubeVersionToSemver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClusterIsSingleNode(t *testing.T) {
	testCases := []struct {
		testName string
		cluster  *v1alpha1.Cluster
		want     bool
	}{
		{
			testName: "cluster with single node",
			cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
						Count: 1,
					},
				},
			},
			want: true,
		},
		{
			testName: "cluster with cp and worker",
			cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
						Count: 1,
					},
					WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{
						{
							Count: ptr.Int(3),
						},
					},
				},
			},
			want: false,
		},
		{
			testName: "cluster with multiple cp",
			cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
						Count: 3,
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.cluster.IsSingleNode()).To(Equal(tt.want))
		})
	}
}
