package docker_test

import (
	"context"
	"testing"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
)

func TestControlPlaneSpecNewCluster(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := clusterSpec()
	client := test.NewFakeKubeClient()

	cp, err := docker.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(wantCluster()))
	g.Expect(cp.KubeadmControlPlane).NotTo(BeNil())
	g.Expect(cp.EtcdCluster).NotTo(BeNil())
	g.Expect(cp.ProviderCluster).NotTo(BeNil())
	g.Expect(cp.ProviderMachineTemplate).NotTo(BeNil())
	g.Expect(cp.ProviderMachineTemplate.Name).To(Equal("test-cluster-control-plane-1"))
	g.Expect(cp.KubeadmControlPlane).NotTo(BeNil())
	g.Expect(cp.EtcdMachineTemplate).NotTo(BeNil())
	g.Expect(cp.EtcdMachineTemplate.Name).To(Equal("test-cluster-etcd-1"))
}

func TestControlPlaneSpecUpdateMachineTemplates(t *testing.T) {
	g := NewWithT(t)
	logger := test.NewNullLogger()
	ctx := context.Background()
	spec := clusterSpec()
	client := test.NewFakeKubeClient(
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: constants.EksaSystemNamespace,
			},
			Spec: controlplanev1.KubeadmControlPlaneSpec{
				MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
					InfrastructureRef: corev1.ObjectReference{
						Name:      "test-cluster-control-plane-1",
						Namespace: constants.EksaSystemNamespace,
					},
				},
			},
		},
		&etcdv1.EtcdadmCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-etcd",
				Namespace: constants.EksaSystemNamespace,
			},
			Spec: etcdv1.EtcdadmClusterSpec{
				InfrastructureTemplate: corev1.ObjectReference{
					Name: "test-cluster-etcd-1",
				},
			},
		},
		&dockerv1.DockerMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-control-plane-1",
				Namespace: constants.EksaSystemNamespace,
			},
			Spec: dockerv1.DockerMachineTemplateSpec{
				Template: dockerv1.DockerMachineTemplateResource{
					Spec: dockerv1.DockerMachineSpec{
						CustomImage: "old-custom-image",
					},
				},
			},
		},
		&dockerv1.DockerMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-etcd-1",
				Namespace: constants.EksaSystemNamespace,
			},
			Spec: dockerv1.DockerMachineTemplateSpec{
				Template: dockerv1.DockerMachineTemplateResource{
					Spec: dockerv1.DockerMachineSpec{
						CustomImage: "old-custom-image-etcd",
					},
				},
			},
		},
	)

	cp, err := docker.ControlPlaneSpec(ctx, logger, client, spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp).NotTo(BeNil())
	g.Expect(cp.Cluster).To(Equal(wantCluster()))
	g.Expect(cp.KubeadmControlPlane).NotTo(BeNil())
	g.Expect(cp.EtcdCluster).NotTo(BeNil())
	g.Expect(cp.ProviderCluster).NotTo(BeNil())
	g.Expect(cp.ProviderMachineTemplate).NotTo(BeNil())
	g.Expect(cp.ProviderMachineTemplate.Name).To(Equal("test-cluster-control-plane-2"))
	g.Expect(cp.KubeadmControlPlane).NotTo(BeNil())
	g.Expect(cp.EtcdMachineTemplate).NotTo(BeNil())
	g.Expect(cp.EtcdMachineTemplate.Name).To(Equal("test-cluster-etcd-2"))
}

func clusterSpec() *cluster.Spec {
	return test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "test-cluster"
		s.Cluster.Spec.KubernetesVersion = "1.19"
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.0.0/16"}
		s.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"10.128.0.0/12"}
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
		s.VersionsBundle = versionsBundle
		s.Cluster.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{Count: 3}
		s.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 3, MachineGroupRef: &v1alpha1.Ref{Name: "test-cluster"}, Name: "md-0"}}
	})
}

func wantCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "eksa-system",
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{
				APIServerPort: nil,
				Services: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{"10.128.0.0/12"},
				},
				Pods: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{"192.168.0.0/16"},
				},
				ServiceDomain: "cluster.local",
			},
			ControlPlaneRef: &corev1.ObjectReference{
				Kind:       "KubeadmControlPlane",
				Namespace:  "eksa-system",
				Name:       "test-cluster",
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			},
			ManagedExternalEtcdRef: &corev1.ObjectReference{
				Kind:       "EtcdadmCluster",
				Namespace:  "eksa-system",
				Name:       "test-cluster-etcd",
				APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
			},
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "DockerCluster",
				Namespace:  "eksa-system",
				Name:       "test-cluster",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
		},
	}
}
