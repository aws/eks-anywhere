package v1alpha1_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestSnowMachineConfigIPPoolRefs(t *testing.T) {
	m := &v1alpha1.SnowMachineConfig{
		Spec: v1alpha1.SnowMachineConfigSpec{
			Network: v1alpha1.SnowNetwork{
				DirectNetworkInterfaces: []v1alpha1.SnowDirectNetworkInterface{
					{
						IPPoolRef: &v1alpha1.Ref{
							Kind: v1alpha1.SnowIPPoolKind,
							Name: "ip-pool-1",
						},
					},
					{
						IPPoolRef: &v1alpha1.Ref{
							Kind: v1alpha1.SnowIPPoolKind,
							Name: "ip-pool-2",
						},
					},
					{
						IPPoolRef: &v1alpha1.Ref{
							Kind: v1alpha1.SnowIPPoolKind,
							Name: "ip-pool-1", // test duplicates
						},
					},
				},
			},
		},
	}

	want := []v1alpha1.Ref{
		{
			Kind: v1alpha1.SnowIPPoolKind,
			Name: "ip-pool-1",
		},
		{
			Kind: v1alpha1.SnowIPPoolKind,
			Name: "ip-pool-2",
		},
	}

	got := m.IPPoolRefs()

	if !v1alpha1.RefSliceEqual(got, want) {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}
