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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/awsiamauth/reconciler"
	reconcilermocks "github.com/aws/eks-anywhere/pkg/awsiamauth/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	cryptomocks "github.com/aws/eks-anywhere/pkg/crypto/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestEnsureCASecretSecretFound(t *testing.T) {
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

func TestEnsureCASecretSecretNotFound(t *testing.T) {
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
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcileDeleteSuccess(t *testing.T) {
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

	kubeSec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsiamauth.KubeconfigSecretName(cluster.Name),
			Namespace: constants.EksaSystemNamespace,
		},
	}
	objs := []runtime.Object{sec, kubeSec}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()
	r := newReconciler(t, cl)

	err := r.ReconcileDelete(ctx, nullLog(), cluster)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestReconcileDeleteNoSecretError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}
	cb := fake.NewClientBuilder()
	cl := cb.Build()
	r := newReconciler(t, cl)

	err := r.ReconcileDelete(ctx, nullLog(), cluster)
	g.Expect(err).ToNot(HaveOccurred())
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

func TestReconcileBuildClusterSpecError(t *testing.T) {
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

func TestReconcileKCPObjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	generateUUID := uuid.New
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)

	bundle := test.Bundle()
	eksdRelease := test.EksdRelease("1-22")
	eksaRelease := test.EKSARelease()
	objs := []runtime.Object{bundle, eksaRelease, eksdRelease}
	cb := fake.NewClientBuilder()
	scheme := runtime.NewScheme()
	_ = releasev1.AddToScheme(scheme)
	_ = eksdv1.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	_ = controlplanev1.AddToScheme(scheme)
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()
	version := test.DevEksaVersion()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.22",
			BundlesRef: &anywherev1.BundlesRef{
				Name:       bundle.Name,
				Namespace:  bundle.Namespace,
				APIVersion: bundle.APIVersion,
			},
			EksaVersion: &version,
		},
	}

	r := reconciler.New(certs, generateUUID, cl, remoteClientRegistry)
	result, err := r.Reconcile(ctx, nullLog(), cluster)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(controller.ResultWithRequeue(5 * time.Second)))
}

func TestReconcileRemoteGetClientError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	generateUUID := uuid.New
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)

	bundle := test.Bundle()
	eksdRelease := test.EksdRelease("1-22")
	eksaRelease := test.EKSARelease()
	version := test.DevEksaVersion()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.22",
			BundlesRef: &anywherev1.BundlesRef{
				Name:       bundle.Name,
				Namespace:  bundle.Namespace,
				APIVersion: bundle.APIVersion,
			},
			EksaVersion: &version,
		},
	}
	kcpVersion := "test"
	kcp := test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
		kcp.Name = cluster.Name
		kcp.Spec.Version = kcpVersion
		kcp.Status = controlplanev1.KubeadmControlPlaneStatus{
			Conditions: clusterv1.Conditions{
				{
					Type:               clusterapi.ReadyCondition,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now()),
				},
			},
			Version: pointer.String(kcpVersion),
		}
	})
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsiamauth.CASecretName(cluster.Name),
			Namespace: constants.EksaSystemNamespace,
		},
	}
	objs := []runtime.Object{bundle, eksdRelease, kcp, sec, eksaRelease}
	cb := fake.NewClientBuilder()
	scheme := runtime.NewScheme()
	_ = releasev1.AddToScheme(scheme)
	_ = eksdv1.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = controlplanev1.AddToScheme(scheme)
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	remoteClientRegistry.EXPECT().GetClient(context.Background(), gomock.AssignableToTypeOf(client.ObjectKey{})).Return(nil, errors.New("client error"))

	r := reconciler.New(certs, generateUUID, cl, remoteClientRegistry)
	result, err := r.Reconcile(ctx, nullLog(), cluster)
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcileConfigMapNotFoundApplyError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	generateUUID := uuid.New
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)

	bundle := test.Bundle()
	eksaRelease := test.EKSARelease()
	eksdRelease := test.EksdRelease("1-22")
	version := test.DevEksaVersion()
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.22",
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Endpoint: &anywherev1.Endpoint{
					Host: "1.2.3.4",
				},
			},
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
			EksaVersion: &version,
		},
	}
	kcpVersion := "test"
	kcp := test.KubeadmControlPlane(func(kcp *controlplanev1.KubeadmControlPlane) {
		kcp.Name = cluster.Name
		kcp.Spec.Version = kcpVersion
		kcp.Status = controlplanev1.KubeadmControlPlaneStatus{
			Conditions: clusterv1.Conditions{
				{
					Type:               clusterapi.ReadyCondition,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now()),
				},
			},
			Version: pointer.String(kcpVersion),
		}
	})
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsiamauth.CASecretName(cluster.Name),
			Namespace: constants.EksaSystemNamespace,
		},
	}
	caSec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterapi.ClusterCASecretName(cluster.Name),
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string][]byte{
			"tls.crt": []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FUR"),
		},
	}
	awsiamconfig := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aws-config",
			Namespace: "eksa-system",
		},
	}
	objs := []runtime.Object{bundle, eksdRelease, kcp, sec, awsiamconfig, caSec, eksaRelease}
	cb := fake.NewClientBuilder()
	scheme := runtime.NewScheme()
	_ = anywherev1.AddToScheme(scheme)
	_ = releasev1.AddToScheme(scheme)
	_ = eksdv1.AddToScheme(scheme)
	_ = clusterv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = controlplanev1.AddToScheme(scheme)
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	rCb := fake.NewClientBuilder()
	rCl := rCb.Build()
	remoteClientRegistry.EXPECT().GetClient(context.Background(), gomock.AssignableToTypeOf(client.ObjectKey{})).Return(rCl, nil)

	r := reconciler.New(certs, generateUUID, cl, remoteClientRegistry)
	result, err := r.Reconcile(ctx, nullLog(), cluster)
	g.Expect(err).To(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}
