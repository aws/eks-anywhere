package clusters_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

const controlPlaneInitalizationInProgressReason = "The first control plane instance is not available yet"

func TestUpdateClusterStatusForControlPlane(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name              string
		kcp               *controlplanev1.KubeadmControlPlane
		controlPlaneCount int
		conditions        []anywherev1.Condition
		wantCondition     *anywherev1.Condition
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
				Message:  controlPlaneInitalizationInProgressReason,
			},
		},
		{
			name: "control plane already initialized",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Conditions = clusterv1.Conditions{
					{
						Type:   controlplanev1.AvailableCondition,
						Status: "True",
					},
				}
			}),
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
				Message:  controlPlaneInitalizationInProgressReason,
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
		{
			name:              "control plane not initialized",
			kcp:               &controlplanev1.KubeadmControlPlane{},
			controlPlaneCount: 1,
			conditions: []anywherev1.Condition{
				{
					Type:     anywherev1.ControlPlaneInitializedCondition,
					Status:   "False",
					Severity: clusterv1.ConditionSeverityInfo,
					Reason:   anywherev1.ControlPlaneInitializationInProgressReason,
					Message:  controlPlaneInitalizationInProgressReason,
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Severity: clusterv1.ConditionSeverityInfo,
				Reason:   anywherev1.ControlPlaneInitializationInProgressReason,
				Message:  controlPlaneInitalizationInProgressReason,
			},
		},
		{
			name: "kubeadmcontrolplane status out of date",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Generation = 1
				kcp.Status.ObservedGeneration = 2
			}),
			controlPlaneCount: 1,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.OutdatedInformationReason,
				Severity: clusterv1.ConditionSeverityInfo,
			},
		},
		{
			name: "scaling up control plane nodes",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Replicas = 1
				kcp.Status.UpdatedReplicas = 1
				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:     clusterv1.ReadyCondition,
						Status:   "False",
						Severity: clusterv1.ConditionSeverityInfo,
					},
				}
			}),
			controlPlaneCount: 3,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.ScalingUpReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Scaling up control plane nodes, 3 expected (1 actual)",
			},
		},
		{
			name: "scaling down control plane nodes",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Replicas = 3
				kcp.Status.UpdatedReplicas = 3

				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:     clusterv1.ReadyCondition,
						Status:   "False",
						Severity: clusterv1.ConditionSeverityInfo,
					},
				}
			}),
			controlPlaneCount: 1,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.ScalingDownReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Scaling down control plane nodes",
			},
		},
		{
			name: "control plane replicas out of date",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.ReadyReplicas = 3
				kcp.Status.Replicas = 3
				kcp.Status.UpdatedReplicas = 1

				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:     clusterv1.ReadyCondition,
						Status:   "False",
						Severity: clusterv1.ConditionSeverityInfo,
					},
				}
			}),
			controlPlaneCount: 3,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.RollingUpgradeInProgress,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Control plane nodes not up-to-date yet",
			},
		},
		{
			name: "control plane nodes provisioning in progress",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Replicas = 3
				kcp.Status.ReadyReplicas = 2
				kcp.Status.UpdatedReplicas = 3

				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:   clusterv1.ReadyCondition,
						Status: "True",
					},
				}
			}),
			controlPlaneCount: 3,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.NodesNotReadyReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Control plane nodes not ready yet",
			},
		},
		{
			name: "control plane ready",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Replicas = 3
				kcp.Status.ReadyReplicas = 3
				kcp.Status.UpdatedReplicas = 3

				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:   clusterv1.ReadyCondition,
						Status: "True",
					},
				}
			}),
			controlPlaneCount: 3,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.ControlPlaneReadyCondition,
				Status: "True",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cluster := test.NewClusterSpec().Cluster
			cluster.Name = "test-cluster"
			cluster.Namespace = constants.EksaSystemNamespace
			cluster.Spec.ControlPlaneConfiguration.Count = tt.controlPlaneCount
			cluster.Status.Conditions = tt.conditions

			objs := []runtime.Object{}

			var client client.Client
			if tt.kcp != nil {
				tt.kcp.Name = cluster.Name
				tt.kcp.Namespace = cluster.Namespace
				objs = append(objs, tt.kcp)
			}

			client = fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

			err := clusters.UpdateClusterStatusForControlPlane(ctx, client, cluster)
			g.Expect(err).To(BeNil())

			condition := conditions.Get(cluster, tt.wantCondition.Type)
			g.Expect(condition).ToNot(BeNil())

			g.Expect(condition.Type).To(Equal(tt.wantCondition.Type))
			g.Expect(condition.Severity).To(Equal(tt.wantCondition.Severity))
			g.Expect(condition.Status).To(Equal(tt.wantCondition.Status))
			g.Expect(condition.Reason).To(Equal(tt.wantCondition.Reason))
			g.Expect(condition.Message).To(ContainSubstring(tt.wantCondition.Message))
		})
	}
}
