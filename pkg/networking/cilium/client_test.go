package cilium_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/networking/cilium/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type retrierTest struct {
	*WithT
	ctx     context.Context
	r       *cilium.RetrierClient
	c       *mocks.MockClient
	cluster *types.Cluster
}

func newRetrierTest(t *testing.T) *retrierTest {
	ctrl := gomock.NewController(t)
	c := mocks.NewMockClient(ctrl)
	return &retrierTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		r:     cilium.NewRetrier(c),
		c:     c,
		cluster: &types.Cluster{
			KubeconfigFile: "kubeconfig",
		},
	}
}

func TestRetrierClientApplySuccess(t *testing.T) {
	tt := newRetrierTest(t)
	data := []byte("data")
	tt.c.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(errors.New("error in apply")).Times(5)
	tt.c.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(nil).Times(1)

	tt.Expect(tt.r.Apply(tt.ctx, tt.cluster, data)).To(Succeed(), "retrierClient.apply() should succeed after 6 tries")
}

func TestRetrierClientApplyError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.r = cilium.NewRetrier(tt.c, cilium.RetrierClientRetrier(retrier.NewWithMaxRetries(5, 0)))
	data := []byte("data")
	tt.c.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(errors.New("error in apply")).Times(5)
	tt.c.EXPECT().ApplyKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(nil).AnyTimes()

	tt.Expect(tt.r.Apply(tt.ctx, tt.cluster, data)).To(MatchError(ContainSubstring("error in apply")), "retrierClient.apply() should fail after 5 tries")
}

func TestRetrierClientDeleteSuccess(t *testing.T) {
	tt := newRetrierTest(t)
	data := []byte("data")
	tt.c.EXPECT().DeleteKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(errors.New("error in delete")).Times(5)
	tt.c.EXPECT().DeleteKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(nil).Times(1)

	tt.Expect(tt.r.Delete(tt.ctx, tt.cluster, data)).To(Succeed(), "retrierClient.Delete() should succeed after 6 tries")
}

func TestRetrierClientDeleteError(t *testing.T) {
	tt := newRetrierTest(t)
	tt.r = cilium.NewRetrier(tt.c, cilium.RetrierClientRetrier(retrier.NewWithMaxRetries(5, 0)))
	data := []byte("data")
	tt.c.EXPECT().DeleteKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(errors.New("error in delete")).Times(5)
	tt.c.EXPECT().DeleteKubeSpecFromBytes(tt.ctx, tt.cluster, data).Return(nil).AnyTimes()

	tt.Expect(tt.r.Delete(tt.ctx, tt.cluster, data)).To(MatchError(ContainSubstring("error in delete")), "retrierClient.Delete() should fail after 5 tries")
}

type waitForCiliumTest struct {
	*retrierTest
	ciliumDaemonSet, preflightDaemonSet   *v1.DaemonSet
	ciliumDeployment, preflightDeployment *v1.Deployment
}

func newWaitForCiliumTest(t *testing.T) *waitForCiliumTest {
	return &waitForCiliumTest{
		retrierTest: newRetrierTest(t),
		ciliumDaemonSet: &v1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ds",
			},
			Status: v1.DaemonSetStatus{
				DesiredNumberScheduled: 5,
				NumberReady:            5,
			},
		},
		preflightDaemonSet: &v1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ds-pre",
			},
			Status: v1.DaemonSetStatus{
				DesiredNumberScheduled: 5,
				NumberReady:            5,
			},
		},
		ciliumDeployment: &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dep",
			},
			Status: v1.DeploymentStatus{
				Replicas:      5,
				ReadyReplicas: 5,
			},
		},
		preflightDeployment: &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dep-pre",
			},
			Status: v1.DeploymentStatus{
				Replicas:      5,
				ReadyReplicas: 5,
			},
		},
	}
}

func TestRetrierClientWaitForPreflightDaemonSetSuccess(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(nil, errors.New("error in get")).Times(5)
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(tt.ciliumDaemonSet, nil)
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium-pre-flight-check", "kube-system", tt.cluster.KubeconfigFile).Return(tt.preflightDaemonSet, nil)

	tt.Expect(tt.r.WaitForPreflightDaemonSet(tt.ctx, tt.cluster)).To(Succeed(), "retrierClient.waitForPreflightDaemonSet() should succeed after 6 tries")
}

func TestRetrierClientWaitForPreflightDaemonSetError(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.r = cilium.NewRetrier(tt.c, cilium.RetrierClientRetrier(retrier.NewWithMaxRetries(5, 0)))
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(nil, errors.New("error in get")).Times(5)
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(tt.ciliumDaemonSet, nil).AnyTimes()
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium-pre-flight-check", "kube-system", tt.cluster.KubeconfigFile).Return(tt.preflightDaemonSet, nil).AnyTimes()

	tt.Expect(tt.r.WaitForPreflightDaemonSet(tt.ctx, tt.cluster)).To(MatchError(ContainSubstring("error in get")), "retrierClient.waitForPreflightDaemonSet() should fail after 5 tries")
}

func TestRetrierClientRolloutRestartDaemonSetSuccess(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.c.EXPECT().RolloutRestartDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(errors.New("error in rollout")).Times(5)
	tt.c.EXPECT().RolloutRestartDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(nil)

	tt.Expect(tt.r.RolloutRestartCiliumDaemonSet(tt.ctx, tt.cluster)).To(Succeed(), "retrierClient.RolloutRestartDaemonSet() should succeed after 6 tries")
}

func TestRetrierClientRolloutRestartDaemonSetError(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.r = cilium.NewRetrier(tt.c, cilium.RetrierClientRetrier(retrier.NewWithMaxRetries(5, 0)))
	tt.c.EXPECT().RolloutRestartDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(errors.New("error in rollout")).Times(5)
	tt.c.EXPECT().RolloutRestartDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(nil).AnyTimes()

	tt.Expect(tt.r.RolloutRestartCiliumDaemonSet(tt.ctx, tt.cluster)).To(MatchError(ContainSubstring("error in rollout")), "retrierClient.RolloutRestartCiliumDaemonSet() should fail after 5 tries")
}

func TestRetrierClientWaitForPreflightDeploymentSuccess(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.c.EXPECT().GetDeployment(tt.ctx, "cilium-pre-flight-check", "kube-system", tt.cluster.KubeconfigFile).Return(nil, errors.New("error in get")).Times(5)
	tt.c.EXPECT().GetDeployment(tt.ctx, "cilium-pre-flight-check", "kube-system", tt.cluster.KubeconfigFile).Return(tt.preflightDeployment, nil)

	tt.Expect(tt.r.WaitForPreflightDeployment(tt.ctx, tt.cluster)).To(Succeed(), "retrierClient.waitForPreflightDeployment() should succeed after 6 tries")
}

func TestRetrierClientWaitForPreflightDeploymentError(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.r = cilium.NewRetrier(tt.c, cilium.RetrierClientRetrier(retrier.NewWithMaxRetries(5, 0)))
	tt.c.EXPECT().GetDeployment(tt.ctx, "cilium-pre-flight-check", "kube-system", tt.cluster.KubeconfigFile).Return(nil, errors.New("error in get")).Times(5)
	tt.c.EXPECT().GetDeployment(tt.ctx, "cilium-pre-flight-check", "kube-system", tt.cluster.KubeconfigFile).Return(tt.preflightDeployment, nil).AnyTimes()

	tt.Expect(tt.r.WaitForPreflightDeployment(tt.ctx, tt.cluster)).To(MatchError(ContainSubstring("error in get")), "retrierClient.waitForPreflightDeployment() should fail after 5 tries")
}

func TestRetrierClientWaitForCiliumDaemonSetSuccess(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(nil, errors.New("error in get")).Times(5)
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(tt.ciliumDaemonSet, nil)

	tt.Expect(tt.r.WaitForCiliumDaemonSet(tt.ctx, tt.cluster)).To(Succeed(), "retrierClient.waitForCiliumDaemonSet() should succeed after 6 tries")
}

func TestRetrierClientWaitForCiliumDaemonSetError(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.r = cilium.NewRetrier(tt.c, cilium.RetrierClientRetrier(retrier.NewWithMaxRetries(5, 0)))
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(nil, errors.New("error in get")).Times(5)
	tt.c.EXPECT().GetDaemonSet(tt.ctx, "cilium", "kube-system", tt.cluster.KubeconfigFile).Return(tt.ciliumDaemonSet, nil).AnyTimes()

	tt.Expect(tt.r.WaitForCiliumDaemonSet(tt.ctx, tt.cluster)).To(MatchError(ContainSubstring("error in get")), "retrierClient.waitForCiliumDaemonSet() should fail after 5 tries")
}

func TestRetrierClientWaitForCiliumDeploymentSuccess(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.c.EXPECT().GetDeployment(tt.ctx, "cilium-operator", "kube-system", tt.cluster.KubeconfigFile).Return(nil, errors.New("error in get")).Times(5)
	tt.c.EXPECT().GetDeployment(tt.ctx, "cilium-operator", "kube-system", tt.cluster.KubeconfigFile).Return(tt.ciliumDeployment, nil)

	tt.Expect(tt.r.WaitForCiliumDeployment(tt.ctx, tt.cluster)).To(Succeed(), "retrierClient.waitForCiliumDeployment() should succeed after 6 tries")
}

func TestRetrierClientWaitForCiliumDeploymentError(t *testing.T) {
	tt := newWaitForCiliumTest(t)
	tt.r = cilium.NewRetrier(tt.c, cilium.RetrierClientRetrier(retrier.NewWithMaxRetries(5, 0)))
	tt.c.EXPECT().GetDeployment(tt.ctx, "cilium-operator", "kube-system", tt.cluster.KubeconfigFile).Return(nil, errors.New("error in get")).Times(5)
	tt.c.EXPECT().GetDeployment(tt.ctx, "cilium-operator", "kube-system", tt.cluster.KubeconfigFile).Return(tt.ciliumDeployment, nil).AnyTimes()

	tt.Expect(tt.r.WaitForCiliumDeployment(tt.ctx, tt.cluster)).To(MatchError(ContainSubstring("error in get")), "retrierClient.waitForCiliumDeployment() should fail after 5 tries")
}
