package cluster_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"credentials": []byte("creds"),
			"ca-bundle":   []byte("certs"),
		},
		Type: "Opaque",
	}
	datacenter := &anywherev1.SnowDatacenterConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: "default",
		},
		Spec: anywherev1.SnowDatacenterConfigSpec{
			IdentityRef: anywherev1.Ref{
				Kind: "Secret",
				Name: secret.Name,
			},
		},
	}
	machineControlPlane := &anywherev1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-1",
			Namespace: "default",
		},
		Spec: anywherev1.SnowMachineConfigSpec{
			Network: anywherev1.SnowNetwork{
				DirectNetworkInterfaces: []anywherev1.SnowDirectNetworkInterface{
					{
						Index: 1,
						IPPoolRef: &anywherev1.Ref{
							Kind: anywherev1.SnowIPPoolKind,
							Name: "ip-pool-1",
						},
						Primary: true,
					},
					{
						Index: 2,
						IPPoolRef: &anywherev1.Ref{
							Kind: anywherev1.SnowIPPoolKind,
							Name: "ip-pool-2",
						},
						Primary: false,
					},
					{
						Index:   3,
						Primary: false,
					},
				},
			},
		},
	}

	machineWorker := &anywherev1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-2",
			Namespace: "default",
		},
		Spec: anywherev1.SnowMachineConfigSpec{
			Network: anywherev1.SnowNetwork{
				DirectNetworkInterfaces: []anywherev1.SnowDirectNetworkInterface{
					{
						Index: 1,
						IPPoolRef: &anywherev1.Ref{
							Kind: anywherev1.SnowIPPoolKind,
							Name: "ip-pool-1",
						},
						Primary: true,
					},
				},
			},
		},
	}

	pool1 := &anywherev1.SnowIPPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ip-pool-1",
		},
		Spec: anywherev1.SnowIPPoolSpec{
			Pools: []anywherev1.IPPool{
				{
					IPStart: "start-1",
					IPEnd:   "end-1",
					Subnet:  "subnet-1",
					Gateway: "gateway-1",
				},
			},
		},
	}

	pool2 := &anywherev1.SnowIPPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ip-pool-2",
		},
		Spec: anywherev1.SnowIPPoolSpec{
			Pools: []anywherev1.IPPool{
				{
					IPStart: "start-2",
					IPEnd:   "end-2",
					Subnet:  "subnet-2",
					Gateway: "gateway-2",
				},
			},
		},
	}

	client.EXPECT().Get(ctx, "datacenter", "default", &anywherev1.SnowDatacenterConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			d := obj.(*anywherev1.SnowDatacenterConfig)
			d.ObjectMeta = datacenter.ObjectMeta
			d.Spec = datacenter.Spec
			return nil
		},
	)

	client.EXPECT().Get(ctx, "machine-1", "default", &anywherev1.SnowMachineConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			m := obj.(*anywherev1.SnowMachineConfig)
			m.ObjectMeta = machineControlPlane.ObjectMeta
			m.Spec = machineControlPlane.Spec
			return nil
		},
	)

	client.EXPECT().Get(ctx, "ip-pool-1", "default", &anywherev1.SnowIPPool{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			p := obj.(*anywherev1.SnowIPPool)
			p.ObjectMeta = pool1.ObjectMeta
			p.Spec = pool1.Spec
			return nil
		},
	)

	client.EXPECT().Get(ctx, "ip-pool-2", "default", &anywherev1.SnowIPPool{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			p := obj.(*anywherev1.SnowIPPool)
			p.ObjectMeta = pool2.ObjectMeta
			p.Spec = pool2.Spec
			return nil
		},
	)

	client.EXPECT().Get(ctx, secret.Name, "default", &corev1.Secret{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			d := obj.(*corev1.Secret)
			d.ObjectMeta = secret.ObjectMeta
			d.TypeMeta = secret.TypeMeta
			d.Data = secret.Data
			d.Type = secret.Type
			return nil
		},
	)

	client.EXPECT().Get(ctx, "machine-2", "default", &anywherev1.SnowMachineConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			m := obj.(*anywherev1.SnowMachineConfig)
			m.ObjectMeta = machineWorker.ObjectMeta
			m.Spec = machineWorker.Spec
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
	g.Expect(config.SnowCredentialsSecret).To(Equal(secret))
	g.Expect(config.SnowIPPools["ip-pool-1"]).To(Equal(pool1))
	g.Expect(config.SnowIPPools["ip-pool-2"]).To(Equal(pool2))
}

func TestDefaultConfigClientBuilderSnowClusterGetIPPoolError(t *testing.T) {
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
		},
	}

	machineControlPlane := &anywherev1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine-1",
			Namespace: "default",
		},
		Spec: anywherev1.SnowMachineConfigSpec{
			Network: anywherev1.SnowNetwork{
				DirectNetworkInterfaces: []anywherev1.SnowDirectNetworkInterface{
					{
						IPPoolRef: &anywherev1.Ref{
							Kind: anywherev1.SnowIPPoolKind,
							Name: "ip-pool-1",
						},
					},
				},
			},
		},
	}

	client.EXPECT().Get(ctx, "datacenter", "default", &anywherev1.SnowDatacenterConfig{}).Return(nil)

	client.EXPECT().Get(ctx, "machine-1", "default", &anywherev1.SnowMachineConfig{}).Return(nil).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			m := obj.(*anywherev1.SnowMachineConfig)
			m.ObjectMeta = machineControlPlane.ObjectMeta
			m.Spec = machineControlPlane.Spec
			return nil
		},
	)

	client.EXPECT().Get(ctx, "ip-pool-1", "default", &anywherev1.SnowIPPool{}).Return(errors.New("error get ip pool"))

	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).To(MatchError(ContainSubstring("error get ip pool")))
	g.Expect(config).To(BeNil())
}

func TestParseConfigMissingSnowDatacenter(t *testing.T) {
	g := NewWithT(t)
	got, err := cluster.ParseConfigFromFile("testdata/cluster_snow_missing_datacenter.yaml")

	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(got.DockerDatacenter).To(BeNil())
}

func TestSetSnowDatacenterIndentityRefDefault(t *testing.T) {
	tests := []struct {
		name   string
		before *anywherev1.SnowDatacenterConfig
		after  *anywherev1.SnowDatacenterConfig
	}{
		{
			name: "identity ref empty",
			before: &anywherev1.SnowDatacenterConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: anywherev1.SnowDatacenterConfigSpec{},
			},
			after: &anywherev1.SnowDatacenterConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: anywherev1.SnowDatacenterConfigSpec{
					IdentityRef: anywherev1.Ref{
						Name: "test-snow-credentials",
						Kind: "Secret",
					},
				},
			},
		},
		{
			name: "identity ref exists",
			before: &anywherev1.SnowDatacenterConfig{
				Spec: anywherev1.SnowDatacenterConfigSpec{
					IdentityRef: anywherev1.Ref{
						Name: "creds-1",
						Kind: "Secret",
					},
				},
			},
			after: &anywherev1.SnowDatacenterConfig{
				Spec: anywherev1.SnowDatacenterConfigSpec{
					IdentityRef: anywherev1.Ref{
						Name: "creds-1",
						Kind: "Secret",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			cluster.SetSnowDatacenterIndentityRefDefault(tt.before)
			g.Expect(tt.before).To(Equal(tt.after))
		})
	}
}

func TestValidateSnowMachineRefExistsError(t *testing.T) {
	g := NewWithT(t)
	c := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       anywherev1.ClusterKind,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eksa-unit-test",
				Namespace: "ns-1",
			},
			Spec: anywherev1.ClusterSpec{
				WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
					{
						MachineGroupRef: &anywherev1.Ref{
							Name: "worker-not-exists",
							Kind: "SnowMachineConfig",
						},
					},
				},
			},
		},
		SnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
			"worker-1": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "worker-1",
				},
			},
		},
	}
	g.Expect(cluster.ValidateConfig(c)).To(
		MatchError(ContainSubstring("unable to find SnowMachineConfig worker-not-exists")),
	)
}

func TestValidateSnowUnstackedEtcdWithDHCPError(t *testing.T) {
	g := NewWithT(t)
	c := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       anywherev1.ClusterKind,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eksa-unit-test",
				Namespace: "ns-1",
			},
			Spec: anywherev1.ClusterSpec{
				DatacenterRef: anywherev1.Ref{
					Kind: anywherev1.SnowDatacenterKind,
				},
				ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
					MachineGroupRef: &anywherev1.Ref{
						Name: "etcd-1",
					},
				},
			},
		},
		SnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
			"etcd-1": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "etcd-1",
				},
				Spec: anywherev1.SnowMachineConfigSpec{
					Network: anywherev1.SnowNetwork{
						DirectNetworkInterfaces: []anywherev1.SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    true,
								Primary: true,
							},
						},
					},
				},
			},
		},
	}
	g.Expect(cluster.ValidateConfig(c)).To(
		MatchError(ContainSubstring("creating unstacked etcd machine with DHCP is not supported for snow")),
	)
}

func TestValidateSnowUnstackedEtcdMissIPPoolError(t *testing.T) {
	g := NewWithT(t)
	c := &cluster.Config{
		Cluster: &anywherev1.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       anywherev1.ClusterKind,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "eksa-unit-test",
				Namespace: "ns-1",
			},
			Spec: anywherev1.ClusterSpec{
				DatacenterRef: anywherev1.Ref{
					Kind: anywherev1.SnowDatacenterKind,
				},
				ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
					MachineGroupRef: &anywherev1.Ref{
						Name: "etcd-1",
					},
				},
			},
		},
		SnowMachineConfigs: map[string]*anywherev1.SnowMachineConfig{
			"etcd-1": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "etcd-1",
				},
				Spec: anywherev1.SnowMachineConfigSpec{
					Network: anywherev1.SnowNetwork{
						DirectNetworkInterfaces: []anywherev1.SnowDirectNetworkInterface{
							{
								Index:   1,
								DHCP:    false,
								Primary: true,
							},
						},
					},
				},
			},
		},
	}
	g.Expect(cluster.ValidateConfig(c)).To(
		MatchError(ContainSubstring("snow machine config ip pool must be specified when using static IP")),
	)
}
