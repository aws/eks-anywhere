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
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/reconciler"
	tinkerbellreconcilermocks "github.com/aws/eks-anywhere/pkg/providers/tinkerbell/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	workloadClusterName = "workload-cluster"
	clusterNamespace    = "test-namespace"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	// TODO: remove after diskExtractor has been refactored and removed.
	features.ClearCache()
	t.Setenv(features.TinkerbellUseDiskExtractorDefaultDiskEnvVar, "true")
	//
	tt := newReconcilerTest(t)

	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
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
	tt.createAllObjs()

	result, err := tt.reconciler().ValidateDatacenterConfig(tt.ctx, test.NewNullLogger(), tt.buildScope())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.cleanup()
}

func TestReconcilerValidateDatacenterConfigMissingManagementCluster(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Spec.ManagementCluster.Name = "nonexistent-management-cluster"
	tt.createAllObjs()

	result, err := tt.reconciler().ValidateDatacenterConfig(tt.ctx, test.NewNullLogger(), tt.buildScope())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(tt.cluster.Status.FailureMessage).To(HaveValue(ContainSubstring("\"nonexistent-management-cluster\" not found")))
	tt.cleanup()
}

func TestReconcilerValidateDatacenterConfigMissingManagementDatacenter(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.managementCluster.Spec.DatacenterRef.Name = "nonexistent-datacenter"
	tt.createAllObjs()

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
	tt.createAllObjs()

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
}

func TestReconcilerReconcileControlPlaneSuccess(t *testing.T) {
	// TODO: remove after diskExtractor has been refactored and removed.
	features.ClearCache()
	t.Setenv(features.TinkerbellUseDiskExtractorDefaultDiskEnvVar, "true")
	//
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, test.NewNullLogger(), tt.buildScope())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadClusterName,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&tinkerbellv1.TinkerbellMachineTemplate{
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
		},
	)
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
}

func TestReconcilerReconcileControlPlaneSuccessRegistryMirrorAuthentication(t *testing.T) {
	// TODO: remove after diskExtractor has been refactored and removed.
	features.ClearCache()
	t.Setenv(features.TinkerbellUseDiskExtractorDefaultDiskEnvVar, "true")
	//
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
	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, test.NewNullLogger(), scope)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadClusterName,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&tinkerbellv1.TinkerbellMachineTemplate{
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
		},
	)
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
}

func TestReconcilerReconcileControlPlaneFailure(t *testing.T) {
	// TODO: remove after diskExtractor has been refactored and removed.
	features.ClearCache()
	t.Setenv(features.TinkerbellUseDiskExtractorDefaultDiskEnvVar, "true")
	//
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	scope := tt.buildScope()
	scope.ClusterSpec.Cluster = scope.ClusterSpec.Cluster.DeepCopy()
	scope.ClusterSpec.Cluster.Name = ""

	_, err := tt.reconciler().ReconcileControlPlane(tt.ctx, test.NewNullLogger(), scope)

	tt.Expect(err).To(MatchError(ContainSubstring("resource name may not be empty")))
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
}

func TestReconcilerReconcileWorkerNodesSuccess(t *testing.T) {
	// TODO: remove after diskExtractor has been refactored and removed.
	features.ClearCache()
	t.Setenv(features.TinkerbellUseDiskExtractorDefaultDiskEnvVar, "true")
	//

	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	tt.cluster.SetSelfManaged()
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
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
}

func TestReconcilerReconcileWorkersSuccess(t *testing.T) {
	// TODO: remove after diskExtractor has been refactored and removed.
	features.ClearCache()
	t.Setenv(features.TinkerbellUseDiskExtractorDefaultDiskEnvVar, "true")
	//

	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.createAllObjs()

	result, err := tt.reconciler().ReconcileWorkers(tt.ctx, test.NewNullLogger(), tt.buildScope())

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
}

func TestReconcilerReconcileWorkerNodesFailure(t *testing.T) {
	// TODO: remove after diskExtractor has been refactored and removed.
	features.ClearCache()
	t.Setenv(features.TinkerbellUseDiskExtractorDefaultDiskEnvVar, "true")
	//

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
}

func TestReconcilerValidateHardwareCountFail(t *testing.T) {
	tt := newReconcilerTest(t)
	logger := test.NewNullLogger()

	tt.cluster.Name = "invalidCluster"
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, &tinkv1alpha1.Hardware{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hw1",
			Namespace: constants.EksaSystemNamespace,
			Labels: map[string]string{
				"type": "cp",
			},
		},
		Spec: tinkv1alpha1.HardwareSpec{
			Metadata: &tinkv1alpha1.HardwareMetadata{
				Instance: &tinkv1alpha1.MetadataInstance{
					ID: "foo",
				},
			},
		},
	},
	)

	tt.withFakeClient()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, gomock.Any()).Return(controller.Result{}, nil)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("minimum hardware count not met for selector '{\"type\":\"worker\"}': have 0, require 1"))
}

func TestReconcilerValidateHardwareNoHardware(t *testing.T) {
	tt := newReconcilerTest(t)
	logger := test.NewNullLogger()

	tt.cluster.Name = "invalidCluster"
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, &tinkv1alpha1.Hardware{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hw1",
			Labels: map[string]string{
				hardware.OwnerNameLabel: "cluster",
				"type":                  "cp",
			},
		},
		Spec: tinkv1alpha1.HardwareSpec{
			Metadata: &tinkv1alpha1.HardwareMetadata{
				Instance: &tinkv1alpha1.MetadataInstance{
					ID: "foo",
				},
			},
		},
	},
	)

	tt.withFakeClient()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, gomock.Any()).Return(controller.Result{}, nil)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("minimum hardware count not met for selector '{\"type\":\"cp\"}': have 0, require 1"))
}

func TestReconcilerCheckRufioMachinesContactableFail(t *testing.T) {
	tt := newReconcilerTest(t)
	logger := test.NewNullLogger()

	tt.cluster.Name = "invalidCluster"
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, &rufiov1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bmc1",
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
	},
	)

	tt.withFakeClient()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, gomock.Any()).Return(controller.Result{}, nil)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("bmc connection failure"))
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
