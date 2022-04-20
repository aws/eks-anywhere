package stack_test

import (
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func getVersionsBundle() cluster.VersionsBundle {
	return cluster.VersionsBundle{
		VersionsBundle: &v1alpha1.VersionsBundle{
			Tinkerbell: v1alpha1.TinkerbellBundle{
				TinkServer: v1alpha1.Image{
					URI: "public.ecr.aws/eks-anywhere/tink-server:latest",
				},
				Hegel: v1alpha1.Image{
					URI: "public.ecr.aws/eks-anywhere/hegel:latest",
				},
				Pbnj: v1alpha1.Image{
					URI: "public.ecr.aws/eks-anywhere/pbnj:latest",
				},
			},
		},
	}
}

func TestGenerateDatabaseManifestSuccess(t *testing.T) {
	tinkerbellStack := stack.NewTinkerbellStack(getVersionsBundle(), "")
	db, err := tinkerbellStack.GenerateDatabaseManifest()
	if err != nil {
		t.Fatalf("failed to generate tink manifest: %v", err)
	}

	test.AssertContentToFile(t, string(db), "testdata/database_expected_manifest.yaml")
}

func TestGenerateTinkManifestSuccess(t *testing.T) {
	tinkerbellStack := stack.NewTinkerbellStack(getVersionsBundle(), "")
	tink, err := tinkerbellStack.GenerateTinkManifest()
	if err != nil {
		t.Fatalf("failed to generate tink manifest: %v", err)
	}

	test.AssertContentToFile(t, string(tink), "testdata/tink_expected_manifest.yaml")
}

func TestGenerateHegelManifestSuccess(t *testing.T) {
	tinkerbellStack := stack.NewTinkerbellStack(getVersionsBundle(), "")
	hegel, err := tinkerbellStack.GenerateHegelManifest()
	if err != nil {
		t.Fatalf("failed to generate tink manifest: %v", err)
	}

	test.AssertContentToFile(t, string(hegel), "testdata/hegel_expected_manifest.yaml")
}

func TestGeneratePbnjManifestSuccess(t *testing.T) {
	tinkerbellStack := stack.NewTinkerbellStack(getVersionsBundle(), "")
	pbnj, err := tinkerbellStack.GeneratePbnjManifest()
	if err != nil {
		t.Fatalf("failed to generate tink manifest: %v", err)
	}

	test.AssertContentToFile(t, string(pbnj), "testdata/pbnj_expected_manifest.yaml")
}
