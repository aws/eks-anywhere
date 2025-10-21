package clusters_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestCheckControlPlaneReadyItIsReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	eksaCluster := eksaCluster()
	kcpVersion := "test"
	kcp := kcpObject(func(k *v1beta1.KubeadmControlPlane) {
		k.Spec.Version = kcpVersion
		k.Status.Conditions = clusterv1.Conditions{
			{
				Type:   clusterapi.ReadyCondition,
				Status: corev1.ConditionTrue,
			},
		}
		k.Status.Version = pointer.String(kcpVersion)
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
	kcp := kcpObject(func(k *v1beta1.KubeadmControlPlane) {
		k.Status = v1beta1.KubeadmControlPlaneStatus{
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
	kcp := kcpObject(func(k *v1beta1.KubeadmControlPlane) {
		k.Status.Conditions = clusterv1.Conditions{
			{
				Type:   clusterapi.ReadyCondition,
				Status: corev1.ConditionFalse,
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
	kcp := kcpObject(func(k *v1beta1.KubeadmControlPlane) {
		k.Status.Conditions = clusterv1.Conditions{
			{
				Type:   clusterapi.ReadyCondition,
				Status: corev1.ConditionTrue,
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
	kcp := kcpObject(func(k *v1beta1.KubeadmControlPlane) {
		k.Status.Conditions = clusterv1.Conditions{
			{
				Type:   clusterapi.ReadyCondition,
				Status: corev1.ConditionTrue,
			},
		}
		k.Status.Version = pointer.String("test")
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

type kcpObjectOpt func(*v1beta1.KubeadmControlPlane)

func kcpObject(opts ...kcpObjectOpt) *v1beta1.KubeadmControlPlane {
	k := &v1beta1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: v1beta1.GroupVersion.String(),
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
