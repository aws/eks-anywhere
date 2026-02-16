package clusters_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestCheckControlPlaneReadyItIsReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	kcpVersion := "test"
	kcp := kcpObject(func(k *controlplanev1beta2.KubeadmControlPlane) {
		k.Spec.Version = kcpVersion
		k.Status.Conditions = []metav1.Condition{
			{
				Type:   clusterv1beta2.AvailableCondition,
				Status: metav1.ConditionTrue,
			},
		}
		k.Status.Version = kcpVersion
	})

	client := fake.NewClientBuilder().WithObjects(eksaCluster, kcp).Build()

	result, err := clusters.CheckControlPlaneReady(ctx, client, test.NewNullLogger(), eksaCluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestCheckControlPlaneReadyNoKcp(t *testing.T) {
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

func TestCheckControlPlaneNotReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	kcp := kcpObject(func(k *controlplanev1beta2.KubeadmControlPlane) {
		k.Status = controlplanev1beta2.KubeadmControlPlaneStatus{
			ObservedGeneration: 2,
		}
	})

	client := fake.NewClientBuilder().WithObjects(eksaCluster, kcp).Build()

	result, err := clusters.CheckControlPlaneReady(ctx, client, test.NewNullLogger(), eksaCluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(
		controller.Result{Result: &controllerruntime.Result{RequeueAfter: 5 * time.Second}}),
	)
}

func TestCheckControlPlaneReadyConditionStatusNotReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	kcp := kcpObject(func(k *controlplanev1beta2.KubeadmControlPlane) {
		k.Status.Conditions = []metav1.Condition{
			{
				Type:   clusterv1beta2.ReadyCondition,
				Status: metav1.ConditionFalse,
			},
		}
	})

	client := fake.NewClientBuilder().WithObjects(eksaCluster, kcp).Build()

	result, err := clusters.CheckControlPlaneReady(ctx, client, test.NewNullLogger(), eksaCluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(
		controller.Result{Result: &controllerruntime.Result{RequeueAfter: 30 * time.Second}}),
	)
}

func TestCheckControlPlaneVersionNilStatusNotReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	kcp := kcpObject(func(k *controlplanev1beta2.KubeadmControlPlane) {
		k.Status.Conditions = []metav1.Condition{
			{
				Type:   clusterv1beta2.ReadyCondition,
				Status: metav1.ConditionTrue,
			},
		}
	})

	client := fake.NewClientBuilder().WithObjects(eksaCluster, kcp).Build()

	result, err := clusters.CheckControlPlaneReady(ctx, client, test.NewNullLogger(), eksaCluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(
		controller.Result{Result: &controllerruntime.Result{RequeueAfter: 30 * time.Second}}),
	)
}

func TestCheckControlPlaneVersionMismatchStatusNotReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	kcp := kcpObject(func(k *controlplanev1beta2.KubeadmControlPlane) {
		k.Status.Conditions = []metav1.Condition{
			{
				Type:   clusterv1beta2.ReadyCondition,
				Status: metav1.ConditionTrue,
			},
		}
		k.Status.Version = "test"
	})

	client := fake.NewClientBuilder().WithObjects(eksaCluster, kcp).Build()

	result, err := clusters.CheckControlPlaneReady(ctx, client, test.NewNullLogger(), eksaCluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(result).To(Equal(
		controller.Result{Result: &controllerruntime.Result{RequeueAfter: 30 * time.Second}}),
	)
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

type kcpObjectOpt func(*controlplanev1beta2.KubeadmControlPlane)

func kcpObject(opts ...kcpObjectOpt) *controlplanev1beta2.KubeadmControlPlane {
	k := &controlplanev1beta2.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: controlplanev1beta2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: constants.EksaSystemNamespace,
		},
	}

	for _, opt := range opts {
		opt(k)
	}

	return k
}
