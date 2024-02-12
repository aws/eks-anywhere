package clusters_test

import (
	"context"
	"fmt"
	"testing"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const controlPlaneInitalizationInProgressReason = "The first control plane instance is not available yet"

func TestUpdateClusterStatusForControlPlane(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name                string
		kcp                 *controlplanev1.KubeadmControlPlane
		controlPlaneCount   int
		conditions          []anywherev1.Condition
		wantCondition       *anywherev1.Condition
		externalEtcdCount   int
		externalEtcdCluster *etcdv1.EtcdadmCluster
		capiCluster         *clusterv1.Cluster
		upgradeType         anywherev1.UpgradeRolloutStrategyType
	}{
		{
			name:                "kcp is nil",
			kcp:                 nil,
			controlPlaneCount:   1,
			conditions:          []anywherev1.Condition{},
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
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
			controlPlaneCount: 1,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
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
			controlPlaneCount:   1,
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
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
			controlPlaneCount:   1,
			conditions:          []anywherev1.Condition{},
			externalEtcdCluster: nil,
			externalEtcdCount:   0,
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
			controlPlaneCount:   1,
			conditions:          []anywherev1.Condition{},
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
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
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Severity: clusterv1.ConditionSeverityInfo,
				Reason:   anywherev1.ControlPlaneInitializationInProgressReason,
				Message:  controlPlaneInitalizationInProgressReason,
			},
		},
		{
			name:              "controlplaneready, kcp is nil",
			kcp:               nil,
			controlPlaneCount: 1,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
				{
					Type:   anywherev1.ControlPlaneReadyCondition,
					Status: "True",
				},
			},
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.ControlPlaneReadyCondition,
				Status: "True",
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
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
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
			externalEtcdCount: 0,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			externalEtcdCluster: nil,
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
			externalEtcdCount: 0,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			externalEtcdCluster: nil,
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
			externalEtcdCount: 0,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			externalEtcdCluster: nil,
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.RollingUpgradeInProgress,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Control plane nodes not up-to-date yet",
			},
		},
		{
			name: "control plane replicas out of date, inplace upgrade",
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
			externalEtcdCount: 0,
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			externalEtcdCluster: nil,
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.InPlaceUpgradeInProgress,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Control plane nodes not up-to-date yet",
			},
			upgradeType: anywherev1.InPlaceStrategyType,
		},
		{
			name: "control plane nodes not ready yet",
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
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.NodesNotReadyReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Control plane nodes not ready yet",
			},
		},
		{
			name: "control plane components unhealthy",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Replicas = 3
				kcp.Status.ReadyReplicas = 3
				kcp.Status.UpdatedReplicas = 3

				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:   clusterv1.ReadyCondition,
						Status: "True",
					},
					{
						Type:     controlplanev1.ControlPlaneComponentsHealthyCondition,
						Reason:   controlplanev1.ControlPlaneComponentsUnhealthyReason,
						Severity: clusterv1.ConditionSeverityError,
						Message:  "test message",
						Status:   "False",
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
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Reason:   anywherev1.ControlPlaneComponentsUnhealthyReason,
				Severity: clusterv1.ConditionSeverityError,
				Message:  "test message",
				Status:   "False",
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
			externalEtcdCount:   0,
			externalEtcdCluster: nil,
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.ControlPlaneReadyCondition,
				Status: "True",
			},
		},
		{
			name: "with external etcd ready",
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
			externalEtcdCount: 1,
			externalEtcdCluster: &etcdv1.EtcdadmCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-cluster-etcd",
					Namespace:  constants.EksaSystemNamespace,
					Generation: 2,
				},
				Status: etcdv1.EtcdadmClusterStatus{
					Ready:              true,
					ObservedGeneration: 2,
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.ControlPlaneReadyCondition,
				Status: "True",
			},
			capiCluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: clusterv1.ClusterSpec{
					ManagedExternalEtcdRef: &corev1.ObjectReference{
						Kind: "EtcdadmCluster",
						Name: fmt.Sprintf("%s-etcd", "test-cluster"),
					},
				},
			},
		},
		{
			name: "with external etcd not ready",
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
			externalEtcdCount: 1,
			externalEtcdCluster: &etcdv1.EtcdadmCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-cluster-etcd",
					Namespace:  constants.EksaSystemNamespace,
					Generation: 2,
				},
				Status: etcdv1.EtcdadmClusterStatus{
					ObservedGeneration: 2,
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Reason:   anywherev1.RollingUpgradeInProgress,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Etcd is not ready",
				Status:   "False",
			},
			capiCluster: &clusterv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: clusterv1.ClusterSpec{
					ManagedExternalEtcdRef: &corev1.ObjectReference{
						Kind: "EtcdadmCluster",
						Name: fmt.Sprintf("%s-etcd", "test-cluster"),
					},
				},
			},
		},
		{
			name: "with external etcd, etcd not reconciled",
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
			externalEtcdCount:   1,
			externalEtcdCluster: &etcdv1.EtcdadmCluster{},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Reason:   anywherev1.ExternalEtcdNotAvailable,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Etcd cluster is not available",
				Status:   "False",
			},
			capiCluster: &clusterv1.Cluster{},
		},
		{
			name: "with external etcd, malformed etcd",
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
			externalEtcdCount:   1,
			externalEtcdCluster: &etcdv1.EtcdadmCluster{},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Reason:   anywherev1.ExternalEtcdNotAvailable,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Etcd cluster is not available",
				Status:   "False",
			},
			capiCluster: &clusterv1.Cluster{},
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
			if tt.externalEtcdCount > 0 {
				cluster.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{
					Count: tt.externalEtcdCount,
					MachineGroupRef: &anywherev1.Ref{
						Name: fmt.Sprintf("%s-etcd", cluster.Name),
					},
				}

				tt.externalEtcdCluster.Name = fmt.Sprintf("%s-etcd", cluster.Name)
				tt.externalEtcdCluster.Namespace = cluster.Namespace

				objs = append(objs, tt.capiCluster)
				objs = append(objs, tt.externalEtcdCluster)
			}
			if tt.upgradeType != "" {
				cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &anywherev1.ControlPlaneUpgradeRolloutStrategy{
					Type: tt.upgradeType,
				}
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

func TestUpdateClusterStatusForControlPlaneError(t *testing.T) {
	g := NewWithT(t)
	cluster := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
				Count: 1,
			},
		},
	}
	capiCluster := &clusterv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: clusterv1.ClusterSpec{
			ManagedExternalEtcdRef: &corev1.ObjectReference{
				Kind: "EtcdadmCluster",
				Name: fmt.Sprintf("%s-etcd", "test-cluster"),
			},
		},
	}
	objs := []runtime.Object{}
	objs = append(objs, capiCluster)
	objs = append(objs, &controlplanev1.KubeadmControlPlane{})
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	err := clusters.UpdateClusterStatusForControlPlane(context.Background(), client, cluster)
	g.Expect(err).NotTo(BeNil())
}

func TestUpdateClusterStatusForWorkers(t *testing.T) {
	cluster := test.NewClusterSpec().Cluster
	clusterName := "test-cluster"
	g := NewWithT(t)

	tests := []struct {
		name                          string
		machineDeployments            []clusterv1.MachineDeployment
		workerNodeGroupConfigurations []anywherev1.WorkerNodeGroupConfiguration
		conditions                    []anywherev1.Condition
		wantCondition                 *anywherev1.Condition
		wantErr                       string
		upgradeType                   anywherev1.UpgradeRolloutStrategyType
	}{
		{
			name:                          "workers not ready, control plane not initialized",
			workerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{},
			machineDeployments:            []clusterv1.MachineDeployment{},
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
				Type:     anywherev1.WorkersReadyCondition,
				Status:   "False",
				Reason:   anywherev1.ControlPlaneNotInitializedReason,
				Severity: clusterv1.ConditionSeverityInfo,
			},
		},
		{
			name: "workers not ready, outdated information, one group",
			workerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
				},
			},
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-0"
					md.ObjectMeta.Generation = 1
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.ObservedGeneration = 0
				}),
			},
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyCondition,
				Status:   "False",
				Reason:   anywherev1.OutdatedInformationReason,
				Severity: clusterv1.ConditionSeverityInfo,
			},
		},
		{
			name: "workers not ready, outdated information, two groups",
			workerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
				},
				{
					Count: ptr.Int(1),
				},
			},
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-0"
					md.ObjectMeta.Generation = 1
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.ObservedGeneration = 0
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-1"
					md.ObjectMeta.Generation = 1
					md.Status.ObservedGeneration = 1
				}),
			},
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyCondition,
				Status:   "False",
				Reason:   anywherev1.OutdatedInformationReason,
				Severity: clusterv1.ConditionSeverityInfo,
			},
		},
		{
			name: "workers not ready, nodes not up to date ",
			workerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
				},
				{
					Count: ptr.Int(2),
				},
			},
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-0"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 1
					md.Status.ReadyReplicas = 1
					md.Status.UpdatedReplicas = 1
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-1"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 2
					md.Status.ReadyReplicas = 2
					md.Status.UpdatedReplicas = 1
				}),
			},
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyCondition,
				Status:   "False",
				Reason:   anywherev1.RollingUpgradeInProgress,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Worker nodes not up-to-date yet",
			},
		},
		{
			name: "workers not ready, nodes not up to date ",
			workerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
				},
				{
					Count: ptr.Int(2),
				},
			},
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-0"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 1
					md.Status.ReadyReplicas = 1
					md.Status.UpdatedReplicas = 1
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-1"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 2
					md.Status.ReadyReplicas = 2
					md.Status.UpdatedReplicas = 1
				}),
			},
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyCondition,
				Status:   "False",
				Reason:   anywherev1.InPlaceUpgradeInProgress,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Worker nodes not up-to-date yet",
			},
			upgradeType: anywherev1.InPlaceStrategyType,
		},
		{
			name: "workers not ready, scaling up",
			workerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
				},
				{
					Count: ptr.Int(2),
				},
			},
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-0"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 0
					md.Status.ReadyReplicas = 0
					md.Status.UpdatedReplicas = 0
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-1"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 2
					md.Status.ReadyReplicas = 2
					md.Status.UpdatedReplicas = 2
				}),
			},
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyCondition,
				Status:   "False",
				Reason:   anywherev1.ScalingUpReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Scaling up worker nodes",
			},
		},
		{
			name: "workers not ready, scaling down",
			workerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(2),
				},
				{
					Count: ptr.Int(1),
				},
			},
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-0"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 2
					md.Status.ReadyReplicas = 2
					md.Status.UpdatedReplicas = 2
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-1"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 2
					md.Status.ReadyReplicas = 2
					md.Status.UpdatedReplicas = 2
				}),
			},
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyCondition,
				Status:   "False",
				Reason:   anywherev1.ScalingDownReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Scaling down worker nodes",
			},
		},
		{
			name: "workers not ready, nodes not ready yet",
			workerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
				},
				{
					Count: ptr.Int(2),
				},
			},
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-0"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.ReadyReplicas = 1
					md.Status.Replicas = 1
					md.Status.UpdatedReplicas = 1
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-1"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.ReadyReplicas = 0
					md.Status.Replicas = 2
					md.Status.UpdatedReplicas = 2
				}),
			},
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyCondition,
				Status:   "False",
				Reason:   anywherev1.NodesNotReadyReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Worker nodes not ready yet",
			},
		},

		{
			name: "workers ready",
			workerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
				},
				{
					Count: ptr.Int(2),
				},
			},
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-0"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 1
					md.Status.ReadyReplicas = 1
					md.Status.UpdatedReplicas = 1
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Name = "md-1"
					md.ObjectMeta.Labels = map[string]string{
						clusterv1.ClusterNameLabel: clusterName,
					}
					md.Status.Replicas = 2
					md.Status.ReadyReplicas = 2
					md.Status.UpdatedReplicas = 2
				}),
			},
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneInitializedCondition,
					Status: "True",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.WorkersReadyCondition,
				Status: "True",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cluster.Name = clusterName
			cluster.Namespace = constants.EksaSystemNamespace
			cluster.Spec.WorkerNodeGroupConfigurations = tt.workerNodeGroupConfigurations
			cluster.Status.Conditions = tt.conditions

			if tt.upgradeType != "" {
				cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &anywherev1.ControlPlaneUpgradeRolloutStrategy{
					Type: tt.upgradeType,
				}
			}

			objs := []runtime.Object{}

			var client client.Client
			for _, md := range tt.machineDeployments {
				objs = append(objs, md.DeepCopy())
			}

			client = fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

			err := clusters.UpdateClusterStatusForWorkers(ctx, client, cluster)
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

func TestUpdateClusterStatusForCNI(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name          string
		spec          *cluster.Spec
		conditions    []anywherev1.Condition
		wantCondition *anywherev1.Condition
		skipUpgrade   *bool
		wantErr       string
	}{
		{
			name:        "control plane is not ready",
			skipUpgrade: ptr.Bool(false),
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ControlPlaneReadyCondition,
					Status: "False",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.DefaultCNIConfiguredCondition,
				Reason:   anywherev1.ControlPlaneNotReadyReason,
				Status:   "False",
				Severity: clusterv1.ConditionSeverityInfo,
			},
		},
		{
			name:        "control plane is not ready, default cni initialized",
			skipUpgrade: ptr.Bool(false),
			conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.DefaultCNIConfiguredCondition,
					Status: "True",
				},
				{
					Type:   anywherev1.ControlPlaneReadyCondition,
					Status: "False",
				},
			},
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.DefaultCNIConfiguredCondition,
				Status: "True",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			spec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = "management-cluster"
				s.Cluster.Spec.ClusterNetwork = anywherev1.ClusterNetwork{
					CNIConfig: &anywherev1.CNIConfig{
						Cilium: &anywherev1.CiliumConfig{
							SkipUpgrade: tt.skipUpgrade,
						},
					},
				}

				s.Cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}
				s.Cluster.Status.Conditions = tt.conditions
			})

			clusters.UpdateClusterStatusForCNI(ctx, spec.Cluster)

			if tt.wantCondition != nil {
				condition := conditions.Get(spec.Cluster, tt.wantCondition.Type)
				g.Expect(condition).ToNot(BeNil())
				g.Expect(condition.Type).To(Equal(tt.wantCondition.Type))
				g.Expect(condition.Severity).To(Equal(tt.wantCondition.Severity))
				g.Expect(condition.Status).To(Equal(tt.wantCondition.Status))
				g.Expect(condition.Reason).To(Equal(tt.wantCondition.Reason))
				g.Expect(condition.Message).To(ContainSubstring(tt.wantCondition.Message))
			}
		})
	}
}
