package clustermanager_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/types"
)

type deleterTest struct {
	Gomega
	tb            testing.TB
	clientFactory *mocks.MockClientFactory
	ctx           context.Context
	spec          *cluster.Spec
	client        kubernetes.Client
	log           logr.Logger
	mgmtCluster   types.Cluster
}

func newDeleterTest(tb testing.TB) *deleterTest {
	ctrl := gomock.NewController(tb)
	return &deleterTest{
		tb:            tb,
		Gomega:        NewWithT(tb),
		clientFactory: mocks.NewMockClientFactory(ctrl),
		ctx:           context.Background(),
		spec:          test.VSphereClusterSpec(tb, tb.Name()),
		log:           test.NewNullLogger(),
		mgmtCluster: types.Cluster{
			KubeconfigFile: "my-config",
		},
	}
}

func (a *deleterTest) buildClient(objs ...kubernetes.Object) {
	a.client = test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(objs)...)
	a.clientFactory.EXPECT().BuildClientFromKubeconfig(a.mgmtCluster.KubeconfigFile).Return(a.client, nil)
}

func TestDeleterRunClusterDeleteSuccess(t *testing.T) {
	tt := newDeleterTest(t)
	tt.spec.Cluster.Namespace = "default"
	tt.buildClient(tt.spec.Cluster)
	a := clustermanager.NewDeleter(tt.log, tt.clientFactory,
		clustermanager.WithDeleterRetryBackOff(time.Millisecond),
		clustermanager.WithDeleterNoTimeouts(),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(Succeed())
}

func TestDeleterRunErrorBuildingClient(t *testing.T) {
	tt := newDeleterTest(t)
	tt.client = test.NewFakeKubeClientAlwaysError()
	tt.clientFactory.EXPECT().BuildClientFromKubeconfig(tt.mgmtCluster.KubeconfigFile).Return(nil, errors.New("bad client"))
	a := clustermanager.NewDeleter(tt.log, tt.clientFactory,
		clustermanager.WithDeleterApplyClusterTimeout(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(MatchError(ContainSubstring("building client to delete cluster")))
}

func TestDeleterRunErrorDeleting(t *testing.T) {
	tt := newDeleterTest(t)
	tt.client = test.NewFakeKubeClientAlwaysError()
	tt.clientFactory.EXPECT().BuildClientFromKubeconfig(tt.mgmtCluster.KubeconfigFile).Return(tt.client, nil)
	a := clustermanager.NewDeleter(tt.log, tt.clientFactory,
		clustermanager.WithDeleterApplyClusterTimeout(0),
	)

	tt.Expect(a.Run(tt.ctx, tt.spec, tt.mgmtCluster)).To(MatchError(ContainSubstring("deleting cluster")))
}
