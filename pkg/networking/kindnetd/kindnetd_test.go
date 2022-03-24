package kindnetd_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking/kindnetd"
	"github.com/aws/eks-anywhere/pkg/networking/kindnetd/mocks"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type kindnetdTest struct {
	*WithT
	k      *kindnetd.Kindnetd
	client *mocks.MockClient
}

func newKindnetdTest(t *testing.T) *kindnetdTest {
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	return &kindnetdTest{
		WithT:  NewWithT(t),
		client: client,
		k:      kindnetd.NewKindnetd(client),
	}
}

func TestKindnetdGenerateManifestSuccess(t *testing.T) {
	tt := newKindnetdTest(t)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.1.0/24"}
		s.VersionsBundle.Kindnetd = KindnetdBundle
	})

	gotFileContent, err := tt.k.GenerateManifest(context.Background(), clusterSpec, []string{})
	if err != nil {
		t.Fatalf("Kindnetd.GenerateManifestFile() error = %v, wantErr nil", err)
	}

	test.AssertContentToFile(t, string(gotFileContent), "testdata/expected_kindnetd_manifest.yaml")
}

var KindnetdBundle = v1alpha1.KindnetdBundle{
	Manifest: v1alpha1.Manifest{
		URI: "testdata/kindnetd_manifest.yaml",
	},
}

func TestKindnetdGenerateManifestWriterError(t *testing.T) {
	tt := newKindnetdTest(t)
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.Kindnetd.Manifest.URI = "testdata/missing_manifest.yaml"
	})

	if _, err := tt.k.GenerateManifest(context.Background(), clusterSpec, []string{}); err == nil {
		t.Fatalf("Kindnetd.GenerateManifestFile() error = nil, want not nil")
	}
}
