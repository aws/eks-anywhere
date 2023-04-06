package reconciler_test

import (
	"context"
	"github.com/pkg/errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	clusterspec "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/reconciler"
	cloudstackreconcilermocks "github.com/aws/eks-anywhere/pkg/providers/cloudstack/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	clusterNamespace = "test-namespace"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	tt := newReconcilerTest(t)

	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	remoteClient := env.Client()

	tt.ipValidator.EXPECT().ValidateControlPlaneIP(tt.ctx, logger, tt.buildSpec()).Return(controller.Result{}, nil)
	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: "workload-cluster", Namespace: "eksa-system"},
	).Return(remoteClient, nil).Times(1)
	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, tt.buildSpec())

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcileCNISuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	remoteClient := fake.NewClientBuilder().Build()
	spec := tt.buildSpec()

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: "workload-cluster", Namespace: "eksa-system"},
	).Return(remoteClient, nil)
	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, spec)

	result, err := tt.reconciler().ReconcileCNI(tt.ctx, logger, spec)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcileCNIErrorClientRegistry(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.withFakeClient()

	logger := test.NewNullLogger()
	spec := tt.buildSpec()

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: "workload-cluster", Namespace: "eksa-system"},
	).Return(nil, errors.New("building client"))

	result, err := tt.reconciler().ReconcileCNI(tt.ctx, logger, spec)

	tt.Expect(err).To(MatchError(ContainSubstring("building client")))
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func (tt *reconcilerTest) withFakeClient() {
	tt.client = fake.NewClientBuilder().WithObjects(clientutil.ObjectsToClientObjects(tt.allObjs())...).Build()
}

func (tt *reconcilerTest) allObjs() []client.Object {
	objs := make([]client.Object, 0, len(tt.eksaSupportObjs)+3)
	objs = append(objs, tt.eksaSupportObjs...)
	objs = append(objs, tt.cluster, tt.machineConfigControlPlane, tt.machineConfigWorker)

	return objs
}

func (tt *reconcilerTest) reconciler() *reconciler.Reconciler {
	return reconciler.New(tt.client, tt.ipValidator, tt.cniReconciler, tt.remoteClientRegistry)
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
}

func newReconcilerTest(t testing.TB) *reconcilerTest {
	ctrl := gomock.NewController(t)
	c := env.Client()

	ipValidator := cloudstackreconcilermocks.NewMockIPValidator(ctrl)
	cniReconciler := cloudstackreconcilermocks.NewMockCNIReconciler(ctrl)
	remoteClientRegistry := cloudstackreconcilermocks.NewMockRemoteClientRegistry(ctrl)

	bundle := test.Bundle()

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
	})

	machineConfigCP := machineConfig(func(m *anywherev1.CloudStackMachineConfig) {
		m.Name = "cp-machine-config"
	})
	machineConfigWN := machineConfig(func(m *anywherev1.CloudStackMachineConfig) {
		m.Name = "worker-machine-config"
	})

	workloadClusterDatacenter := dataCenter(func(d *anywherev1.CloudStackDatacenterConfig) {})

	cluster := cloudstackCluster(func(c *anywherev1.Cluster) {
		c.Name = "workload-cluster"
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
	})

	tt := &reconcilerTest{
		t:           t,
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		ipValidator: ipValidator,
		client:      c,
		eksaSupportObjs: []client.Object{
			test.Namespace(clusterNamespace),
			test.Namespace(constants.EksaSystemNamespace),
			managementCluster,
			workloadClusterDatacenter,
			bundle,
			test.EksdRelease(),
		},
		cluster:                   cluster,
		datacenterConfig:          workloadClusterDatacenter,
		machineConfigControlPlane: machineConfigCP,
		machineConfigWorker:       machineConfigWN,
		cniReconciler:             cniReconciler,
		remoteClientRegistry:      remoteClientRegistry,
	}

	return tt
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
