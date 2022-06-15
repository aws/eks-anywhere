package clients_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/controllers/controllers/clients"
	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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

	client := clients.NewKubeClient(cl)
	receiveCluster := &anywherev1.Cluster{}
	g.Expect(client.Get(ctx, "my-cluster", "default", receiveCluster)).To(Succeed())
	g.Expect(receiveCluster).To(Equal(cluster))
}

func TestKubeClientGetNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects().Build()

	client := clients.NewKubeClient(cl)
	receiveCluster := &anywherev1.Cluster{}
	g.Expect(client.Get(ctx, "my-cluster", "default", receiveCluster)).Error()
}
