package controllers_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestClusterReconcilerEnsureOwnerReferences(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.Spec.IdentityProviderRefs = []anywherev1.Ref{
		{
			Kind: anywherev1.OIDCConfigKind,
			Name: "my-oidc",
		},
		{
			Kind: anywherev1.AWSIamConfigKind,
			Name: "my-iam",
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	oidc := &anywherev1.OIDCConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-oidc",
			Namespace: cluster.Namespace,
		},
	}
	awsIAM := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-iam",
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       cluster.Name,
				},
			},
		},
	}
	bundles := createBundle()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}
	eksaRelease := test.EKSARelease()

	objs := []runtime.Object{cluster, managementCluster, oidc, awsIAM, bundles, secret, eksaRelease}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	iam := newMockAWSIamConfigReconciler(t)
	iam.EXPECT().EnsureCASecret(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)
	iam.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	pcc := newMockPackagesClient(t)
	pcc.EXPECT().Reconcile(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	mhc := newMockMachineHealthCheckReconciler(t)
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), iam, validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))

	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: bundles.Namespace, Name: bundles.Name}, bundles)).To(Succeed())
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: constants.EksaSystemNamespace, Name: cluster.Name + "-kubeconfig"}, secret)).To(Succeed())

	g.Expect(err).NotTo(HaveOccurred())

	newOidc := &anywherev1.OIDCConfig{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: "my-oidc"}, newOidc)).To(Succeed())
	g.Expect(newOidc.OwnerReferences).To(HaveLen(1))
	g.Expect(newOidc.OwnerReferences[0].Name).To(Equal(cluster.Name))

	newAWSIam := &anywherev1.AWSIamConfig{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: "my-iam"}, newAWSIam)).To(Succeed())
	g.Expect(newAWSIam.OwnerReferences).To(HaveLen(1))
	g.Expect(newAWSIam.OwnerReferences[0]).To(Equal(awsIAM.OwnerReferences[0]))
}

func TestClusterReconcilerReconcileChildObjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "my-namespace",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "my-namespace",
		},
	}
	cluster.Spec.IdentityProviderRefs = []anywherev1.Ref{
		{
			Kind: anywherev1.OIDCConfigKind,
			Name: "my-oidc",
		},
		{
			Kind: anywherev1.AWSIamConfigKind,
			Name: "my-iam",
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	objs := []runtime.Object{cluster, managementCluster}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()
	api := envtest.NewAPIExpecter(t, cl)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), newMockClusterValidator(t), nil, newMockMachineHealthCheckReconciler(t), nil)
	g.Expect(r.Reconcile(ctx, clusterRequest(cluster))).Error().To(MatchError(ContainSubstring("not found")))
	c := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Status.FailureMessage).To(HaveValue(Equal(
			"Dependent cluster objects don't exist: oidcconfigs.anywhere.eks.amazonaws.com \"my-oidc\" not found",
		)))
		g.Expect(c.Status.FailureReason).To(HaveValue(Equal(anywherev1.MissingDependentObjectsReason)))
	})
}

func TestClusterReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewClusterReconciler(client, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), newMockClusterValidator(t), nil, nil, nil)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager(), env.Manager().GetLogger())).To(Succeed())
}

func TestClusterReconcilerManagementClusterNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-management-cluster",
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "my-namespace",
		},
	}
	cluster.SetManagedBy("my-management-cluster-2")

	objs := []runtime.Object{cluster, managementCluster}
	cb := fake.NewClientBuilder()
	cb.WithIndex(&anywherev1.Cluster{}, "metadata.name", clientutil.ClusterNameIndexer)
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()
	api := envtest.NewAPIExpecter(t, cl)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), newMockClusterValidator(t), nil, nil, nil)
	g.Expect(r.Reconcile(ctx, clusterRequest(cluster))).Error().To(BeNil())

	c := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Status.FailureMessage).To(HaveValue(Equal("Management cluster my-management-cluster-2 does not exist")))
		g.Expect(c.Status.FailureReason).To(HaveValue(Equal(anywherev1.ManagementClusterRefInvalidReason)))
	})
}

func TestClusterReconcilerSetBundlesRef(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-management-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "bundles-1",
				Namespace: "default",
			},
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "bundles-1",
				Namespace: "default",
			},
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}
	bundles := createBundle()

	objs := []runtime.Object{cluster, managementCluster, secret, bundles}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	mgmtCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: managementCluster.Name}, mgmtCluster)).To(Succeed())
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Spec.BundlesRef.Namespace, Name: cluster.Spec.BundlesRef.Name}, bundles)).To(Succeed())
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: constants.EksaSystemNamespace, Name: cluster.Name + "-kubeconfig"}, secret)).To(Succeed())
	pcc := newMockPackagesClient(t)
	pcc.EXPECT().Reconcile(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	mhc := newMockMachineHealthCheckReconciler(t)
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).ToNot(HaveOccurred())

	newCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: "my-cluster"}, newCluster)).To(Succeed())
	g.Expect(newCluster.Spec.BundlesRef).To(Equal(mgmtCluster.Spec.BundlesRef))
}

func TestClusterReconcilerSetDefaultEksaVersion(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	bundles := createBundle()
	cluster.SetManagedBy("my-management-cluster")

	objs := []runtime.Object{cluster, managementCluster, test.EKSARelease(), bundles}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	mgmtCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: managementCluster.Name}, mgmtCluster)).To(Succeed())
	pcc := newMockPackagesClient(t)
	pcc.EXPECT().Reconcile(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	mhc := newMockMachineHealthCheckReconciler(t)
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).ToNot(HaveOccurred())

	newCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: "my-cluster"}, newCluster)).To(Succeed())
	g.Expect(newCluster.Spec.EksaVersion).To(Equal(mgmtCluster.Spec.EksaVersion))
}

func TestClusterReconcilerWorkloadClusterMgmtClusterNameFail(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "my-namespace",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "my-namespace",
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")
	// clusterSpec := &c.Spec{
	// 	Config: &c.Config{
	// 		Cluster: cluster,
	// 	},
	// }

	objs := []runtime.Object{cluster, managementCluster}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).
		Return(errors.New("test error"))

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), validator, nil, nil, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(HaveOccurred())

	api := envtest.NewAPIExpecter(t, cl)
	c := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Status.FailureMessage).To(HaveValue(Equal("test error")))
		g.Expect(c.Status.FailureReason).To(HaveValue(Equal(anywherev1.ManagementClusterRefInvalidReason)))
	})
}

func TestClusterReconcilerNoBundleFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := anywherev1.EksaVersion("v0.22.0")

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	kcp := testKubeadmControlPlaneFromCluster(cluster)

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(controller)

	clusterValidator := mocks.NewMockClusterValidator(controller)
	registry := newRegistryMock(providerReconciler)
	eksaReleaseV022 := test.EKSARelease()
	eksaReleaseV022.Name = "eksa-v0-22-0"
	eksaReleaseV022.Spec.Version = "eksa-v0-22-0"
	c := fake.NewClientBuilder().WithRuntimeObjects(cluster, kcp, eksaReleaseV022).
		WithStatusSubresource(cluster).
		Build()
	mockPkgs := mocks.NewMockPackagesClient(controller)

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, mockPkgs, mhcReconciler, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(MatchError(ContainSubstring("getting bundle for cluster")))
}

func TestClusterReconcilerFailSignatureValidation(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := anywherev1.EksaVersion("v0.22.0")

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	eksaReleaseV022 := test.EKSARelease()
	eksaReleaseV022.Name = "eksa-v0-22-0"
	eksaReleaseV022.Spec.Version = "eksa-v0-22-0"
	bundles := createBundle()
	bundles.Spec.VersionsBundles[0].KubeVersion = "1.25"

	kcp := testKubeadmControlPlaneFromCluster(cluster)

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(controller)

	clusterValidator := mocks.NewMockClusterValidator(controller)
	registry := newRegistryMock(providerReconciler)
	eksdRelease := createEKSDRelease()
	c := fake.NewClientBuilder().WithRuntimeObjects(cluster, kcp, eksaReleaseV022, bundles, eksdRelease).
		WithStatusSubresource(cluster).
		Build()
	mockPkgs := mocks.NewMockPackagesClient(controller)

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, mockPkgs, mhcReconciler, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(MatchError(ContainSubstring("validating bundle signature")))
}

func TestClusterReconcilerUpdatesCertificateStatusSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
			Conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ReadyCondition,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	kcp := testKubeadmControlPlaneFromCluster(cluster)
	eksaRelease := test.EKSARelease()
	bundles := createBundle()

	ctrl := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(ctrl)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	clusterValidator := mocks.NewMockClusterValidator(ctrl)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(ctrl)
	mockPkgs := mocks.NewMockPackagesClient(ctrl)

	providerReconciler.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)
	mhcReconciler.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	registry := newRegistryMock(providerReconciler)
	c := fake.NewClientBuilder().WithRuntimeObjects(cluster, kcp, eksaRelease, bundles).
		WithStatusSubresource(cluster).
		Build()

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, mockPkgs, mhcReconciler, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).ToNot(HaveOccurred())
}

// MockClient that fails on MachineList operations.
type MockClient struct {
	client.Client
}

func (m *MockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	// Check if we're trying to list MachineList objects
	if _, ok := list.(*clusterv1.MachineList); ok {
		return fmt.Errorf("simulated client error during machine list operation")
	}
	// For all other list operations, delegate to the real client
	return m.Client.List(ctx, list, opts...)
}

func TestClusterReconcilerUpdatesCertificateStatusError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
			Conditions: []anywherev1.Condition{
				{
					Type:   anywherev1.ReadyCondition,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	kcp := testKubeadmControlPlaneFromCluster(cluster)
	eksaRelease := test.EKSARelease()
	bundles := createBundle()

	ctrl := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(ctrl)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	clusterValidator := mocks.NewMockClusterValidator(ctrl)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(ctrl)
	mockPkgs := mocks.NewMockPackagesClient(ctrl)

	providerReconciler.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)
	mhcReconciler.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	registry := newRegistryMock(providerReconciler)
	c := fake.NewClientBuilder().WithRuntimeObjects(cluster, kcp, eksaRelease, bundles).
		WithStatusSubresource(cluster).
		Build()
	failingClient := &MockClient{Client: c}

	r := controllers.NewClusterReconciler(failingClient, registry, iam, clusterValidator, mockPkgs, mhcReconciler, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(HaveOccurred())
}

func newRegistryForDummyProviderReconciler() controllers.ProviderClusterReconcilerRegistry {
	return newRegistryMock(dummyProviderReconciler{})
}

func newRegistryMock(reconciler clusters.ProviderClusterReconciler) dummyProviderReconcilerRegistry {
	return dummyProviderReconcilerRegistry{
		reconciler: reconciler,
	}
}

type dummyProviderReconcilerRegistry struct {
	reconciler clusters.ProviderClusterReconciler
}

func (d dummyProviderReconcilerRegistry) Get(_ string) clusters.ProviderClusterReconciler {
	return d.reconciler
}

type dummyProviderReconciler struct{}

func (dummyProviderReconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	return controller.Result{}, nil
}

func (dummyProviderReconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	return controller.Result{}, nil
}

func clusterRequest(cluster *anywherev1.Cluster) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}
}

func nullLog() logr.Logger {
	return logr.New(logf.NullLogSink{})
}

func newMockAWSIamConfigReconciler(t *testing.T) *mocks.MockAWSIamConfigReconciler {
	ctrl := gomock.NewController(t)
	return mocks.NewMockAWSIamConfigReconciler(ctrl)
}

func newMockClusterValidator(t *testing.T) *mocks.MockClusterValidator {
	ctrl := gomock.NewController(t)
	return mocks.NewMockClusterValidator(ctrl)
}

func newMockPackagesClient(t *testing.T) *mocks.MockPackagesClient {
	ctrl := gomock.NewController(t)
	return mocks.NewMockPackagesClient(ctrl)
}

func newMockMachineHealthCheckReconciler(t *testing.T) *mocks.MockMachineHealthCheckReconciler {
	ctrl := gomock.NewController(t)
	return mocks.NewMockMachineHealthCheckReconciler(ctrl)
}

func TestClusterReconcilerCleanupOrphanedAWSIamConfigSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
			UID:       "cluster-uid-123",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	// Create an orphaned AWS IAM config owned by this cluster
	orphanedAWSIam := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "orphaned-aws-iam",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
		},
	}

	bundles := createBundle()
	eksaRelease := test.EKSARelease()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}

	objs := []runtime.Object{cluster, managementCluster, orphanedAWSIam, bundles, eksaRelease, secret}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	ctrl := gomock.NewController(t)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	validator := mocks.NewMockClusterValidator(ctrl)
	pcc := mocks.NewMockPackagesClient(ctrl)
	mhc := mocks.NewMockMachineHealthCheckReconciler(ctrl)

	// Expect cleanup calls for orphaned AWS IAM config
	iam.EXPECT().ReconcileWorkloadClusterDelete(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any(), orphanedAWSIam).Return(nil)
	iam.EXPECT().ReconcileDelete(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)

	// Since cleanup happens in postClusterProviderReconcile, the reconciliation continues
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)
	pcc.EXPECT().Reconcile(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), iam, validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).ToNot(HaveOccurred())

	// Verify the orphaned AWS IAM config was deleted
	deletedConfig := &anywherev1.AWSIamConfig{}
	err = cl.Get(ctx, client.ObjectKey{Namespace: orphanedAWSIam.Namespace, Name: orphanedAWSIam.Name}, deletedConfig)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("not found"))
}

func TestClusterReconcilerCleanupOrphanedAWSIamConfigListError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
			UID:       "cluster-uid-123",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	bundles := createBundle()
	eksaRelease := test.EKSARelease()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}

	objs := []runtime.Object{cluster, managementCluster, bundles, eksaRelease, secret}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	// Create a mock client that fails on AWSIamConfigList operations
	mockClient := &MockAWSIamConfigListClient{Client: cl}

	ctrl := gomock.NewController(t)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	validator := mocks.NewMockClusterValidator(ctrl)
	pcc := mocks.NewMockPackagesClient(ctrl)
	mhc := mocks.NewMockMachineHealthCheckReconciler(ctrl)

	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)

	r := controllers.NewClusterReconciler(mockClient, newRegistryForDummyProviderReconciler(), iam, validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("listing AWSIamConfig objects"))
}

func TestClusterReconcilerCleanupOrphanedAWSIamConfigWorkloadCleanupError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
			UID:       "cluster-uid-123",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	// Create an orphaned AWS IAM config owned by this cluster
	orphanedAWSIam := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "orphaned-aws-iam",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
		},
	}

	bundles := createBundle()
	eksaRelease := test.EKSARelease()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}

	objs := []runtime.Object{cluster, managementCluster, orphanedAWSIam, bundles, eksaRelease, secret}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	ctrl := gomock.NewController(t)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	validator := mocks.NewMockClusterValidator(ctrl)
	pcc := mocks.NewMockPackagesClient(ctrl)
	mhc := mocks.NewMockMachineHealthCheckReconciler(ctrl)

	// Expect workload cleanup to fail
	workloadCleanupError := errors.New("workload cleanup failed")
	iam.EXPECT().ReconcileWorkloadClusterDelete(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any(), orphanedAWSIam).Return(workloadCleanupError)

	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), iam, validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("cleaning up workload cluster AWS IAM resources"))
}

func TestClusterReconcilerCleanupOrphanedAWSIamConfigManagementCleanupError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
			UID:       "cluster-uid-123",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	// Create an orphaned AWS IAM config owned by this cluster
	orphanedAWSIam := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "orphaned-aws-iam",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
		},
	}

	bundles := createBundle()
	eksaRelease := test.EKSARelease()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}

	objs := []runtime.Object{cluster, managementCluster, orphanedAWSIam, bundles, eksaRelease, secret}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	ctrl := gomock.NewController(t)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	validator := mocks.NewMockClusterValidator(ctrl)
	pcc := mocks.NewMockPackagesClient(ctrl)
	mhc := mocks.NewMockMachineHealthCheckReconciler(ctrl)

	// Expect workload cleanup to succeed but management cleanup to fail
	managementCleanupError := errors.New("management cleanup failed")
	iam.EXPECT().ReconcileWorkloadClusterDelete(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any(), orphanedAWSIam).Return(nil)
	iam.EXPECT().ReconcileDelete(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(managementCleanupError)

	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), iam, validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("cleaning up management cluster AWS IAM resources"))
}

func TestClusterReconcilerCleanupOrphanedAWSIamConfigDeleteError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
			UID:       "cluster-uid-123",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	// Create an orphaned AWS IAM config owned by this cluster
	orphanedAWSIam := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "orphaned-aws-iam",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
		},
	}

	bundles := createBundle()
	eksaRelease := test.EKSARelease()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}

	objs := []runtime.Object{cluster, managementCluster, orphanedAWSIam, bundles, eksaRelease, secret}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	// Create a mock client that fails on Delete operations for AWSIamConfig
	mockClient := &MockAWSIamConfigDeleteClient{Client: cl}

	ctrl := gomock.NewController(t)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	validator := mocks.NewMockClusterValidator(ctrl)
	pcc := mocks.NewMockPackagesClient(ctrl)
	mhc := mocks.NewMockMachineHealthCheckReconciler(ctrl)

	// Expect cleanup calls to succeed
	iam.EXPECT().ReconcileWorkloadClusterDelete(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any(), orphanedAWSIam).Return(nil)
	iam.EXPECT().ReconcileDelete(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)

	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)

	r := controllers.NewClusterReconciler(mockClient, newRegistryForDummyProviderReconciler(), iam, validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("deleting orphaned AWSIamConfig"))
}

func TestClusterReconcilerNoOrphanedAWSIamConfig(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
			UID:       "cluster-uid-123",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	bundles := createBundle()
	eksaRelease := test.EKSARelease()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}

	objs := []runtime.Object{cluster, bundles, eksaRelease, secret}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	ctrl := gomock.NewController(t)
	iam := mocks.NewMockAWSIamConfigReconciler(ctrl)
	validator := mocks.NewMockClusterValidator(ctrl)
	pcc := mocks.NewMockPackagesClient(ctrl)
	mhc := mocks.NewMockMachineHealthCheckReconciler(ctrl)

	// No AWS IAM cleanup calls should be made since there are no orphaned configs
	// This is a self-managed cluster, so no validation or packages reconciliation
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.Any()).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), iam, validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).ToNot(HaveOccurred())
}

// MockAWSIamConfigListClient that fails on AWSIamConfigList operations.
type MockAWSIamConfigListClient struct {
	client.Client
}

func (m *MockAWSIamConfigListClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	// Check if we're trying to list AWSIamConfigList objects
	if _, ok := list.(*anywherev1.AWSIamConfigList); ok {
		return fmt.Errorf("simulated client error during AWSIamConfig list operation")
	}
	// For all other list operations, delegate to the real client
	return m.Client.List(ctx, list, opts...)
}

// MockAWSIamConfigDeleteClient that fails on Delete operations for AWSIamConfig.
type MockAWSIamConfigDeleteClient struct {
	client.Client
}

func (m *MockAWSIamConfigDeleteClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	// Check if we're trying to delete an AWSIamConfig object
	if _, ok := obj.(*anywherev1.AWSIamConfig); ok {
		return fmt.Errorf("simulated client error during AWSIamConfig delete operation")
	}
	// For all other delete operations, delegate to the real client
	return m.Client.Delete(ctx, obj, opts...)
}
