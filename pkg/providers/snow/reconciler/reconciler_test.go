package reconciler_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
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
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
	"github.com/aws/eks-anywhere/pkg/providers/snow/reconciler"
	"github.com/aws/eks-anywhere/pkg/providers/snow/reconciler/mocks"
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

	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.kcp)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	remoteClient := fake.NewClientBuilder().Build()

	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, tt.buildSpec()).Return(controller.Result{}, nil)

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: tt.cluster.Name, Namespace: "eksa-system"},
	).Return(remoteClient, nil)
	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, tt.buildSpec())

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.Expect(tt.cluster.Status.FailureMessage).To(BeNil())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeNil())
}

func TestReconcilerReconcileWorkerNodesSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "my-management-cluster"
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
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx,
		&bootstrapv1.KubeadmConfigTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-management-cluster-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&snowv1.AWSSnowMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-management-cluster-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&clusterv1.MachineDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-management-cluster-md-0",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&snowv1.AWSSnowIPPool{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ip-pool",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
}

func TestReconcilerValidateMachineConfigsInvalidWorkerMachineConfig(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.machineConfigWorker.Status.SpecValid = false
	m := "Something wrong"
	tt.machineConfigWorker.Status.FailureMessage = &m
	tt.withFakeClient()

	result, err := tt.reconciler().ValidateMachineConfigs(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(tt.cluster.Status.FailureMessage).ToNot(BeZero())
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("Invalid worker-machine-config SnowMachineConfig"))
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("Something wrong"))
	tt.Expect(tt.cluster.Status.FailureReason).ToNot(BeZero())
	tt.Expect(*tt.cluster.Status.FailureReason).To(HaveValue(Equal(anywherev1.MachineConfigInvalidReason)))
}

func TestReconcilerValidateMachineConfigsInvalidControlPlaneMachineConfig(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.machineConfigControlPlane.Status.SpecValid = false
	m := "Something wrong"
	tt.machineConfigControlPlane.Status.FailureMessage = &m
	tt.withFakeClient()

	result, err := tt.reconciler().ValidateMachineConfigs(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(tt.cluster.Status.FailureMessage).ToNot(BeZero())
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("Invalid cp-machine-config SnowMachineConfig"))
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("Something wrong"))
	tt.Expect(tt.cluster.Status.FailureReason).ToNot(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(HaveValue(Equal(anywherev1.MachineConfigInvalidReason)))
}

func TestReconcilerValidateMachineConfigsMachineConfigNotValidated(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.machineConfigWorker.Status.SpecValid = false
	tt.withFakeClient()

	result, err := tt.reconciler().ValidateMachineConfigs(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(result).To(Equal(controller.Result{Result: &reconcile.Result{}}), "result should stop reconciliation")
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeNil())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeNil())
}

func TestReconcilerReconcileWorkers(t *testing.T) {
	t.Skip("Flaky (https://github.com/aws/eks-anywhere/issues/6999)")

	tt := newReconcilerTest(t)
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)

	tt.createAllObjs()

	result, err := tt.reconciler().ReconcileWorkers(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcilerReconcileControlPlane(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: "eksa-system",
			},
		},
	)

	tt.ShouldEventuallyExist(tt.ctx,
		&snowv1.AWSSnowMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-control-plane-1",
				Namespace: "eksa-system",
			},
		},
	)

	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.ShouldEventuallyExist(tt.ctx, capiCluster)

	tt.ShouldEventuallyExist(tt.ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: tt.cluster.Name + "-snow-credentials", Namespace: "eksa-system"}})
}

func TestReconcilerCheckControlPlaneReadyItIsReady(t *testing.T) {
	tt := newReconcilerTest(t)
	kcpVersion := "test"
	tt.kcp.Spec.Version = kcpVersion
	tt.kcp.Status = controlplanev1.KubeadmControlPlaneStatus{
		Conditions: clusterv1.Conditions{
			{
				Type:               clusterapi.ReadyCondition,
				Status:             corev1.ConditionTrue,
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		},
		Version: pointer.String(kcpVersion),
	}
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.kcp)
	tt.withFakeClient()

	result, err := tt.reconciler().CheckControlPlaneReady(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcilerReconcileCNISuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	remoteClient := fake.NewClientBuilder().Build()
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

func TestReconcilerReconcileCNIErrorClientRegistry(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
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

type reconcilerTest struct {
	t testing.TB
	*WithT
	*envtest.APIExpecter
	ctx                       context.Context
	cniReconciler             *mocks.MockCNIReconciler
	remoteClientRegistry      *mocks.MockRemoteClientRegistry
	ipValidator               *mocks.MockIPValidator
	cluster                   *anywherev1.Cluster
	client                    client.Client
	env                       *envtest.Environment
	eksaSupportObjs           []client.Object
	machineConfigControlPlane *anywherev1.SnowMachineConfig
	machineConfigWorker       *anywherev1.SnowMachineConfig
	kcp                       *controlplanev1.KubeadmControlPlane
}

func newReconcilerTest(t testing.TB) *reconcilerTest {
	ctrl := gomock.NewController(t)
	cniReconciler := mocks.NewMockCNIReconciler(ctrl)
	remoteClientRegistry := mocks.NewMockRemoteClientRegistry(ctrl)
	ipValidator := mocks.NewMockIPValidator(ctrl)
	c := env.Client()

	bundle := test.Bundle()
	version := test.DevEksaVersion()

	managementCluster := snowCluster(func(c *anywherev1.Cluster) {
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

	machineConfigCP := snowMachineConfig(func(m *anywherev1.SnowMachineConfig) {
		m.Name = "cp-machine-config"
		m.Status.SpecValid = true
	})

	ipPool := ipPool()
	machineConfigWN := snowMachineConfig(func(m *anywherev1.SnowMachineConfig) {
		m.Name = "worker-machine-config"
		m.Spec.Network.DirectNetworkInterfaces[0].DHCP = false
		m.Spec.Network.DirectNetworkInterfaces[0].IPPoolRef = &anywherev1.Ref{
			Name: ipPool.Name,
			Kind: ipPool.Kind,
		}
		m.Status.SpecValid = true
	})

	credentialsSecret := credentialsSecret()
	workloadClusterDatacenter := snowDataCenter(func(d *anywherev1.SnowDatacenterConfig) {
		d.Spec.IdentityRef = anywherev1.Ref{
			Kind: "Secret",
			Name: credentialsSecret.Name,
		}
	})

	cluster := snowCluster(func(c *anywherev1.Cluster) {
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
				Kind: "SnowMachineConfig",
				Name: machineConfigCP.Name,
			},
		}
		c.Spec.DatacenterRef = anywherev1.Ref{
			Kind: anywherev1.SnowDatacenterKind,
			Name: workloadClusterDatacenter.Name,
		}

		c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations,
			anywherev1.WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
				MachineGroupRef: &anywherev1.Ref{
					Kind: "SnowMachineConfig",
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

	tt := &reconcilerTest{
		t:                    t,
		WithT:                NewWithT(t),
		APIExpecter:          envtest.NewAPIExpecter(t, c),
		ctx:                  context.Background(),
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		ipValidator:          ipValidator,
		client:               c,
		env:                  env,
		eksaSupportObjs: []client.Object{
			test.Namespace(clusterNamespace),
			test.Namespace(constants.EksaSystemNamespace),
			managementCluster,
			workloadClusterDatacenter,
			bundle,
			test.EksdRelease("1-22"),
			credentialsSecret,
			ipPool,
			test.EKSARelease(),
		},
		cluster:                   cluster,
		machineConfigControlPlane: machineConfigCP,
		machineConfigWorker:       machineConfigWN,
		kcp:                       kcp,
	}

	t.Cleanup(tt.cleanup)
	return tt
}

func (tt *reconcilerTest) cleanup() {
	tt.DeleteAndWait(tt.ctx, tt.allObjs()...)
	tt.DeleteAllOfAndWait(tt.ctx, &bootstrapv1.KubeadmConfigTemplate{})
	tt.DeleteAllOfAndWait(tt.ctx, &snowv1.AWSSnowMachineTemplate{})
	tt.DeleteAllOfAndWait(tt.ctx, &clusterv1.MachineDeployment{})
}

func (tt *reconcilerTest) buildSpec() *clusterspec.Spec {
	tt.t.Helper()
	spec, err := clusterspec.BuildSpec(tt.ctx, clientutil.NewKubeClient(tt.client), tt.cluster)
	tt.Expect(err).NotTo(HaveOccurred())

	return spec
}

func (tt *reconcilerTest) withFakeClient() {
	tt.client = fake.NewClientBuilder().WithObjects(clientutil.ObjectsToClientObjects(tt.allObjs())...).Build()
}

func (tt *reconcilerTest) reconciler() *reconciler.Reconciler {
	return reconciler.New(tt.client, tt.cniReconciler, tt.remoteClientRegistry, tt.ipValidator)
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

type clusterOpt func(*anywherev1.Cluster)

func snowCluster(opts ...clusterOpt) *anywherev1.Cluster {
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

type datacenterOpt func(*anywherev1.SnowDatacenterConfig)

func snowDataCenter(opts ...datacenterOpt) *anywherev1.SnowDatacenterConfig {
	d := &anywherev1.SnowDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.SnowDatacenterKind,
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

type snowMachineOpt func(*anywherev1.SnowMachineConfig)

func snowMachineConfig(opts ...snowMachineOpt) *anywherev1.SnowMachineConfig {
	m := &anywherev1.SnowMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.SnowMachineConfigKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterNamespace,
		},
		Spec: anywherev1.SnowMachineConfigSpec{
			PhysicalNetworkConnector: anywherev1.SFPPlus,
			OSFamily:                 anywherev1.Ubuntu,
			Network: anywherev1.SnowNetwork{
				DirectNetworkInterfaces: []anywherev1.SnowDirectNetworkInterface{
					{
						Index:   1,
						Primary: true,
						DHCP:    true,
					},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func credentialsSecret() *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-snow-credentials",
			Namespace: clusterNamespace,
		},
		Data: map[string][]byte{
			"credentials": []byte("creds"),
			"ca-bundle":   []byte("certs"),
		},
		Type: "Opaque",
	}
}

func ipPool() *anywherev1.SnowIPPool {
	return &anywherev1.SnowIPPool{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.SnowIPPoolKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ip-pool",
			Namespace: clusterNamespace,
		},
		Spec: anywherev1.SnowIPPoolSpec{
			Pools: []anywherev1.IPPool{
				{
					IPStart: "start",
					IPEnd:   "end",
					Gateway: "gateway",
					Subnet:  "subnet",
				},
			},
		},
	}
}
