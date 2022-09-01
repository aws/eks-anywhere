package clusters_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestCheckControlPlaneReadyItIsReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	capiCluster := capiCluster(func(c *clusterv1.Cluster) {
		c.Status.Conditions = clusterv1.Conditions{
			{
				Type:   clusterapi.ControlPlaneReadyCondition,
				Status: corev1.ConditionTrue,
			},
		}
	})

	client := fake.NewClientBuilder().WithObjects(eksaCluster, capiCluster).Build()

	result, err := clusters.CheckControlPlaneReady(ctx, client, test.NewNullLogger(), eksaCluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestCheckControlPlaneReadyNoCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()

	client := fake.NewClientBuilder().WithObjects(eksaCluster).Build()

	result, err := clusters.CheckControlPlaneReady(ctx, client, test.NewNullLogger(), eksaCluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(
		controller.Result{Result: &controllerruntime.Result{RequeueAfter: 5 * time.Second}}),
	)
}

func TestCheckControlPlaneReadyNotReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	capiCluster := capiCluster()

	client := fake.NewClientBuilder().WithObjects(eksaCluster, capiCluster).Build()

	result, err := clusters.CheckControlPlaneReady(ctx, client, test.NewNullLogger(), eksaCluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(
		controller.Result{Result: &controllerruntime.Result{RequeueAfter: 30 * time.Second}}),
	)
}

func TestCheckControlPlaneReadyErrorReading(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()

	// This should make the client fail because CRDs are not registered
	client := fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()

	_, err := clusters.CheckControlPlaneReady(ctx, client, test.NewNullLogger(), eksaCluster)
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

type capiClusterOpt func(*clusterv1.Cluster)

func capiCluster(opts ...capiClusterOpt) *clusterv1.Cluster {
	c := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
