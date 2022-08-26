package reconciler_test

import (
	"context"
	"testing"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	clusterspec "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/providers/snow/reconciler"
	"github.com/aws/eks-anywhere/pkg/providers/snow/reconciler/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	clusterNamespace = "test-namespace"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	result, err := tt.reconciler().Reconcile(tt.ctx, test.NewNullLogger(), tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcilerValidateMachineConfigsInvalidWorkerMachineConfig(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.machineConfigWorker.Status.SpecValid = false
	m := "Something wrong"
	tt.machineConfigWorker.Status.FailureMessage = &m
	tt.withFakeClient()

	_, err := tt.reconciler().ValidateMachineConfigs(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(tt.cluster.Status.FailureMessage).ToNot(BeZero())
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("SnowMachineConfig worker-machine-config is invalid"))
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("Something wrong"))
}

func TestReconcilerValidateMachineConfigsInvalidControlPlaneMachineConfig(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.machineConfigControlPlane.Status.SpecValid = false
	m := "Something wrong"
	tt.machineConfigControlPlane.Status.FailureMessage = &m
	tt.withFakeClient()

	_, err := tt.reconciler().ValidateMachineConfigs(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(tt.cluster.Status.FailureMessage).ToNot(BeZero())
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("SnowMachineConfig cp-machine-config is invalid"))
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("Something wrong"))
}

func TestReconcilerReconcileWorkers(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	result, err := tt.reconciler().ReconcileWorkers(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

type reconcilerTest struct {
	t testing.TB
	*WithT
	ctx                       context.Context
	cniReconciler             *mocks.MockCNIReconciler
	cluster                   *anywherev1.Cluster
	client                    client.Client
	env                       *envtest.Environment
	eksaSupportObjs           []envtest.Object
	machineConfigControlPlane *anywherev1.SnowMachineConfig
	machineConfigWorker       *anywherev1.SnowMachineConfig
}

func newReconcilerTest(t testing.TB) *reconcilerTest {
	ctrl := gomock.NewController(t)
	cniReconciler := mocks.NewMockCNIReconciler(ctrl)
	client := env.Client()

	bundle := createBundle()

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
	})

	machineConfigCP := snowMachineConfig(func(m *anywherev1.SnowMachineConfig) {
		m.Name = "cp-machine-config"
		m.Status.SpecValid = true
	})
	machineConfigWN := snowMachineConfig(func(m *anywherev1.SnowMachineConfig) {
		m.Name = "worker-machine-config"
		m.Status.SpecValid = true
	})

	cluster := snowCluster(func(c *anywherev1.Cluster) {
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
				Kind: "SnowMachineConfig",
				Name: machineConfigCP.Name,
			},
		}

		c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations,
			anywherev1.WorkerNodeGroupConfiguration{
				Count: 1,
				MachineGroupRef: &anywherev1.Ref{
					Kind: "SnowMachineConfig",
					Name: machineConfigWN.Name,
				},
				Name:   "md-0",
				Labels: nil,
			},
		)
	})

	tt := &reconcilerTest{
		t:             t,
		WithT:         NewWithT(t),
		ctx:           context.Background(),
		cniReconciler: cniReconciler,
		client:        client,
		env:           env,
		eksaSupportObjs: []envtest.Object{
			namespace(clusterNamespace),
			namespace(constants.EksaSystemNamespace),
			managementCluster,
			snowDataCenter(),
			bundle,
			eksdRelease(),
		},
		cluster:                   cluster,
		machineConfigControlPlane: machineConfigCP,
		machineConfigWorker:       machineConfigWN,
	}

	t.Cleanup(tt.cleanup)
	return tt
}

func (tt *reconcilerTest) cleanup() {
	for _, obj := range tt.allObjs() {
		_, isNamespace := obj.(*v1.Namespace)
		if isNamespace {
			// namespaces can't be deleted with envtest
			continue
		}
		tt.Expect(tt.client.Delete(tt.ctx, obj)).To(Succeed())
		key := client.ObjectKeyFromObject(obj)
		for {
			if err := tt.client.Get(tt.ctx, key, obj); apierrors.IsNotFound(err) {
				break
			}
		}
	}
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
	return reconciler.New(tt.client, tt.cniReconciler)
}

func (tt *reconcilerTest) createAllObjs() {
	tt.t.Helper()
	envtest.CreateObjs(tt.ctx, tt.t, tt.client, tt.allObjs()...)
}

func (tt *reconcilerTest) allObjs() []envtest.Object {
	objs := make([]envtest.Object, 0, len(tt.eksaSupportObjs)+3)
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
			DatacenterRef: anywherev1.Ref{
				Kind: "SnowDatacenterConfig",
				Name: "datacenter",
			},
			KubernetesVersion: "1.20",
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func snowDataCenter() *anywherev1.SnowDatacenterConfig {
	return &anywherev1.SnowDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.SnowDatacenterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: clusterNamespace,
		},
	}
}

func createBundle() *releasev1.Bundles {
	return &releasev1.Bundles{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Bundles",
			APIVersion: releasev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bundle-1",
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
					CertManager:            releasev1.CertManagerBundle{},
					ClusterAPI:             releasev1.CoreClusterAPI{},
					Bootstrap:              releasev1.KubeadmBootstrapBundle{},
					ControlPlane:           releasev1.KubeadmControlPlaneBundle{},
					Aws:                    releasev1.AwsBundle{},
					VSphere:                releasev1.VSphereBundle{},
					Docker:                 releasev1.DockerBundle{},
					Eksa:                   releasev1.EksaBundle{},
					Cilium:                 releasev1.CiliumBundle{},
					Kindnetd:               releasev1.KindnetdBundle{},
					Flux:                   releasev1.FluxBundle{},
					BottleRocketBootstrap:  releasev1.BottlerocketBootstrapBundle{},
					BottleRocketAdmin:      releasev1.BottlerocketAdminBundle{},
					ExternalEtcdBootstrap:  releasev1.EtcdadmBootstrapBundle{},
					ExternalEtcdController: releasev1.EtcdadmControllerBundle{},
					Tinkerbell:             releasev1.TinkerbellBundle{},
				},
			},
		},
	}
}

func eksdRelease() *eksdv1.Release {
	return &eksdv1.Release{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Release",
			APIVersion: "distro.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "eksa-system",
		},
		Spec: eksdv1.ReleaseSpec{
			Number: 1,
		},
		Status: eksdv1.ReleaseStatus{
			Components: []eksdv1.Component{
				{
					Assets: []eksdv1.Asset{
						{
							Name:  "etcd-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "node-driver-registrar-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "livenessprobe-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "external-attacher-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "external-provisioner-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "pause-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "aws-iam-authenticator-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "coredns-image",
							Image: &eksdv1.AssetImage{},
						},
						{
							Name:  "kube-apiserver-image",
							Image: &eksdv1.AssetImage{},
						},
					},
				},
			},
		},
	}
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
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func namespace(name string) *v1.Namespace {
	return &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
