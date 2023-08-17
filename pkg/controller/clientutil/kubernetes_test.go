package clientutil_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test/envtest"
	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

func TestKubeClientGet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
	}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(cluster).Build()

	client := clientutil.NewKubeClient(cl)
	receiveCluster := &anywherev1.Cluster{}
	g.Expect(client.Get(ctx, "my-cluster", "default", receiveCluster)).To(Succeed())
	g.Expect(receiveCluster).To(Equal(cluster))
}

func TestKubeClientGetNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects().Build()

	client := clientutil.NewKubeClient(cl)
	receiveCluster := &anywherev1.Cluster{}
	g.Expect(client.Get(ctx, "my-cluster", "default", receiveCluster)).Error()
}

func TestKubeClientList(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster1 := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
	}
	cluster2 := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-2",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
	}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(cluster1, cluster2).Build()

	client := clientutil.NewKubeClient(cl)
	receiveClusters := &anywherev1.ClusterList{}
	g.Expect(client.List(ctx, receiveClusters)).To(Succeed())
	g.Expect(receiveClusters.Items).To(ConsistOf(*cluster1, *cluster2))
}

func TestKubeClientCreate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
	}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects().Build()

	client := clientutil.NewKubeClient(cl)
	g.Expect(client.Create(ctx, cluster)).To(Succeed())

	api := envtest.NewAPIExpecter(t, cl)
	api.ShouldEventuallyExist(ctx, cluster)
}

func TestKubeClientUpdate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
	}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(cluster).Build()

	client := clientutil.NewKubeClient(cl)

	updatedCluster := cluster.DeepCopy()
	updatedCluster.Spec.KubernetesVersion = anywherev1.Kube126
	g.Expect(client.Update(ctx, updatedCluster)).To(Succeed())

	api := envtest.NewAPIExpecter(t, cl)
	api.ShouldEventuallyMatch(ctx, cluster, func(g Gomega) {
		g.Expect(cluster).To(BeComparableTo(updatedCluster))
	})
}

func TestKubeClientDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
	}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(cluster).Build()

	client := clientutil.NewKubeClient(cl)

	g.Expect(client.Delete(ctx, cluster)).To(Succeed())

	api := envtest.NewAPIExpecter(t, cl)
	api.ShouldEventuallyNotExist(ctx, cluster)
}

func TestKubeClientDeleteAllOf(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster1 := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
	}
	cluster2 := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-2",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
	}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(cluster1, cluster2).Build()

	client := clientutil.NewKubeClient(cl)
	opts := &kubernetes.DeleteAllOfOptions{
		Namespace: "default",
	}
	g.Expect(client.DeleteAllOf(ctx, &anywherev1.Cluster{}, opts)).To(Succeed())

	api := envtest.NewAPIExpecter(t, cl)
	api.ShouldEventuallyNotExist(ctx, cluster1)
	api.ShouldEventuallyNotExist(ctx, cluster2)
}

func TestKubeClientApplyServerSide(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	fieldManager := "my-manager"
	cluster1 := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
	}

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(cluster1).Build()

	v := anywherev1.EksaVersion("v0.15.0")
	cluster1.Spec.EksaVersion = &v

	client := clientutil.NewKubeClient(cl)
	opts := kubernetes.ApplyServerSideOptions{
		ForceOwnership: true,
	}
	g.Expect(client.ApplyServerSide(ctx, fieldManager, cluster1, opts)).To(Succeed())

	c := envtest.CloneNameNamespace(cluster1)
	api := envtest.NewAPIExpecter(t, cl)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Spec.EksaVersion).To(Equal(cluster1.Spec.EksaVersion))
	})
}
