package tinkerbell_test

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
)

type ValidClusterSpecBuilder struct {
	ControlPlaneMachineName    string
	ExternalEtcdMachineName    string
	WorkerNodeGroupMachineName string
	Namespace                  string
}

func NewDefaultValidClusterSpecBuilder() ValidClusterSpecBuilder {
	return ValidClusterSpecBuilder{
		ControlPlaneMachineName:    "control-plane",
		ExternalEtcdMachineName:    "external-etcd",
		WorkerNodeGroupMachineName: "worker-node-group",
		Namespace:                  "namespace",
	}
}

func (b ValidClusterSpecBuilder) Build() *tinkerbell.ClusterSpec {
	return &tinkerbell.ClusterSpec{
		Spec: &cluster.Spec{
			Config: &cluster.Config{
				Cluster: &v1alpha1.Cluster{
					Spec: v1alpha1.ClusterSpec{
						ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
							Count: 1,
							MachineGroupRef: &v1alpha1.Ref{
								Kind: v1alpha1.TinkerbellMachineConfigKind,
								Name: b.ControlPlaneMachineName,
							},
							Endpoint: &v1alpha1.Endpoint{
								Host: "1.1.1.1",
							},
						},
						ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{
							Count: 1,
							MachineGroupRef: &v1alpha1.Ref{
								Kind: v1alpha1.TinkerbellMachineConfigKind,
								Name: b.ExternalEtcdMachineName,
							},
						},
						WorkerNodeGroupConfigurations: []v1alpha1.WorkerNodeGroupConfiguration{
							{
								Name:  "worker-node-group-0",
								Count: 1,
								MachineGroupRef: &v1alpha1.Ref{
									Kind: v1alpha1.TinkerbellMachineConfigKind,
									Name: b.WorkerNodeGroupMachineName,
								},
							},
						},
						DatacenterRef: v1alpha1.Ref{
							Kind: v1alpha1.TinkerbellDatacenterKind,
							Name: "tinkerbell-data-center",
						},
					},
				},
			},
		},
		DatacenterConfig: &v1alpha1.TinkerbellDatacenterConfig{
			ObjectMeta: v1.ObjectMeta{
				Name:      "datacenter-config",
				Namespace: b.Namespace,
			},
			Spec: v1alpha1.TinkerbellDatacenterConfigSpec{
				TinkerbellIP: "1.1.1.1",
			},
		},
		MachineConfigs: map[string]*v1alpha1.TinkerbellMachineConfig{
			b.ControlPlaneMachineName: {
				ObjectMeta: v1.ObjectMeta{
					Name:      b.ControlPlaneMachineName,
					Namespace: b.Namespace,
				},
				Spec: v1alpha1.TinkerbellMachineConfigSpec{
					OSFamily: v1alpha1.Ubuntu,
				},
			},
			b.ExternalEtcdMachineName: {
				ObjectMeta: v1.ObjectMeta{
					Name:      b.ExternalEtcdMachineName,
					Namespace: b.Namespace,
				},
				Spec: v1alpha1.TinkerbellMachineConfigSpec{
					OSFamily: v1alpha1.Ubuntu,
				},
			},
			b.WorkerNodeGroupMachineName: {
				ObjectMeta: v1.ObjectMeta{
					Name:      b.WorkerNodeGroupMachineName,
					Namespace: b.Namespace,
				},
				Spec: v1alpha1.TinkerbellMachineConfigSpec{
					OSFamily: v1alpha1.Ubuntu,
				},
			},
		},
	}
}
