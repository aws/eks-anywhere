package controllers_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers/controllers"
	"github.com/aws/eks-anywhere/controllers/controllers/clusters"
	"github.com/aws/eks-anywhere/controllers/controllers/reconciler"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

func TestClusterReconcilerEnsureOwnerReferences(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

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
	objs := []runtime.Object{cluster, oidc, awsIAM}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	r := controllers.NewClusterReconciler(cl, logr.New(logf.NullLogSink{}), cl.Scheme(), nil, nil, newDummyProviderReconcilerBuilder())
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

func TestClusterReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewClusterReconciler(client, logf.Log, client.Scheme(), nil, nil, newDummyProviderReconcilerBuilder())

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func newDummyProviderReconcilerBuilder() controllers.BuildProviderReconciler {
	return func(datacenterKind string, client client.Client, log logr.Logger, validator *vsphere.Validator, defaulter *vsphere.Defaulter, tracker *remote.ClusterCacheTracker) (clusters.ProviderClusterReconciler, error) {
		return dummyProviderReconciler{}, nil
	}
}

type dummyProviderReconciler struct{}

func (dummyProviderReconciler) Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (reconciler.Result, error) {
	return reconciler.Result{}, nil
}

func clusterRequest(cluster *anywherev1.Cluster) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}
}
