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

func TestParseConfigMissingVSphereDatacenter(t *testing.T) {
	g := NewWithT(t)
	got, err := cluster.ParseConfigFromFile("testdata/cluster_vsphere_missing_datacenter.yaml")

	g.Expect(err).To(Not(HaveOccurred()))

	g.Expect(got.VSphereDatacenter).To(BeNil())
}

func TestDefaultConfigClientBuilderVSphereCluster(t *testing.T) {
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
				Kind: anywherev1.VSphereDatacenterKind,
				Name: "datacenter",
			},
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.VSphereMachineConfigKind,
					Name: "machine-1",
				},
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					MachineGroupRef: &anywherev1.Ref{
						Kind: anywherev1.VSphereMachineConfigKind,
						Name: "machine-2",
					},
				},
				{
					MachineGroupRef: &anywherev1.Ref{
						Kind: anywherev1.CloudStackMachineConfigKind, // Should not process this one
						Name: "machine-3",
					},
				},
			},
		},
	}
	datacenter := &anywherev1.VSphereDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: "default",
		},
	}
	machineControlPlane := &anywherev1.VSphereMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-1",
			Namespace: "default",
		},
	}

	machineWorker := &anywherev1.VSphereMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-2",
			Namespace: "default",
		},
	}

	client.EXPECT().Get(ctx, "datacenter", "default", &anywherev1.VSphereDatacenterConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			d := obj.(*anywherev1.VSphereDatacenterConfig)
			d.ObjectMeta = datacenter.ObjectMeta
			d.Spec = datacenter.Spec
			return nil
		},
	)

	client.EXPECT().Get(ctx, "machine-1", "default", &anywherev1.VSphereMachineConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			m := obj.(*anywherev1.VSphereMachineConfig)
			m.ObjectMeta = machineControlPlane.ObjectMeta
			return nil
		},
	)

	client.EXPECT().Get(ctx, "machine-2", "default", &anywherev1.VSphereMachineConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			m := obj.(*anywherev1.VSphereMachineConfig)
			m.ObjectMeta = machineWorker.ObjectMeta
			return nil
		},
	)

	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.Cluster).To(Equal(cluster))
	g.Expect(config.VSphereDatacenter).To(Equal(datacenter))
	g.Expect(len(config.VSphereMachineConfigs)).To(Equal(2))
	g.Expect(config.VSphereMachineConfigs["machine-1"]).To(Equal(machineControlPlane))
	g.Expect(config.VSphereMachineConfigs["machine-2"]).To(Equal(machineWorker))
}
