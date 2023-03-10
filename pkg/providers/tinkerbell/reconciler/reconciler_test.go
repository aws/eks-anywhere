package reconciler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	rufiov1alpha1 "github.com/tinkerbell/rufio/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	clusterspec "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/reconciler"
	tinkerbellreconcilermocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	workloadClusterName = "workload-cluster"
	clusterNamespace    = "test-namespace"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	tt := newReconcilerTest(t)

	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tinkHardware("hw1", "cp"))
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tinkHardware("hw2", "worker"))
	tt.createAllObjs()

	logger := test.NewNullLogger()
	remoteClient := env.Client()

	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, tt.buildSpec()).Return(controller.Result{}, nil)

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: workloadClusterName, Namespace: constants.EksaSystemNamespace},
	).Return(remoteClient, nil)
	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, tt.buildSpec())

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.cleanup()
}

func TestReconcilerValidateDatacenterConfigSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	result, err := tt.reconciler().ValidateDatacenterConfig(tt.ctx, test.NewNullLogger(), tt.buildScope())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.cleanup()
}

func TestReconcilerValidateDatacenterConfigMissingManagementCluster(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Spec.ManagementCluster.Name = "nonexistent-management-cluster"
	tt.withFakeClient()

	result, err := tt.reconciler().ValidateDatacenterConfig(tt.ctx, test.NewNullLogger(), tt.buildScope())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(tt.cluster.Status.FailureMessage).To(HaveValue(ContainSubstring("\"nonexistent-management-cluster\" not found")))
	tt.cleanup()
}

func TestReconcilerValidateDatacenterConfigMissingManagementDatacenter(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.managementCluster.Spec.DatacenterRef.Name = "nonexistent-datacenter"
	tt.withFakeClient()

	result, err := tt.reconciler().ValidateDatacenterConfig(tt.ctx, test.NewNullLogger(), tt.buildScope())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(tt.cluster.Status.FailureMessage).To(HaveValue(ContainSubstring("\"nonexistent-datacenter\" not found")))
	tt.cleanup()
}

func TestReconcilerValidateDatacenterConfigIpMismatch(t *testing.T) {
	tt := newReconcilerTest(t)
	managementDatacenterConfig := dataCenter(func(d *anywherev1.TinkerbellDatacenterConfig) {
		d.Name = "ip-mismatch-datacenter"
		d.Spec.TinkerbellIP = "3.3.3.3"
	})
	tt.managementCluster.Spec.DatacenterRef.Name = managementDatacenterConfig.Name
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, managementDatacenterConfig)
	tt.withFakeClient()

	result, err := tt.reconciler().ValidateDatacenterConfig(tt.ctx, test.NewNullLogger(), tt.buildScope())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(tt.cluster.Status.FailureMessage).To(HaveValue(ContainSubstring("workload cluster Tinkerbell IP must match managment cluster Tinkerbell IP")))
	tt.cleanup()
}

func TestReconcileCNISuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	remoteClient := fake.NewClientBuilder().Build()
	scope := tt.buildScope()

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: workloadClusterName, Namespace: constants.EksaSystemNamespace},
	).Return(remoteClient, nil)
	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, scope.ClusterSpec)

	result, err := tt.reconciler().ReconcileCNI(tt.ctx, logger, scope)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.cleanup()
}

func TestReconcileCNIErrorClientRegistry(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	spec := tt.buildScope()

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: workloadClusterName, Namespace: constants.EksaSystemNamespace},
	).Return(nil, errors.New("building client"))

	result, err := tt.reconciler().ReconcileCNI(tt.ctx, logger, spec)

	tt.Expect(err).To(MatchError(ContainSubstring("building client")))
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.cleanup()
}

func TestReconcilerReconcileControlPlaneScaleSuccess(t *testing.T) {
	tt := newReconcilerTest(t)

	tt.createAllObjs()
	scope := tt.buildScope()
	scope.ClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count = 2
	logger := test.NewNullLogger()
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	_, err = tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	_, _ = tt.reconciler().OmitMachineTemplate(tt.ctx, logger, scope)
	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, logger, scope)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	kcp := &controlplanev1.KubeadmControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadClusterName,
			Namespace: constants.EksaSystemNamespace,
		},
	}
	tt.ShouldEventuallyMatch(tt.ctx, kcp,
		func(g Gomega) {
			g.Expect(kcp.Spec.Replicas).To(HaveValue(BeEquivalentTo(2)))
		})
	tt.ShouldEventuallyExist(tt.ctx, controlPlaneMachineTemplate())
	tt.cleanup()
}

func TestReconcilerReconcileControlPlaneSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	scope := tt.buildScope()
	logger := test.NewNullLogger()
	scope.ControlPlane = tinkerbellCP(workloadClusterName)
	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, logger, scope)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx, kubeadmControlPlane())
	tt.ShouldEventuallyExist(tt.ctx, controlPlaneMachineTemplate())

	tt.ShouldEventuallyExist(tt.ctx, &tinkerbellv1.TinkerbellCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TinkerbellCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadClusterName,
			Namespace: constants.EksaSystemNamespace,
		},
	})

	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = workloadClusterName
	})
	tt.ShouldEventuallyExist(tt.ctx, capiCluster)
	tt.ShouldEventuallyNotExist(tt.ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "registry-credentials", Namespace: constants.EksaSystemNamespace}})
	tt.cleanup()
}

func TestReconcilerReconcileControlPlaneSuccessRegistryMirrorAuthentication(t *testing.T) {
	t.Setenv("REGISTRY_USERNAME", "username")
	t.Setenv("REGISTRY_PASSWORD", "password")
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	scope := tt.buildScope()
	scope.ClusterSpec.Cluster.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{
		Authenticate: true,
		Endpoint:     "1.2.3.4",
		Port:         "65536",
	}
	logger := test.NewNullLogger()
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	_, err = tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, logger, scope)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx, kubeadmControlPlane())
	tt.ShouldEventuallyExist(tt.ctx, controlPlaneMachineTemplate())
	tt.ShouldEventuallyExist(tt.ctx, &tinkerbellv1.TinkerbellCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TinkerbellCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadClusterName,
			Namespace: constants.EksaSystemNamespace,
		},
	})

	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = workloadClusterName
	})
	tt.ShouldEventuallyExist(tt.ctx, capiCluster)
	tt.ShouldEventuallyExist(tt.ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "registry-credentials", Namespace: constants.EksaSystemNamespace}})
	tt.cleanup()
}

func TestReconcilerReconcileControlPlaneFailure(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	scope := tt.buildScope()
	scope.ClusterSpec.Cluster = scope.ClusterSpec.Cluster.DeepCopy()
	scope.ClusterSpec.Cluster.Name = ""
	logger := test.NewNullLogger()
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())

	_, err = tt.reconciler().ReconcileControlPlane(tt.ctx, logger, scope)

	tt.Expect(err).To(MatchError(ContainSubstring("resource name may not be empty")))
	tt.cleanup()
}

func TestReconcilerValidateClusterSpecInvalidDatacenterConfig(t *testing.T) {
	tt := newReconcilerTest(t)

	logger := test.NewNullLogger()

	tt.cluster.Name = "invalidCluster"
	tt.cluster.Spec.KubernetesVersion = "1.22"
	tt.datacenterConfig.Spec.TinkerbellIP = ""

	tt.withFakeClient()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, gomock.Any()).Return(controller.Result{}, nil)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("missing spec.tinkerbellIP field"))
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.cleanup()
}

func TestReconcilerValidateClusterSpecInvalidOSFamily(t *testing.T) {
	tt := newReconcilerTest(t)

	logger := test.NewNullLogger()

	tt.cluster.Name = "invalidCluster"
	tt.machineConfigWorker.Spec.OSFamily = "invalidOS"

	tt.withFakeClient()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, gomock.Any()).Return(controller.Result{}, nil)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("unsupported spec.osFamily (invalidOS); Please use one of the following: ubuntu, redhat, bottlerocket"))
	tt.cleanup()
}

func TestReconcilerReconcileWorkerNodesSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	tt.cluster.SetSelfManaged()
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tinkHardware("hw1", "cp"))
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tinkHardware("hw2", "worker"))
	tt.createAllObjs()

	logger := test.NewNullLogger()

	result, err := tt.reconciler().ReconcileWorkerNodes(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx,
		&bootstrapv1.KubeadmConfigTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      capiCluster.Name + "-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&tinkerbellv1.TinkerbellMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      capiCluster.Name + "-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&clusterv1.MachineDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      capiCluster.Name + "-md-0",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.cleanup()
}

func TestReconcilerReconcileWorkersScaleSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	mt := &tinkerbellv1.TinkerbellMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tt.cluster.Name + "-md-0-1",
			Namespace: constants.EksaSystemNamespace,
		},
	}
	tt.createAllObjs()

	scope := tt.buildScope()
	scope.ClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(2)
	logger := test.NewNullLogger()
	scope.ControlPlane = tinkerbellCP(tt.cluster.Name)
	scope.Workers = tinkWorker(tt.cluster.Name, func(w *tinkerbell.Workers) {
		w.Groups[0].MachineDeployment.Spec.Replicas = ptr.Int32(2)
	})
	_, err := tt.reconciler().OmitMachineTemplate(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	result, err := tt.reconciler().ReconcileWorkers(tt.ctx, logger, scope)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx, mt)
	md := &clusterv1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tt.cluster.Name + "-md-0",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: clusterv1.MachineDeploymentSpec{
			Replicas: ptr.Int32(2),
		},
	}
	tt.ShouldEventuallyMatch(tt.ctx, md,
		func(g Gomega) {
			g.Expect(md.Spec.Replicas).To(HaveValue(BeEquivalentTo(2)))
		})
	tt.cleanup()
}

func TestReconcilerReconcileWorkersSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.createAllObjs()

	scope := tt.buildScope()
	logger := test.NewNullLogger()
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	_, err = tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	result, err := tt.reconciler().ReconcileWorkers(tt.ctx, logger, scope)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx,
		&clusterv1.MachineDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-md-0",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&bootstrapv1.KubeadmConfigTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&tinkerbellv1.TinkerbellMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.cleanup()
}

func TestReconcilerReconcileWorkerNodesFailure(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	tt.cluster.SetSelfManaged()
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.cluster.Spec.KubernetesVersion = ""
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.createAllObjs()

	logger := test.NewNullLogger()

	_, err := tt.reconciler().ReconcileWorkerNodes(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(MatchError(ContainSubstring("building cluster Spec for worker node reconcile")))
	tt.cleanup()
}

func TestReconcilerValidateHardwareCountNewClusterFail(t *testing.T) {
	tt := newReconcilerTest(t)
	logger := test.NewNullLogger()

	tt.cluster.Name = "invalidCluster"
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tinkHardware("hw1", "cp"))

	tt.withFakeClient()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, gomock.Any()).Return(controller.Result{}, nil)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("minimum hardware count not met for selector '{\"type\":\"worker\"}': have 0, require 1"))
	tt.cleanup()
}

func TestReconcilerValidateHardwareCountRollingUpdateFail(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	scope := tt.buildScope()
	scope.ClusterSpec.VersionsBundle.KubeDistro.Kubernetes.Tag = "1.23"
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	_, err = tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	result, err := tt.reconciler().ValidateHardware(tt.ctx, logger, scope)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("minimum hardware count not met for selector"))
	tt.cleanup()
}

func TestReconcilerValidateHardwareScalingUpdateFail(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	scope := tt.buildScope()
	scope.ClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(2)
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	op, err := tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(op).To(Equal(reconciler.ScaleOperation))
	result, err := tt.reconciler().ValidateHardware(tt.ctx, logger, scope)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("minimum hardware count not met for selector '{\"type\":\"worker\"}': have 0, require 1"))
	tt.cleanup()
}

func TestReconcilerValidateHardwareNoHardware(t *testing.T) {
	tt := newReconcilerTest(t)
	logger := test.NewNullLogger()

	tt.cluster.Name = "invalidCluster"
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tinkHardware("hw1", "worker"))

	tt.withFakeClient()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, gomock.Any()).Return(controller.Result{}, nil)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("minimum hardware count not met for selector '{\"type\":\"cp\"}': have 0, require 1"))
	tt.cleanup()
}

func TestReconcilerValidateRufioMachinesFail(t *testing.T) {
	tt := newReconcilerTest(t)
	logger := test.NewNullLogger()

	tt.cluster.Name = "invalidCluster"
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, &rufiov1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bmc0",
			Namespace: constants.EksaSystemNamespace,
		},
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, &rufiov1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bmc1",
			Namespace: constants.EksaSystemNamespace,
		},
		Status: rufiov1alpha1.MachineStatus{
			Conditions: []rufiov1alpha1.MachineCondition{
				{
					Type:   rufiov1alpha1.Contactable,
					Status: rufiov1alpha1.ConditionTrue,
				},
			},
		},
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, &rufiov1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bmc2",
			Namespace: constants.EksaSystemNamespace,
		},
		Status: rufiov1alpha1.MachineStatus{
			Conditions: []rufiov1alpha1.MachineCondition{
				{
					Type:    rufiov1alpha1.Contactable,
					Status:  rufiov1alpha1.ConditionFalse,
					Message: "bmc connection failure",
				},
			},
		},
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tinkHardware("hw1", "cp"))
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tinkHardware("hw2", "worker"))

	tt.withFakeClient()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, gomock.Any()).Return(controller.Result{}, nil)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("bmc connection failure"))
	tt.cleanup()
}

func TestReconcilerGenerateSpec(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	logger := test.NewNullLogger()
	scope := tt.buildScope()
	result, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(scope.ControlPlane).To(Equal(tinkerbellCP(workloadClusterName)))
	tt.Expect(scope.Workers).To(Equal(tinkWorker(workloadClusterName)))
	tt.cleanup()
}

func TestReconciler_DetectOperationK8sVersionUpgrade(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	scope := tt.buildScope()
	scope.ClusterSpec.VersionsBundle.KubeDistro.Kubernetes.Tag = "1.23"
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	op, err := tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(op).To(Equal(reconciler.K8sVersionUpgradeOperation))
	tt.cleanup()
}

func TestReconciler_DetectOperationExistingWorkerNodeGroupScaleUpdate(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	scope := tt.buildScope()
	scope.ClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(2)
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	op, err := tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(op).To(Equal(reconciler.ScaleOperation))
	tt.cleanup()
}

func TestReconciler_DetectOperationNewWorkerNodeGroupScaleUpdate(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	scope := tt.buildScope()
	scope.ClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = append(scope.ClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations,
		anywherev1.WorkerNodeGroupConfiguration{
			Count: ptr.Int(1),
			MachineGroupRef: &anywherev1.Ref{
				Kind: anywherev1.TinkerbellMachineConfigKind,
				Name: tt.machineConfigWorker.Name,
			},
			Name:   "md-1",
			Labels: nil,
		},
	)
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	op, err := tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(op).To(Equal(reconciler.ScaleOperation))
	tt.cleanup()
}

func TestReconciler_DetectOperationNoChanges(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	scope := tt.buildScope()
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	op, err := tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).To(BeNil())
	tt.Expect(op).To(Equal(reconciler.NoChange))
	tt.cleanup()
}

func TestReconciler_DetectOperationNewCluster(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	logger := test.NewNullLogger()
	scope := tt.buildScope()
	scope.ClusterSpec.Cluster.Name = "new-cluster"
	_, err := tt.reconciler().GenerateSpec(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	op, err := tt.reconciler().DetectOperation(tt.ctx, logger, scope)
	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(op).To(Equal(reconciler.NewClusterOperation))
	tt.cleanup()
}

func TestReconciler_DetectOperationFail(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.client = fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()
	_, err := tt.reconciler().DetectOperation(tt.ctx, test.NewNullLogger(), &reconciler.Scope{ClusterSpec: &clusterspec.Spec{Config: &clusterspec.Config{Cluster: &anywherev1.Cluster{}}}})
	tt.Expect(err).To(MatchError(ContainSubstring("no kind is registered for the type")))
}

func TestWorkerReplicasDiffFail(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.client = fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()
	scope := &reconciler.Scope{
		ClusterSpec: &clusterspec.Spec{
			Config: &clusterspec.Config{
				Cluster: &anywherev1.Cluster{
					Spec: anywherev1.ClusterSpec{
						WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
							{
								Name: "md-0",
							},
						},
					},
				},
			},
		},
	}

	_, err := tt.reconciler().WorkerReplicasDiff(tt.ctx, scope)
	tt.Expect(err).To(MatchError(ContainSubstring("no kind is registered for the type")))
}

func (tt *reconcilerTest) withFakeClient() {
	tt.client = fake.NewClientBuilder().WithObjects(clientutil.ObjectsToClientObjects(tt.allObjs())...).Build()
}

func (tt *reconcilerTest) reconciler() *reconciler.Reconciler {
	return reconciler.New(tt.client, tt.cniReconciler, tt.remoteClientRegistry, tt.ipValidator)
}

func (tt *reconcilerTest) buildScope() *reconciler.Scope {
	tt.t.Helper()
	spec, err := clusterspec.BuildSpec(tt.ctx, clientutil.NewKubeClient(tt.client), tt.cluster)
	tt.Expect(err).NotTo(HaveOccurred())

	return reconciler.NewScope(spec)
}

func (tt *reconcilerTest) buildSpec() *clusterspec.Spec {
	tt.t.Helper()
	spec, err := clusterspec.BuildSpec(tt.ctx, clientutil.NewKubeClient(tt.client), tt.cluster)
	tt.Expect(err).NotTo(HaveOccurred())

	return spec
}

func (tt *reconcilerTest) createAllObjs() {
	tt.t.Helper()
	envtest.CreateObjs(tt.ctx, tt.t, tt.client, tt.allObjs()...)
}

func (tt *reconcilerTest) allObjs() []client.Object {
	objs := make([]client.Object, 0, len(tt.eksaSupportObjs)+3)
	objs = append(objs, tt.eksaSupportObjs...)
	objs = append(objs, tt.cluster, tt.machineConfigControlPlane, tt.machineConfigWorker, tt.managementCluster)

	return objs
}

type reconcilerTest struct {
	t testing.TB
	*WithT
	*envtest.APIExpecter
	ctx                       context.Context
	cluster                   *anywherev1.Cluster
	managementCluster         *anywherev1.Cluster
	client                    client.Client
	eksaSupportObjs           []client.Object
	datacenterConfig          *anywherev1.TinkerbellDatacenterConfig
	machineConfigControlPlane *anywherev1.TinkerbellMachineConfig
	machineConfigWorker       *anywherev1.TinkerbellMachineConfig
	ipValidator               *tinkerbellreconcilermocks.MockIPValidator
	cniReconciler             *tinkerbellreconcilermocks.MockCNIReconciler
	remoteClientRegistry      *tinkerbellreconcilermocks.MockRemoteClientRegistry
}

func newReconcilerTest(t testing.TB) *reconcilerTest {
	ctrl := gomock.NewController(t)
	c := env.Client()

	cniReconciler := tinkerbellreconcilermocks.NewMockCNIReconciler(ctrl)
	remoteClientRegistry := tinkerbellreconcilermocks.NewMockRemoteClientRegistry(ctrl)
	ipValidator := tinkerbellreconcilermocks.NewMockIPValidator(ctrl)

	bundle := test.Bundle()

	managementClusterDatacenter := dataCenter(func(d *anywherev1.TinkerbellDatacenterConfig) {
		d.Name = "management-datacenter"
	})

	managementCluster := tinkerbellCluster(func(c *anywherev1.Cluster) {
		c.Name = "management-cluster"
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: c.Name,
		}
		c.Spec.BundlesRef = &anywherev1.BundlesRef{
			Name:       bundle.Name,
			Namespace:  bundle.Namespace,
			APIVersion: bundle.APIVersion,
		}
		c.Spec.DatacenterRef = anywherev1.Ref{
			Kind: anywherev1.TinkerbellDatacenterKind,
			Name: managementClusterDatacenter.Name,
		}
	})

	machineConfigCP := machineConfig(func(m *anywherev1.TinkerbellMachineConfig) {
		m.Name = "cp-machine-config"
		m.Spec.HardwareSelector = anywherev1.HardwareSelector{"type": "cp"}
	})
	machineConfigWN := machineConfig(func(m *anywherev1.TinkerbellMachineConfig) {
		m.Name = "worker-machine-config"
		m.Spec.HardwareSelector = anywherev1.HardwareSelector{"type": "worker"}
	})

	workloadClusterDatacenter := dataCenter(func(d *anywherev1.TinkerbellDatacenterConfig) {})

	cluster := tinkerbellCluster(func(c *anywherev1.Cluster) {
		c.Name = workloadClusterName
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: managementCluster.Name,
		}
		c.Spec.BundlesRef = &anywherev1.BundlesRef{
			Name:       bundle.Name,
			Namespace:  bundle.Namespace,
			APIVersion: bundle.APIVersion,
		}
		c.Spec.ControlPlaneConfiguration = anywherev1.ControlPlaneConfiguration{
			Count: 1,
			Endpoint: &anywherev1.Endpoint{
				Host: "1.1.1.1",
			},
			MachineGroupRef: &anywherev1.Ref{
				Kind: anywherev1.TinkerbellMachineConfigKind,
				Name: machineConfigCP.Name,
			},
		}
		c.Spec.DatacenterRef = anywherev1.Ref{
			Kind: anywherev1.TinkerbellDatacenterKind,
			Name: workloadClusterDatacenter.Name,
		}

		c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations,
			anywherev1.WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.TinkerbellMachineConfigKind,
					Name: machineConfigWN.Name,
				},
				Name:   "md-0",
				Labels: nil,
			},
		)
	})

	tt := &reconcilerTest{
		t:                    t,
		WithT:                NewWithT(t),
		APIExpecter:          envtest.NewAPIExpecter(t, c),
		ctx:                  context.Background(),
		ipValidator:          ipValidator,
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		client:               c,
		eksaSupportObjs: []client.Object{
			test.Namespace(clusterNamespace),
			test.Namespace(constants.EksaSystemNamespace),
			workloadClusterDatacenter,
			managementClusterDatacenter,
			bundle,
			test.EksdRelease(),
		},
		cluster:                   cluster,
		managementCluster:         managementCluster,
		datacenterConfig:          workloadClusterDatacenter,
		machineConfigControlPlane: machineConfigCP,
		machineConfigWorker:       machineConfigWN,
	}

	t.Cleanup(tt.cleanup)
	return tt
}

func (tt *reconcilerTest) cleanup() {
	tt.DeleteAndWait(tt.ctx, tt.allObjs()...)
	tt.DeleteAllOfAndWait(tt.ctx, &bootstrapv1.KubeadmConfigTemplate{})
	tt.DeleteAllOfAndWait(tt.ctx, &clusterv1.Cluster{})
	tt.DeleteAllOfAndWait(tt.ctx, &controlplanev1.KubeadmControlPlane{})
	tt.DeleteAllOfAndWait(tt.ctx, &tinkerbellv1.TinkerbellCluster{})
	tt.DeleteAllOfAndWait(tt.ctx, &tinkerbellv1.TinkerbellMachineTemplate{})
	tt.DeleteAllOfAndWait(tt.ctx, &clusterv1.MachineDeployment{})
}

type clusterOpt func(*anywherev1.Cluster)

func tinkerbellCluster(opts ...clusterOpt) *anywherev1.Cluster {
	c := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterNamespace,
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.22",
			ClusterNetwork: anywherev1.ClusterNetwork{
				Pods: anywherev1.Pods{
					CidrBlocks: []string{"0.0.0.0"},
				},
				Services: anywherev1.Services{
					CidrBlocks: []string{"0.0.0.0"},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type datacenterOpt func(config *anywherev1.TinkerbellDatacenterConfig)

func dataCenter(opts ...datacenterOpt) *anywherev1.TinkerbellDatacenterConfig {
	d := &anywherev1.TinkerbellDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.TinkerbellDatacenterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: clusterNamespace,
		},
		Spec: anywherev1.TinkerbellDatacenterConfigSpec{
			TinkerbellIP: "2.2.2.2",
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

type tinkerbellMachineOpt func(config *anywherev1.TinkerbellMachineConfig)

func machineConfig(opts ...tinkerbellMachineOpt) *anywherev1.TinkerbellMachineConfig {
	m := &anywherev1.TinkerbellMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.TinkerbellMachineConfigKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterNamespace,
		},
		Spec: anywherev1.TinkerbellMachineConfigSpec{
			OSFamily: "bottlerocket",
			HardwareSelector: anywherev1.HardwareSelector{
				"key": "cp",
			},
			Users: []anywherev1.UserConfiguration{
				{
					Name:              "user",
					SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8ZEibIrz1AUBKDvmDiWLs9f5DnOerC4qPITiDtSOuPAsxgZbRMavBfVTxodMdAkYRYlXxK6PqNo0ve0qcOV2yvpxH1OogasMMetck6BlM/dIoo3vEY4ZoG9DuVRIf9Iry5gJKbpMDYWpx1IGZrDMOFcIM20ii2qLQQk5hfq9OqdqhToEJFixdgJt/y/zt6Koy3kix+XsnrVdAHgWAq4CZuwt1G6JUAqrpob3H8vPmL7aS+35ktf0pHBm6nYoxRhslnWMUb/7vpzWiq+fUBIm2LYqvrnm7t3fRqFx7p2sZqAm2jDNivyYXwRXkoQPR96zvGeMtuQ5BVGPpsDfVudSW21+pEXHI0GINtTbua7Ogz7wtpVywSvHraRgdFOeY9mkXPzvm2IhoqNrteck2GErwqSqb19mPz6LnHueK0u7i6WuQWJn0CUoCtyMGIrowXSviK8qgHXKrmfTWATmCkbtosnLskNdYuOw8bKxq5S4WgdQVhPps2TiMSZ bottlerocket@ip-10-2-0-6"},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func kubeadmControlPlane() *controlplanev1.KubeadmControlPlane {
	return &controlplanev1.KubeadmControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadClusterName,
			Namespace: constants.EksaSystemNamespace,
		},
	}
}

func controlPlaneMachineTemplate() *tinkerbellv1.TinkerbellMachineTemplate {
	return &tinkerbellv1.TinkerbellMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workloadClusterName + "-control-plane-1",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: tinkerbellv1.TinkerbellMachineTemplateSpec{
			Template: tinkerbellv1.TinkerbellMachineTemplateResource{
				Spec: tinkerbellv1.TinkerbellMachineSpec{
					HardwareAffinity: &tinkerbellv1.HardwareAffinity{
						Required: []tinkerbellv1.HardwareAffinityTerm{
							{
								LabelSelector: metav1.LabelSelector{MatchLabels: map[string]string{}},
							},
						},
					},
				},
			},
		},
	}
}

func tinkHardware(hardwareName, labelType string) *tinkv1alpha1.Hardware {
	return &tinkv1alpha1.Hardware{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Hardware",
			APIVersion: "tinkerbell.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hardwareName,
			Namespace: constants.EksaSystemNamespace,
			Labels: map[string]string{
				"type": labelType,
			},
		},
		Spec: tinkv1alpha1.HardwareSpec{
			Metadata: &tinkv1alpha1.HardwareMetadata{
				Instance: &tinkv1alpha1.MetadataInstance{
					ID: "foo",
				},
			},
		},
	}
}

type cpOpt func(plane *tinkerbell.ControlPlane)

func tinkerbellCP(clusterName string, opts ...cpOpt) *tinkerbell.ControlPlane {
	cp := &tinkerbell.ControlPlane{
		BaseControlPlane: tinkerbell.BaseControlPlane{
			Cluster: &clusterv1.Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Cluster",
					APIVersion: "cluster.x-k8s.io/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName,
					Namespace: constants.EksaSystemNamespace,
					Labels:    map[string]string{"cluster.x-k8s.io/cluster-name": workloadClusterName},
				},
				Spec: clusterv1.ClusterSpec{
					ClusterNetwork: &clusterv1.ClusterNetwork{
						Services: &clusterv1.NetworkRanges{
							CIDRBlocks: []string{"0.0.0.0"},
						},
						Pods: &clusterv1.NetworkRanges{
							CIDRBlocks: []string{"0.0.0.0"},
						},
					},
					ControlPlaneEndpoint: clusterv1.APIEndpoint{
						Host: "1.1.1.1",
						Port: 6443,
					},
					ControlPlaneRef: &corev1.ObjectReference{
						Kind:       "KubeadmControlPlane",
						Name:       clusterName,
						APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
					},
					InfrastructureRef: &corev1.ObjectReference{
						Kind:       "TinkerbellCluster",
						Name:       clusterName,
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
					},
				},
				Status: clusterv1.ClusterStatus{},
			},
			KubeadmControlPlane: &controlplanev1.KubeadmControlPlane{
				TypeMeta: metav1.TypeMeta{
					Kind:       "KubeadmControlPlane",
					APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName,
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: controlplanev1.KubeadmControlPlaneSpec{
					Replicas: ptr.Int32(1),
					Version:  "v1.19.8",
					MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
						InfrastructureRef: corev1.ObjectReference{
							Kind:       "TinkerbellMachineTemplate",
							Name:       "workload-cluster-control-plane-1",
							APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
						},
					},
					KubeadmConfigSpec: bootstrapv1.KubeadmConfigSpec{
						ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
							ImageRepository: "public.ecr.aws/eks-distro/kubernetes",
							Etcd: bootstrapv1.Etcd{
								Local: &bootstrapv1.LocalEtcd{
									ImageMeta: bootstrapv1.ImageMeta{
										ImageRepository: "",
										ImageTag:        "",
									},
									DataDir:        "",
									ExtraArgs:      nil,
									ServerCertSANs: nil,
									PeerCertSANs:   nil,
								},
							},
							ControllerManager: bootstrapv1.ControlPlaneComponent{
								ExtraVolumes: []bootstrapv1.HostPathMount{
									{
										Name:      "kubeconfig",
										HostPath:  "/var/lib/kubeadm/controller-manager.conf",
										MountPath: "/etc/kubernetes/controller-manager.conf",
										ReadOnly:  true,
										PathType:  "File",
									},
								},
							},
							Scheduler: bootstrapv1.ControlPlaneComponent{
								ExtraVolumes: []bootstrapv1.HostPathMount{
									{
										Name:      "kubeconfig",
										HostPath:  "/var/lib/kubeadm/scheduler.conf",
										MountPath: "/etc/kubernetes/scheduler.conf",
										ReadOnly:  true,
										PathType:  "File",
									},
								},
							},
							CertificatesDir: "/var/lib/kubeadm/pki",
						},
						InitConfiguration: &bootstrapv1.InitConfiguration{
							NodeRegistration: bootstrapv1.NodeRegistrationOptions{
								KubeletExtraArgs: map[string]string{
									"read-only-port":    "0",
									"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
									"anonymous-auth":    "false",
									"provider-id":       "PROVIDER_ID",
								},
							},
						},
						JoinConfiguration: &bootstrapv1.JoinConfiguration{
							NodeRegistration: bootstrapv1.NodeRegistrationOptions{
								KubeletExtraArgs: map[string]string{
									"anonymous-auth":    "false",
									"provider-id":       "PROVIDER_ID",
									"read-only-port":    "0",
									"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								},
								IgnorePreflightErrors: []string{"DirAvailable--etc-kubernetes-manifests"},
							},
							CACertPath:                            "",
							Discovery:                             bootstrapv1.Discovery{},
							ControlPlane:                          nil,
							SkipPhases:                            nil,
							Patches:                               nil,
							BottlerocketCustomHostContainers:      nil,
							BottlerocketCustomBootstrapContainers: nil,
						},
						Files: []bootstrapv1.File{
							{
								Path:        "/etc/kubernetes/manifests/kube-vip.yaml",
								Owner:       "root:root",
								Permissions: "",
								Encoding:    "",
								Append:      false,
								Content:     "apiVersion: v1\nkind: Pod\nmetadata:\n  creationTimestamp: null\n  name: kube-vip\n  namespace: kube-system\nspec:\n  containers:\n  - args:\n    - manager\n    env:\n    - name: vip_arp\n      value: \"true\"\n    - name: port\n      value: \"6443\"\n    - name: vip_cidr\n      value: \"32\"\n    - name: cp_enable\n      value: \"true\"\n    - name: cp_namespace\n      value: kube-system\n    - name: vip_ddns\n      value: \"false\"\n    - name: vip_leaderelection\n      value: \"true\"\n    - name: vip_leaseduration\n      value: \"15\"\n    - name: vip_renewdeadline\n      value: \"10\"\n    - name: vip_retryperiod\n      value: \"2\"\n    - name: address\n      value: 1.1.1.1\n    image: \n    imagePullPolicy: IfNotPresent\n    name: kube-vip\n    resources: {}\n    securityContext:\n      capabilities:\n        add:\n        - NET_ADMIN\n        - NET_RAW\n    volumeMounts:\n    - mountPath: /etc/kubernetes/admin.conf\n      name: kubeconfig\n  hostNetwork: true\n  volumes:\n  - hostPath:\n      path: /etc/kubernetes/admin.conf\n    name: kubeconfig\nstatus: {}\n",
								ContentFrom: nil,
							},
						},
						Users: []bootstrapv1.User{
							{
								Name:              "user",
								Sudo:              ptr.String("ALL=(ALL) NOPASSWD:ALL"),
								SSHAuthorizedKeys: []string{"[ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8ZEibIrz1AUBKDvmDiWLs9f5DnOerC4qPITiDtSOuPAsxgZbRMavBfVTxodMdAkYRYlXxK6PqNo0ve0qcOV2yvpxH1OogasMMetck6BlM/dIoo3vEY4ZoG9DuVRIf9Iry5gJKbpMDYWpx1IGZrDMOFcIM20ii2qLQQk5hfq9OqdqhToEJFixdgJt/y/zt6Koy3kix+XsnrVdAHgWAq4CZuwt1G6JUAqrpob3H8vPmL7aS+35ktf0pHBm6nYoxRhslnWMUb/7vpzWiq+fUBIm2LYqvrnm7t3fRqFx7p2sZqAm2jDNivyYXwRXkoQPR96zvGeMtuQ5BVGPpsDfVudSW21+pEXHI0GINtTbua7Ogz7wtpVywSvHraRgdFOeY9mkXPzvm2IhoqNrteck2GErwqSqb19mPz6LnHueK0u7i6WuQWJn0CUoCtyMGIrowXSviK8qgHXKrmfTWATmCkbtosnLskNdYuOw8bKxq5S4WgdQVhPps2TiMSZ bottlerocket@ip-10-2-0-6]"},
							},
						},
						Format: "bottlerocket",
					},
				},
			},
			ProviderCluster: &tinkerbellv1.TinkerbellCluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TinkerbellCluster",
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterName,
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: tinkerbellv1.TinkerbellClusterSpec{
					ImageLookupFormat:       "--kube-v1.19.8.raw.gz",
					ImageLookupBaseRegistry: "/",
				},
			},
			ControlPlaneMachineTemplate: &tinkerbellv1.TinkerbellMachineTemplate{
				TypeMeta: metav1.TypeMeta{
					Kind:       "TinkerbellMachineTemplate",
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "workload-cluster-control-plane-1",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: tinkerbellv1.TinkerbellMachineTemplateSpec{
					Template: tinkerbellv1.TinkerbellMachineTemplateResource{
						Spec: tinkerbellv1.TinkerbellMachineSpec{
							TemplateOverride: "global_timeout: 6000\nid: \"\"\nname: workload-cluster\ntasks:\n- actions:\n  - environment:\n      COMPRESSED: \"true\"\n      DEST_DISK: '{{ index .Hardware.Disks 0 }}'\n      IMG_URL: \"\"\n    image: \"\"\n    name: stream-image\n    timeout: 600\n  - environment:\n      BOOTCONFIG_CONTENTS: kernel {}\n      DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}'\n      DEST_PATH: /bootconfig.data\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0644\"\n      UID: \"0\"\n    image: \"\"\n    name: write-bootconfig\n    pid: host\n    timeout: 90\n  - environment:\n      DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}'\n      DEST_PATH: /user-data.toml\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      HEGEL_URLS: http://2.2.2.2:50061,http://2.2.2.2:50061\n      MODE: \"0644\"\n      UID: \"0\"\n    image: \"\"\n    name: write-user-data\n    pid: host\n    timeout: 90\n  - environment:\n      DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}'\n      DEST_PATH: /net.toml\n      DIRMODE: \"0755\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      IFNAME: eno1\n      MODE: \"0644\"\n      STATIC_BOTTLEROCKET: \"true\"\n      UID: \"0\"\n    image: \"\"\n    name: write-netplan\n    pid: host\n    timeout: 90\n  - image: \"\"\n    name: reboot-image\n    pid: host\n    timeout: 90\n    volumes:\n    - /worker:/worker\n  name: workload-cluster\n  volumes:\n  - /dev:/dev\n  - /dev/console:/dev/console\n  - /lib/firmware:/lib/firmware:ro\n  worker: '{{.device_1}}'\nversion: \"0.1\"\n",
							HardwareAffinity: &tinkerbellv1.HardwareAffinity{
								Required: []tinkerbellv1.HardwareAffinityTerm{
									{LabelSelector: metav1.LabelSelector{
										MatchLabels: map[string]string{"type": "cp"},
									}},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, opt := range opts {
		opt(cp)
	}
	return cp
}

type workerOpt func(*tinkerbell.Workers)

func tinkWorker(clusterName string, opts ...workerOpt) *tinkerbell.Workers {
	w := &tinkerbell.Workers{
		Groups: []clusterapi.WorkerGroup[*tinkerbellv1.TinkerbellMachineTemplate]{
			{
				KubeadmConfigTemplate: &bootstrapv1.KubeadmConfigTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "KubeadmConfigTemplate",
						APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterName + "-md-0-1",
						Namespace: constants.EksaSystemNamespace,
					},
					Spec: bootstrapv1.KubeadmConfigTemplateSpec{
						Template: bootstrapv1.KubeadmConfigTemplateResource{
							Spec: bootstrapv1.KubeadmConfigSpec{
								Users: []bootstrapv1.User{
									{
										Name:              "user",
										Sudo:              ptr.String("ALL=(ALL) NOPASSWD:ALL"),
										SSHAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8ZEibIrz1AUBKDvmDiWLs9f5DnOerC4qPITiDtSOuPAsxgZbRMavBfVTxodMdAkYRYlXxK6PqNo0ve0qcOV2yvpxH1OogasMMetck6BlM/dIoo3vEY4ZoG9DuVRIf9Iry5gJKbpMDYWpx1IGZrDMOFcIM20ii2qLQQk5hfq9OqdqhToEJFixdgJt/y/zt6Koy3kix+XsnrVdAHgWAq4CZuwt1G6JUAqrpob3H8vPmL7aS+35ktf0pHBm6nYoxRhslnWMUb/7vpzWiq+fUBIm2LYqvrnm7t3fRqFx7p2sZqAm2jDNivyYXwRXkoQPR96zvGeMtuQ5BVGPpsDfVudSW21+pEXHI0GINtTbua7Ogz7wtpVywSvHraRgdFOeY9mkXPzvm2IhoqNrteck2GErwqSqb19mPz6LnHueK0u7i6WuQWJn0CUoCtyMGIrowXSviK8qgHXKrmfTWATmCkbtosnLskNdYuOw8bKxq5S4WgdQVhPps2TiMSZ bottlerocket@ip-10-2-0-6"},
									},
								},
								Format: "bottlerocket",
								JoinConfiguration: &bootstrapv1.JoinConfiguration{
									Pause: bootstrapv1.Pause{
										ImageMeta: bootstrapv1.ImageMeta{
											ImageRepository: "",
											ImageTag:        "",
										},
									},
									BottlerocketBootstrap: bootstrapv1.BottlerocketBootstrap{
										ImageMeta: bootstrapv1.ImageMeta{
											ImageRepository: "",
											ImageTag:        "",
										},
									},
									BottlerocketAdmin: bootstrapv1.BottlerocketAdmin{
										ImageMeta: bootstrapv1.ImageMeta{
											ImageRepository: "",
											ImageTag:        "",
										},
									},
									BottlerocketControl: bootstrapv1.BottlerocketControl{
										ImageMeta: bootstrapv1.ImageMeta{
											ImageRepository: "",
											ImageTag:        "",
										},
									},
									NodeRegistration: bootstrapv1.NodeRegistrationOptions{
										KubeletExtraArgs: map[string]string{
											"anonymous-auth":    "false",
											"provider-id":       "PROVIDER_ID",
											"read-only-port":    "0",
											"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
										},
									},
								},
							},
						},
					},
				},
				MachineDeployment: &clusterv1.MachineDeployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "MachineDeployment",
						APIVersion: "cluster.x-k8s.io/v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterName + "-md-0",
						Namespace: constants.EksaSystemNamespace,
						Labels: map[string]string{
							"pool":                          "md-0",
							"cluster.x-k8s.io/cluster-name": clusterName,
						},
					},
					Spec: clusterv1.MachineDeploymentSpec{
						ClusterName: workloadClusterName,
						Replicas:    ptr.Int32(1),
						Selector: metav1.LabelSelector{
							MatchLabels: map[string]string{},
						},
						Template: clusterv1.MachineTemplateSpec{
							ObjectMeta: clusterv1.ObjectMeta{
								Labels: map[string]string{
									"pool":                          "md-0",
									"cluster.x-k8s.io/cluster-name": clusterName,
								},
							},
							Spec: clusterv1.MachineSpec{
								ClusterName: clusterName,
								Bootstrap: clusterv1.Bootstrap{
									ConfigRef: &corev1.ObjectReference{
										Kind:       "KubeadmConfigTemplate",
										Name:       clusterName + "-md-0-1",
										APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
									},
								},
								InfrastructureRef: corev1.ObjectReference{
									Kind:       "TinkerbellMachineTemplate",
									Name:       clusterName + "-md-0-1",
									APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
								},
								Version: ptr.String("v1.19.8"),
							},
						},
					},
				},
				ProviderMachineTemplate: &tinkerbellv1.TinkerbellMachineTemplate{
					TypeMeta: metav1.TypeMeta{
						Kind:       "TinkerbellMachineTemplate",
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterName + "-md-0-1",
						Namespace: constants.EksaSystemNamespace,
					},
					Spec: tinkerbellv1.TinkerbellMachineTemplateSpec{
						Template: tinkerbellv1.TinkerbellMachineTemplateResource{Spec: tinkerbellv1.TinkerbellMachineSpec{
							TemplateOverride: "global_timeout: 6000\nid: \"\"\nname: " + clusterName + "\ntasks:\n- actions:\n  - environment:\n      COMPRESSED: \"true\"\n      DEST_DISK: '{{ index .Hardware.Disks 0 }}'\n      IMG_URL: \"\"\n    image: \"\"\n    name: stream-image\n    timeout: 600\n  - environment:\n      BOOTCONFIG_CONTENTS: kernel {}\n      DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}'\n      DEST_PATH: /bootconfig.data\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      MODE: \"0644\"\n      UID: \"0\"\n    image: \"\"\n    name: write-bootconfig\n    pid: host\n    timeout: 90\n  - environment:\n      DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}'\n      DEST_PATH: /user-data.toml\n      DIRMODE: \"0700\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      HEGEL_URLS: http://2.2.2.2:50061,http://2.2.2.2:50061\n      MODE: \"0644\"\n      UID: \"0\"\n    image: \"\"\n    name: write-user-data\n    pid: host\n    timeout: 90\n  - environment:\n      DEST_DISK: '{{ formatPartition ( index .Hardware.Disks 0 ) 12 }}'\n      DEST_PATH: /net.toml\n      DIRMODE: \"0755\"\n      FS_TYPE: ext4\n      GID: \"0\"\n      IFNAME: eno1\n      MODE: \"0644\"\n      STATIC_BOTTLEROCKET: \"true\"\n      UID: \"0\"\n    image: \"\"\n    name: write-netplan\n    pid: host\n    timeout: 90\n  - image: \"\"\n    name: reboot-image\n    pid: host\n    timeout: 90\n    volumes:\n    - /worker:/worker\n  name: workload-cluster\n  volumes:\n  - /dev:/dev\n  - /dev/console:/dev/console\n  - /lib/firmware:/lib/firmware:ro\n  worker: '{{.device_1}}'\nversion: \"0.1\"\n",
							HardwareAffinity: &tinkerbellv1.HardwareAffinity{
								Required: []tinkerbellv1.HardwareAffinityTerm{
									{
										LabelSelector: metav1.LabelSelector{
											MatchLabels: map[string]string{"type": "worker"},
										},
									},
								},
							},
						}},
					},
				},
			},
		},
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}
