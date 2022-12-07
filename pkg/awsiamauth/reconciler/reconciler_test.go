package reconciler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/awsiamauth/reconciler"
	reconcilermocks "github.com/aws/eks-anywhere/pkg/awsiamauth/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	cryptomocks "github.com/aws/eks-anywhere/pkg/crypto/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestEnsureCASecret_SecretFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsiamauth.CASecretName(cluster.Name),
			Namespace: constants.EksaSystemNamespace,
		},
	}

	objs := []runtime.Object{sec}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()
	r := newReconciler(t, cl)

	result, err := r.EnsureCASecret(ctx, nullLog(), cluster)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestEnsureCASecret_SecretNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cb := fake.NewClientBuilder()
	scheme := runtime.NewScheme()
	_ = anywherev1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	cl := cb.WithScheme(scheme).Build()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}
	r := newReconciler(t, cl)

	result, err := r.EnsureCASecret(ctx, nullLog(), cluster)
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
	g.Expect(result).To(Equal(controller.Result{}))
}

func newReconciler(t *testing.T, client client.WithWatch) *reconciler.Reconciler {
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	generateUUID := uuid.New
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)

	certs.EXPECT().GenerateIamAuthSelfSignCertKeyPair().Return([]byte("ca-cert"), []byte("ca-key"), nil).MinTimes(0).MaxTimes(1)

	return reconciler.New(certs, generateUUID, client, remoteClientRegistry)
}

func nullLog() logr.Logger {
	return logr.New(logf.NullLogSink{})
}

func TestReconcile_BuildSpecError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	generateUUID := uuid.New
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)
	cb := fake.NewClientBuilder()
	cl := cb.Build()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}

	r := reconciler.New(certs, generateUUID, cl, remoteClientRegistry)
	result, err := r.Reconcile(ctx, nullLog(), cluster)
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcile_CPReadyRequeue(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	generateUUID := uuid.New
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)

	bundle := test.Bundle()
	eksdRelease := test.EksdRelease()
	objs := []runtime.Object{bundle, eksdRelease}
	cb := fake.NewClientBuilder()
	scheme := runtime.NewScheme()
	_ = releasev1.AddToScheme(scheme)
	_ = eksdv1.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.20",
			BundlesRef: &anywherev1.BundlesRef{
				Name:       bundle.Name,
				Namespace:  bundle.Namespace,
				APIVersion: bundle.APIVersion,
			},
		},
	}

	r := reconciler.New(certs, generateUUID, cl, remoteClientRegistry)
	result, err := r.Reconcile(ctx, nullLog(), cluster)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(controller.ResultWithRequeue(5 * time.Second)))
}

func TestReconcile_ensureCASecretOwnerRef_NoSecret(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	generateUUID := uuid.New
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)

	bundle := test.Bundle()
	eksdRelease := test.EksdRelease()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.20",
			BundlesRef: &anywherev1.BundlesRef{
				Name:       bundle.Name,
				Namespace:  bundle.Namespace,
				APIVersion: bundle.APIVersion,
			},
		},
	}
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = cluster.Name
	})

	objs := []runtime.Object{bundle, eksdRelease, capiCluster}
	cb := fake.NewClientBuilder()
	scheme := runtime.NewScheme()
	_ = releasev1.AddToScheme(scheme)
	_ = eksdv1.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	r := reconciler.New(certs, generateUUID, cl, remoteClientRegistry)
	result, err := r.Reconcile(ctx, nullLog(), cluster)
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcile_RemoteGetClient_Error(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	generateUUID := uuid.New
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)

	bundle := test.Bundle()
	eksdRelease := test.EksdRelease()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.20",
			BundlesRef: &anywherev1.BundlesRef{
				Name:       bundle.Name,
				Namespace:  bundle.Namespace,
				APIVersion: bundle.APIVersion,
			},
		},
	}
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = cluster.Name
	})
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsiamauth.CASecretName(cluster.Name),
			Namespace: constants.EksaSystemNamespace,
		},
	}
	objs := []runtime.Object{bundle, eksdRelease, capiCluster, sec}
	cb := fake.NewClientBuilder()
	scheme := runtime.NewScheme()
	_ = releasev1.AddToScheme(scheme)
	_ = eksdv1.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	remoteClientRegistry.EXPECT().GetClient(context.Background(), gomock.AssignableToTypeOf(client.ObjectKey{})).Return(nil, errors.New("client error"))

	r := reconciler.New(certs, generateUUID, cl, remoteClientRegistry)
	result, err := r.Reconcile(ctx, nullLog(), cluster)
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcile_ConfigMap_NotFound_ApplyError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	generateUUID := uuid.New
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)

	bundle := test.Bundle()
	eksdRelease := test.EksdRelease()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.20",
			BundlesRef: &anywherev1.BundlesRef{
				Name:       bundle.Name,
				Namespace:  bundle.Namespace,
				APIVersion: bundle.APIVersion,
			},
			IdentityProviderRefs: []anywherev1.Ref{
				{
					Name: "aws-config",
					Kind: "AWSIamConfig",
				},
			},
		},
	}
	capiCluster := test.CAPICluster(func(c *clusterv1.Cluster) {
		c.Name = cluster.Name
	})
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsiamauth.CASecretName(cluster.Name),
			Namespace: constants.EksaSystemNamespace,
		},
	}
	awsiamconfig := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aws-config",
			Namespace: "eksa-system",
		},
	}
	objs := []runtime.Object{bundle, eksdRelease, capiCluster, sec, awsiamconfig}
	cb := fake.NewClientBuilder()
	scheme := runtime.NewScheme()
	_ = anywherev1.AddToScheme(scheme)
	_ = releasev1.AddToScheme(scheme)
	_ = eksdv1.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	rCb := fake.NewClientBuilder()
	rCl := rCb.Build()
	remoteClientRegistry.EXPECT().GetClient(context.Background(), gomock.AssignableToTypeOf(client.ObjectKey{})).Return(rCl, nil)

	r := reconciler.New(certs, generateUUID, cl, remoteClientRegistry)
	result, err := r.Reconcile(ctx, nullLog(), cluster)
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}
