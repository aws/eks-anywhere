package clustermanager_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"k8s.io/utils/pointer"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	mockswriter "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	mockskubeconfig "github.com/aws/eks-anywhere/pkg/kubeconfig/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

type clusterCreatorTest struct {
	*WithT
	ctx              context.Context
	spec             *cluster.Spec
	mgmtCluster      *types.Cluster
	applier          *mocks.MockClusterApplier
	writer           *mockswriter.MockFileWriter
	kubeconfigWriter *mockskubeconfig.MockWriter
}

func newClusterCreator(t *testing.T, clusterName string) (*clustermanager.ClusterCreator, *clusterCreatorTest) {
	ctrl := gomock.NewController(t)
	cct := &clusterCreatorTest{
		WithT:            NewWithT(t),
		applier:          mocks.NewMockClusterApplier(ctrl),
		writer:           mockswriter.NewMockFileWriter(ctrl),
		kubeconfigWriter: mockskubeconfig.NewMockWriter(ctrl),
		spec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Name = clusterName
		}),
		ctx: context.Background(),
		mgmtCluster: &types.Cluster{
			KubeconfigFile: "my-config",
		},
	}

	cc := clustermanager.NewClusterCreator(cct.applier, cct.kubeconfigWriter, cct.writer)

	return cc, cct
}

func (cct *clusterCreatorTest) expectFileCreate(fileName, path string, w io.WriteCloser) {
	cct.writer.EXPECT().Create(fileName, gomock.AssignableToTypeOf([]filewriter.FileOptionsFunc{})).Return(w, path, nil)
}

func (cct *clusterCreatorTest) expectWriteKubeconfig(clusterName string, w io.Writer) {
	cct.kubeconfigWriter.EXPECT().WriteKubeconfig(cct.ctx, clusterName, cct.mgmtCluster.KubeconfigFile, w).Return(nil)
}

func (cct *clusterCreatorTest) expectApplierRun() {
	cct.applier.EXPECT().Run(cct.ctx, cct.spec, *cct.mgmtCluster).Return(nil)
}

func TestClusterCreatorCreateSync(t *testing.T) {
	clusterName := "testCluster"
	clusCreator, tt := newClusterCreator(t, clusterName)
	path := "testpath"
	writer := os.NewFile(uintptr(*pointer.Uint(0)), "test")
	tt.expectApplierRun()
	tt.expectWriteKubeconfig(clusterName, writer)
	fileName := fmt.Sprintf("%s-eks-a-cluster.kubeconfig", clusterName)
	tt.expectFileCreate(fileName, path, writer)
	_, err := clusCreator.CreateSync(tt.ctx, tt.spec, tt.mgmtCluster)
	tt.Expect(err).To(BeNil())
}
