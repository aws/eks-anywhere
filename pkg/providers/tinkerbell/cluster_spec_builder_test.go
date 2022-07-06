package tinkerbell_test

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
)

type ClusterSpecBuilder struct {
	ControlPlaneMachineName    string
	ExternalEtcdMachineName    string
	WorkerNodeGroupMachineName string
	Namespace                  string
	IncludeHardwareSelectors   bool
	AdditionalNodeGroups       []v1alpha1.WorkerNodeGroupConfiguration
}

func DefaultClusterSpecBuilder() ClusterSpecBuilder {
	return ClusterSpecBuilder{
		ControlPlaneMachineName:    "control-plane",
		ExternalEtcdMachineName:    "external-etcd",
		WorkerNodeGroupMachineName: "worker-node-group",
		Namespace:                  "namespace",
		IncludeHardwareSelectors:   true,
	}
}

func DefaultClusterSpec() *tinkerbell.ClusterSpec {
	return DefaultClusterSpecBuilder().Build()
}

func (b *ClusterSpecBuilder) WithoutHardwareSelectors() {
	b.IncludeHardwareSelectors = false
}

func (b *ClusterSpecBuilder) WithAdditionalNodeGroups(groups ...v1alpha1.WorkerNodeGroupConfiguration) {
	b.AdditionalNodeGroups = groups
}

func (b ClusterSpecBuilder) Build() *tinkerbell.ClusterSpec {
	spec := &tinkerbell.ClusterSpec{
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
							{
								Name:  "worker-node-group-1",
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
					HardwareSelector: v1alpha1.HardwareSelector{"type": "cp"},
					OSFamily:         v1alpha1.Ubuntu,
				},
			},
			b.ExternalEtcdMachineName: {
				ObjectMeta: v1.ObjectMeta{
					Name:      b.ExternalEtcdMachineName,
					Namespace: b.Namespace,
				},
				Spec: v1alpha1.TinkerbellMachineConfigSpec{
					HardwareSelector: v1alpha1.HardwareSelector{"type": "etcd"},
					OSFamily:         v1alpha1.Ubuntu,
				},
			},
			b.WorkerNodeGroupMachineName: {
				ObjectMeta: v1.ObjectMeta{
					Name:      b.WorkerNodeGroupMachineName,
					Namespace: b.Namespace,
				},
				Spec: v1alpha1.TinkerbellMachineConfigSpec{
					HardwareSelector: v1alpha1.HardwareSelector{"type": "worker"},
					OSFamily:         v1alpha1.Ubuntu,
				},
			},
		},
	}

	if !b.IncludeHardwareSelectors {
		for _, config := range spec.MachineConfigs {
			config.Spec.HardwareSelector = v1alpha1.HardwareSelector{}
		}
	}

	spec.Config.Cluster.Spec.WorkerNodeGroupConfigurations = append(
		spec.Config.Cluster.Spec.WorkerNodeGroupConfigurations,
		b.AdditionalNodeGroups...,
	)

	return spec
}
