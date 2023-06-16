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
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestUpdateControlPlaneInitializedCondition(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name          string
		kcp           *controlplanev1.KubeadmControlPlane
		conditions    []anywherev1.Condition
		wantCondition *anywherev1.Condition
		wantErr       string
	}{
		{
			name:       "kcp is nil",
			kcp:        nil,
			conditions: []anywherev1.Condition{},
			wantCondition: &anywherev1.Condition{
				Type:     "ControlPlaneInitialized",
				Status:   "False",
				Severity: clusterv1.ConditionSeverityInfo,
				Reason:   anywherev1.WaitingForControlPlaneInitializedReason,
				Message:  anywherev1.FirstControlPlaneUnavailableMessage,
			},
			wantErr: "",
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
			wantErr: "",
		},
		{
			name:          "kcp missing available condition ",
			kcp:           test.KubeadmControlPlane(),
			wantCondition: nil,
			wantErr:       "unable to read condition",
		},
		{
			name: "kcp available condition with error",
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
			wantErr: "",
		},
		{
			name: "kcp  available condtion false",
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
				Reason:   anywherev1.WaitingForControlPlaneInitializedReason,
				Message:  anywherev1.FirstControlPlaneUnavailableMessage,
			},
			wantErr: "",
		},
		{
			name: "kcp available condition true",
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
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := test.NewClusterSpec().Cluster
			cluster.Status.Conditions = tt.conditions

			err := clusters.UpdateControlPlaneInitializedCondition(cluster, tt.kcp)

			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(err).To(BeNil())

				condition := conditions.Get(cluster, anywherev1.ControlPlaneInitializedCondition)
				g.Expect(condition).ToNot(BeNil())

				condition.LastTransitionTime = v1.Time{}
				g.Expect(condition).To(Equal(tt.wantCondition))
			}
		})
	}
}

func TestUpdateControlPlaneReadyCondition(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name          string
		kcp           *controlplanev1.KubeadmControlPlane
		spec          *cluster.Spec
		wantCondition *anywherev1.Condition
		wantErr       string
	}{
		{
			name: "control plane not initialized",
			kcp:  &controlplanev1.KubeadmControlPlane{},
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:     anywherev1.ControlPlaneInitializedCondition,
						Status:   "False",
						Severity: clusterv1.ConditionSeverityInfo,
						Reason:   anywherev1.WaitingForControlPlaneInitializedReason,
						Message:  anywherev1.FirstControlPlaneUnavailableMessage,
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Severity: clusterv1.ConditionSeverityInfo,
				Reason:   anywherev1.WaitingForControlPlaneInitializedReason,
				Message:  anywherev1.FirstControlPlaneUnavailableMessage,
			},
			wantErr: "",
		},
		{
			name: "control plane initialized but kcp nil",
			kcp:  nil,
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneInitializedCondition,
						Status: "True",
					},
				}
			}),
			wantCondition: nil,
			wantErr:       "expected reference to kubeadmcontrolplane, but got nil instead",
		},
		{
			name: "control plane pending update",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Generation = 1
				kcp.Status.ObservedGeneration = 2
			}),
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneInitializedCondition,
						Status: "True",
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.PendingUpdateReason,
				Severity: clusterv1.ConditionSeverityInfo,
			},
			wantErr: "",
		},
		{
			name: "kcp missing ready condition",
			kcp:  test.KubeadmControlPlane(),
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneInitializedCondition,
						Status: "True",
					},
				}
			}),
			wantCondition: nil,
			wantErr:       "unable to read condition",
		},
		{
			name: "kcp ready condition with error",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:     clusterv1.ReadyCondition,
						Status:   "False",
						Severity: clusterv1.ConditionSeverityError,
						Reason:   "TestReason",
						Message:  "test message",
					},
				}
			}),
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneInitializedCondition,
						Status: "True",
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Severity: clusterv1.ConditionSeverityError,
				Reason:   "TestReason",
				Message:  "test message",
			},
			wantErr: "",
		},
		{
			name: "control plane replicas out of date",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
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
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneInitializedCondition,
						Status: "True",
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.WaitingForControlPlaneReadyReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Control plane nodes not up-to-date yet",
			},
			wantErr: "",
		},
		{
			name: "control plane not ready yet",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.ReadyReplicas = 1
				kcp.Status.UpdatedReplicas = 3
				kcp.Status.Replicas = 3
				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:     clusterv1.ReadyCondition,
						Status:   "False",
						Severity: clusterv1.ConditionSeverityInfo,
					},
				}
			}),
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneInitializedCondition,
						Status: "True",
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.ControlPlaneReadyCondition,
				Status:   "False",
				Reason:   anywherev1.WaitingForControlPlaneReadyReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Control plane nodes not ready yet",
			},
			wantErr: "",
		},
		{
			name: "control plane all ready",
			kcp: test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				kcp.Status.ReadyReplicas = 3
				kcp.Status.Conditions = []clusterv1.Condition{
					{
						Type:   clusterv1.ReadyCondition,
						Status: "True",
					},
				}
			}),
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.ControlPlaneConfiguration.Count = 3
				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneInitializedCondition,
						Status: "True",
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.ControlPlaneReadyCondition,
				Status: "True",
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := tt.spec.Cluster
			err := clusters.UpdateControlPlaneReadyCondition(cluster, tt.kcp)

			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(err).To(BeNil())

				condition := conditions.Get(cluster, anywherev1.ControlPlaneReadyCondition)
				g.Expect(condition).ToNot(BeNil())

				message := condition.Message
				wantMessage := tt.wantCondition.Message

				// Ignoring these fields for comparison sake, they're not too important for testing
				condition.Message = ""
				condition.LastTransitionTime = v1.Time{}
				tt.wantCondition.Message = ""

				g.Expect(condition).To(Equal(tt.wantCondition))
				g.Expect(message).To(ContainSubstring(wantMessage))
			}
		})
	}
}

func TestUpdateWorkersReadyCondition(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name               string
		spec               *cluster.Spec
		machineDeployments []clusterv1.MachineDeployment
		wantCondition      *anywherev1.Condition
		wantErr            string
	}{
		{
			name: "workers pending update",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						Count: ptr.Int(1),
					},
					{
						Count: ptr.Int(1),
					},
				}
			}),
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Generation = 1
					md.Status.ObservedGeneration = 0
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.ObjectMeta.Generation = 1
					md.Status.ObservedGeneration = 1
				}),
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyConditon,
				Status:   "False",
				Reason:   anywherev1.PendingUpdateReason,
				Severity: clusterv1.ConditionSeverityInfo,
			},
		},
		{
			name: "workers not ready yet",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						Count: ptr.Int(1),
					},
					{
						Count: ptr.Int(2),
					},
				}
			}),
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.Status.ReadyReplicas = 1
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.Status.ReadyReplicas = 0
				}),
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyConditon,
				Status:   "False",
				Reason:   anywherev1.WaitingForWorkersReadyReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Worker nodes not ready yet",
			},
		},
		{
			name: "worker out of date ",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						Count: ptr.Int(1),
					},
					{
						Count: ptr.Int(2),
					},
				}
			}),
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.Status.UpdatedReplicas = 1
					md.Status.Replicas = 1
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.Status.Replicas = 3
					md.Status.UpdatedReplicas = 1
				}),
			},
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.WorkersReadyConditon,
				Status:   "False",
				Reason:   anywherev1.WaitingForWorkersReadyReason,
				Severity: clusterv1.ConditionSeverityInfo,
				Message:  "Worker nodes not up-to-date yet",
			},
		},
		{
			name: "workers ready",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
					{
						Count: ptr.Int(1),
					},
					{
						Count: ptr.Int(2),
					},
				}
			}),
			machineDeployments: []clusterv1.MachineDeployment{
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.Status.ReadyReplicas = 1
					md.Status.UpdatedReplicas = 1
				}),
				*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
					md.Status.ReadyReplicas = 2
					md.Status.UpdatedReplicas = 2
				}),
			},
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.WorkersReadyConditon,
				Status: "True",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := tt.spec.Cluster
			err := clusters.UpdateWorkersReadyCondition(cluster, tt.machineDeployments)

			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				condition := conditions.Get(cluster, anywherev1.WorkersReadyConditon)
				g.Expect(condition).ToNot(BeNil())

				message := condition.Message
				wantMessage := tt.wantCondition.Message

				// Ignoring these fields for comparison sake, they're not too important for testing
				condition.Message = ""
				condition.LastTransitionTime = v1.Time{}
				tt.wantCondition.Message = ""

				g.Expect(err).To(BeNil())
				g.Expect(condition).To(Equal(tt.wantCondition))
				g.Expect(message).To(ContainSubstring(wantMessage))
			}
		})
	}
}

func TestUpdateSelfManagedClusterDefaultCNIConfiguredCondition(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		name                    string
		spec                    *cluster.Spec
		wantCondition           *anywherev1.Condition
		checkLastTransitionTime bool
		wantErr                 string
	}{
		{
			name: "no condition updates",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.ClusterNetwork = anywherev1.ClusterNetwork{
					CNIConfig: &anywherev1.CNIConfig{
						Cilium: &anywherev1.CiliumConfig{},
					},
				}

				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneReadyCondition,
						Status: "False",
					},
					{
						Type:               anywherev1.DefaultCNIConfiguredCondition,
						Reason:             anywherev1.WaitingForDefaultCNIConfiguredReason,
						Status:             "False",
						LastTransitionTime: v1.Time{},
						Severity:           clusterv1.ConditionSeverityInfo,
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:               anywherev1.DefaultCNIConfiguredCondition,
				Reason:             anywherev1.WaitingForDefaultCNIConfiguredReason,
				Status:             "False",
				LastTransitionTime: v1.Time{},
				Severity:           clusterv1.ConditionSeverityInfo,
			},
			checkLastTransitionTime: true,
		},
		{
			name: "cilium is unmanaged",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.ClusterNetwork = anywherev1.ClusterNetwork{
					CNIConfig: &anywherev1.CNIConfig{
						Cilium: &anywherev1.CiliumConfig{
							SkipUpgrade: ptr.Bool(true),
						},
					},
				}

				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneReadyCondition,
						Status: "True",
					},
					{
						Type:     anywherev1.DefaultCNIConfiguredCondition,
						Reason:   anywherev1.WaitingForDefaultCNIConfiguredReason,
						Status:   "False",
						Severity: clusterv1.ConditionSeverityInfo,
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:     anywherev1.DefaultCNIConfiguredCondition,
				Reason:   anywherev1.SkipUpgradesForDefaultCNIConfiguredReason,
				Status:   "False",
				Severity: clusterv1.ConditionSeverityWarning,
			},
		},
		{
			name: "cilium is managed",
			spec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Spec.ClusterNetwork = anywherev1.ClusterNetwork{
					CNIConfig: &anywherev1.CNIConfig{
						Cilium: &anywherev1.CiliumConfig{},
					},
				}

				s.Cluster.Status.Conditions = []anywherev1.Condition{
					{
						Type:   anywherev1.ControlPlaneReadyCondition,
						Status: "True",
					},
					{
						Type:     anywherev1.DefaultCNIConfiguredCondition,
						Reason:   anywherev1.WaitingForDefaultCNIConfiguredReason,
						Status:   "False",
						Severity: clusterv1.ConditionSeverityInfo,
					},
				}
			}),
			wantCondition: &anywherev1.Condition{
				Type:   anywherev1.DefaultCNIConfiguredCondition,
				Status: "True",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cluster := tt.spec.Cluster
			clusters.UpdateSelfManagedClusterDefaultCNIConfiguredCondition(cluster)
			condition := conditions.Get(cluster, anywherev1.DefaultCNIConfiguredCondition)
			g.Expect(condition).ToNot(BeNil())

			// Ignoring these fields for comparison sake, they're not too important for testing
			condition.Message = ""

			if !tt.checkLastTransitionTime {
				condition.LastTransitionTime = v1.Time{}
			}
			g.Expect(condition).To(Equal(tt.wantCondition))
		})
	}
}
