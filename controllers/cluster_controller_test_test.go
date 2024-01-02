package controllers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestClusterReconcilerEnsureOwnerReferences(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()
	bundlesRef := &anywherev1.BundlesRef{
		Name:      "my-bundles-ref",
		Namespace: "my-namespace",
	}

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
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "v1.25",
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
	bundles := &v1alpha1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-bundles-ref",
			Namespace: cluster.Namespace,
		},
		Spec: v1alpha1.BundlesSpec{
			VersionsBundles: []v1alpha1.VersionsBundle{
				{
					KubeVersion: "v1.25",
					PackageController: v1alpha1.PackageBundle{
						HelmChart: v1alpha1.Image{},
					},
				},
			},
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}
	objs := []runtime.Object{cluster, managementCluster, oidc, awsIAM, bundles, secret, test.EKSARelease()}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	iam := newMockAWSIamConfigReconciler(t)
	iam.EXPECT().EnsureCASecret(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)
	iam.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	pcc := newMockPackagesClient(t)
	pcc.EXPECT().Reconcile(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	mhc := newMockMachineHealthCheckReconciler(t)
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), iam, validator, pcc, mhc)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))

	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: bundlesRef.Namespace, Name: bundlesRef.Name}, bundles)).To(Succeed())
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
	cl := cb.WithRuntimeObjects(objs...).Build()
	api := envtest.NewAPIExpecter(t, cl)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), newMockClusterValidator(t), nil, newMockMachineHealthCheckReconciler(t))
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
	r := controllers.NewClusterReconciler(client, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), newMockClusterValidator(t), nil, nil)

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
	cl := cb.WithRuntimeObjects(objs...).Build()
	api := envtest.NewAPIExpecter(t, cl)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), newMockClusterValidator(t), nil, nil)
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
				Name:      "my-bundles-ref",
				Namespace: "my-namespace",
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
			KubernetesVersion: "v1.25",
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "my-bundles-ref",
				Namespace: "my-namespace",
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
	bundles := &v1alpha1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-bundles-ref",
			Namespace: cluster.Spec.BundlesRef.Namespace,
		},
		Spec: v1alpha1.BundlesSpec{
			VersionsBundles: []v1alpha1.VersionsBundle{
				{
					KubeVersion: "v1.25",
					PackageController: v1alpha1.PackageBundle{
						HelmChart: v1alpha1.Image{},
					},
				},
			},
		},
	}

	objs := []runtime.Object{cluster, managementCluster, secret, bundles}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

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

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), validator, pcc, mhc)
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
			Name: "my-management-cluster",
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
			Name: "my-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "v1.25",
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	objs := []runtime.Object{cluster, managementCluster, test.EKSARelease()}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	mgmtCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: managementCluster.Name}, mgmtCluster)).To(Succeed())
	pcc := newMockPackagesClient(t)
	pcc.EXPECT().Reconcile(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	mhc := newMockMachineHealthCheckReconciler(t)
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), validator, pcc, mhc)
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
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).
		Return(errors.New("test error"))

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), validator, nil, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(HaveOccurred())

	api := envtest.NewAPIExpecter(t, cl)
	c := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Status.FailureMessage).To(HaveValue(Equal("test error")))
		g.Expect(c.Status.FailureReason).To(HaveValue(Equal(anywherev1.ManagementClusterRefInvalidReason)))
	})
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

func (dummyProviderReconciler) ReconcileWorkerNodes(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
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
