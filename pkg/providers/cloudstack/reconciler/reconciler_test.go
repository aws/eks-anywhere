package reconciler_test

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
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
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/reconciler"
	cloudstackreconcilermocks "github.com/aws/eks-anywhere/pkg/providers/cloudstack/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	clusterNamespace = "test-namespace"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	t.Skip("Flaky (https://github.com/aws/eks-anywhere/issues/6996)")

	tt := newReconcilerTest(t)
	// We want to check that the cluster status is cleaned up if validations are passed
	tt.cluster.SetFailure(anywherev1.FailureReasonType("InvalidCluster"), "invalid cluster")

	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.secret, tt.kcp)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	remoteClient := env.Client()

	spec := tt.buildSpec()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, tt.buildSpec()).Return(controller.Result{}, nil)
	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: tt.cluster.Name, Namespace: constants.EksaSystemNamespace},
	).Return(remoteClient, nil).Times(1)

	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, spec)
	ctrl := gomock.NewController(t)
	validator := cloudstack.NewMockProviderValidator(ctrl)
	tt.validatorRegistry.EXPECT().Get(tt.execConfig).Return(validator, nil).Times(1)
	validator.EXPECT().ValidateClusterMachineConfigs(tt.ctx, spec).Return(nil).Times(1)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeNil())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeNil())
}

func TestReconcilerValidateDatacenterConfigRequeue(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.datacenterConfig.Status.SpecValid = false

	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.secret)
	tt.createAllObjs()

	logger := test.NewNullLogger()

	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, tt.buildSpec()).Return(controller.Result{}, nil)
	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)
	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(result).To(Equal(controller.ResultWithReturn()))
	tt.Expect(tt.datacenterConfig.Status.FailureMessage).To(BeNil())
}

func TestReconcilerValidateDatacenterConfigFail(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.datacenterConfig.Status.SpecValid = false
	tt.datacenterConfig.Status.FailureMessage = ptr.String("Invalid CloudStackDatacenterConfig")

	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.secret)
	tt.createAllObjs()

	logger := test.NewNullLogger()

	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, tt.buildSpec()).Return(controller.Result{}, nil)

	_, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)
	tt.Expect(err).To(BeNil())
	tt.Expect(&tt.datacenterConfig.Status.FailureMessage).To(HaveValue(Equal("Invalid CloudStackDatacenterConfig")))
}

func TestReconcilerValidateMachineConfigInvalidSecret(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	spec := tt.buildSpec()
	logger := test.NewNullLogger()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, spec).Return(controller.Result{}, nil)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("Secret \"global\" not found")))
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeNil())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeNil())
}

func TestReconcilerValidateMachineConfigGetValidatorFail(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.secret)
	tt.createAllObjs()

	spec := tt.buildSpec()
	logger := test.NewNullLogger()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, spec).Return(controller.Result{}, nil)

	errMsg := "building cmk executable: nil exec config for CloudMonkey, unable to proceed"
	tt.validatorRegistry.EXPECT().Get(tt.execConfig).Return(nil, errors.New(errMsg)).Times(1)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring(errMsg)))
	tt.Expect(result).To(Equal(controller.Result{}))
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeNil())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeNil())
}

func TestReconcilerValidateMachineConfigFail(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.secret)
	tt.createAllObjs()

	spec := tt.buildSpec()
	logger := test.NewNullLogger()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, spec).Return(controller.Result{}, nil)

	ctrl := gomock.NewController(t)
	validator := cloudstack.NewMockProviderValidator(ctrl)
	tt.validatorRegistry.EXPECT().Get(tt.execConfig).Return(validator, nil).Times(1)
	errMsg := "Invalid CloudStackMachineConfig: validating service offering"
	validator.EXPECT().ValidateClusterMachineConfigs(tt.ctx, spec).Return(errors.New(errMsg)).Times(1)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)
	tt.Expect(err).To(BeNil())
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(tt.cluster.Status.FailureMessage).To(HaveValue(ContainSubstring(errMsg)))
	tt.Expect(tt.cluster.Status.FailureReason).To(HaveValue(Equal(anywherev1.MachineConfigInvalidReason)))
}

func TestReconcilerControlPlaneIsNotReady(t *testing.T) {
	t.Skip("Flaky (https://github.com/aws/eks-anywhere/issues/7000)")

	tt := newReconcilerTest(t)

	tt.kcp.Status = controlplanev1.KubeadmControlPlaneStatus{
		Conditions: clusterv1.Conditions{
			{
				Type:               clusterapi.ReadyCondition,
				Status:             corev1.ConditionFalse,
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		},
		ObservedGeneration: 2,
	}
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.kcp, tt.secret)
	tt.createAllObjs()

	logger := test.NewNullLogger()

	spec := tt.buildSpec()
	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, spec).Return(controller.Result{}, nil)
	ctrl := gomock.NewController(t)
	validator := cloudstack.NewMockProviderValidator(ctrl)
	tt.validatorRegistry.EXPECT().Get(tt.execConfig).Return(validator, nil).Times(1)
	validator.EXPECT().ValidateClusterMachineConfigs(tt.ctx, spec).Return(nil).Times(1)
	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.ResultWithRequeue(30 * time.Second)))
}

func TestReconcileControlPlaneUnstackedEtcdSuccess(t *testing.T) {
	t.Skip("Flaky (https://github.com/aws/eks-anywhere/issues/7001)")

	tt := newReconcilerTest(t)
	tt.cluster.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{
		Count: 1,
		MachineGroupRef: &anywherev1.Ref{
			Kind: anywherev1.CloudStackMachineConfigKind,
			Name: tt.machineConfigControlPlane.Name,
		},
	}
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.secret)
	tt.createAllObjs()
	logger := test.NewNullLogger()
	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, logger, tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx, &cloudstackv1.CloudStackCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tt.cluster.Name,
			Namespace: constants.EksaSystemNamespace,
		},
	})
	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&cloudstackv1.CloudStackMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-control-plane-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&cloudstackv1.CloudStackMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-etcd-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&etcdv1.EtcdadmCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-etcd",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
}

func TestReconcileControlPlaneStackedEtcdSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.secret)
	tt.createAllObjs()
	logger := test.NewNullLogger()
	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, logger, tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})

	tt.ShouldEventuallyExist(tt.ctx, capiCluster)
	tt.ShouldEventuallyExist(tt.ctx, &cloudstackv1.CloudStackCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tt.cluster.Name,
			Namespace: constants.EksaSystemNamespace,
		},
	})
	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&cloudstackv1.CloudStackMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-control-plane-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyNotExist(tt.ctx,
		&cloudstackv1.CloudStackMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-etcd-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyNotExist(tt.ctx,
		&etcdv1.EtcdadmCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-etcd",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
}

func TestReconcileCNISuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	remoteClient := fake.NewClientBuilder().Build()
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.secret)
	spec := tt.buildSpec()

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: tt.cluster.Name, Namespace: "eksa-system"},
	).Return(remoteClient, nil)
	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, spec)

	result, err := tt.reconciler().ReconcileCNI(tt.ctx, logger, spec)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcileCNIErrorClientRegistry(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.secret)
	spec := tt.buildSpec()

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: tt.cluster.Name, Namespace: "eksa-system"},
	).Return(nil, errors.New("building client"))

	result, err := tt.reconciler().ReconcileCNI(tt.ctx, logger, spec)

	tt.Expect(err).To(MatchError(ContainSubstring("building client")))
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcilerReconcileWorkersSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster, tt.secret)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	result, err := tt.reconciler().ReconcileWorkers(tt.ctx, logger, tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
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
		&cloudstackv1.CloudStackMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
}

func TestReconcilerReconcileWorkersFailure(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	tt.cluster.SetSelfManaged()
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster, tt.secret)
	tt.createAllObjs()
	clusterSpec := tt.buildSpec()
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = ptr.Int(int(math.Inf(1)))

	logger := test.NewNullLogger()

	_, err := tt.reconciler().ReconcileWorkers(tt.ctx, logger, clusterSpec)

	tt.Expect(err).To(MatchError(ContainSubstring("Generate worker node CAPI spec")))
}

func TestReconcilerReconcileWorkerNodesSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	tt.cluster.SetSelfManaged()
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster, tt.secret)
	tt.createAllObjs()

	logger := test.NewNullLogger()

	result, err := tt.reconciler().ReconcileWorkerNodes(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
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
		&cloudstackv1.CloudStackMachineTemplate{
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

func TestReconcilerReconcileWorkerNodesFailure(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "mgmt-cluster"
	tt.cluster.SetSelfManaged()
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.cluster.Spec.KubernetesVersion = ""
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster, tt.secret)
	tt.createAllObjs()

	logger := test.NewNullLogger()

	_, err := tt.reconciler().ReconcileWorkerNodes(tt.ctx, logger, tt.cluster)

	tt.Expect(err).To(MatchError(ContainSubstring("building cluster Spec for worker node reconcile")))
}

func (tt *reconcilerTest) withFakeClient() {
	tt.client = fake.NewClientBuilder().WithObjects(clientutil.ObjectsToClientObjects(tt.allObjs())...).Build()
}

func (tt *reconcilerTest) createAllObjs() {
	tt.t.Helper()
	envtest.CreateObjs(tt.ctx, tt.t, tt.client, tt.allObjs()...)
}

func (tt *reconcilerTest) allObjs() []client.Object {
	objs := make([]client.Object, 0, len(tt.eksaSupportObjs)+3)
	objs = append(objs, tt.eksaSupportObjs...)
	objs = append(objs, tt.cluster, tt.machineConfigControlPlane, tt.machineConfigWorker)

	return objs
}

func (tt *reconcilerTest) reconciler() *reconciler.Reconciler {
	return reconciler.New(tt.client, tt.ipValidator, tt.cniReconciler, tt.remoteClientRegistry, tt.validatorRegistry)
}

func (tt *reconcilerTest) buildSpec() *clusterspec.Spec {
	tt.t.Helper()
	spec, err := clusterspec.BuildSpec(tt.ctx, clientutil.NewKubeClient(tt.client), tt.cluster)
	tt.Expect(err).NotTo(HaveOccurred())

	return spec
}

type reconcilerTest struct {
	t testing.TB
	*WithT
	*envtest.APIExpecter
	ctx                       context.Context
	cluster                   *anywherev1.Cluster
	client                    client.Client
	eksaSupportObjs           []client.Object
	datacenterConfig          *anywherev1.CloudStackDatacenterConfig
	machineConfigControlPlane *anywherev1.CloudStackMachineConfig
	machineConfigWorker       *anywherev1.CloudStackMachineConfig
	ipValidator               *cloudstackreconcilermocks.MockIPValidator
	cniReconciler             *cloudstackreconcilermocks.MockCNIReconciler
	remoteClientRegistry      *cloudstackreconcilermocks.MockRemoteClientRegistry
	validatorRegistry         *cloudstack.MockValidatorRegistry
	execConfig                *decoder.CloudStackExecConfig
	secret                    *corev1.Secret
	kcp                       *controlplanev1.KubeadmControlPlane
}

func newReconcilerTest(t testing.TB) *reconcilerTest {
	ctrl := gomock.NewController(t)
	c := env.Client()

	ipValidator := cloudstackreconcilermocks.NewMockIPValidator(ctrl)
	cniReconciler := cloudstackreconcilermocks.NewMockCNIReconciler(ctrl)
	remoteClientRegistry := cloudstackreconcilermocks.NewMockRemoteClientRegistry(ctrl)
	validatorRegistry := cloudstack.NewMockValidatorRegistry(ctrl)
	execConfig := &decoder.CloudStackExecConfig{
		Profiles: []decoder.CloudStackProfileConfig{
			{
				Name:          "global",
				ApiKey:        "test-key1",
				SecretKey:     "test-secret1",
				ManagementUrl: "http://1.1.1.1:8080/client/api",
			},
		},
	}

	bundle := test.Bundle()
	version := test.DevEksaVersion()

	managementCluster := cloudstackCluster(func(c *anywherev1.Cluster) {
		c.Name = "management-cluster"
		c.Spec.ManagementCluster = anywherev1.ManagementCluster{
			Name: c.Name,
		}
		c.Spec.BundlesRef = &anywherev1.BundlesRef{
			Name:       bundle.Name,
			Namespace:  bundle.Namespace,
			APIVersion: bundle.APIVersion,
		}
		c.Spec.EksaVersion = &version
	})

	machineConfigCP := machineConfig(func(m *anywherev1.CloudStackMachineConfig) {
		m.Name = "cp-machine-config"
		m.Spec.Template.Name = "kubernetes-1-22"
		m.Spec.Users = append(m.Spec.Users,
			anywherev1.UserConfiguration{
				Name:              "user",
				SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8ZEibIrz1AUBKDvmDiWLs9f5DnOerC4qPITiDtSOuPAsxgZbRMavBfVTxodMdAkYRYlXxK6PqNo0ve0qcOV2yvpxH1OogasMMetck6BlM/dIoo3vEY4ZoG9DuVRIf9Iry5gJKbpMDYWpx1IGZrDMOFcIM20ii2qLQQk5hfq9OqdqhToEJFixdgJt/y/zt6Koy3kix+XsnrVdAHgWAq4CZuwt1G6JUAqrpob3H8vPmL7aS+35ktf0pHBm6nYoxRhslnWMUb/7vpzWiq+fUBIm2LYqvrnm7t3fRqFx7p2sZqAm2jDNivyYXwRXkoQPR96zvGeMtuQ5BVGPpsDfVudSW21+pEXHI0GINtTbua7Ogz7wtpVywSvHraRgdFOeY9mkXPzvm2IhoqNrteck2GErwqSqb19mPz6LnHueK0u7i6WuQWJn0CUoCtyMGIrowXSviK8qgHXKrmfTWATmCkbtosnLskNdYuOw8bKxq5S4WgdQVhPps2TiMSZndjX5NTr8= ubuntu@ip-10-2-0-6"},
			})
	})
	machineConfigWN := machineConfig(func(m *anywherev1.CloudStackMachineConfig) {
		m.Name = "worker-machine-config"
		m.Spec.Template.Name = "kubernetes-1-22"
		m.Spec.Users = append(m.Spec.Users,
			anywherev1.UserConfiguration{
				Name:              "user",
				SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8ZEibIrz1AUBKDvmDiWLs9f5DnOerC4qPITiDtSOuPAsxgZbRMavBfVTxodMdAkYRYlXxK6PqNo0ve0qcOV2yvpxH1OogasMMetck6BlM/dIoo3vEY4ZoG9DuVRIf9Iry5gJKbpMDYWpx1IGZrDMOFcIM20ii2qLQQk5hfq9OqdqhToEJFixdgJt/y/zt6Koy3kix+XsnrVdAHgWAq4CZuwt1G6JUAqrpob3H8vPmL7aS+35ktf0pHBm6nYoxRhslnWMUb/7vpzWiq+fUBIm2LYqvrnm7t3fRqFx7p2sZqAm2jDNivyYXwRXkoQPR96zvGeMtuQ5BVGPpsDfVudSW21+pEXHI0GINtTbua7Ogz7wtpVywSvHraRgdFOeY9mkXPzvm2IhoqNrteck2GErwqSqb19mPz6LnHueK0u7i6WuQWJn0CUoCtyMGIrowXSviK8qgHXKrmfTWATmCkbtosnLskNdYuOw8bKxq5S4WgdQVhPps2TiMSZndjX5NTr8= ubuntu@ip-10-2-0-6"},
			})
	})

	workloadClusterDatacenter := dataCenter(func(d *anywherev1.CloudStackDatacenterConfig) {
		d.Spec.AvailabilityZones = append(d.Spec.AvailabilityZones,
			anywherev1.CloudStackAvailabilityZone{
				Name:           "test-zone",
				CredentialsRef: "global",
			})
		d.Status.SpecValid = true
	})

	cluster := cloudstackCluster(func(c *anywherev1.Cluster) {
		c.Name = strings.ToLower(t.Name())
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
				Kind: anywherev1.CloudStackMachineConfigKind,
				Name: machineConfigCP.Name,
			},
		}
		c.Spec.DatacenterRef = anywherev1.Ref{
			Kind: anywherev1.CloudStackDatacenterKind,
			Name: workloadClusterDatacenter.Name,
		}

		c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations,
			anywherev1.WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.CloudStackMachineConfigKind,
					Name: machineConfigWN.Name,
				},
				Name:   "md-0",
				Labels: nil,
			},
		)

		c.Spec.EksaVersion = &version
	})

	kcp := test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
		kcp.Name = cluster.Name
		kcp.Spec = controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: corev1.ObjectReference{
					Name: fmt.Sprintf("%s-control-plane-1", cluster.Name),
				},
			},
		}
		kcp.Status = controlplanev1.KubeadmControlPlaneStatus{
			Conditions: clusterv1.Conditions{
				{
					Type:               clusterapi.ReadyCondition,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now()),
				},
			},
			ObservedGeneration: 2,
		}
	})

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "global",
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string][]byte{
			decoder.APIKeyKey:    []byte("test-key1"),
			decoder.APIUrlKey:    []byte("http://1.1.1.1:8080/client/api"),
			decoder.SecretKeyKey: []byte("test-secret1"),
		},
	}

	tt := &reconcilerTest{
		t:           t,
		WithT:       NewWithT(t),
		APIExpecter: envtest.NewAPIExpecter(t, c),
		ctx:         context.Background(),
		ipValidator: ipValidator,
		client:      c,
		eksaSupportObjs: []client.Object{
			test.Namespace(clusterNamespace),
			test.Namespace(constants.EksaSystemNamespace),
			managementCluster,
			workloadClusterDatacenter,
			bundle,
			test.EksdRelease("1-22"),
			test.EKSARelease(),
		},
		cluster:                   cluster,
		datacenterConfig:          workloadClusterDatacenter,
		machineConfigControlPlane: machineConfigCP,
		machineConfigWorker:       machineConfigWN,
		cniReconciler:             cniReconciler,
		remoteClientRegistry:      remoteClientRegistry,
		validatorRegistry:         validatorRegistry,
		execConfig:                execConfig,
		secret:                    secret,
		kcp:                       kcp,
	}

	t.Cleanup(tt.cleanup)
	return tt
}

func (tt *reconcilerTest) cleanup() {
	tt.DeleteAndWait(tt.ctx, tt.allObjs()...)
	tt.DeleteAllOfAndWait(tt.ctx, &bootstrapv1.KubeadmConfigTemplate{})
	tt.DeleteAllOfAndWait(tt.ctx, &clusterv1.Cluster{})
	tt.DeleteAllOfAndWait(tt.ctx, &clusterv1.MachineDeployment{})
	tt.DeleteAllOfAndWait(tt.ctx, &cloudstackv1.CloudStackCluster{})
	tt.DeleteAllOfAndWait(tt.ctx, &controlplanev1.KubeadmControlPlane{})
	tt.DeleteAndWait(tt.ctx, &cloudstackv1.CloudStackMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tt.cluster.Name + "-etcd-1",
			Namespace: "eksa-system",
		},
	})
	tt.DeleteAndWait(tt.ctx, &etcdv1.EtcdadmCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tt.cluster.Name + "-etcd",
			Namespace: "eksa-system",
		},
	})
}

type clusterOpt func(*anywherev1.Cluster)

func cloudstackCluster(opts ...clusterOpt) *anywherev1.Cluster {
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

type datacenterOpt func(config *anywherev1.CloudStackDatacenterConfig)

func dataCenter(opts ...datacenterOpt) *anywherev1.CloudStackDatacenterConfig {
	d := &anywherev1.CloudStackDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.CloudStackDatacenterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: clusterNamespace,
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

type cloudstackMachineOpt func(config *anywherev1.CloudStackMachineConfig)

func machineConfig(opts ...cloudstackMachineOpt) *anywherev1.CloudStackMachineConfig {
	m := &anywherev1.CloudStackMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.CloudStackMachineConfigKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterNamespace,
		},
		Spec: anywherev1.CloudStackMachineConfigSpec{},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}
