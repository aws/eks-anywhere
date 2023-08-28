package kindnetd_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/networking/kindnetd"
	"github.com/aws/eks-anywhere/pkg/networking/kindnetd/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type kindnetdTest struct {
	*WithT
	ctx     context.Context
	k       *kindnetd.Kindnetd
	cluster *types.Cluster
	client  *mocks.MockClient
	reader  manifests.FileReader
	spec    *cluster.Spec
}

func newKindnetdTest(t *testing.T) *kindnetdTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	reader := files.NewReader()
	return &kindnetdTest{
		WithT:  NewWithT(t),
		ctx:    context.Background(),
		client: client,
		cluster: &types.Cluster{
			Name:           "w-cluster",
			KubeconfigFile: "config.kubeconfig",
		},
		reader: reader,
		k:      kindnetd.NewKindnetd(client, reader),
		spec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.1.0/24"}
			s.VersionsBundles["1.19"].Kindnetd = kindnetdBundle
		}),
	}
}

func TestKindnetdInstallSuccess(t *testing.T) {
	tt := newKindnetdTest(t)
	tt.client.EXPECT().ApplyKubeSpecFromBytes(
		tt.ctx,
		tt.cluster,
		test.MatchFile("testdata/expected_kindnetd_manifest.yaml"),
	)

	tt.Expect(tt.k.Install(tt.ctx, tt.cluster, tt.spec, nil)).To(Succeed())
}

var kindnetdBundle = v1alpha1.KindnetdBundle{
	Manifest: v1alpha1.Manifest{
		URI: "testdata/kindnetd_manifest.yaml",
	},
}
