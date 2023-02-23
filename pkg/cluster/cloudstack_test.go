package cluster_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestParseConfigMissingCloudstackDatacenter(t *testing.T) {
	g := NewWithT(t)
	got, err := cluster.ParseConfigFromFile("testdata/cluster_cloudstack_missing_datacenter.yaml")

	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(got.CloudStackDatacenter).To(BeNil())
}

func TestDefaultConfigClientBuilderBuildCloudStackClusterSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	cluster := cloudStackCluster()

	datacenter := cloudStackDatacenter()
	machineControlPlane := cloudStackMachineConfig("cp-machine-1")
	machineWorker1 := cloudStackMachineConfig("worker-machine-1")
	machineWorker2 := cloudStackMachineConfig("worker-machine-2")

	client := test.NewFakeKubeClient(datacenter, machineControlPlane, machineWorker1, machineWorker2)
	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.Cluster).To(Equal(cluster))
	g.Expect(config.CloudStackDatacenter).To(Equal(datacenter))
	g.Expect(config.CloudStackMachineConfigs["cp-machine-1"]).To(Equal(machineControlPlane))
	g.Expect(config.CloudStackMachineConfigs["worker-machine-1"]).To(Equal(machineWorker1))
}

func TestDefaultConfigClientBuilderBuildCloudStackClusterFailure(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	cluster := cloudStackCluster()

	datacenter := cloudStackDatacenter()
	machineControlPlane := cloudStackMachineConfig("cp-machine-1")
	machineWorker1 := cloudStackMachineConfig("worker-machine-1")

	tests := []struct {
		name    string
		objects []client.Object
		wantErr string
	}{
		{
			name: "missing machine config",
			objects: []client.Object{
				datacenter,
				machineControlPlane,
			},
			wantErr: "cloudstackmachineconfigs.anywhere.eks.amazonaws.com \"worker-machine-1\" not found",
		},
		{
			name: "missing datacenter config",
			objects: []client.Object{
				machineControlPlane,
				machineWorker1,
			},
			wantErr: "cloudstackdatacenterconfigs.anywhere.eks.amazonaws.com \"datacenter\" not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := test.NewFakeKubeClient(tt.objects...)
			_, err := b.Build(ctx, client, cluster)
			g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
		})
	}
}

func cloudStackCluster() *anywherev1.Cluster {
	return &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			DatacenterRef: anywherev1.Ref{
				Kind: anywherev1.CloudStackDatacenterKind,
				Name: "datacenter",
			},
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 1,
				MachineGroupRef: &anywherev1.Ref{
					Name: "cp-machine-1",
					Kind: anywherev1.CloudStackMachineConfigKind,
				},
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Name: "md-0",
					MachineGroupRef: &anywherev1.Ref{
						Name: "worker-machine-1",
						Kind: anywherev1.CloudStackMachineConfigKind,
					},
				},
				{
					Name: "md-1",
					MachineGroupRef: &anywherev1.Ref{
						Name: "worker-machine-2",
						Kind: anywherev1.VSphereMachineConfigKind, // Should not process this one
					},
				},
			},
		},
	}
}

func cloudStackDatacenter() *anywherev1.CloudStackDatacenterConfig {
	return &anywherev1.CloudStackDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.CloudStackDatacenterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: "default",
		},
	}
}

func cloudStackMachineConfig(name string) *anywherev1.CloudStackMachineConfig {
	return &anywherev1.CloudStackMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.CloudStackMachineConfigKind,
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}
}
