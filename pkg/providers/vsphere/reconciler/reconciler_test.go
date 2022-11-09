package reconciler_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
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
	"github.com/aws/eks-anywhere/pkg/govmomi"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/reconciler"
	vspherereconcilermocks "github.com/aws/eks-anywhere/pkg/providers/vsphere/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	clusterNamespace = "test-namespace"
)

func TestReconcilerReconcileSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	capiCluster := capiCluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.createAllObjs()

	logger := test.NewNullLogger()
	remoteClient := fake.NewClientBuilder().Build()

	tt.govcClient.EXPECT().ValidateVCenterSetupMachineConfig(tt.ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	tt.govcClient.EXPECT().ValidateVCenterSetupMachineConfig(tt.ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	tt.govcClient.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, gomock.Any()).Return("test", nil)
	tt.govcClient.EXPECT().GetTags(tt.ctx, tt.machineConfigControlPlane.Spec.Template).Return([]string{"os:ubuntu", fmt.Sprintf("eksdRelease:%s", tt.bundle.Spec.VersionsBundles[0].EksD.Name)}, nil)
	tt.govcClient.EXPECT().GetWorkloadAvailableSpace(tt.ctx, tt.machineConfigControlPlane.Spec.Datastore).Return(100.0, nil).Times(2)

	tt.remoteClientRegistry.EXPECT().GetClient(
		tt.ctx, client.ObjectKey{Name: "workload-cluster", Namespace: "eksa-system"},
	).Return(remoteClient, nil)
	tt.cniReconciler.EXPECT().Reconcile(tt.ctx, logger, remoteClient, tt.buildSpec())

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcilerReconcileWorkerNodesSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.cluster.Name = "my-management-cluster"
	tt.cluster.SetSelfManaged()
	capiCluster := capiCluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.createAllObjs()

	logger := test.NewNullLogger()

	tt.govcClient.EXPECT().ValidateVCenterSetupMachineConfig(tt.ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	tt.govcClient.EXPECT().ValidateVCenterSetupMachineConfig(tt.ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	tt.govcClient.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, gomock.Any()).Return("test", nil)
	tt.govcClient.EXPECT().GetTags(tt.ctx, tt.machineConfigControlPlane.Spec.Template).Return([]string{"os:ubuntu", fmt.Sprintf("eksdRelease:%s", tt.bundle.Spec.VersionsBundles[0].EksD.Name)}, nil)
	tt.govcClient.EXPECT().GetWorkloadAvailableSpace(tt.ctx, tt.machineConfigControlPlane.Spec.Datastore).Return(100.0, nil).Times(2)

	result, err := tt.reconciler().ReconcileWorkerNodes(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
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
		&vspherev1.VSphereMachineTemplate{
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
}

func TestReconcilerFailToSetUpMachineConfigCP(t *testing.T) {
	tt := newReconcilerTest(t)
	logger := test.NewNullLogger()
	tt.withFakeClient()

	tt.govcClient.EXPECT().ValidateVCenterSetupMachineConfig(tt.ctx, tt.datacenterConfig, tt.machineConfigControlPlane, gomock.Any()).Return(fmt.Errorf("error"))
	tt.govcClient.EXPECT().ValidateVCenterSetupMachineConfig(tt.ctx, tt.datacenterConfig, tt.machineConfigWorker, gomock.Any()).Return(nil).MaxTimes(1)
	tt.govcClient.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, tt.machineConfigControlPlane).Return("test", nil).Times(0)
	tt.govcClient.EXPECT().GetTags(tt.ctx, tt.machineConfigControlPlane.Spec.Template).Return([]string{"os:ubuntu", fmt.Sprintf("eksdRelease:%s", tt.bundle.Spec.VersionsBundles[0].EksD.Name)}, nil).Times(0)
	tt.govcClient.EXPECT().GetWorkloadAvailableSpace(tt.ctx, tt.machineConfigControlPlane.Spec.Datastore).Return(100.0, nil).Times(0)

	result, err := tt.reconciler().ValidateMachineConfigs(tt.ctx, logger, tt.buildSpec())
	g := NewWithT(t)
	g.Expect(err).To(MatchError(ContainSubstring("validating vCenter setup for VSphereMachineConfig")))
	g.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("validating vCenter setup for VSphereMachineConfig"))
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcilerControlPlaneIsNotReady(t *testing.T) {
	tt := newReconcilerTest(t)
	capiCluster := capiCluster(func(c *clusterv1.Cluster) {
		c.Name = tt.cluster.Name
	})
	capiCluster.Status.Conditions = clusterv1.Conditions{
		{
			Type:               clusterapi.ControlPlaneReadyCondition,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.NewTime(time.Now()),
		},
	}
	tt.eksaSupportObjs = append(tt.eksaSupportObjs, capiCluster)
	tt.createAllObjs()

	logger := test.NewNullLogger()

	tt.govcClient.EXPECT().ValidateVCenterSetupMachineConfig(tt.ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	tt.govcClient.EXPECT().ValidateVCenterSetupMachineConfig(tt.ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	tt.govcClient.EXPECT().SearchTemplate(tt.ctx, tt.datacenterConfig.Spec.Datacenter, gomock.Any()).Return("test", nil)
	tt.govcClient.EXPECT().GetTags(tt.ctx, tt.machineConfigControlPlane.Spec.Template).Return([]string{"os:ubuntu", fmt.Sprintf("eksdRelease:%s", tt.bundle.Spec.VersionsBundles[0].EksD.Name)}, nil)
	tt.govcClient.EXPECT().GetWorkloadAvailableSpace(tt.ctx, tt.machineConfigControlPlane.Spec.Datastore).Return(100.0, nil).Times(2)

	result, err := tt.reconciler().Reconcile(tt.ctx, logger, tt.cluster)

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.ResultWithRequeue(30 * time.Second)))
}

func TestReconcilerReconcileWorkersSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	result, err := tt.reconciler().ReconcileWorkers(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcilerInvalidDatacenterConfig(t *testing.T) {
	tt := newReconcilerTest(t)
	logger := test.NewNullLogger()
	tt.datacenterConfig.Status.SpecValid = false
	m := "Something wrong"
	tt.datacenterConfig.Status.FailureMessage = &m
	tt.withFakeClient()

	_, err := tt.reconciler().ValidateDatacenterConfig(tt.ctx, logger, tt.buildSpec())

	tt.Expect(err).To(BeNil(), "error should be nil to prevent requeue")
	tt.Expect(*tt.cluster.Status.FailureMessage).To(ContainSubstring("Something wrong"))
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

func TestReconcilerReconcileControlPlaneSuccess(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()

	result, err := tt.reconciler().ReconcileControlPlane(tt.ctx, test.NewNullLogger(), tt.buildSpec())

	tt.Expect(err).NotTo(HaveOccurred())
	tt.Expect(tt.cluster.Status.FailureMessage).To(BeZero())
	tt.Expect(result).To(Equal(controller.Result{}))

	tt.ShouldEventuallyExist(tt.ctx,
		&addonsv1.ClusterResourceSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workload-cluster-crs-0",
				Namespace: "eksa-system",
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&controlplanev1.KubeadmControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workload-cluster",
				Namespace: "eksa-system",
			},
		},
	)
	tt.ShouldEventuallyExist(tt.ctx,
		&vspherev1.VSphereMachineTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workload-cluster-control-plane-1",
				Namespace: "eksa-system",
			},
		},
	)
	capiCluster := capiCluster(func(c *clusterv1.Cluster) {
		c.Name = "workload-cluster"
	})
	tt.ShouldEventuallyExist(tt.ctx, capiCluster)
}

func TestReconcilerReconcileControlPlaneFailure(t *testing.T) {
	tt := newReconcilerTest(t)
	tt.createAllObjs()
	spec := tt.buildSpec()
	spec.Cluster.Spec.KubernetesVersion = ""

	_, err := tt.reconciler().ReconcileControlPlane(tt.ctx, test.NewNullLogger(), spec)

	tt.Expect(err).To(HaveOccurred())
}

type reconcilerTest struct {
	t testing.TB
	*WithT
	*envtest.APIExpecter
	ctx                       context.Context
	cniReconciler             *vspherereconcilermocks.MockCNIReconciler
	govcClient                *mocks.MockProviderGovcClient
	validator                 *vsphere.Validator
	defaulter                 *vsphere.Defaulter
	remoteClientRegistry      *vspherereconcilermocks.MockRemoteClientRegistry
	cluster                   *anywherev1.Cluster
	client                    client.Client
	env                       *envtest.Environment
	bundle                    *releasev1.Bundles
	eksaSupportObjs           []client.Object
	datacenterConfig          *anywherev1.VSphereDatacenterConfig
	machineConfigControlPlane *anywherev1.VSphereMachineConfig
	machineConfigWorker       *anywherev1.VSphereMachineConfig
}

func newReconcilerTest(t testing.TB) *reconcilerTest {
	ctrl := gomock.NewController(t)
	cniReconciler := vspherereconcilermocks.NewMockCNIReconciler(ctrl)
	remoteClientRegistry := vspherereconcilermocks.NewMockRemoteClientRegistry(ctrl)
	c := env.Client()

	govcClient := mocks.NewMockProviderGovcClient(ctrl)
	vcb := govmomi.NewVMOMIClientBuilder()
	validator := vsphere.NewValidator(govcClient, &networkutils.DefaultNetClient{}, vcb)
	defaulter := vsphere.NewDefaulter(govcClient)

	bundle := createBundle()

	managementCluster := vsphereCluster(func(c *anywherev1.Cluster) {
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

	machineConfigCP := machineConfig(func(m *anywherev1.VSphereMachineConfig) {
		m.Name = "cp-machine-config"
	})
	machineConfigWN := machineConfig(func(m *anywherev1.VSphereMachineConfig) {
		m.Name = "worker-machine-config"
	})

	credentialsSecret := createSecret()
	workloadClusterDatacenter := dataCenter(func(d *anywherev1.VSphereDatacenterConfig) {
		d.Status.SpecValid = true
	})

	cluster := vsphereCluster(func(c *anywherev1.Cluster) {
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
				Kind: anywherev1.VSphereMachineConfigKind,
				Name: machineConfigCP.Name,
			},
		}
		c.Spec.DatacenterRef = anywherev1.Ref{
			Kind: anywherev1.VSphereDatacenterKind,
			Name: workloadClusterDatacenter.Name,
		}

		c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations,
			anywherev1.WorkerNodeGroupConfiguration{
				Count: ptr.Int(1),
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.VSphereMachineConfigKind,
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
		cniReconciler:        cniReconciler,
		govcClient:           govcClient,
		validator:            validator,
		defaulter:            defaulter,
		remoteClientRegistry: remoteClientRegistry,
		client:               c,
		env:                  env,
		eksaSupportObjs: []client.Object{
			namespace(clusterNamespace),
			namespace(constants.EksaSystemNamespace),
			managementCluster,
			workloadClusterDatacenter,
			bundle,
			eksdRelease(),
			credentialsSecret,
		},
		bundle:                    bundle,
		cluster:                   cluster,
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
	tt.DeleteAllOfAndWait(tt.ctx, &vspherev1.VSphereMachineTemplate{})
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
	return reconciler.New(tt.client, tt.validator, tt.defaulter, tt.cniReconciler, tt.remoteClientRegistry)
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

func vsphereCluster(opts ...clusterOpt) *anywherev1.Cluster {
	c := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterNamespace,
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.20",
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

type datacenterOpt func(config *anywherev1.VSphereDatacenterConfig)

func dataCenter(opts ...datacenterOpt) *anywherev1.VSphereDatacenterConfig {
	d := &anywherev1.VSphereDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.VSphereDatacenterKind,
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

func createBundle() *releasev1.Bundles {
	return &releasev1.Bundles{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Bundles",
			APIVersion: releasev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bundles-1",
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
							Name: "kube-apiserver-image",
							Image: &eksdv1.AssetImage{
								URI: "public.ecr.aws/eks-distro/kubernetes/kube-apiserver:v1.19.8",
							},
						},
					},
				},
			},
		},
	}
}

type vsphereMachineOpt func(config *anywherev1.VSphereMachineConfig)

func machineConfig(opts ...vsphereMachineOpt) *anywherev1.VSphereMachineConfig {
	m := &anywherev1.VSphereMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.VSphereMachineConfigKind,
			APIVersion: anywherev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterNamespace,
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
					SshAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC8ZEibIrz1AUBKDvmDiWLs9f5DnOerC4qPITiDtSOuPAsxgZbRMavBfVTxodMdAkYRYlXxK6PqNo0ve0qcOV2yvpxH1OogasMMetck6BlM/dIoo3vEY4ZoG9DuVRIf9Iry5gJKbpMDYWpx1IGZrDMOFcIM20ii2qLQQk5hfq9OqdqhToEJFixdgJt/y/zt6Koy3kix+XsnrVdAHgWAq4CZuwt1G6JUAqrpob3H8vPmL7aS+35ktf0pHBm6nYoxRhslnWMUb/7vpzWiq+fUBIm2LYqvrnm7t3fRqFx7p2sZqAm2jDNivyYXwRXkoQPR96zvGeMtuQ5BVGPpsDfVudSW21+pEXHI0GINtTbua7Ogz7wtpVywSvHraRgdFOeY9mkXPzvm2IhoqNrteck2GErwqSqb19mPz6LnHueK0u7i6WuQWJn0CUoCtyMGIrowXSviK8qgHXKrmfTWATmCkbtosnLskNdYuOw8bKxq5S4WgdQVhPps2TiMSZndjX5NTr8= ubuntu@ip-10-2-0-6"},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func namespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

type capiClusterOpt func(*clusterv1.Cluster)

func capiCluster(opts ...capiClusterOpt) *clusterv1.Cluster {
	c := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: clusterv1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Status: clusterv1.ClusterStatus{
			Conditions: clusterv1.Conditions{
				{
					Type:               clusterapi.ControlPlaneReadyCondition,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now()),
				},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func createSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "eksa-system",
			Name:      vsphere.CredentialsObjectName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Data: map[string][]byte{
			"username": []byte("test"),
			"password": []byte("test"),
		},
	}
}
