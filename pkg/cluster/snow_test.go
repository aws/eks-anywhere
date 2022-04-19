package cluster

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestSetSnowMachineConfigsAnnotations(t *testing.T) {
	tests := []struct {
		name                   string
		config                 *Config
		wantSnowMachineConfigs map[string]*v1alpha1.SnowMachineConfig
	}{
		{
			name: "workload cluster with external etcd",
			config: &Config{
				Cluster: &v1alpha1.Cluster{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-cluster",
					},
					Spec: v1alpha1.ClusterSpec{
						ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
							MachineGroupRef: &v1alpha1.Ref{
								Name: "cp-machine",
							},
						},
						ExternalEtcdConfiguration: &v1alpha1.ExternalEtcdConfiguration{
							MachineGroupRef: &v1alpha1.Ref{
								Name: "etcd-machine",
							},
						},
						ManagementCluster: v1alpha1.ManagementCluster{
							Name: "mgmt-cluster",
						},
					},
				},
				SnowMachineConfigs: map[string]*v1alpha1.SnowMachineConfig{
					"cp-machine": {
						ObjectMeta: v1.ObjectMeta{
							Name: "cp-machine",
						},
					},
					"etcd-machine": {
						ObjectMeta: v1.ObjectMeta{
							Name: "etcd-machine",
						},
					},
				},
			},
			wantSnowMachineConfigs: map[string]*v1alpha1.SnowMachineConfig{
				"cp-machine": {
					ObjectMeta: v1.ObjectMeta{
						Name: "cp-machine",
						Annotations: map[string]string{
							"anywhere.eks.amazonaws.com/control-plane": "true",
							"anywhere.eks.amazonaws.com/managed-by":    "mgmt-cluster",
						},
					},
				},
				"etcd-machine": {
					ObjectMeta: v1.ObjectMeta{
						Name: "etcd-machine",
						Annotations: map[string]string{
							"anywhere.eks.amazonaws.com/etcd":       "true",
							"anywhere.eks.amazonaws.com/managed-by": "mgmt-cluster",
						},
					},
				},
			},
		},
		{
			name: "management cluster",
			config: &Config{
				Cluster: &v1alpha1.Cluster{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-cluster",
					},
					Spec: v1alpha1.ClusterSpec{
						ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
							MachineGroupRef: &v1alpha1.Ref{
								Name: "cp-machine",
							},
						},
					},
				},
				SnowMachineConfigs: map[string]*v1alpha1.SnowMachineConfig{
					"cp-machine": {
						ObjectMeta: v1.ObjectMeta{
							Name: "cp-machine",
						},
					},
				},
			},
			wantSnowMachineConfigs: map[string]*v1alpha1.SnowMachineConfig{
				"cp-machine": {
					ObjectMeta: v1.ObjectMeta{
						Name: "cp-machine",
						Annotations: map[string]string{
							"anywhere.eks.amazonaws.com/control-plane": "true",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := SetSnowMachineConfigsAnnotations(tt.config)
			g.Expect(err).To(Succeed())
			g.Expect(tt.config.SnowMachineConfigs).To(Equal(tt.wantSnowMachineConfigs))
		})
	}
}
