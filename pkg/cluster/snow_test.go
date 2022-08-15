package cluster_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
)

func TestSetSnowMachineConfigsAnnotations(t *testing.T) {
	tests := []struct {
		name                   string
		config                 *cluster.Config
		wantSnowMachineConfigs map[string]*anywherev1.SnowMachineConfig
	}{
		{
			name: "workload cluster with external etcd",
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-cluster",
					},
					Spec: anywherev1.ClusterSpec{
						ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
							MachineGroupRef: &anywherev1.Ref{
								Name: "cp-machine",
							},
						},
						ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
							MachineGroupRef: &anywherev1.Ref{
								Name: "etcd-machine",
							},
						},
						ManagementCluster: anywherev1.ManagementCluster{
							Name: "mgmt-cluster",
						},
					},
				},
				SnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
					"cp-machine": {
						ObjectMeta: metav1.ObjectMeta{
							Name: "cp-machine",
						},
					},
					"etcd-machine": {
						ObjectMeta: metav1.ObjectMeta{
							Name: "etcd-machine",
						},
					},
				},
			},
			wantSnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
				"cp-machine": {
					ObjectMeta: metav1.ObjectMeta{
						Name: "cp-machine",
						Annotations: map[string]string{
							"anywhere.eks.amazonaws.com/control-plane": "true",
							"anywhere.eks.amazonaws.com/managed-by":    "mgmt-cluster",
						},
					},
				},
				"etcd-machine": {
					ObjectMeta: metav1.ObjectMeta{
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
			config: &cluster.Config{
				Cluster: &anywherev1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-cluster",
					},
					Spec: anywherev1.ClusterSpec{
						ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
							MachineGroupRef: &anywherev1.Ref{
								Name: "cp-machine",
							},
						},
					},
				},
				SnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
					"cp-machine": {
						ObjectMeta: metav1.ObjectMeta{
							Name: "cp-machine",
						},
					},
				},
			},
			wantSnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
				"cp-machine": {
					ObjectMeta: metav1.ObjectMeta{
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
			err := cluster.SetSnowMachineConfigsAnnotations(tt.config)
			g.Expect(err).To(Succeed())
			g.Expect(tt.config.SnowMachineConfigs).To(Equal(tt.wantSnowMachineConfigs))
		})
	}
}

func TestDefaultConfigClientBuilderSnowCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			DatacenterRef: anywherev1.Ref{
				Kind: anywherev1.SnowDatacenterKind,
				Name: "datacenter",
			},
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.SnowMachineConfigKind,
					Name: "machine-1",
				},
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					MachineGroupRef: &anywherev1.Ref{
						Kind: anywherev1.SnowMachineConfigKind,
						Name: "machine-2",
					},
				},
				{
					MachineGroupRef: &anywherev1.Ref{
						Kind: anywherev1.VSphereMachineConfigKind,
						Name: "machine-3",
					},
				},
			},
		},
	}
	datacenter := &anywherev1.SnowDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: "default",
		},
	}
	machineControlPlane := &anywherev1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-1",
			Namespace: "default",
		},
	}

	machineWorker := &anywherev1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-2",
			Namespace: "default",
		},
	}

	client.EXPECT().Get(ctx, "datacenter", "default", &anywherev1.SnowDatacenterConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			d := obj.(*anywherev1.SnowDatacenterConfig)
			d.ObjectMeta = datacenter.ObjectMeta
			return nil
		},
	)

	client.EXPECT().Get(ctx, "machine-1", "default", &anywherev1.SnowMachineConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			m := obj.(*anywherev1.SnowMachineConfig)
			m.ObjectMeta = machineControlPlane.ObjectMeta
			return nil
		},
	)

	client.EXPECT().Get(ctx, "machine-2", "default", &anywherev1.SnowMachineConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			m := obj.(*anywherev1.SnowMachineConfig)
			m.ObjectMeta = machineWorker.ObjectMeta
			return nil
		},
	)

	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.Cluster).To(Equal(cluster))
	g.Expect(config.SnowDatacenter).To(Equal(datacenter))
	g.Expect(len(config.SnowMachineConfigs)).To(Equal(2))
	g.Expect(config.SnowMachineConfigs["machine-1"]).To(Equal(machineControlPlane))
	g.Expect(config.SnowMachineConfigs["machine-2"]).To(Equal(machineWorker))
}

func TestParseConfigMissingSnowDatacenter(t *testing.T) {
	g := NewWithT(t)
	got, err := cluster.ParseConfigFromFile("testdata/cluster_snow_missing_datacenter.yaml")

	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(got.DockerDatacenter).To(BeNil())
}
