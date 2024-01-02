package controllers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/govmomi"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	vspheremocks "github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
	vspherereconciler "github.com/aws/eks-anywhere/pkg/providers/vsphere/reconciler"
	vspherereconcilermocks "github.com/aws/eks-anywhere/pkg/providers/vsphere/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var clusterName = "test-cluster"

var controlPlaneInitalizationInProgressReason = "The first control plane instance is not available yet"

type vsphereClusterReconcilerTest struct {
	govcClient *vspheremocks.MockProviderGovcClient
	reconciler *controllers.ClusterReconciler
	client     client.Client
}

func testKubeadmControlPlaneFromCluster(cluster *anywherev1.Cluster) *controlplanev1.KubeadmControlPlane {
	k := controller.CAPIKubeadmControlPlaneKey(cluster)
	return test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
		kcp.Name = k.Name
		kcp.Namespace = k.Namespace
		expectedReplicas := int32(cluster.Spec.ControlPlaneConfiguration.Count)

		kcp.Status = controlplanev1.KubeadmControlPlaneStatus{
			Replicas:        expectedReplicas,
			UpdatedReplicas: expectedReplicas,
			ReadyReplicas:   expectedReplicas,
			Conditions: clusterv1.Conditions{
				{
					Type:   controlplanev1.ControlPlaneComponentsHealthyCondition,
					Status: apiv1.ConditionStatus("True"),
				},
				{
					Type:   controlplanev1.AvailableCondition,
					Status: apiv1.ConditionStatus("True"),
				},
				{
					Type:   clusterv1.ReadyCondition,
					Status: apiv1.ConditionStatus("True"),
				},
			},
		}
	})
}

func machineDeploymentsFromCluster(cluster *anywherev1.Cluster) []clusterv1.MachineDeployment {
	return []clusterv1.MachineDeployment{
		*test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
			md.ObjectMeta.Name = "md-0"
			md.ObjectMeta.Labels = map[string]string{
				clusterv1.ClusterNameLabel: cluster.Name,
			}
			md.Status.Replicas = 1
			md.Status.ReadyReplicas = 1
			md.Status.UpdatedReplicas = 1
		}),
	}
}

func newVsphereClusterReconcilerTest(t *testing.T, objs ...runtime.Object) *vsphereClusterReconcilerTest {
	ctrl := gomock.NewController(t)
	govcClient := vspheremocks.NewMockProviderGovcClient(ctrl)

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	clusterValidator := mocks.NewMockClusterValidator(ctrl)

	vcb := govmomi.NewVMOMIClientBuilder()

	validator := vsphere.NewValidator(govcClient, vcb)
	defaulter := vsphere.NewDefaulter(govcClient)
	cniReconciler := vspherereconcilermocks.NewMockCNIReconciler(ctrl)
	ipValidator := vspherereconcilermocks.NewMockIPValidator(ctrl)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(ctrl)

	reconciler := vspherereconciler.New(
		cl,
		validator,
		defaulter,
		cniReconciler,
		nil,
		ipValidator,
	)
	registry := clusters.NewProviderClusterReconcilerRegistryBuilder().
		Add(anywherev1.VSphereDatacenterKind, reconciler).
		Build()

	mockPkgs := mocks.NewMockPackagesClient(ctrl)
	mockPkgs.EXPECT().
		ReconcileDelete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	r := controllers.NewClusterReconciler(cl, &registry, iam, clusterValidator, mockPkgs, mhcReconciler)

	return &vsphereClusterReconcilerTest{
		govcClient: govcClient,
		reconciler: r,
		client:     cl,
	}
}

func TestClusterReconcilerReconcileSelfManagedCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	selfManagedCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-management-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			BundlesRef: &anywherev1.BundlesRef{
				Name: "my-bundles-ref",
			},
			EksaVersion: &version,
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				},
			},
			MachineHealthCheck: &anywherev1.MachineHealthCheck{
				UnhealthyMachineTimeout: &metav1.Duration{
					Duration: constants.DefaultUnhealthyMachineTimeout,
				},
				NodeStartupTimeout: &metav1.Duration{
					Duration: constants.DefaultNodeStartupTimeout,
				},
			},
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	kcp := testKubeadmControlPlaneFromCluster(selfManagedCluster)

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(controller)

	clusterValidator := mocks.NewMockClusterValidator(controller)
	registry := newRegistryMock(providerReconciler)
	c := fake.NewClientBuilder().WithRuntimeObjects(selfManagedCluster, kcp, test.EKSARelease()).Build()
	mockPkgs := mocks.NewMockPackagesClient(controller)
	providerReconciler.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(selfManagedCluster))
	mhcReconciler.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(selfManagedCluster)).Return(nil)

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, mockPkgs, mhcReconciler)
	result, err := r.Reconcile(ctx, clusterRequest(selfManagedCluster))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))
}

func TestClusterReconcilerReconcileConditions(t *testing.T) {
	testCases := []struct {
		testName                string
		skipCNIUpgrade          bool
		cniUpgradeInProgress    bool
		kcpStatus               controlplanev1.KubeadmControlPlaneStatus
		machineDeploymentStatus clusterv1.MachineDeploymentStatus
		result                  ctrl.Result
		wantConditions          []anywherev1.Condition
	}{
		{
			testName:             "cluster not ready, control plane not initialized",
			skipCNIUpgrade:       false,
			cniUpgradeInProgress: false,
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("False"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{},
			wantConditions: []anywherev1.Condition{
				*conditions.FalseCondition(anywherev1.ControlPlaneInitializedCondition, anywherev1.ControlPlaneInitializationInProgressReason, clusterv1.ConditionSeverityInfo, controlPlaneInitalizationInProgressReason),
				*conditions.FalseCondition(anywherev1.ControlPlaneReadyCondition, anywherev1.ControlPlaneInitializationInProgressReason, clusterv1.ConditionSeverityInfo, controlPlaneInitalizationInProgressReason),
				*conditions.FalseCondition(anywherev1.DefaultCNIConfiguredCondition, anywherev1.ControlPlaneNotReadyReason, clusterv1.ConditionSeverityInfo, ""),
				*conditions.FalseCondition(anywherev1.WorkersReadyCondition, anywherev1.ControlPlaneNotInitializedReason, clusterv1.ConditionSeverityInfo, ""),
				*conditions.FalseCondition(anywherev1.ReadyCondition, anywherev1.ControlPlaneInitializationInProgressReason, clusterv1.ConditionSeverityInfo, controlPlaneInitalizationInProgressReason),
			},
			result: ctrl.Result{Requeue: false, RequeueAfter: 10 * time.Second},
		},
		{
			testName:             "cluster not ready, control plane initialized",
			skipCNIUpgrade:       false,
			cniUpgradeInProgress: false,
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("False"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{},
			wantConditions: []anywherev1.Condition{
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
				*conditions.FalseCondition(anywherev1.ControlPlaneReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up control plane nodes, 1 expected (0 actual)"),
				*conditions.FalseCondition(anywherev1.DefaultCNIConfiguredCondition, anywherev1.ControlPlaneNotReadyReason, clusterv1.ConditionSeverityInfo, ""),
				*conditions.FalseCondition(anywherev1.WorkersReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up worker nodes, 1 expected (0 actual)"),
				*conditions.FalseCondition(anywherev1.ReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up control plane nodes, 1 expected (0 actual)"),
			},
			result: ctrl.Result{Requeue: false, RequeueAfter: 10 * time.Second},
		},
		{
			testName:             "cluster not ready, control plane ready",
			skipCNIUpgrade:       false,
			cniUpgradeInProgress: false,
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.ControlPlaneComponentsHealthyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{},
			wantConditions: []anywherev1.Condition{
				*conditions.FalseCondition(anywherev1.ReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up worker nodes, 1 expected (0 actual)"),
				*conditions.TrueCondition(anywherev1.ControlPlaneReadyCondition),
				*conditions.TrueCondition(anywherev1.DefaultCNIConfiguredCondition),
				*conditions.FalseCondition(anywherev1.WorkersReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up worker nodes, 1 expected (0 actual)"),
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
			},
			result: ctrl.Result{Requeue: false, RequeueAfter: 10 * time.Second},
		},
		{
			testName:             "cluster ready",
			skipCNIUpgrade:       false,
			cniUpgradeInProgress: false,
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.ControlPlaneComponentsHealthyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
			},
			wantConditions: []anywherev1.Condition{
				*conditions.TrueCondition(anywherev1.ReadyCondition),
				*conditions.TrueCondition(anywherev1.ControlPlaneReadyCondition),
				*conditions.TrueCondition(anywherev1.DefaultCNIConfiguredCondition),
				*conditions.TrueCondition(anywherev1.WorkersReadyCondition),
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
			},
			result: ctrl.Result{},
		},
		{
			testName:             "cluster ready, skip upgrades for default cni",
			skipCNIUpgrade:       true,
			cniUpgradeInProgress: false,
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.ControlPlaneComponentsHealthyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
			},
			wantConditions: []anywherev1.Condition{
				*conditions.TrueCondition(anywherev1.ReadyCondition),
				*conditions.TrueCondition(anywherev1.ControlPlaneReadyCondition),
				*conditions.FalseCondition(anywherev1.DefaultCNIConfiguredCondition, anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, clusterv1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades"),
				*conditions.TrueCondition(anywherev1.WorkersReadyCondition),
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
			},
			result: ctrl.Result{},
		},
		{
			testName:             "cluster not ready, default cni upgrade in progress",
			skipCNIUpgrade:       false,
			cniUpgradeInProgress: true,
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.ControlPlaneComponentsHealthyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
			},
			wantConditions: []anywherev1.Condition{
				*conditions.FalseCondition(anywherev1.ReadyCondition, anywherev1.DefaultCNIUpgradeInProgressReason, clusterv1.ConditionSeverityInfo, "Cilium version upgrade needed"),
				*conditions.TrueCondition(anywherev1.ControlPlaneReadyCondition),
				*conditions.FalseCondition(anywherev1.DefaultCNIConfiguredCondition, anywherev1.DefaultCNIUpgradeInProgressReason, clusterv1.ConditionSeverityInfo, "Cilium version upgrade needed"),
				*conditions.TrueCondition(anywherev1.WorkersReadyCondition),
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
			},
			result: ctrl.Result{Requeue: false, RequeueAfter: 10 * time.Second},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			config, bundles := baseTestVsphereCluster()
			config.Cluster.Name = "test-cluster"
			config.Cluster.Generation = 2
			config.Cluster.Status.ObservedGeneration = 1
			config.Cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}

			config.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(tt.skipCNIUpgrade)

			mgmt := config.DeepCopy()
			mgmt.Cluster.Name = "management-cluster"
			config.Cluster.Spec.MachineHealthCheck = &anywherev1.MachineHealthCheck{
				UnhealthyMachineTimeout: &metav1.Duration{
					Duration: constants.DefaultUnhealthyMachineTimeout,
				},
				NodeStartupTimeout: &metav1.Duration{
					Duration: constants.DefaultNodeStartupTimeout,
				},
			}

			g := NewWithT(t)

			objs := make([]runtime.Object, 0, 4+len(config.ChildObjects()))
			kcp := test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				k := controller.CAPIKubeadmControlPlaneKey(config.Cluster)
				kcp.Name = k.Name
				kcp.Namespace = k.Namespace
				kcp.Status = tt.kcpStatus
			})

			md1 := test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
				md.ObjectMeta.Labels = map[string]string{
					clusterv1.ClusterNameLabel: config.Cluster.Name,
				}
				md.Status = tt.machineDeploymentStatus
			})

			objs = append(objs, config.Cluster, bundles, kcp, md1, mgmt.Cluster, test.EKSARelease())

			for _, o := range config.ChildObjects() {
				objs = append(objs, o)
			}

			testClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

			mockCtrl := gomock.NewController(t)
			providerReconciler := mocks.NewMockProviderClusterReconciler(mockCtrl)
			iam := mocks.NewMockAWSIamConfigReconciler(mockCtrl)
			clusterValidator := mocks.NewMockClusterValidator(mockCtrl)
			registry := newRegistryMock(providerReconciler)
			mockPkgs := mocks.NewMockPackagesClient(mockCtrl)
			mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(mockCtrl)

			ctx := context.Background()
			log := testr.New(t)
			logCtx := ctrl.LoggerInto(ctx, log)

			iam.EXPECT().EnsureCASecret(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(controller.Result{}, nil)
			iam.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(controller.Result{}, nil)
			providerReconciler.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Times(1).Do(
				func(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) {
					kcpReadyCondition := conditions.Get(kcp, clusterv1.ReadyCondition)
					if kcpReadyCondition == nil ||
						(kcpReadyCondition != nil && kcpReadyCondition.Status == "False") {
						conditions.MarkFalse(cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.ControlPlaneNotReadyReason, clusterv1.ConditionSeverityInfo, "")
						return
					}

					if tt.skipCNIUpgrade {
						conditions.MarkFalse(cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, clusterv1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades")
						return
					}

					if tt.cniUpgradeInProgress {
						conditions.MarkFalse(cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.DefaultCNIUpgradeInProgressReason, clusterv1.ConditionSeverityInfo, "Cilium version upgrade needed")
						return
					}

					conditions.MarkTrue(cluster, anywherev1.DefaultCNIConfiguredCondition)
				},
			)
			clusterValidator.EXPECT().ValidateManagementClusterName(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(nil)

			mockPkgs.EXPECT().Reconcile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

			mhcReconciler.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(nil)

			r := controllers.NewClusterReconciler(testClient, registry, iam, clusterValidator, mockPkgs, mhcReconciler)

			result, err := r.Reconcile(logCtx, clusterRequest(config.Cluster))

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(tt.result))

			api := envtest.NewAPIExpecter(t, testClient)
			c := envtest.CloneNameNamespace(config.Cluster)
			api.ShouldEventuallyMatch(logCtx, c, func(g Gomega) {
				g.Expect(c.Status.ObservedGeneration).To(
					Equal(c.Generation), "status generation should have been updated to the metadata generation's value",
				)
			})

			api.ShouldEventuallyMatch(logCtx, c, func(g Gomega) {
				for _, wantCondition := range tt.wantConditions {
					condition := conditions.Get(c, wantCondition.Type)
					g.Expect(condition).ToNot(BeNil())
					g.Expect(condition).To((conditions.HaveSameStateOf(&wantCondition)))
				}
			})
		})
	}
}

func TestClusterReconcilerReconcileSelfManagedClusterConditions(t *testing.T) {
	testCases := []struct {
		testName                string
		skipCNIUpgrade          bool
		kcpStatus               controlplanev1.KubeadmControlPlaneStatus
		machineDeploymentStatus clusterv1.MachineDeploymentStatus
		result                  ctrl.Result
		wantConditions          []anywherev1.Condition
	}{
		{
			testName: "cluster not ready, control plane not ready",
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("False"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{},
			skipCNIUpgrade:          false,
			wantConditions: []anywherev1.Condition{
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
				*conditions.FalseCondition(anywherev1.ControlPlaneReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up control plane nodes, 1 expected (0 actual)"),
				*conditions.FalseCondition(anywherev1.DefaultCNIConfiguredCondition, anywherev1.ControlPlaneNotReadyReason, clusterv1.ConditionSeverityInfo, ""),
				*conditions.FalseCondition(anywherev1.WorkersReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up worker nodes, 1 expected (0 actual)"),
				*conditions.FalseCondition(anywherev1.ReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up control plane nodes, 1 expected (0 actual)"),
			},
			result: ctrl.Result{Requeue: false, RequeueAfter: 10 * time.Second},
		},
		{
			testName: "cluster not ready, control plane ready",
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.ControlPlaneComponentsHealthyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{},
			skipCNIUpgrade:          false,
			wantConditions: []anywherev1.Condition{
				*conditions.FalseCondition(anywherev1.ReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up worker nodes, 1 expected (0 actual)"),
				*conditions.TrueCondition(anywherev1.ControlPlaneReadyCondition),
				*conditions.TrueCondition(anywherev1.DefaultCNIConfiguredCondition),
				*conditions.FalseCondition(anywherev1.WorkersReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up worker nodes, 1 expected (0 actual)"),
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
			},
			result: ctrl.Result{Requeue: false, RequeueAfter: 10 * time.Second},
		},
		{
			testName: "cluster not ready, skip upgrades for default cni",
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.ControlPlaneComponentsHealthyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{},
			skipCNIUpgrade:          true,
			wantConditions: []anywherev1.Condition{
				*conditions.FalseCondition(anywherev1.ReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up worker nodes, 1 expected (0 actual)"),
				*conditions.TrueCondition(anywherev1.ControlPlaneReadyCondition),
				*conditions.FalseCondition(anywherev1.DefaultCNIConfiguredCondition, anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, clusterv1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades"),
				*conditions.FalseCondition(anywherev1.WorkersReadyCondition, anywherev1.ScalingUpReason, clusterv1.ConditionSeverityInfo, "Scaling up worker nodes, 1 expected (0 actual)"),
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
			},
			result: ctrl.Result{Requeue: false, RequeueAfter: 10 * time.Second},
		},
		{
			testName: "cluster ready, skip default cni upgrades",
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.ControlPlaneComponentsHealthyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
			},
			skipCNIUpgrade: true,
			wantConditions: []anywherev1.Condition{
				*conditions.TrueCondition(anywherev1.ReadyCondition),
				*conditions.TrueCondition(anywherev1.ControlPlaneReadyCondition),
				*conditions.FalseCondition(anywherev1.DefaultCNIConfiguredCondition, anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, clusterv1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades"),
				*conditions.TrueCondition(anywherev1.WorkersReadyCondition),
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
			},
			result: ctrl.Result{},
		},
		{
			testName: "cluster ready",
			kcpStatus: controlplanev1.KubeadmControlPlaneStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
				Conditions: clusterv1.Conditions{
					{
						Type:   controlplanev1.ControlPlaneComponentsHealthyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   controlplanev1.AvailableCondition,
						Status: apiv1.ConditionStatus("True"),
					},
					{
						Type:   clusterv1.ReadyCondition,
						Status: apiv1.ConditionStatus("True"),
					},
				},
			},
			machineDeploymentStatus: clusterv1.MachineDeploymentStatus{
				ReadyReplicas:   1,
				Replicas:        1,
				UpdatedReplicas: 1,
			},
			skipCNIUpgrade: false,
			wantConditions: []anywherev1.Condition{
				*conditions.TrueCondition(anywherev1.ReadyCondition),
				*conditions.TrueCondition(anywherev1.ControlPlaneReadyCondition),
				*conditions.TrueCondition(anywherev1.DefaultCNIConfiguredCondition),
				*conditions.TrueCondition(anywherev1.WorkersReadyCondition),
				*conditions.TrueCondition(anywherev1.ControlPlaneInitializedCondition),
			},
			result: ctrl.Result{},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			config, bundles := baseTestVsphereCluster()
			clusterName := "test-cluster"
			config.Cluster.Name = clusterName
			config.Cluster.Generation = 2
			config.Cluster.Status.ObservedGeneration = 1
			config.Cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: clusterName}

			config.Cluster.Spec.ClusterNetwork.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(tt.skipCNIUpgrade)

			g := NewWithT(t)

			objs := make([]runtime.Object, 0, 4+len(config.ChildObjects()))
			kcp := test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
				k := controller.CAPIKubeadmControlPlaneKey(config.Cluster)
				kcp.Name = k.Name
				kcp.Namespace = k.Namespace
				kcp.Status = tt.kcpStatus
			})

			md1 := test.MachineDeployment(func(md *clusterv1.MachineDeployment) {
				md.ObjectMeta.Labels = map[string]string{
					clusterv1.ClusterNameLabel: config.Cluster.Name,
				}
				md.Status = tt.machineDeploymentStatus
			})

			objs = append(objs, config.Cluster, bundles, kcp, md1, test.EKSARelease())
			for _, o := range config.ChildObjects() {
				objs = append(objs, o)
			}
			testClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

			mockCtrl := gomock.NewController(t)
			providerReconciler := mocks.NewMockProviderClusterReconciler(mockCtrl)
			iam := mocks.NewMockAWSIamConfigReconciler(mockCtrl)
			clusterValidator := mocks.NewMockClusterValidator(mockCtrl)
			registry := newRegistryMock(providerReconciler)
			mockPkgs := mocks.NewMockPackagesClient(mockCtrl)
			mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(mockCtrl)

			ctx := context.Background()
			log := testr.New(t)
			logCtx := ctrl.LoggerInto(ctx, log)

			iam.EXPECT().EnsureCASecret(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(controller.Result{}, nil)
			iam.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(controller.Result{}, nil)

			providerReconciler.EXPECT().Reconcile(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Do(
				func(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) {
					kcpReadyCondition := conditions.Get(kcp, clusterv1.ReadyCondition)
					if kcpReadyCondition == nil ||
						(kcpReadyCondition != nil && kcpReadyCondition.Status == "False") {
						conditions.MarkFalse(cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.ControlPlaneNotReadyReason, clusterv1.ConditionSeverityInfo, "")
						return
					}

					if tt.skipCNIUpgrade {
						conditions.MarkFalse(cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.SkipUpgradesForDefaultCNIConfiguredReason, clusterv1.ConditionSeverityWarning, "Configured to skip default Cilium CNI upgrades")
						return
					}

					conditions.MarkTrue(cluster, anywherev1.DefaultCNIConfiguredCondition)
				},
			)
			mhcReconciler.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(nil)

			r := controllers.NewClusterReconciler(testClient, registry, iam, clusterValidator, mockPkgs, mhcReconciler)

			result, err := r.Reconcile(logCtx, clusterRequest(config.Cluster))

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(tt.result))

			api := envtest.NewAPIExpecter(t, testClient)
			c := envtest.CloneNameNamespace(config.Cluster)
			api.ShouldEventuallyMatch(logCtx, c, func(g Gomega) {
				g.Expect(c.Status.ObservedGeneration).To(
					Equal(c.Generation), "status generation should have been updated to the metadata generation's value",
				)
			})

			api.ShouldEventuallyMatch(logCtx, c, func(g Gomega) {
				for _, wantCondition := range tt.wantConditions {
					condition := conditions.Get(c, wantCondition.Type)
					g.Expect(condition).ToNot(BeNil())
					g.Expect(condition).To((conditions.HaveSameStateOf(&wantCondition)))
				}
			})
		})
	}
}

func TestClusterReconcilerReconcileGenerations(t *testing.T) {
	testCases := []struct {
		testName                  string
		clusterGeneration         int64
		childReconciledGeneration int64
		reconciledGeneration      int64

		datacenterGeneration          int64
		cpMachineConfigGeneration     int64
		workerMachineConfigGeneration int64
		oidcGeneration                int64
		awsIAMGeneration              int64

		wantReconciliation            bool
		wantChildReconciledGeneration int64
	}{
		{
			testName:                      "matching generation, matching aggregated generation",
			clusterGeneration:             2,
			reconciledGeneration:          2,
			childReconciledGeneration:     12,
			datacenterGeneration:          1,
			cpMachineConfigGeneration:     2,
			workerMachineConfigGeneration: 5,
			oidcGeneration:                3,
			awsIAMGeneration:              1,
			wantReconciliation:            false,
			wantChildReconciledGeneration: 12,
		},
		{
			testName:                      "matching generation, non-matching aggregated generation",
			clusterGeneration:             2,
			reconciledGeneration:          2,
			childReconciledGeneration:     10,
			datacenterGeneration:          1,
			cpMachineConfigGeneration:     2,
			workerMachineConfigGeneration: 5,
			oidcGeneration:                3,
			awsIAMGeneration:              1,
			wantReconciliation:            true,
			wantChildReconciledGeneration: 12,
		},
		{
			testName:                      "non-matching generation, matching aggregated generation",
			clusterGeneration:             3,
			reconciledGeneration:          2,
			childReconciledGeneration:     12,
			datacenterGeneration:          1,
			cpMachineConfigGeneration:     2,
			workerMachineConfigGeneration: 5,
			oidcGeneration:                3,
			awsIAMGeneration:              1,
			wantReconciliation:            true,
			wantChildReconciledGeneration: 12,
		},
		{
			testName:                      "non-matching generation, non-matching aggregated generation",
			clusterGeneration:             3,
			reconciledGeneration:          2,
			childReconciledGeneration:     12,
			datacenterGeneration:          1,
			cpMachineConfigGeneration:     2,
			workerMachineConfigGeneration: 5,
			oidcGeneration:                3,
			awsIAMGeneration:              3,
			wantReconciliation:            true,
			wantChildReconciledGeneration: 14,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			config, bundles := baseTestVsphereCluster()
			version := test.DevEksaVersion()
			config.Cluster.Spec.EksaVersion = &version
			config.Cluster.Generation = tt.clusterGeneration
			config.Cluster.Status.ObservedGeneration = tt.clusterGeneration
			config.Cluster.Status.ReconciledGeneration = tt.reconciledGeneration
			config.Cluster.Status.ReconciledGeneration = tt.reconciledGeneration
			config.Cluster.Status.ChildrenReconciledGeneration = tt.childReconciledGeneration

			config.VSphereDatacenter.Generation = tt.datacenterGeneration
			cpMachine := config.VSphereMachineConfigs[config.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
			cpMachine.Generation = tt.cpMachineConfigGeneration
			workerMachineConfig := config.VSphereMachineConfigs[config.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
			workerMachineConfig.Generation = tt.workerMachineConfigGeneration

			for _, oidc := range config.OIDCConfigs {
				oidc.Generation = tt.oidcGeneration
			}
			for _, awsIAM := range config.AWSIAMConfigs {
				awsIAM.Generation = tt.awsIAMGeneration
			}

			kcp := testKubeadmControlPlaneFromCluster(config.Cluster)
			machineDeployments := machineDeploymentsFromCluster(config.Cluster)

			g := NewWithT(t)
			ctx := context.Background()

			objs := make([]runtime.Object, 0, 7+len(machineDeployments))
			objs = append(objs, config.Cluster, bundles, test.EKSARelease())
			for _, o := range config.ChildObjects() {
				objs = append(objs, o)
			}

			objs = append(objs, kcp)

			for _, obj := range machineDeployments {
				objs = append(objs, obj.DeepCopy())
			}

			client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
			mockCtrl := gomock.NewController(t)
			providerReconciler := mocks.NewMockProviderClusterReconciler(mockCtrl)
			iam := mocks.NewMockAWSIamConfigReconciler(mockCtrl)
			clusterValidator := mocks.NewMockClusterValidator(mockCtrl)
			registry := newRegistryMock(providerReconciler)
			mockPkgs := mocks.NewMockPackagesClient(mockCtrl)
			mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(mockCtrl)

			if tt.wantReconciliation {
				iam.EXPECT().EnsureCASecret(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(config.Cluster)).Return(controller.Result{}, nil)
				iam.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(config.Cluster)).Return(controller.Result{}, nil)
				providerReconciler.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Times(1)
				mhcReconciler.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(nil)

			} else {
				providerReconciler.EXPECT().Reconcile(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			}

			r := controllers.NewClusterReconciler(client, registry, iam, clusterValidator, mockPkgs, mhcReconciler)

			result, err := r.Reconcile(ctx, clusterRequest(config.Cluster))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(ctrl.Result{}))

			api := envtest.NewAPIExpecter(t, client)
			c := envtest.CloneNameNamespace(config.Cluster)
			api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
				g.Expect(c.Status.ReconciledGeneration).To(
					Equal(c.Generation), "status generation should have been updated to the metadata generation's value",
				)

				g.Expect(c.Status.ChildrenReconciledGeneration).To(
					Equal(tt.wantChildReconciledGeneration), "status children generation should have been updated to the aggregated generation's value",
				)
			})
		})
	}
}

func TestClusterReconcilerReconcilePausedCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	managementCluster := vsphereCluster()
	managementCluster.Name = "management-cluster"
	cluster := vsphereCluster()
	cluster.SetManagedBy(managementCluster.Name)
	capiCluster := newCAPICluster(cluster.Name, cluster.Namespace)
	kcp := testKubeadmControlPlaneFromCluster(cluster)
	machineDeployments := machineDeploymentsFromCluster(cluster)

	objs := make([]runtime.Object, 0, 5)
	objs = append(objs, managementCluster, cluster, capiCluster, kcp)
	for _, md := range machineDeployments {
		objs = append(objs, md.DeepCopy())
	}

	// Mark as paused
	cluster.PauseReconcile()

	c := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	ctrl := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(ctrl)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	clusterValidator := mocks.NewMockClusterValidator(ctrl)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(ctrl)

	registry := newRegistryMock(providerReconciler)
	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, nil, mhcReconciler)
	g.Expect(r.Reconcile(ctx, clusterRequest(cluster))).To(Equal(reconcile.Result{}))
	api := envtest.NewAPIExpecter(t, c)

	cl := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, cl, func(g Gomega) {
		g.Expect(
			controllerutil.ContainsFinalizer(cluster, controllers.ClusterFinalizerName),
		).To(BeFalse(), "Cluster should not have the finalizer added")
	})
}

func TestClusterReconcilerReconcileDeletedSelfManagedCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	deleteTimestamp := metav1.NewTime(time.Now())
	selfManagedCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "my-management-cluster",
			DeletionTimestamp: &deleteTimestamp,
		},
		Spec: anywherev1.ClusterSpec{
			BundlesRef: &anywherev1.BundlesRef{
				Name: "my-bundles-ref",
			},
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				},
			},
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	clusterValidator := mocks.NewMockClusterValidator(controller)
	registry := newRegistryMock(providerReconciler)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(controller)

	c := fake.NewClientBuilder().WithRuntimeObjects(selfManagedCluster).Build()

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, nil, mhcReconciler)
	_, err := r.Reconcile(ctx, clusterRequest(selfManagedCluster))
	g.Expect(err).To(MatchError(ContainSubstring("deleting self-managed clusters is not supported")))
}

func TestClusterReconcilerReconcileSelfManagedClusterRegAuthFailNoSecret(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	selfManagedCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-management-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			BundlesRef: &anywherev1.BundlesRef{
				Name: "my-bundles-ref",
			},
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				},
			},
			RegistryMirrorConfiguration: &anywherev1.RegistryMirrorConfiguration{
				Authenticate: true,
			},
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	clusterValidator := mocks.NewMockClusterValidator(controller)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(controller)

	registry := newRegistryMock(providerReconciler)
	c := fake.NewClientBuilder().WithRuntimeObjects(selfManagedCluster, test.EKSARelease()).Build()

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, nil, mhcReconciler)
	_, err := r.Reconcile(ctx, clusterRequest(selfManagedCluster))
	g.Expect(err).To(MatchError(ContainSubstring("fetching registry auth secret")))
}

func TestClusterReconcilerDeleteExistingCAPIClusterSuccess(t *testing.T) {
	secret := createSecret()
	managementCluster := vsphereCluster()
	managementCluster.Name = "management-cluster"
	cluster := vsphereCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}
	now := metav1.Now()
	cluster.DeletionTimestamp = &now

	datacenterConfig := vsphereDataCenter(cluster)
	bundle := createBundle(managementCluster)
	machineConfigCP := vsphereCPMachineConfig()
	machineConfigWN := vsphereWorkerMachineConfig()

	capiCluster := newCAPICluster(cluster.Name, cluster.Namespace)

	objs := []runtime.Object{cluster, datacenterConfig, secret, bundle, machineConfigCP, machineConfigWN, managementCluster, capiCluster}

	tt := newVsphereClusterReconcilerTest(t, objs...)

	req := clusterRequest(cluster)

	ctx := context.Background()

	_, err := tt.reconciler.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	apiCluster := &clusterv1.Cluster{}

	err = tt.client.Get(context.TODO(), req.NamespacedName, apiCluster)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected apierrors.IsNotFound but got: (%v)", err)
	}
	if apiCluster.Status.FailureMessage != nil {
		t.Errorf("Expected failure message to be nil. FailureMessage:%s", *apiCluster.Status.FailureMessage)
	}
	if apiCluster.Status.FailureReason != nil {
		t.Errorf("Expected failure message to be nil. FailureReason:%s", *apiCluster.Status.FailureReason)
	}
}

func TestClusterReconcilerReconcileDeletePausedCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	managementCluster := vsphereCluster()
	managementCluster.Name = "management-cluster"
	cluster := vsphereCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}
	controllerutil.AddFinalizer(cluster, controllers.ClusterFinalizerName)
	capiCluster := newCAPICluster(cluster.Name, cluster.Namespace)
	kcp := testKubeadmControlPlaneFromCluster(cluster)
	machineDeployments := machineDeploymentsFromCluster(cluster)

	objs := make([]runtime.Object, 0, 5)
	objs = append(objs, managementCluster, cluster, capiCluster, kcp)
	for _, md := range machineDeployments {
		objs = append(objs, md.DeepCopy())
	}

	// Mark cluster for deletion
	now := metav1.Now()
	cluster.DeletionTimestamp = &now

	// Mark as paused
	cluster.PauseReconcile()

	controller := gomock.NewController(t)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	clusterValidator := mocks.NewMockClusterValidator(controller)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(controller)

	c := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	r := controllers.NewClusterReconciler(c, newRegistryForDummyProviderReconciler(), iam, clusterValidator, nil, mhcReconciler)
	g.Expect(r.Reconcile(ctx, clusterRequest(cluster))).To(Equal(reconcile.Result{}))
	api := envtest.NewAPIExpecter(t, c)

	cl := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, cl, func(g Gomega) {
		g.Expect(
			controllerutil.ContainsFinalizer(cluster, controllers.ClusterFinalizerName),
		).To(BeTrue(), "Cluster should still have the finalizer")
	})

	capiCl := envtest.CloneNameNamespace(capiCluster)
	api.ShouldEventuallyMatch(ctx, capiCl, func(g Gomega) {
		g.Expect(
			capiCluster.DeletionTimestamp.IsZero(),
		).To(BeTrue(), "CAPI cluster should exist and not be marked for deletion")
	})
}

func TestClusterReconcilerReconcileDeleteClusterManagedByCLI(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	managementCluster := vsphereCluster()
	managementCluster.Name = "management-cluster"
	cluster := vsphereCluster()
	cluster.SetManagedBy(managementCluster.Name)
	controllerutil.AddFinalizer(cluster, controllers.ClusterFinalizerName)
	capiCluster := newCAPICluster(cluster.Name, cluster.Namespace)

	// Mark cluster for deletion
	now := metav1.Now()
	cluster.DeletionTimestamp = &now

	// Mark as managed by CLI
	cluster.Annotations[anywherev1.ManagedByCLIAnnotation] = "true"

	c := fake.NewClientBuilder().WithRuntimeObjects(
		managementCluster, cluster, capiCluster,
	).Build()
	controller := gomock.NewController(t)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	clusterValidator := mocks.NewMockClusterValidator(controller)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(controller)

	r := controllers.NewClusterReconciler(c, newRegistryForDummyProviderReconciler(), iam, clusterValidator, nil, mhcReconciler)
	g.Expect(r.Reconcile(ctx, clusterRequest(cluster))).To(Equal(reconcile.Result{}))
	api := envtest.NewAPIExpecter(t, c)

	cl := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyNotExist(ctx, cl)

	capiCl := envtest.CloneNameNamespace(capiCluster)
	api.ShouldEventuallyMatch(ctx, capiCl, func(g Gomega) {
		g.Expect(
			capiCluster.DeletionTimestamp.IsZero(),
		).To(BeTrue(), "CAPI cluster should exist and not be marked for deletion")
	})
}

func TestClusterReconcilerDeleteNoCAPIClusterSuccess(t *testing.T) {
	g := NewWithT(t)

	secret := createSecret()
	managementCluster := vsphereCluster()
	managementCluster.Name = "management-cluster"
	cluster := vsphereCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}
	now := metav1.Now()
	cluster.DeletionTimestamp = &now

	datacenterConfig := vsphereDataCenter(cluster)
	bundle := createBundle(managementCluster)
	machineConfigCP := vsphereCPMachineConfig()
	machineConfigWN := vsphereWorkerMachineConfig()

	objs := []runtime.Object{cluster, datacenterConfig, secret, bundle, machineConfigCP, machineConfigWN, managementCluster}

	g.Expect(cluster.Finalizers).NotTo(ContainElement(controllers.ClusterFinalizerName))

	tt := newVsphereClusterReconcilerTest(t, objs...)

	req := clusterRequest(cluster)

	ctx := context.Background()

	controllerutil.AddFinalizer(cluster, controllers.ClusterFinalizerName)
	_, err := tt.reconciler.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	apiCluster := &anywherev1.Cluster{}

	err = tt.client.Get(context.TODO(), req.NamespacedName, apiCluster)
	if err != nil {
		t.Fatalf("get cluster: (%v)", err)
	}

	if apiCluster.Status.FailureMessage != nil {
		t.Errorf("Expected failure message to be nil. FailureMessage:%s", *apiCluster.Status.FailureMessage)
	}
	if apiCluster.Status.FailureReason != nil {
		t.Errorf("Expected failure reason to be nil. FailureReason:%s", *apiCluster.Status.FailureReason)
	}
}

func TestClusterReconcilerSkipDontInstallPackagesOnSelfManaged(t *testing.T) {
	ctx := context.Background()
	version := test.DevEksaVersion()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "my-namespace",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "v1.25",
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "my-bundles-ref",
				Namespace: "my-namespace",
			},
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				},
			},
			ManagementCluster: anywherev1.ManagementCluster{
				Name: "",
			},
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	objs := []runtime.Object{cluster, test.EKSARelease()}
	cb := fake.NewClientBuilder()
	mockClient := cb.WithRuntimeObjects(objs...).Build()
	nullRegistry := newRegistryForDummyProviderReconciler()

	ctrl := gomock.NewController(t)
	mockPkgs := mocks.NewMockPackagesClient(ctrl)
	mockPkgs.EXPECT().ReconcileDelete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(ctrl)
	mhcReconciler.EXPECT().Reconcile(ctx, gomock.Any(), sameName(cluster)).Return(nil)
	r := controllers.NewClusterReconciler(mockClient, nullRegistry, nil, nil, mockPkgs, mhcReconciler)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	if err != nil {
		t.Fatalf("expected err to be nil, got %s", err)
	}
}

func TestClusterReconcilerDontDeletePackagesOnSelfManaged(t *testing.T) {
	ctx := context.Background()
	deleteTime := metav1.NewTime(time.Now().Add(-1 * time.Second))
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "my-cluster",
			Namespace:         "my-namespace",
			DeletionTimestamp: &deleteTime,
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "v1.25",
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "my-bundles-ref",
				Namespace: "my-namespace",
			},
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				},
			},
			ManagementCluster: anywherev1.ManagementCluster{
				Name: "",
			},
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	objs := []runtime.Object{cluster}
	cb := fake.NewClientBuilder()
	mockClient := cb.WithRuntimeObjects(objs...).Build()
	nullRegistry := newRegistryForDummyProviderReconciler()

	ctrl := gomock.NewController(t)
	// At the moment, Reconcile won't get this far, but if the time comes when
	// deleting self-managed clusters via full cluster lifecycle happens, we
	// need to be aware and adapt appropriately.
	mockPkgs := mocks.NewMockPackagesClient(ctrl)
	mockPkgs.EXPECT().ReconcileDelete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	r := controllers.NewClusterReconciler(mockClient, nullRegistry, nil, nil, mockPkgs, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	if err == nil || !strings.Contains(err.Error(), "deleting self-managed clusters is not supported") {
		t.Fatalf("unexpected error %s", err)
	}
}

func TestClusterReconcilerPackagesDeletion(s *testing.T) {
	newTestCluster := func() *anywherev1.Cluster {
		deleteTime := metav1.NewTime(time.Now().Add(-1 * time.Second))
		return &anywherev1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "my-workload-cluster",
				Namespace:         "my-namespace",
				DeletionTimestamp: &deleteTime,
			},
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: "v1.25",
				BundlesRef: &anywherev1.BundlesRef{
					Name:      "my-bundles-ref",
					Namespace: "my-namespace",
				},
				ClusterNetwork: anywherev1.ClusterNetwork{
					CNIConfig: &anywherev1.CNIConfig{
						Cilium: &anywherev1.CiliumConfig{},
					},
				},
				ManagementCluster: anywherev1.ManagementCluster{
					Name: "my-management-cluster",
				},
			},
			Status: anywherev1.ClusterStatus{
				ReconciledGeneration: 1,
			},
		}
	}

	s.Run("errors when packages client errors", func(t *testing.T) {
		ctx := context.Background()
		log := testr.New(t)
		logCtx := ctrl.LoggerInto(ctx, log)
		cluster := newTestCluster()
		cluster.Spec.BundlesRef.Name = "non-existent"
		ctrl := gomock.NewController(t)
		objs := []runtime.Object{cluster}
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		nullRegistry := newRegistryForDummyProviderReconciler()
		mockPkgs := mocks.NewMockPackagesClient(ctrl)
		mockPkgs.EXPECT().ReconcileDelete(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test error"))
		mockIAM := mocks.NewMockAWSIamConfigReconciler(ctrl)
		mockValid := mocks.NewMockClusterValidator(ctrl)
		mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(ctrl)

		r := controllers.NewClusterReconciler(fakeClient, nullRegistry, mockIAM, mockValid, mockPkgs, mhcReconciler)
		_, err := r.Reconcile(logCtx, clusterRequest(cluster))
		if err == nil || !strings.Contains(err.Error(), "test error") {
			t.Errorf("expected packages client deletion error, got %s", err)
		}
	})
}

func TestClusterReconcilerPackagesInstall(s *testing.T) {
	version := test.DevEksaVersion()
	newTestCluster := func() *anywherev1.Cluster {
		return &anywherev1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-workload-cluster",
				Namespace: "my-namespace",
			},
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: "v1.25",
				BundlesRef: &anywherev1.BundlesRef{
					Name:      "my-bundles-ref",
					Namespace: "my-namespace",
				},
				ClusterNetwork: anywherev1.ClusterNetwork{
					CNIConfig: &anywherev1.CNIConfig{
						Cilium: &anywherev1.CiliumConfig{},
					},
				},
				ManagementCluster: anywherev1.ManagementCluster{
					Name: "my-management-cluster",
				},
				EksaVersion: &version,
				MachineHealthCheck: &anywherev1.MachineHealthCheck{
					UnhealthyMachineTimeout: &metav1.Duration{
						Duration: constants.DefaultUnhealthyMachineTimeout,
					},
					NodeStartupTimeout: &metav1.Duration{
						Duration: constants.DefaultNodeStartupTimeout,
					},
				},
			},
			Status: anywherev1.ClusterStatus{
				ReconciledGeneration: 1,
			},
		}
	}

	s.Run("skips installation when disabled via cluster spec", func(t *testing.T) {
		ctx := context.Background()
		log := testr.New(t)
		logCtx := ctrl.LoggerInto(ctx, log)
		cluster := newTestCluster()
		cluster.Spec.Packages = &anywherev1.PackageConfiguration{Disable: true}
		ctrl := gomock.NewController(t)
		bundles := createBundle(cluster)
		bundles.Spec.VersionsBundles[0].KubeVersion = string(cluster.Spec.KubernetesVersion)
		bundles.ObjectMeta.Name = cluster.Spec.BundlesRef.Name
		bundles.ObjectMeta.Namespace = cluster.Spec.BundlesRef.Namespace
		secret := &apiv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: constants.EksaSystemNamespace,
				Name:      cluster.Name + "-kubeconfig",
			},
		}
		mgmt := cluster.DeepCopy()
		mgmt.Name = "my-management-cluster"
		objs := []runtime.Object{cluster, bundles, secret, mgmt, test.EKSARelease()}
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		nullRegistry := newRegistryForDummyProviderReconciler()
		mockIAM := mocks.NewMockAWSIamConfigReconciler(ctrl)
		mockValid := mocks.NewMockClusterValidator(ctrl)
		mockValid.EXPECT().ValidateManagementClusterName(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)
		mockPkgs := mocks.NewMockPackagesClient(ctrl)
		mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(ctrl)

		mhcReconciler.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(cluster)).Return(nil)

		mockPkgs.EXPECT().
			EnableFullLifecycle(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		r := controllers.NewClusterReconciler(fakeClient, nullRegistry, mockIAM, mockValid, mockPkgs, mhcReconciler)
		_, err := r.Reconcile(logCtx, clusterRequest(cluster))
		if err != nil {
			t.Errorf("expected nil error, got %s", err)
		}
	})
}

func TestClusterReconcilerValidateManagementEksaVersionFail(t *testing.T) {
	lower := anywherev1.EksaVersion("v0.1.0")
	higher := anywherev1.EksaVersion("v0.5.0")
	config, _ := baseTestVsphereCluster()
	config.Cluster.Name = "test-cluster"
	config.Cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}
	config.Cluster.Spec.BundlesRef = nil
	config.Cluster.Spec.EksaVersion = &higher

	mgmt := config.DeepCopy()
	mgmt.Cluster.Name = "management-cluster"
	mgmt.Cluster.Spec.BundlesRef = nil
	mgmt.Cluster.Spec.EksaVersion = &lower

	g := NewWithT(t)

	objs := make([]runtime.Object, 0, 4+len(config.ChildObjects()))
	objs = append(objs, config.Cluster, mgmt.Cluster)

	for _, o := range config.ChildObjects() {
		objs = append(objs, o)
	}

	testClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	mockCtrl := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(mockCtrl)
	iam := mocks.NewMockAWSIamConfigReconciler(mockCtrl)
	clusterValidator := mocks.NewMockClusterValidator(mockCtrl)
	registry := newRegistryMock(providerReconciler)
	mockPkgs := mocks.NewMockPackagesClient(mockCtrl)

	ctx := context.Background()
	log := testr.New(t)
	logCtx := ctrl.LoggerInto(ctx, log)

	iam.EXPECT().EnsureCASecret(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(controller.Result{}, nil)
	clusterValidator.EXPECT().ValidateManagementClusterName(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(nil)

	r := controllers.NewClusterReconciler(testClient, registry, iam, clusterValidator, mockPkgs, nil)

	_, err := r.Reconcile(logCtx, clusterRequest(config.Cluster))

	g.Expect(err).To(HaveOccurred())
}

func TestClusterReconcilerNotAvailableEksaVersion(t *testing.T) {
	version := test.DevEksaVersion()
	config, _ := baseTestVsphereCluster()
	config.Cluster.Name = "test-cluster"
	config.Cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}
	config.Cluster.Spec.BundlesRef = nil
	config.Cluster.Spec.EksaVersion = &version

	mgmt := config.DeepCopy()
	mgmt.Cluster.Name = "management-cluster"
	mgmt.Cluster.Spec.BundlesRef = nil
	mgmt.Cluster.Spec.EksaVersion = &version

	g := NewWithT(t)

	objs := make([]runtime.Object, 0, 4+len(config.ChildObjects()))
	objs = append(objs, config.Cluster, mgmt.Cluster)

	for _, o := range config.ChildObjects() {
		objs = append(objs, o)
	}

	testClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()

	mockCtrl := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(mockCtrl)
	iam := mocks.NewMockAWSIamConfigReconciler(mockCtrl)
	clusterValidator := mocks.NewMockClusterValidator(mockCtrl)
	registry := newRegistryMock(providerReconciler)
	mockPkgs := mocks.NewMockPackagesClient(mockCtrl)

	ctx := context.Background()
	log := testr.New(t)
	logCtx := ctrl.LoggerInto(ctx, log)

	iam.EXPECT().EnsureCASecret(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(controller.Result{}, nil)
	clusterValidator.EXPECT().ValidateManagementClusterName(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(config.Cluster)).Return(nil)

	r := controllers.NewClusterReconciler(testClient, registry, iam, clusterValidator, mockPkgs, nil)

	req := clusterRequest(config.Cluster)
	_, err := r.Reconcile(logCtx, req)

	g.Expect(err).To(HaveOccurred())
	eksaCluster := &anywherev1.Cluster{}

	expectedError := fmt.Sprintf("eksarelease %v could not be found on the management cluster", *config.Cluster.Spec.EksaVersion)
	g.Expect(testClient.Get(ctx, req.NamespacedName, eksaCluster)).To(Succeed())

	g.Expect(eksaCluster.Status.FailureReason).To(HaveValue(Equal(anywherev1.EksaVersionInvalidReason)))
	g.Expect(eksaCluster.Status.FailureMessage).To(HaveValue(Equal(expectedError)))
}

func vsphereWorkerMachineConfig() *anywherev1.VSphereMachineConfig {
	return &anywherev1.VSphereMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereMachineConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-wn",
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       name,
				},
			},
		},
		Spec: anywherev1.VSphereMachineConfigSpec{
			DiskGiB:           40,
			Datastore:         "test",
			Folder:            "test",
			NumCPUs:           2,
			MemoryMiB:         16,
			OSFamily:          "ubuntu",
			ResourcePool:      "test",
			StoragePolicyName: "test",
			Template:          "test",
			Users: []anywherev1.UserConfiguration{
				{
					Name:              "user",
					SshAuthorizedKeys: []string{"ABC"},
				},
			},
		},
		Status: anywherev1.VSphereMachineConfigStatus{},
	}
}

func newCAPICluster(name, namespace string) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func vsphereCPMachineConfig() *anywherev1.VSphereMachineConfig {
	return &anywherev1.VSphereMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereMachineConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-cp",
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       name,
				},
			},
		},
		Spec: anywherev1.VSphereMachineConfigSpec{
			DiskGiB:           40,
			Datastore:         "test",
			Folder:            "test",
			NumCPUs:           2,
			MemoryMiB:         16,
			OSFamily:          "ubuntu",
			ResourcePool:      "test",
			StoragePolicyName: "test",
			Template:          "test",
			Users: []anywherev1.UserConfiguration{
				{
					Name:              "user",
					SshAuthorizedKeys: []string{"ABC"},
				},
			},
		},
		Status: anywherev1.VSphereMachineConfigStatus{},
	}
}

func createBundle(cluster *anywherev1.Cluster) *releasev1.Bundles {
	return &releasev1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: "default",
		},
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.20",
					EksD: releasev1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "testdata/release.yaml",
						KubeVersion:    "1.20",
					},
					CertManager:                releasev1.CertManagerBundle{},
					ClusterAPI:                 releasev1.CoreClusterAPI{},
					Bootstrap:                  releasev1.KubeadmBootstrapBundle{},
					ControlPlane:               releasev1.KubeadmControlPlaneBundle{},
					VSphere:                    releasev1.VSphereBundle{},
					Docker:                     releasev1.DockerBundle{},
					Eksa:                       releasev1.EksaBundle{},
					Cilium:                     releasev1.CiliumBundle{},
					Kindnetd:                   releasev1.KindnetdBundle{},
					Flux:                       releasev1.FluxBundle{},
					BottleRocketHostContainers: releasev1.BottlerocketHostContainersBundle{},
					ExternalEtcdBootstrap:      releasev1.EtcdadmBootstrapBundle{},
					ExternalEtcdController:     releasev1.EtcdadmControllerBundle{},
					Tinkerbell:                 releasev1.TinkerbellBundle{},
				},
			},
		},
	}
}

func vsphereDataCenter(cluster *anywherev1.Cluster) *anywherev1.VSphereDatacenterConfig {
	return &anywherev1.VSphereDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereDatacenterConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       cluster.Name,
				},
			},
		},
		Spec: anywherev1.VSphereDatacenterConfigSpec{
			Thumbprint: "aaa",
			Server:     "ssss",
			Datacenter: "daaa",
			Network:    "networkA",
		},
		Status: anywherev1.VSphereDatacenterConfigStatus{
			SpecValid: true,
		},
	}
}

func vsphereCluster() *anywherev1.Cluster {
	version := test.DevEksaVersion()
	return &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: namespace,
		},
		Spec: anywherev1.ClusterSpec{
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				},
			},
			DatacenterRef: anywherev1.Ref{
				Kind: "VSphereDatacenterConfig",
				Name: "datacenter",
			},
			KubernetesVersion: "1.20",
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &anywherev1.Endpoint{
					Host: "1.1.1.1",
				},
				MachineGroupRef: &anywherev1.Ref{
					Kind: "VSphereMachineConfig",
					Name: clusterName + "-cp",
				},
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
					MachineGroupRef: &anywherev1.Ref{
						Kind: "VSphereMachineConfig",
						Name: clusterName + "-wn",
					},
					Name:   "md-0",
					Labels: nil,
				},
			},
			EksaVersion: &version,
			MachineHealthCheck: &anywherev1.MachineHealthCheck{
				UnhealthyMachineTimeout: &metav1.Duration{
					Duration: constants.DefaultUnhealthyMachineTimeout,
				},
				NodeStartupTimeout: &metav1.Duration{
					Duration: constants.DefaultNodeStartupTimeout,
				},
			},
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
}

func createSecret() *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "eksa-system",
			Name:      vsphere.CredentialsObjectName,
		},
		Data: map[string][]byte{
			"username": []byte("test"),
			"password": []byte("test"),
		},
	}
}

type sameNameCluster struct{ c *anywherev1.Cluster }

func sameName(c *anywherev1.Cluster) gomock.Matcher {
	return &sameNameCluster{c}
}

func (s *sameNameCluster) Matches(x interface{}) bool {
	cluster, ok := x.(*anywherev1.Cluster)
	if !ok {
		return false
	}

	return s.c.Name == cluster.Name && s.c.Namespace == cluster.Namespace
}

func (s *sameNameCluster) String() string {
	return fmt.Sprintf("has name %s and namespace %s", s.c.Name, s.c.Namespace)
}

func baseTestVsphereCluster() (*cluster.Config, *releasev1.Bundles) {
	config := &cluster.Config{
		VSphereMachineConfigs: map[string]*anywherev1.VSphereMachineConfig{},
		OIDCConfigs:           map[string]*anywherev1.OIDCConfig{},
		AWSIAMConfigs:         map[string]*anywherev1.AWSIamConfig{},
	}

	config.Cluster = vsphereCluster()
	config.VSphereDatacenter = vsphereDataCenter(config.Cluster)

	machineConfigCP := vsphereCPMachineConfig()
	machineConfigWorker := vsphereWorkerMachineConfig()
	config.VSphereMachineConfigs[machineConfigCP.Name] = machineConfigCP
	config.VSphereMachineConfigs[machineConfigWorker.Name] = machineConfigWorker

	config.Cluster.Spec.IdentityProviderRefs = []anywherev1.Ref{
		{
			Kind: anywherev1.OIDCConfigKind,
			Name: "my-oidc",
		},
		{
			Kind: anywherev1.AWSIamConfigKind,
			Name: "my-iam",
		},
	}

	oidc := &anywherev1.OIDCConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-oidc",
			Namespace: config.Cluster.Namespace,
		},
	}
	awsIAM := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-iam",
			Namespace: config.Cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       config.Cluster.Name,
				},
			},
		},
	}

	config.AWSIAMConfigs[awsIAM.Name] = awsIAM
	config.OIDCConfigs[oidc.Name] = oidc

	bundles := &releasev1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-bundles-ref",
			Namespace: config.Cluster.Namespace,
		},
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "v1.25",
					PackageController: releasev1.PackageBundle{
						HelmChart: releasev1.Image{},
					},
				},
			},
		},
	}

	config.Cluster.Spec.BundlesRef = &anywherev1.BundlesRef{
		Name:      bundles.Name,
		Namespace: bundles.Namespace,
	}

	return config, bundles
}
