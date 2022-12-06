package controllers_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestClusterReconcilerEnsureOwnerReferences(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "my-namespace",
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
	objs := []runtime.Object{cluster, managementCluster, oidc, awsIAM}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	iam := newMockAWSIamConfigReconciler(t)
	iam.EXPECT().EnsureCASecret(ctx, gomock.AssignableToTypeOf(logr.Logger{}), cl, gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)
	iam.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), cl, gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), iam)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
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

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "my-namespace",
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

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t))
	g.Expect(r.Reconcile(ctx, clusterRequest(cluster))).Error().To(MatchError(ContainSubstring("not found")))
	c := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Status.FailureMessage).To(HaveValue(Equal(
			"Dependent cluster objects don't exist: oidcconfigs.anywhere.eks.amazonaws.com \"my-oidc\" not found",
		)))
	})
}

func TestClusterReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewClusterReconciler(client, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t))

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
	cluster.SetManagedBy("my-management-cluster")

	objs := []runtime.Object{cluster, managementCluster}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()
	api := envtest.NewAPIExpecter(t, cl)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t))
	g.Expect(r.Reconcile(ctx, clusterRequest(cluster))).Error().To(MatchError(ContainSubstring("\"my-management-cluster\" not found")))
	c := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Status.FailureMessage).To(HaveValue(Equal("Management cluster my-management-cluster does not exist")))
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
				Name: "my-bundles-ref",
			},
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	objs := []runtime.Object{cluster, managementCluster}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	mgmtCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: managementCluster.Name}, mgmtCluster)).To(Succeed())

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t))
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).ToNot(HaveOccurred())

	newCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: "my-cluster"}, newCluster)).To(Succeed())
	g.Expect(newCluster.Spec.BundlesRef).To(Equal(mgmtCluster.Spec.BundlesRef))
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

func (dummyProviderReconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
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
