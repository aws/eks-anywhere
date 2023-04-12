package awsiamauth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/awsiamauth"
	"github.com/aws/eks-anywhere/pkg/awsiamauth/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type retrierTest struct {
	*WithT
	ctx     context.Context
	r       awsiamauth.RetrierClient
	c       *mocks.MockClient
	cluster *types.Cluster
}

func newRetrierTest(t *testing.T) *retrierTest {
	ctrl := gomock.NewController(t)
	c := mocks.NewMockClient(ctrl)
	return &retrierTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		r:     awsiamauth.NewRetrierClient(c, awsiamauth.RetrierClientRetrier(*retrier.NewWithMaxRetries(5, 0))),
		c:     c,
		cluster: &types.Cluster{
			KubeconfigFile: "kubeconfig",
		},
	}
}

func TestRetrierClientApplySuccess(t *testing.T) {
	tt := newRetrierTest(t)
	data := []byte("data")
	tt.c.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(errors.New("error in apply")).Times(4)
	tt.c.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(nil).Times(1)

	tt.Expect(tt.r.Apply(tt.ctx, tt.cluster, data)).To(Succeed(), "retrierClient.apply() should succeed after 5 tries")
}

func TestRetrierClientApplyError(t *testing.T) {
	tt := newRetrierTest(t)
	data := []byte("data")
	tt.c.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(errors.New("error in apply")).Times(5)
	tt.c.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(nil).AnyTimes()

	tt.Expect(tt.r.Apply(tt.ctx, tt.cluster, data)).To(MatchError(ContainSubstring("error in apply")), "retrierClient.apply() should fail after 5 tries")
}

func TestRetrierClientGetAPIServerURLSuccess(t *testing.T) {
	tt := newRetrierTest(t)
	tt.c.EXPECT().GetApiServerUrl(tt.ctx, tt.cluster).Return("", errors.New("error in GetApiServerUrl")).Times(4)
	tt.c.EXPECT().GetApiServerUrl(tt.ctx, tt.cluster).Return("apiserverurl", nil).Times(1)

	url, err := tt.r.GetAPIServerURL(tt.ctx, tt.cluster)
	tt.Expect(url).To(Equal("apiserverurl"))
	tt.Expect(err).To(Succeed(), "retrierClient.GetApiServerUrl() should succeed after 5 tries")
}

func TestRetrierClientGetAPIServerURLError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.c.EXPECT().GetApiServerUrl(tt.ctx, tt.cluster).Return("", errors.New("error in GetApiServerUrl")).Times(5)
	tt.c.EXPECT().GetApiServerUrl(tt.ctx, tt.cluster).Return("apiserverurl", nil).AnyTimes()

	url, err := tt.r.GetAPIServerURL(tt.ctx, tt.cluster)
	tt.Expect(url).To(Equal(""))
	tt.Expect(err).To(MatchError(ContainSubstring("error in GetApiServerUrl")), "retrierClient.GetApiServerUrl() should fail after 5 tries")
}

func TestRetrierClientGetClusterCACertSuccess(t *testing.T) {
	tt := newRetrierTest(t)
	tt.c.EXPECT().GetObject(tt.ctx, "secret", "test-cluster-ca", "eksa-system", tt.cluster.KubeconfigFile, &corev1.Secret{}).Return(errors.New("error in GetObject")).Times(4)
	tt.c.EXPECT().
		GetObject(tt.ctx, "secret", "test-cluster-ca", "eksa-system", tt.cluster.KubeconfigFile, &corev1.Secret{}).
		DoAndReturn(func(_ context.Context, _, _, _, _ string, obj *corev1.Secret) error {
			obj.Data = map[string][]byte{
				"tls.crt": []byte("cert"),
			}
			return nil
		}).Times(1)

	cert, err := tt.r.GetClusterCACert(tt.ctx, tt.cluster, "test-cluster")
	tt.Expect(cert).To(Equal([]byte("Y2VydA==")))
	tt.Expect(err).To(Succeed(), "retrierClient.GetObject() should succeed after 5 tries")
}

func TestRetrierClientGetClusterCACertError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.c.EXPECT().GetObject(tt.ctx, "secret", "test-cluster-ca", "eksa-system", tt.cluster.KubeconfigFile, &corev1.Secret{}).Return(errors.New("error in GetObject")).Times(5)
	tt.c.EXPECT().GetObject(tt.ctx, "secret", "test-cluster-ca", "eksa-system", tt.cluster.KubeconfigFile, &corev1.Secret{}).Return(nil).AnyTimes()

	cert, err := tt.r.GetClusterCACert(tt.ctx, tt.cluster, "test-cluster")
	tt.Expect(cert).To(BeNil())
	tt.Expect(err).To(MatchError(ContainSubstring("error in GetObject")), "retrierClient.GetObject() should fail after 5 tries")
}

func TestRetrierClientGetClusterCACertNotFound(t *testing.T) {
	tt := newRetrierTest(t)
	tt.c.EXPECT().GetObject(tt.ctx, "secret", "test-cluster-ca", "eksa-system", tt.cluster.KubeconfigFile, &corev1.Secret{}).Return(errors.New("error in GetObject")).Times(4)
	tt.c.EXPECT().
		GetObject(tt.ctx, "secret", "test-cluster-ca", "eksa-system", tt.cluster.KubeconfigFile, &corev1.Secret{}).
		DoAndReturn(func(_ context.Context, _, _, _, _ string, obj *corev1.Secret) error {
			obj.Data = map[string][]byte{
				"tls.crt.invalid": []byte("cert"),
			}
			return nil
		}).Times(1)

	cert, err := tt.r.GetClusterCACert(tt.ctx, tt.cluster, "test-cluster")
	tt.Expect(cert).To(BeNil())
	tt.Expect(err).To(MatchError(ContainSubstring("tls.crt not found in secret [test-cluster-ca]")))
}
