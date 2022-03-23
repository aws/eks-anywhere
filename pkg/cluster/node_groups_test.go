package cluster_test

import (
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestNodeGroupsToDelete(t *testing.T) {
	tests := []struct {
		name         string
		new, current *cluster.Spec
		want         []anywherev1.WorkerNodeGroupConfiguration
	}{
		{
			name: "one worker node group, missing name, no changes",
			current: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						MachineGroupRef: &anywherev1.Ref{
							Kind: anywherev1.VSphereMachineConfigKind,
							Name: "machine-config-1",
						},
					},
				}
			}),
			new: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						Name: "md-0",
						MachineGroupRef: &anywherev1.Ref{
							Kind: anywherev1.VSphereMachineConfigKind,
							Name: "machine-config-1",
						},
					},
				}
			}),
			want: []anywherev1.WorkerNodeGroupConfiguration{},
		},
		{
			name: "one worker node group, missing name, new name is not default",
			current: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						MachineGroupRef: &anywherev1.Ref{
							Kind: anywherev1.VSphereMachineConfigKind,
							Name: "machine-config-1",
						},
					},
				}
			}),
			new: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						Name: "worker-node-group-0",
						MachineGroupRef: &anywherev1.Ref{
							Kind: anywherev1.VSphereMachineConfigKind,
							Name: "machine-config-1",
						},
					},
				}
			}),
			want: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name: "md-0",
					MachineGroupRef: &anywherev1.Ref{
						Kind: anywherev1.VSphereMachineConfigKind,
						Name: "machine-config-1",
					},
				},
			},
		},
		{
			name: "new added, some removed, some stay",
			current: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						Name: "worker-node-group-0",
						MachineGroupRef: &anywherev1.Ref{
							Kind: anywherev1.VSphereMachineConfigKind,
							Name: "machine-config-1",
						},
					},
					{
						Name: "worker-node-group-1",
						MachineGroupRef: &anywherev1.Ref{
							Kind: anywherev1.VSphereMachineConfigKind,
							Name: "machine-config-1",
						},
					},
				}
			}),
			new: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						Name: "worker-node-group-0",
						MachineGroupRef: &anywherev1.Ref{
							Kind: anywherev1.VSphereMachineConfigKind,
							Name: "machine-config-1",
						},
					},
					{
						Name: "worker-node-group-2",
						MachineGroupRef: &anywherev1.Ref{
							Kind: anywherev1.VSphereMachineConfigKind,
							Name: "machine-config-1",
						},
					},
				}
			}),
			want: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name: "worker-node-group-1",
					MachineGroupRef: &anywherev1.Ref{
						Kind: anywherev1.VSphereMachineConfigKind,
						Name: "machine-config-1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cluster.NodeGroupsToDelete(tt.current, tt.new); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodeGroupsToDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}
