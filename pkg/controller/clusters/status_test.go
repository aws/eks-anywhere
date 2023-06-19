package clusters_test

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"

	"github.com/aws/eks-anywhere/internal/test"
	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestUpdateControlPlaneInitializedCondition(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name          string
		kcp           *controlplanev1.KubeadmControlPlane
		conditions    []anywherev1.Condition
		wantCondition *anywherev1.Condition
	}{
		{
			name:       "kcp is nil",
			kcp:        nil,
			conditions: []anywherev1.Condition{},
			wantCondition: &anywherev1.Condition{
				Type:     "ControlPlaneInitialized",
				Status:   "False",
				Severity: clusterv1.ConditionSeverityInfo,
				Reason:   anywherev1.ControlPlaneInitializationInProgressReason,
				Message:  anywherev1.FirstControlPlaneUnavailableMessage,
			},
		},
		{
			name: "control plane already initialized",
			kcp:  nil,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.ControlPlaneInitializedCondition,
				Status: "True",
			},
		},
		{
			name: "kcp status outdated, generations do not match",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.ObjectMeta.Generation = 1
				kcp.Status.ObservedGeneration = 0
			}),
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneInitializedCondition,
				Status:   "False",
				Severity: clusterv1.ConditionSeverityInfo,
				Reason:   anywherev1.OutdatedInformationReason,
				Message:  "",
			},
		},
		{
			name: "kcp not available, condition error",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:     controlplanev1.AvailableCondition,
						Status:   "False",
						Severity: clusterv1.ConditionSeverityError,
						Reason:   "TestReason",
						Message:  "test message",
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneInitializedCondition,
				Status:   "False",
				Severity: clusterv1.ConditionSeverityError,
				Reason:   "TestReason",
				Message:  "test message",
			},
		},
		{
			name: "kcp not availabe yet",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Conditions = clusterv1.Conditions{
					{
						Type:   controlplanev1.AvailableCondition,
						Status: "False",
					},
				}
			}),
			conditions: []anywherev1.Condition{},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneInitializedCondition,
				Status:   "False",
				Severity: clusterv1.ConditionSeverityInfo,
				Reason:   anywherev1.ControlPlaneInitializationInProgressReason,
				Message:  anywherev1.FirstControlPlaneUnavailableMessage,
			},
		},
		{
			name: "kcp available",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Conditions = clusterv1.Conditions{
					{
						Type:   controlplanev1.AvailableCondition,
						Status: "True",
					},
				}
			}),
			conditions: []anywherev1.Condition{},
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.ControlPlaneInitializedCondition,
				Status: "True",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := test.NewClusterSpec().Cluster
			cluster.Status.Conditions = tt.conditions

			clusters.UpdateControlPlaneInitializedCondition(cluster, tt.kcp)

			condition := conditions.Get(cluster, anywherev1.ControlPlaneInitializedCondition)
			g.Expect(condition).ToNot(BeNil())

			condition.LastTransitionTime = v1.Time{}
			g.Expect(condition).To(Equal(tt.wantCondition))
		})
	}
}
