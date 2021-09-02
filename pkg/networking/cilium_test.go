package networking_test

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestCiliumGenerateManifestSuccess(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.Cilium = ciliumBundle
	})

	c := networking.NewCilium()

	gotFileContent, err := c.GenerateManifest(clusterSpec)
	if err != nil {
		t.Fatalf("Cilium.GenerateManifestFile() error = %v, wantErr nil", err)
	}

	test.AssertContentToFile(t, string(gotFileContent), ciliumBundle.Manifest.URI)
}

var ciliumBundle = v1alpha1.CiliumBundle{
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

func TestCiliumGenerateManifestWriterError(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.Cilium.Manifest.URI = "testdata/missing_manifest.yaml"
	})

	c := networking.NewCilium()

	if _, err := c.GenerateManifest(clusterSpec); err == nil {
		t.Fatalf("Cilium.GenerateManifestFile() error = nil, want not nil")
	}
}
