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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/govmomi"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	vspheremocks "github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
	vspherereconciler "github.com/aws/eks-anywhere/pkg/providers/vsphere/reconciler"
	vspherereconcilermocks "github.com/aws/eks-anywhere/pkg/providers/vsphere/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
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
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	clusterValidator := mocks.NewMockClusterValidator(ctrl)

	vcb := govmomi.NewVMOMIClientBuilder()

	validator := vsphere.NewValidator(govcClient, vcb)
	defaulter := vsphere.NewDefaulter(govcClient)
	cniReconciler := vspherereconcilermocks.NewMockCNIReconciler(ctrl)
	ipValidator := vspherereconcilermocks.NewMockIPValidator(ctrl)

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

	r := controllers.NewClusterReconciler(cl, &registry, iam, clusterValidator, mockPkgs)

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
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	clusterValidator := mocks.NewMockClusterValidator(controller)
	registry := newRegistryMock(providerReconciler)
	c := fake.NewClientBuilder().WithRuntimeObjects(selfManagedCluster).Build()
	mockPkgs := mocks.NewMockPackagesClient(controller)
	providerReconciler.EXPECT().ReconcileWorkerNodes(ctx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(selfManagedCluster))

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, mockPkgs)
	result, err := r.Reconcile(ctx, clusterRequest(selfManagedCluster))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))
}

func TestClusterReconcilerReconcilePausedCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	managementCluster := createCluster()
	managementCluster.Name = "management-cluster"
	cluster := createCluster()
	cluster.SetManagedBy(managementCluster.Name)
	capiCluster := newCAPICluster(cluster.Name, cluster.Namespace)

	// Mark as paused
	cluster.PauseReconcile()

	c := fake.NewClientBuilder().WithRuntimeObjects(
		managementCluster, cluster, capiCluster,
	).Build()

	ctrl := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(ctrl)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	clusterValidator := mocks.NewMockClusterValidator(ctrl)
	registry := newRegistryMock(providerReconciler)
	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, nil)
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
		},
	}

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	clusterValidator := mocks.NewMockClusterValidator(controller)
	registry := newRegistryMock(providerReconciler)
	c := fake.NewClientBuilder().WithRuntimeObjects(selfManagedCluster).Build()

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, nil)
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

func TestClusterReconcilerReconcileDeletePausedCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	managementCluster := createCluster()
	managementCluster.Name = "management-cluster"
	cluster := createCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}
	controllerutil.AddFinalizer(cluster, controllers.ClusterFinalizerName)
	capiCluster := newCAPICluster(cluster.Name, cluster.Namespace)

	// Mark cluster for deletion
	now := metav1.Now()
	cluster.DeletionTimestamp = &now

	// Mark as paused
	cluster.PauseReconcile()

	controller := gomock.NewController(t)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	clusterValidator := mocks.NewMockClusterValidator(controller)
	c := fake.NewClientBuilder().WithRuntimeObjects(
		managementCluster, cluster, capiCluster,
	).Build()

	r := controllers.NewClusterReconciler(c, newRegistryForDummyProviderReconciler(), iam, clusterValidator, nil)
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
	managementCluster := createCluster()
	managementCluster.Name = "management-cluster"
	cluster := createCluster()
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

	r := controllers.NewClusterReconciler(c, newRegistryForDummyProviderReconciler(), iam, clusterValidator, nil)
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
}

func TestClusterReconcilerSkipDontInstallPackagesOnSelfManaged(t *testing.T) {
	ctx := context.Background()
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
			ManagementCluster: anywherev1.ManagementCluster{
				Name: "",
			},
		},
	}
	objs := []runtime.Object{cluster}
	cb := fake.NewClientBuilder()
	mockClient := cb.WithRuntimeObjects(objs...).Build()
	nullRegistry := newRegistryForDummyProviderReconciler()

	ctrl := gomock.NewController(t)
	mockPkgs := mocks.NewMockPackagesClient(ctrl)
	mockPkgs.EXPECT().ReconcileDelete(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	r := controllers.NewClusterReconciler(mockClient, nullRegistry, nil, nil, mockPkgs)
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
			ManagementCluster: anywherev1.ManagementCluster{
				Name: "",
			},
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
	r := controllers.NewClusterReconciler(mockClient, nullRegistry, nil, nil, mockPkgs)
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
				ManagementCluster: anywherev1.ManagementCluster{
					Name: "my-management-cluster",
				},
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
		mockPkgs.EXPECT().ReconcileDelete(logCtx, log, gomock.Any(), gomock.Any()).Return(fmt.Errorf("test error"))
		mockIAM := mocks.NewMockAWSIamConfigReconciler(ctrl)
		mockValid := mocks.NewMockClusterValidator(ctrl)

		r := controllers.NewClusterReconciler(fakeClient, nullRegistry, mockIAM, mockValid, mockPkgs)
		_, err := r.Reconcile(logCtx, clusterRequest(cluster))
		if err == nil || !strings.Contains(err.Error(), "test error") {
			t.Errorf("expected packages client deletion error, got %s", err)
		}
	})
}

func TestClusterReconcilerPackagesInstall(s *testing.T) {
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
				ManagementCluster: anywherev1.ManagementCluster{
					Name: "my-management-cluster",
				},
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
		objs := []runtime.Object{cluster, bundles, secret}
		fakeClient := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		nullRegistry := newRegistryForDummyProviderReconciler()
		mockIAM := mocks.NewMockAWSIamConfigReconciler(ctrl)
		mockValid := mocks.NewMockClusterValidator(ctrl)
		mockValid.EXPECT().ValidateManagementClusterName(logCtx, log, gomock.Any()).Return(nil)
		mockPkgs := mocks.NewMockPackagesClient(ctrl)
		mockPkgs.EXPECT().
			EnableFullLifecycle(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		r := controllers.NewClusterReconciler(fakeClient, nullRegistry, mockIAM, mockValid, mockPkgs)
		_, err := r.Reconcile(logCtx, clusterRequest(cluster))
		if err != nil {
			t.Errorf("expected nil error, got %s", err)
		}
	})
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
					CertManager:                v1alpha1.CertManagerBundle{},
					ClusterAPI:                 v1alpha1.CoreClusterAPI{},
					Bootstrap:                  v1alpha1.KubeadmBootstrapBundle{},
					ControlPlane:               v1alpha1.KubeadmControlPlaneBundle{},
					VSphere:                    v1alpha1.VSphereBundle{},
					Docker:                     v1alpha1.DockerBundle{},
					Eksa:                       v1alpha1.EksaBundle{},
					Cilium:                     v1alpha1.CiliumBundle{},
					Kindnetd:                   v1alpha1.KindnetdBundle{},
					Flux:                       v1alpha1.FluxBundle{},
					BottleRocketHostContainers: v1alpha1.BottlerocketHostContainersBundle{},
					ExternalEtcdBootstrap:      v1alpha1.EtcdadmBootstrapBundle{},
					ExternalEtcdController:     v1alpha1.EtcdadmControllerBundle{},
					Tinkerbell:                 v1alpha1.TinkerbellBundle{},
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
