package cilium_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/networking/cilium/mocks"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type ciliumtest struct {
	*WithT
	ctx    context.Context
	client mocks.MockClient
	h      *mocks.MockHelm
	spec   *cluster.Spec
	cilium *cilium.Cilium
}

func newCiliumTest(t *testing.T) *ciliumtest {
	ctrl := gomock.NewController(t)
	h := mocks.NewMockHelm(ctrl)
	client := mocks.NewMockClient(ctrl)
	return &ciliumtest{
		WithT:  NewWithT(t),
		ctx:    context.Background(),
		client: *client,
		h:      h,
		spec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.VersionsBundle.Cilium = v1alpha1.CiliumBundle{
				Cilium: v1alpha1.Image{
					URI: "public.ecr.aws/isovalent/cilium:v1.9.10-eksa.1",
				},
				Operator: v1alpha1.Image{
					URI: "public.ecr.aws/isovalent/operator-generic:v1.9.10-eksa.1",
				},
				Manifest: v1alpha1.Manifest{
					URI: "testdata/cilium_manifest.yaml",
				},
			}
		}),
		cilium: cilium.NewCilium(client, h),
	}
}

func TestCiliumGenerateManifestSuccess(t *testing.T) {
	tt := newCiliumTest(t)

	gotFileContent, err := tt.cilium.GenerateManifest(tt.spec)
	tt.Expect(err).To(Not(HaveOccurred()), "GenerateManifest() should succeed")
	test.AssertContentToFile(t, string(gotFileContent), tt.spec.VersionsBundle.Cilium.Manifest.URI)
}

func TestCiliumGenerateManifestWriterError(t *testing.T) {
	tt := newCiliumTest(t)
	tt.spec.VersionsBundle.Cilium.Manifest.URI = "testdata/missing_manifest.yaml"

	_, err := tt.cilium.GenerateManifest(tt.spec)
	tt.Expect(err).To(MatchError(ContainSubstring("can't load networking manifest [testdata/missing_manifest.yaml]")), "GenerateManifest() should fail with missing file error")
}
