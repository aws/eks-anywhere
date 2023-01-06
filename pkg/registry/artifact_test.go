package registry_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry"
)

func TestArtifact_VersionTag(t *testing.T) {
	artifact := registry.NewArtifact("localhost:8443", "owner/repo", "latest", "")
	assert.Equal(t, ":latest", artifact.Version())
	assert.Equal(t, "localhost:8443/owner/repo:latest", artifact.VersionedImage())
}

func TestArtifact_VersionDigest(t *testing.T) {
	artifact := registry.NewArtifact("localhost:8443", "owner/repo", "", "sha256:0db6a")
	assert.Equal(t, "@sha256:0db6a", artifact.Version())
	assert.Equal(t, "localhost:8443/owner/repo@sha256:0db6a", artifact.VersionedImage())
}
