package stack_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
)

func TestGenerateTinkManifestSuccess(t *testing.T) {
	tink, err := stack.GenerateTinkManifest("tink-server-image", "1.2.3.4")
	if err != nil {
		t.Fatalf("failed to generate tink manifest: %v", err)
	}
	t.Log(string(tink))
}
