package v1alpha1_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func boolPointer(b bool) *bool {
	return &b
}

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
					Count: 3,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: v1alpha1.VSphereMachineConfigKind,
						Name: "eksa-unit-test-1",
					},
				},
				{
					Count: 3,
					MachineGroupRef: &v1alpha1.Ref{
						Kind: v1alpha1.VSphereMachineConfigKind,
						Name: "eksa-unit-test-2",
					},
				},
				{
					Count: 5,
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
			testName: "nil flag",
			cluster:  &v1alpha1.Cluster{},
			want:     true,
		},
		{
			testName: "true flag",
			cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					Management: boolPointer(true),
				},
			},
			want: true,
		},
		{
			testName: "false flag",
			cluster: &v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					Management: boolPointer(false),
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

func setSelfManaged(c *v1alpha1.Cluster, s bool) {
	if s {
		c.SetSelfManaged()
	} else {
		c.SetManagedBy("management-cluster")
	}
}
