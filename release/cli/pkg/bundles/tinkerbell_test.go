package bundles

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	utilversion "k8s.io/apimachinery/pkg/util/version"

	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func TestGetTinkerbellBundleProviderVersion(t *testing.T) {
	commit := strings.Repeat("a", 40)
	tests := []struct {
		name            string
		gitTag          string
		expectedVersion string
		dryRun          bool
	}{
		{name: "release tag", gitTag: "v0.6.6", expectedVersion: "v0.6.6"},
		{name: "commit pin", gitTag: commit, expectedVersion: "v0.7.0-dev-" + commit},
		{name: "release tag dry run", gitTag: "v0.6.6", expectedVersion: "v0.6.6", dryRun: true},
		{name: "commit pin dry run", gitTag: commit, expectedVersion: "v0.7.0-dev-" + commit, dryRun: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifactDir := t.TempDir()
			buildRepo := t.TempDir()
			require.NoError(t, exec.Command("git", "-C", buildRepo, "init", "--initial-branch=main").Run())
			projectDir := filepath.Join(buildRepo, constants.CaptProjectPath)
			require.NoError(t, os.MkdirAll(projectDir, 0o700))
			require.NoError(t, os.WriteFile(filepath.Join(projectDir, "GIT_TAG"), []byte(tt.gitTag), 0o600))
			metadata := "releaseSeries:\n- major: 0\n  minor: 6\n- major: 0\n  minor: 7\n"
			require.NoError(t, os.WriteFile(filepath.Join(artifactDir, "metadata.yaml"), []byte(metadata), 0o600))
			imageURI := fmt.Sprintf("registry.example.com/capt:%s", tt.gitTag)
			digest := "sha256:" + strings.Repeat("b", 64)
			r := &releasetypes.ReleaseConfig{DryRun: tt.dryRun, BuildRepoSource: buildRepo}
			for _, project := range []string{"kube-vip", "envoy", "tinkerbell", "tink", "hub", "hook", "stack", "tinkerbell-chart", "tinkerbell-crds"} {
				r.BundleArtifactsTable.Store(project, []releasetypes.Artifact{})
			}
			r.BundleArtifactsTable.Store("cluster-api-provider-tinkerbell", []releasetypes.Artifact{
				{Image: &releasetypes.ImageArtifact{
					AssetName:         "cluster-api-provider-tinkerbell",
					GitTag:            tt.gitTag,
					SourcedFromBranch: "main",
					ReleaseImageURI:   imageURI,
				}},
				{Manifest: &releasetypes.ManifestArtifact{
					ArtifactPath:  artifactDir,
					ReleaseName:   "metadata.yaml",
					ReleaseCdnURI: "https://artifacts.example.com/capt/" + tt.gitTag + "/metadata.yaml",
				}},
			})
			digests := releasetypes.ImageDigestsTable{}
			digests.Store(imageURI, digest)

			bundle, err := GetTinkerbellBundle(r, digests)
			require.NoError(t, err)
			base, checksum, found := strings.Cut(bundle.Version, "+")
			assert.Equal(t, tt.expectedVersion, base)
			assert.True(t, found)
			assert.Len(t, checksum, 7)
			assert.Equal(t, imageURI, bundle.ClusterAPIController.URI)
			assert.Equal(t, digest, bundle.ClusterAPIController.ImageDigest)
			assert.Equal(t, "https://artifacts.example.com/capt/"+tt.gitTag+"/metadata.yaml", bundle.Metadata.URI)
			parsed, err := utilversion.ParseSemantic(bundle.Version)
			require.NoError(t, err)
			if tt.gitTag == commit {
				assert.True(t, parsed.AtLeast(utilversion.MustParseSemantic("v0.6.6")))
			}
			t.Logf("tinkerbell.version=%s image=%s", bundle.Version, bundle.ClusterAPIController.URI)
		})
	}
}

func TestTinkerbellProviderVersionInvalidMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
		gitTag   string
		want     string
	}{
		{name: "invalid source version", gitTag: "main", want: "must be a semantic version or a full commit SHA"},
		{name: "empty release series", metadata: "releaseSeries: []", want: "has no release series"},
		{name: "malformed metadata", metadata: "releaseSeries: [", want: "parsing CAPT metadata"},
		{name: "negative version", metadata: "releaseSeries:\n- major: -1\n  minor: 7", want: "parsing CAPT metadata"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifact := &releasetypes.ManifestArtifact{ArtifactPath: t.TempDir(), ReleaseName: "metadata.yaml"}
			require.NoError(t, os.WriteFile(filepath.Join(artifact.ArtifactPath, artifact.ReleaseName), []byte(tt.metadata), 0o600))
			gitTag := tt.gitTag
			if gitTag == "" {
				gitTag = strings.Repeat("a", 40)
			}
			_, err := tinkerbellProviderVersion(gitTag, artifact, "abcdef1")
			require.ErrorContains(t, err, tt.want)
		})
	}
	t.Run("missing metadata", func(t *testing.T) {
		_, err := tinkerbellProviderVersion(strings.Repeat("a", 40), nil, "abcdef1")
		require.ErrorContains(t, err, "metadata is required")
	})
}
