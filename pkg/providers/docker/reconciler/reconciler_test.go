package reconciler_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	clusterspec "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/providers/docker/reconciler"
	dockereconcilermocks "github.com/aws/eks-anywhere/pkg/providers/docker/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	clusterNamespace = "test-namespace"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	t.Skip("Flaky (https://github.com/aws/eks-anywhere/issues/6996)")

	tt := newReconcilerTest(t)
	logger := test.NewNullLogger()

	tt.eksaSupportObjs = append(tt.eksaSupportObjs, tt.kcp)
	tt.createAllObjs()

	remoteClient := fake.NewClientBuilder().Build()
	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: tt.cluster.Name, Namespace: constants.EksaSystemNamespace},
	).Return(remoteClient, nil)
	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, tt.buildSpec())

	tt.Expect(tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)).To(Equal(controller.Result{}))
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())

	tt.ShouldEventuallyExist(tt.ctx, tt.kcp)
	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&dockerv1.DockerMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-control-plane-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&dockerv1.DockerCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyNotExist(tt.ctx,
		&dockerv1.DockerMachineTemplate{
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
		&dockerv1.DockerMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
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
		&dockerv1.DockerMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-md-0-1",
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
		&dockerv1.DockerMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-md-0-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
}

func TestReconcilerReconcileWorkersErrorGeneratingSpec(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	spec := tt.buildSpec()
	// this will always return an error since objects are not registered in the scheme
	tt.client = fake.NewClientBuilder().WithScheme(runtime.NewScheme()).Build()

	tt.Expect(
		tt.reconciler().ReconcileWorkers(tt.ctx, test.NewNullLogger(), spec),
	).Error().To(MatchError(ContainSubstring("generating workers spec")))
}

func TestReconcilerReconcileWorkerNodesFail(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "my-management-cluster"
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

func TestReconcileControlPlaneStackedEtcdSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
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
	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&dockerv1.DockerMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-control-plane-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&dockerv1.DockerCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyNotExist(tt.ctx,
		&dockerv1.DockerMachineTemplate{
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

func TestReconcileControlPlaneUnstackedEtcdSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{
		Count: 1,
	}
	tt.createAllObjs()
	logger := test.NewNullLogger()
	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, logger, tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(tt.cluster.Status.FailureReason).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&dockerv1.DockerMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name + "-control-plane-1",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&dockerv1.DockerCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tt.cluster.Name,
				Namespace: constants.EksaSystemNamespace,
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&dockerv1.DockerMachineTemplate{
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

func TestReconcilerReconcileControlPlaneFailure(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	spec := tt.buildSpec()
	spec.Cluster.Spec.KubernetesVersion = ""
	_, err := tt.reconciler().ReconcileControlPlane(tt.ctx, test.NewNullLogger(), spec)
	tt.Expect(err).To(MatchError(ContainSubstring("generating docker control plane yaml spec")))
}

type reconcilerTest struct {
	t testing.TB
	*WithT
	*envtest.APIExpecter
	ctx                  context.Context
	cniReconciler        *dockereconcilermocks.MockCNIReconciler
	remoteClientRegistry *dockereconcilermocks.MockRemoteClientRegistry
	cluster              *anywherev1.Cluster
	client               client.Client
	env                  *envtest.Environment
	eksaSupportObjs      []client.Object
	datacenterConfig     *anywherev1.DockerDatacenterConfig
	kcp                  *controlplanev1.KubeadmControlPlane
}

func newReconcilerTest(t testing.TB) *reconcilerTest {
	ctrl := gomock.NewController(t)
	cniReconciler := dockereconcilermocks.NewMockCNIReconciler(ctrl)
	remoteClientRegistry := dockereconcilermocks.NewMockRemoteClientRegistry(ctrl)
	c := env.Client()

	bundle := test.Bundle()
	version := test.DevEksaVersion()

	managementCluster := dockerCluster(func(c *anywherev1.Cluster) {
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

	workloadClusterDatacenter := dataCenter()

	cluster := dockerCluster(func(c *anywherev1.Cluster) {
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
		}
		c.Spec.DatacenterRef = anywherev1.Ref{
			Kind: anywherev1.DockerDatacenterKind,
			Name: workloadClusterDatacenter.Name,
		}
		c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations,
			anywherev1.WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
				Name:  "md-0",
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
		cluster:              cluster,
		client:               c,
		env:                  env,
		eksaSupportObjs: []client.Object{
			test.Namespace(clusterNamespace),
			test.Namespace(constants.EksaSystemNamespace),
			managementCluster,
			workloadClusterDatacenter,
			bundle,
			test.EksdRelease("1-22"),
			test.EKSARelease(),
		},
		datacenterConfig: workloadClusterDatacenter,
		kcp:              kcp,
	}

	t.Cleanup(tt.cleanup)
	return tt
}

func (tt *reconcilerTest) cleanup() {
	tt.DeleteAndWait(tt.ctx, tt.allObjs()...)

	tt.DeleteAllOfAndWait(tt.ctx, &clusterv1.Cluster{})
	tt.DeleteAllOfAndWait(tt.ctx, &dockerv1.DockerMachineTemplate{})
	tt.DeleteAllOfAndWait(tt.ctx, &etcdv1.EtcdadmCluster{})
}

func (tt *reconcilerTest) reconciler() *reconciler.Reconciler {
	return reconciler.New(tt.client, tt.cniReconciler, tt.remoteClientRegistry)
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

func (tt *reconcilerTest) createAllObjs() {
	tt.t.Helper()
	envtest.CreateObjs(tt.ctx, tt.t, tt.client, tt.allObjs()...)
}

func (tt *reconcilerTest) allObjs() []client.Object {
	objs := make([]client.Object, 0, len(tt.eksaSupportObjs)+1)
	objs = append(objs, tt.eksaSupportObjs...)
	objs = append(objs, tt.cluster)

	return objs
}

type clusterOpt func(*anywherev1.Cluster)

func dockerCluster(opts ...clusterOpt) *anywherev1.Cluster {
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

func dataCenter() *anywherev1.DockerDatacenterConfig {
	return &anywherev1.DockerDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.DockerDatacenterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: clusterNamespace,
		},
	}
}
