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
			{
				Type:   controlplanev1beta2.KubeadmControlPlaneMachinesUpToDateCondition,
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

func TestCheckControlPlaneReadyMachineStatus(t *testing.T) {
	tests := []struct {
		name                string
		reportedVersion     string
		machineStatus       metav1.ConditionStatus
		conditionGeneration int64
		requeueAfter        time.Duration
	}{
		{
			name:                "EKS-D kubelet hash is up to date",
			reportedVersion:     "v1.37.1-eks-abcdef0",
			machineStatus:       metav1.ConditionTrue,
			conditionGeneration: 2,
		},
		{
			name:                "matching version is up to date",
			reportedVersion:     "v1.37.1-eks-1-37-4",
			machineStatus:       metav1.ConditionTrue,
			conditionGeneration: 2,
		},
		{
			name:                "matching version still has a rollout pending",
			reportedVersion:     "v1.37.1-eks-1-37-4",
			machineStatus:       metav1.ConditionFalse,
			conditionGeneration: 2,
			requeueAfter:        30 * time.Second,
		},
		{
			name:                "kubelet version has not caught up",
			reportedVersion:     "v1.36.2-eks-abcdef0",
			machineStatus:       metav1.ConditionFalse,
			conditionGeneration: 2,
			requeueAfter:        30 * time.Second,
		},
		{
			name:                "machine status is unknown",
			reportedVersion:     "v1.37.1-eks-1-37-4",
			machineStatus:       metav1.ConditionUnknown,
			conditionGeneration: 2,
			requeueAfter:        30 * time.Second,
		},
		{
			name:            "machine status is missing",
			reportedVersion: "v1.37.1-eks-1-37-4",
			requeueAfter:    30 * time.Second,
		},
		{
			name:                "machine status predates the desired upgrade",
			reportedVersion:     "v1.37.1-eks-1-37-4",
			machineStatus:       metav1.ConditionTrue,
			conditionGeneration: 1,
			requeueAfter:        30 * time.Second,
		},
		{
			name:                "reported version is missing",
			machineStatus:       metav1.ConditionTrue,
			conditionGeneration: 2,
			requeueAfter:        30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			eksaCluster := eksaCluster()
			kcp := kcpObject(func(k *controlplanev1beta2.KubeadmControlPlane) {
				k.Generation = 2
				k.Spec.Version = "v1.37.1-eks-1-37-4"
				k.Status.ObservedGeneration = 2
				k.Status.Version = tt.reportedVersion
				k.Status.Conditions = []metav1.Condition{
					{
						Type:               clusterv1beta2.AvailableCondition,
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 2,
					},
				}
				if tt.machineStatus != "" {
					k.Status.Conditions = append(k.Status.Conditions, metav1.Condition{
						Type:               controlplanev1beta2.KubeadmControlPlaneMachinesUpToDateCondition,
						Status:             tt.machineStatus,
						ObservedGeneration: tt.conditionGeneration,
					})
				}
			})
			client := fake.NewClientBuilder().WithObjects(eksaCluster, kcp).Build()

			result, err := clusters.CheckControlPlaneReady(context.Background(), client, test.NewNullLogger(), eksaCluster)

			g.Expect(err).NotTo(HaveOccurred())
			expected := controller.Result{}
			if tt.requeueAfter != 0 {
				expected = controller.ResultWithRequeue(tt.requeueAfter)
			}
			g.Expect(result).To(Equal(expected))
		})
	}
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
				Type:   clusterv1beta2.AvailableCondition,
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
				Type:   clusterv1beta2.AvailableCondition,
				Status: metav1.ConditionTrue,
			},
			{
				Type:   controlplanev1beta2.KubeadmControlPlaneMachinesUpToDateCondition,
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
				Type:   clusterv1beta2.AvailableCondition,
				Status: metav1.ConditionTrue,
			},
			{
				Type:   controlplanev1beta2.KubeadmControlPlaneMachinesUpToDateCondition,
				Status: metav1.ConditionFalse,
			},
		}
		k.Spec.Version = "v1.37.1-eks-1-37-4"
		k.Status.Version = "v1.36.2-eks-abcdef0"
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
