package api_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestWithWorkerNodeCount(t *testing.T) {
	tests := []struct {
		name    string
		cluster *anywherev1.Cluster
		want    []anywherev1.WorkerNodeGroupConfiguration
	}{
		{
			name: "with worker node config empty",
			cluster: &anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{},
			},
			want: []anywherev1.WorkerNodeGroupConfiguration{
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
		{
			name: "with worker node config greater than one",
			cluster: &anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							Name:                     "md-0",
							Count:                    ptr.Int(1),
							AutoScalingConfiguration: nil,
							MachineGroupRef:          nil,
							Taints:                   nil,
							Labels:                   nil,
						},
						{
							Name:                     "md-1",
							Count:                    ptr.Int(1),
							AutoScalingConfiguration: nil,
							MachineGroupRef:          nil,
							Taints:                   nil,
							Labels:                   nil,
						},
					},
				},
			},
			want: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name:                     "md-0",
					Count:                    ptr.Int(5),
					AutoScalingConfiguration: nil,
					MachineGroupRef:          nil,
					Taints:                   nil,
					Labels:                   nil,
				},
				{
					Name:                     "md-1",
					Count:                    ptr.Int(1),
					AutoScalingConfiguration: nil,
					MachineGroupRef:          nil,
					Taints:                   nil,
					Labels:                   nil,
				},
			},
		},
		{
			name: "with empty worker node config",
			cluster: &anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{},
			},
			want: []anywherev1.WorkerNodeGroupConfiguration{
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
			api.WithWorkerNodeCount(5)(tt.cluster)
			g := NewWithT(t)
			g.Expect(tt.cluster.Spec.WorkerNodeGroupConfigurations).To(Equal(tt.want))
		})
	}
}

func TestWithWorkerNodeAutoScalingConfig(t *testing.T) {
	expectedAutoScalingConfiguration := &anywherev1.AutoScalingConfiguration{
		MinCount: 1,
		MaxCount: 5,
	}
	tests := []struct {
		name    string
		cluster *anywherev1.Cluster
		want    []anywherev1.WorkerNodeGroupConfiguration
	}{
		{
			name: "with worker node config without autoscaling",
			cluster: &anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							Name:                     "md-0",
							Count:                    ptr.Int(2),
							AutoScalingConfiguration: nil,
							MachineGroupRef:          nil,
							Taints:                   nil,
							Labels:                   nil,
						},
					},
				},
			},
			want: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name:                     "md-0",
					Count:                    ptr.Int(2),
					AutoScalingConfiguration: expectedAutoScalingConfiguration,
					MachineGroupRef:          nil,
					Taints:                   nil,
					Labels:                   nil,
				},
			},
		},
		{
			name: "with worker node config empty",
			cluster: &anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{},
			},
			want: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name:                     "",
					Count:                    ptr.Int(1),
					AutoScalingConfiguration: expectedAutoScalingConfiguration,
					MachineGroupRef:          nil,
					Taints:                   nil,
					Labels:                   nil,
				},
			},
		},
		{
			name: "with worker node config greater than one",
			cluster: &anywherev1.Cluster{
				Spec: anywherev1.ClusterSpec{
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							Name:                     "md-0",
							Count:                    ptr.Int(1),
							AutoScalingConfiguration: nil,
							MachineGroupRef:          nil,
							Taints:                   nil,
							Labels:                   nil,
						},
						{
							Name:                     "md-1",
							Count:                    ptr.Int(1),
							AutoScalingConfiguration: nil,
							MachineGroupRef:          nil,
							Taints:                   nil,
							Labels:                   nil,
						},
					},
				},
			},
			want: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name:                     "md-0",
					Count:                    ptr.Int(1),
					AutoScalingConfiguration: expectedAutoScalingConfiguration,
					MachineGroupRef:          nil,
					Taints:                   nil,
					Labels:                   nil,
				},
				{
					Name:                     "md-1",
					Count:                    ptr.Int(1),
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
			api.WithWorkerNodeAutoScalingConfig(1, 5)(tt.cluster)
			g := NewWithT(t)
			g.Expect(tt.cluster.Spec.WorkerNodeGroupConfigurations).To(Equal(tt.want))
		})
	}
}
