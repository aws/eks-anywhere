package reconciler_test

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/awsiamauth/reconciler"
	reconcilermocks "github.com/aws/eks-anywhere/pkg/awsiamauth/reconciler/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	cryptomocks "github.com/aws/eks-anywhere/pkg/crypto/mocks"
)

func TestReconcileAWSIAMAuthCASecret_SecretFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	r := newReconciler(t)
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      awsiamauth.GetAwsIamAuthCaSecretName("test-cluster"),
			Namespace: constants.EksaSystemNamespace,
		},
	}

	objs := []runtime.Object{sec}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	result, err := r.ReconcileAWSIAMAuthCASecret(ctx, nullLog(), cl, "test-cluster")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(controller.Result{}))
}

func TestReconcileAWSIAMAuthCASecret_SecretNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	r := newReconciler(t)
	cb := fake.NewClientBuilder()
	cl := cb.Build()

	result, err := r.ReconcileAWSIAMAuthCASecret(ctx, nullLog(), cl, "test-cluster")
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
	g.Expect(result).To(Equal(controller.Result{}))
}

func newReconciler(t *testing.T) *reconciler.Reconciler {
	ctrl := gomock.NewController(t)
	certs := cryptomocks.NewMockCertificateGenerator(ctrl)
	clusterID := uuid.MustParse("36db102f-9e1e-4ca4-8300-271d30b14161")
	remoteClientRegistry := reconcilermocks.NewMockRemoteClientRegistry(ctrl)

	certs.EXPECT().GenerateIamAuthSelfSignCertKeyPair().Return([]byte("ca-cert"), []byte("ca-key"), nil).MinTimes(0).MaxTimes(1)

	return reconciler.New(certs, clusterID, remoteClientRegistry)
}

func nullLog() logr.Logger {
	return logr.New(logf.NullLogSink{})
}
