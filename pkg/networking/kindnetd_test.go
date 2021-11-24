package networking_test

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestKindnetdGenerateManifestSuccess(t *testing.T) {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"192.168.1.0/24"}
		s.VersionsBundle.Kindnetd = KindnetdBundle
	})

	c := networking.NewKindnetd()

	gotFileContent, err := c.GenerateManifest(clusterSpec)
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
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.VersionsBundle.Kindnetd.Manifest.URI = "testdata/missing_manifest.yaml"
	})

	c := networking.NewKindnetd()

	if _, err := c.GenerateManifest(clusterSpec); err == nil {
		t.Fatalf("Kindnetd.GenerateManifestFile() error = nil, want not nil")
	}
}
