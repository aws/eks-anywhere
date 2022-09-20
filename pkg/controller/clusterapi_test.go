package controller_test

import (
	"context"
	"testing"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
)

func TestGetCAPIClusterSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	capiCluster := capiCluster()
	client := fake.NewClientBuilder().WithObjects(eksaCluster, capiCluster).Build()

	g.Expect(controller.GetCAPICluster(ctx, client, eksaCluster)).To(Equal(capiCluster))
}

func TestGetCAPIClusterNoCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	client := fake.NewClientBuilder().WithObjects(eksaCluster).Build()

	g.Expect(controller.GetCAPICluster(ctx, client, eksaCluster)).To(BeNil())
}

func TestGetCAPIClusterError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	// This should make the client fail because CRDs are not registered
	client := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()

	_, err := controller.GetCAPICluster(ctx, client, eksaCluster)
	g.Expect(err).To(MatchError(ContainSubstring("no kind is registered for the type")))
}

func TestGetCapiClusterObjectKey(t *testing.T) {
	g := NewWithT(t)
	eksaCluster := eksaCluster()

	expected := types.NamespacedName{
		Name:      "my-cluster",
		Namespace: "eksa-system",
	}

	key := controller.CapiClusterObjectKey(eksaCluster)
	g.Expect(key).To(Equal(expected))
}

func TestGetEtcdClusterSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	etcdCluster := etcdCluster()
	eksaCluster.Spec = anywherev1.ClusterSpec{
		ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
			Count:           5,
			MachineGroupRef: nil,
		},
	}
	client := fake.NewClientBuilder().WithObjects(eksaCluster, etcdCluster).Build()

	g.Expect(controller.GetEtcdCluster(ctx, client, eksaCluster)).To(Equal(etcdCluster))
}

func TestGetEtcdClusterNoCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	client := fake.NewClientBuilder().WithObjects(eksaCluster).Build()

	g.Expect(controller.GetEtcdCluster(ctx, client, eksaCluster)).To(BeNil())
}

func TestGetEtcdClusterError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	// This should make the client fail because CRDs are not registered
	client := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()

	_, err := controller.GetEtcdCluster(ctx, client, eksaCluster)
	g.Expect(err).To(MatchError(ContainSubstring("no kind is registered for the type")))
}

func eksaCluster() *anywherev1.Cluster {
	return &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
	}
}

func capiCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}
}

func etcdCluster() *etcdv1.EtcdadmCluster {
	return &etcdv1.EtcdadmCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EtcdadmCluster",
			APIVersion: etcdv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-etcd",
			Namespace: "eksa-system",
		},
	}
}
