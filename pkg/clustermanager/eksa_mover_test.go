package clustermanager_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/types"
)

type moverTest struct {
	gomega.Gomega
	tb            testing.TB
	clientFactory *mocks.MockClientFactory
	ctx           context.Context
	spec          *cluster.Spec
	fromClient    kubernetes.Client
	toClient      kubernetes.Client
	log           logr.Logger
	mgmtCluster   *types.Cluster
	bootstrap     *types.Cluster
}

func newMoverTest(tb testing.TB) *moverTest {
	ctrl := gomock.NewController(tb)
	return &moverTest{
		tb:            tb,
		Gomega:        gomega.NewWithT(tb),
		clientFactory: mocks.NewMockClientFactory(ctrl),
		ctx:           context.Background(),
		spec:          test.VSphereClusterSpec(tb, tb.Name()),
		log:           test.NewNullLogger(),
		bootstrap: &types.Cluster{
			KubeconfigFile: "bootstrap-config",
		},
		mgmtCluster: &types.Cluster{
			KubeconfigFile: "my-config",
		},
	}
}

func (a *moverTest) buildClients(fromObjs, toObjs []kubernetes.Object) {
	a.fromClient = test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(fromObjs)...)
	a.toClient = test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(toObjs)...)
}

func TestMoverSuccess(t *testing.T) {
	tt := newMoverTest(t)
	objs := tt.spec.ClusterAndChildren()
	tt.buildClients(objs, nil)
	m := clustermanager.NewMover(tt.log, tt.clientFactory,
		clustermanager.WithMoverRetryBackOff(time.Millisecond),
		clustermanager.WithMoverNoTimeouts(),
	)

	tt.Expect(m.Move(tt.ctx, tt.spec, tt.fromClient, tt.toClient)).To(gomega.Succeed())

	for _, obj := range objs {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
		tt.Expect(tt.toClient.Get(tt.ctx, obj.GetName(), obj.GetNamespace(), u)).To(gomega.Succeed())
		original, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		tt.Expect(err).To(gomega.Succeed())
		tt.Expect(u.Object["spec"]).To(gomega.BeComparableTo(original["spec"]))
	}
}

func TestMoverFailReadCluster(t *testing.T) {
	tt := newMoverTest(t)
	tt.buildClients(nil, nil)
	m := clustermanager.NewMover(tt.log, tt.clientFactory,
		clustermanager.WithMoverRetryBackOff(time.Millisecond),
		clustermanager.WithMoverApplyClusterTimeout(time.Millisecond),
	)
	err := m.Move(tt.ctx, tt.spec, tt.fromClient, tt.toClient)

	tt.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("reading cluster from source")))
}

func TestMoverFailGetChildren(t *testing.T) {
	tt := newMoverTest(t)
	objs := []kubernetes.Object{tt.spec.Cluster}
	tt.buildClients(objs, nil)
	m := clustermanager.NewMover(tt.log, tt.clientFactory,
		clustermanager.WithMoverRetryBackOff(time.Millisecond),
		clustermanager.WithMoverApplyClusterTimeout(time.Millisecond),
	)

	err := m.Move(tt.ctx, tt.spec, tt.fromClient, tt.toClient)
	tt.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("reading child object")))
}

func TestMoverAlreadyMoved(t *testing.T) {
	tt := newMoverTest(t)
	objs := tt.spec.ClusterAndChildren()
	tt.buildClients(objs, objs)
	m := clustermanager.NewMover(tt.log, tt.clientFactory,
		clustermanager.WithMoverRetryBackOff(time.Millisecond),
		clustermanager.WithMoverApplyClusterTimeout(time.Millisecond),
	)

	err := m.Move(tt.ctx, tt.spec, tt.fromClient, tt.toClient)
	tt.Expect(err).To(gomega.Succeed())

	for _, obj := range objs {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
		tt.Expect(tt.toClient.Get(tt.ctx, obj.GetName(), obj.GetNamespace(), u)).To(gomega.Succeed())
		original, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		tt.Expect(err).To(gomega.Succeed())
		// the entire object including metadata/status should be equal if the object already exists in dst
		tt.Expect(u.Object).To(gomega.BeComparableTo(original))
	}
}
