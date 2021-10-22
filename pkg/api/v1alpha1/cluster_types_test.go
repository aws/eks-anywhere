package v1alpha1_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
