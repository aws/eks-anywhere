package bootstrapper_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/bootstrapper/mocks"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type retrierTest struct {
	*WithT
	ctx         context.Context
	r           bootstrapper.RetrierClient
	kind        *mocks.MockKindClient
	k8s         *mocks.MockKubernetesClient
	cluster     *types.Cluster
	clusterSpec *cluster.Spec
}

func newRetrierTest(t *testing.T) *retrierTest {
	ctrl := gomock.NewController(t)
	kind := mocks.NewMockKindClient(ctrl)
	k8s := mocks.NewMockKubernetesClient(ctrl)

	return &retrierTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		r:     bootstrapper.NewRetrierClient(kind, k8s, bootstrapper.WithRetrierClientRetrier(*retrier.NewWithMaxRetries(5, 0))),
		kind:  kind,
		k8s:   k8s,
		cluster: &types.Cluster{
			KubeconfigFile: "kubeconfig",
		},
		clusterSpec: test.NewClusterSpec(),
	}
}

func TestRetrierClientApplySuccess(t *testing.T) {
	tt := newRetrierTest(t)
	data := []byte("data")
	tt.k8s.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(errors.New("error in apply")).Times(4)
	tt.k8s.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(nil).Times(1)

	tt.Expect(tt.r.Apply(tt.ctx, tt.cluster, data)).To(Succeed(), "retrierClient.apply() should succeed after 5 tries")
}

func TestRetrierClientApplyError(t *testing.T) {
	tt := newRetrierTest(t)
	data := []byte("data")
	tt.k8s.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(errors.New("error in apply")).Times(5)
	tt.k8s.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(nil).AnyTimes()

	tt.Expect(tt.r.Apply(tt.ctx, tt.cluster, data)).To(MatchError(ContainSubstring("error in apply")), "retrierClient.apply() should fail after 5 tries")
}

func TestRetrierClientCreateNamespaceSuccess(t *testing.T) {
	tt := newRetrierTest(t)
	tt.k8s.EXPECT().CreateNamespaceIfNotPresent(tt.ctx, "kubeconfig", "test-namespace").Return(errors.New("error in CreateNamespaceIfNotPresent")).Times(4)
	tt.k8s.EXPECT().CreateNamespaceIfNotPresent(tt.ctx, "kubeconfig", "test-namespace").Return(nil).Times(1)
	tt.Expect(tt.r.CreateNamespace(tt.ctx, "kubeconfig", "test-namespace")).To(Succeed(), "retrierClient.CreateNamespace() should succeed after 5 tries")
}

func TestRetrierClientCreateNamespaceError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.k8s.EXPECT().CreateNamespaceIfNotPresent(tt.ctx, "kubeconfig", "test-namespace").Return(errors.New("error in CreateNamespaceIfNotPresent")).Times(5)
	tt.k8s.EXPECT().CreateNamespaceIfNotPresent(tt.ctx, "kubeconfig", "test-namespace").Return(nil).AnyTimes()
	tt.Expect(tt.r.CreateNamespace(tt.ctx, "kubeconfig", "test-namespace")).To(MatchError(ContainSubstring("error in CreateNamespace")), "retrierClient.CreateNamespace() should fail after 5 tries")
}

func TestRetrierClientGetCAPIClusterCRDSuccess(t *testing.T) {
	tt := newRetrierTest(t)
	tt.k8s.EXPECT().ValidateClustersCRD(tt.ctx, tt.cluster).Return(errors.New("error in ValidateClustersCRD")).Times(4)
	tt.k8s.EXPECT().ValidateClustersCRD(tt.ctx, tt.cluster).Return(nil).Times(1)
	tt.Expect(tt.r.GetCAPIClusterCRD(tt.ctx, tt.cluster)).To(Succeed(), "retrierClient.GetCAPIClusterCRD() should succeed after 5 tries")
}

func TestRetrierClientGetCAPIClusterCRDError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.k8s.EXPECT().ValidateClustersCRD(tt.ctx, tt.cluster).Return(errors.New("error in ValidateClustersCRD")).Times(5)
	tt.k8s.EXPECT().ValidateClustersCRD(tt.ctx, tt.cluster).Return(nil).AnyTimes()
	tt.Expect(tt.r.GetCAPIClusterCRD(tt.ctx, tt.cluster)).To(MatchError(ContainSubstring("error in ValidateClustersCRD")), "retrierClient.GetCAPIClusterCRD() should fail after 5 tries")
}

func TestRetrierClientGetCAPIClustersSuccess(t *testing.T) {
	tt := newRetrierTest(t)
	tt.k8s.EXPECT().GetClusters(tt.ctx, tt.cluster).Return(nil, errors.New("error in GetClusters")).Times(4)
	tt.k8s.EXPECT().GetClusters(tt.ctx, tt.cluster).Return(nil, nil).Times(1)
	_, err := tt.r.GetCAPIClusters(tt.ctx, tt.cluster)
	tt.Expect(err).To(Succeed(), "retrierClient.GetCAPIClusters() should succeed after 5 tries")
}

func TestRetrierClientGetCAPIClustersError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.k8s.EXPECT().GetClusters(tt.ctx, tt.cluster).Return(nil, errors.New("error in GetClusters")).Times(5)
	tt.k8s.EXPECT().GetClusters(tt.ctx, tt.cluster).Return(nil, nil).AnyTimes()
	_, err := tt.r.GetCAPIClusters(tt.ctx, tt.cluster)
	tt.Expect(err).To(MatchError(ContainSubstring("error in GetClusters")), "retrierClient.GetCAPIClusters() should fail after 5 tries")
}

func TestRetrierClientKindClusterExistsSuccess(t *testing.T) {
	tt := newRetrierTest(t)
	tt.kind.EXPECT().ClusterExists(tt.ctx, "test-cluster").Return(false, errors.New("error in ClusterExists")).Times(4)
	tt.kind.EXPECT().ClusterExists(tt.ctx, "test-cluster").Return(true, nil).Times(1)
	exists, err := tt.r.KindClusterExists(tt.ctx, "test-cluster")
	tt.Expect(exists).To(Equal(true))
	tt.Expect(err).To(Succeed(), "retrierClient.KindClusterExists() should succeed after 5 tries")
}

func TestRetrierClientKindClusterExistsError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.kind.EXPECT().ClusterExists(tt.ctx, "test-cluster").Return(false, errors.New("error in ClusterExists")).Times(5)
	tt.kind.EXPECT().ClusterExists(tt.ctx, "test-cluster").Return(true, nil).AnyTimes()
	_, err := tt.r.KindClusterExists(tt.ctx, "test-cluster")
	tt.Expect(err).To(MatchError(ContainSubstring("error in ClusterExists")), "retrierClient.KindClusterExists() should fail after 5 tries")
}

func TestRetrierClientGetKindClusterKubeconfigSuccess(t *testing.T) {
	tt := newRetrierTest(t)
	tt.kind.EXPECT().GetKubeconfig(tt.ctx, "test-cluster").Return("", errors.New("error in GetKubeconfig")).Times(4)
	tt.kind.EXPECT().GetKubeconfig(tt.ctx, "test-cluster").Return("kubeconfig", nil).Times(1)
	kubeconfig, err := tt.r.GetKindClusterKubeconfig(tt.ctx, "test-cluster")
	tt.Expect(kubeconfig).To(Equal("kubeconfig"))
	tt.Expect(err).To(Succeed(), "retrierClient.GetKindClusterKubeconfig() should succeed after 5 tries")
}

func TestRetrierClientGetKindClusterKubeconfigError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.kind.EXPECT().GetKubeconfig(tt.ctx, "test-cluster").Return("", errors.New("error in GetKubeconfig")).Times(5)
	tt.kind.EXPECT().GetKubeconfig(tt.ctx, "test-cluster").Return("kubeconfig", nil).AnyTimes()
	_, err := tt.r.GetKindClusterKubeconfig(tt.ctx, "test-cluster")
	tt.Expect(err).To(MatchError(ContainSubstring("error in GetKubeconfig")), "retrierClient.GetKindClusterKubeconfig() should fail after 5 tries")
}

func TestRetrierClientDeleteKindClusterSuccess(t *testing.T) {
	tt := newRetrierTest(t)
	tt.kind.EXPECT().DeleteBootstrapCluster(tt.ctx, tt.cluster).Return(errors.New("error in DeleteBootstrapCluster")).Times(4)
	tt.kind.EXPECT().DeleteBootstrapCluster(tt.ctx, tt.cluster).Return(nil).Times(1)
	tt.Expect(tt.r.DeleteKindCluster(tt.ctx, tt.cluster)).To(Succeed(), "retrierClient.DeleteKindCluster() should succeed after 5 tries")
}

func TestRetrierClientDeleteKindClusterError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.kind.EXPECT().DeleteBootstrapCluster(tt.ctx, tt.cluster).Return(errors.New("error in DeleteBootstrapCluster")).Times(5)
	tt.kind.EXPECT().DeleteBootstrapCluster(tt.ctx, tt.cluster).Return(nil).AnyTimes()
	tt.Expect(tt.r.DeleteKindCluster(tt.ctx, tt.cluster)).To(MatchError(ContainSubstring("error in DeleteBootstrapCluster")), "retrierClient.DeleteKindCluster() should fail after 5 tries")
}
