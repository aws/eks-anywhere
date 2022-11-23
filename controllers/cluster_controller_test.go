package controllers_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/govmomi"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	vspheremocks "github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
	vspherereconciler "github.com/aws/eks-anywhere/pkg/providers/vsphere/reconciler"
	vspherereconcilermocks "github.com/aws/eks-anywhere/pkg/providers/vsphere/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	clusterFinalizerName = "clusters.anywhere.eks.amazonaws.com/finalizer"
)

type vsphereClusterReconcilerTest struct {
	govcClient *vspheremocks.MockProviderGovcClient
	reconciler *controllers.ClusterReconciler
	client     client.Client
}

func newVsphereClusterReconcilerTest(t *testing.T, objs ...runtime.Object) *vsphereClusterReconcilerTest {
	ctrl := gomock.NewController(t)
	govcClient := vspheremocks.NewMockProviderGovcClient(ctrl)

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	vcb := govmomi.NewVMOMIClientBuilder()

	validator := vsphere.NewValidator(govcClient, &networkutils.DefaultNetClient{}, vcb)
	defaulter := vsphere.NewDefaulter(govcClient)
	cniReconciler := vspherereconcilermocks.NewMockCNIReconciler(ctrl)

	reconciler := vspherereconciler.New(
		cl,
		validator,
		defaulter,
		cniReconciler,
		nil,
	)
	registry := clusters.NewProviderClusterReconcilerRegistryBuilder().
		Add(anywherev1.VSphereDatacenterKind, reconciler).
		Build()

	r := controllers.NewClusterReconciler(cl, &registry)

	return &vsphereClusterReconcilerTest{
		govcClient: govcClient,
		reconciler: r,
		client:     cl,
	}
}

func TestClusterReconcilerReconcileSelfManagedCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	selfManagedCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-management-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			BundlesRef: &anywherev1.BundlesRef{
				Name: "my-bundles-ref",
			},
		},
	}

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	registry := newRegistryMock(providerReconciler)
	c := fake.NewClientBuilder().WithRuntimeObjects(selfManagedCluster).Build()

	providerReconciler.EXPECT().ReconcileWorkerNodes(ctx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(selfManagedCluster))

	r := controllers.NewClusterReconciler(c, registry)
	result, err := r.Reconcile(ctx, clusterRequest(selfManagedCluster))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))
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
		},
	}

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	registry := newRegistryMock(providerReconciler)
	c := fake.NewClientBuilder().WithRuntimeObjects(selfManagedCluster).Build()

	r := controllers.NewClusterReconciler(c, registry)
	_, err := r.Reconcile(ctx, clusterRequest(selfManagedCluster))
	g.Expect(err).To(MatchError(ContainSubstring("deleting self-managed clusters is not supported")))
}

func TestClusterReconcilerDeleteExistingCAPIClusterSuccess(t *testing.T) {
	secret := createSecret()
	managementCluster := createCluster()
	managementCluster.Name = "management-cluster"
	cluster := createCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}
	now := metav1.Now()
	cluster.DeletionTimestamp = &now

	datacenterConfig := createDataCenter(cluster)
	bundle := createBundle(managementCluster)
	machineConfigCP := createCPMachineConfig()
	machineConfigWN := createWNMachineConfig()

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
}

func TestClusterReconcilerDeleteNoCAPIClusterSuccess(t *testing.T) {
	g := NewWithT(t)

	secret := createSecret()
	managementCluster := createCluster()
	managementCluster.Name = "management-cluster"
	cluster := createCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}
	now := metav1.Now()
	cluster.DeletionTimestamp = &now

	datacenterConfig := createDataCenter(cluster)
	bundle := createBundle(managementCluster)
	machineConfigCP := createCPMachineConfig()
	machineConfigWN := createWNMachineConfig()

	objs := []runtime.Object{cluster, datacenterConfig, secret, bundle, machineConfigCP, machineConfigWN, managementCluster}

	g.Expect(cluster.Finalizers).NotTo(ContainElement(clusterFinalizerName))

	tt := newVsphereClusterReconcilerTest(t, objs...)

	req := clusterRequest(cluster)

	ctx := context.Background()

	controllerutil.AddFinalizer(cluster, clusterFinalizerName)
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
}

func createWNMachineConfig() *anywherev1.VSphereMachineConfig {
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

func createCPMachineConfig() *anywherev1.VSphereMachineConfig {
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

func createBundle(cluster *anywherev1.Cluster) *v1alpha1.Bundles {
	return &v1alpha1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: "default",
		},
		Spec: v1alpha1.BundlesSpec{
			VersionsBundles: []v1alpha1.VersionsBundle{
				{
					KubeVersion: "1.20",
					EksD: v1alpha1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "testdata/release.yaml",
						KubeVersion:    "1.20",
					},
					CertManager:            v1alpha1.CertManagerBundle{},
					ClusterAPI:             v1alpha1.CoreClusterAPI{},
					Bootstrap:              v1alpha1.KubeadmBootstrapBundle{},
					ControlPlane:           v1alpha1.KubeadmControlPlaneBundle{},
					VSphere:                v1alpha1.VSphereBundle{},
					Docker:                 v1alpha1.DockerBundle{},
					Eksa:                   v1alpha1.EksaBundle{},
					Cilium:                 v1alpha1.CiliumBundle{},
					Kindnetd:               v1alpha1.KindnetdBundle{},
					Flux:                   v1alpha1.FluxBundle{},
					BottleRocketBootstrap:  v1alpha1.BottlerocketBootstrapBundle{},
					BottleRocket:           v1alpha1.BottlerocketBundle{},
					ExternalEtcdBootstrap:  v1alpha1.EtcdadmBootstrapBundle{},
					ExternalEtcdController: v1alpha1.EtcdadmControllerBundle{},
					Tinkerbell:             v1alpha1.TinkerbellBundle{},
				},
			},
		},
	}
}

func createDataCenter(cluster *anywherev1.Cluster) *anywherev1.VSphereDatacenterConfig {
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

func createCluster() *anywherev1.Cluster {
	return &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: anywherev1.ClusterSpec{
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
					Name: name + "-cp",
				},
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: ptr.Int(1),
					MachineGroupRef: &anywherev1.Ref{
						Kind: "VSphereMachineConfig",
						Name: name + "-wn",
					},
					Name:   "md-0",
					Labels: nil,
				},
			},
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
