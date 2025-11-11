package bundles

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
)

func TestGetBottlerocketBootstrapContainersBundle(t *testing.T) {
	tests := []struct {
		name               string
		eksDReleaseChannel string
		artifacts          map[string][]releasetypes.Artifact
		imageDigests       map[string]string
		expectedBundle     anywherev1alpha1.BottlerocketBootstrapContainersBundle
		expectError        bool
	}{
		{
			name:               "successful bundle generation",
			eksDReleaseChannel: "1-28",
			artifacts: map[string][]releasetypes.Artifact{
				"bottlerocket-bootstrap-1-28": {
					{
						Image: &releasetypes.ImageArtifact{
							AssetName:       "bottlerocket-bootstrap-vsphere-multi-network",
							ReleaseImageURI: "public.ecr.aws/release-container-registry/bottlerocket-bootstrap-vsphere-multi-network:v1-28-63-eks-a-v0.0.0-dev-build.1",
							OS:              "linux",
							Arch:            []string{"amd64", "arm64"},
						},
					},
				},
			},
			imageDigests: map[string]string{
				"public.ecr.aws/release-container-registry/bottlerocket-bootstrap-vsphere-multi-network:v1-28-63-eks-a-v0.0.0-dev-build.1": "sha256:abcdef123456",
			},
			expectedBundle: anywherev1alpha1.BottlerocketBootstrapContainersBundle{
				VsphereMultiNetworkBootstrap: anywherev1alpha1.Image{
					Name:        "bottlerocket-bootstrap-vsphere-multi-network",
					Description: "Container image for bottlerocket-bootstrap-vsphere-multi-network image",
					OS:          "linux",
					Arch:        []string{"amd64", "arm64"},
					URI:         "public.ecr.aws/release-container-registry/bottlerocket-bootstrap-vsphere-multi-network:v1-28-63-eks-a-v0.0.0-dev-build.1",
					ImageDigest: "sha256:abcdef123456",
				},
			},
			expectError: false,
		},
		{
			name:               "missing artifacts returns empty bundle",
			eksDReleaseChannel: "1-28",
			artifacts:          map[string][]releasetypes.Artifact{},
			imageDigests:       map[string]string{},
			expectedBundle:     anywherev1alpha1.BottlerocketBootstrapContainersBundle{},
			expectError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock release config
			bundleArtifactsTable := releasetypes.ArtifactsTable{}
			for key, artifacts := range tt.artifacts {
				bundleArtifactsTable.Store(key, artifacts)
			}

			imageDigestsTable := releasetypes.ImageDigestsTable{}
			for uri, digest := range tt.imageDigests {
				imageDigestsTable.Store(uri, digest)
			}

			r := &releasetypes.ReleaseConfig{
				BundleArtifactsTable: bundleArtifactsTable,
			}

			// Test the function
			result, err := GetBottlerocketBootstrapContainersBundle(r, tt.eksDReleaseChannel, imageDigestsTable)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBundle, result)
			}
		})
	}
}

func TestGetBottlerocketBootstrapArtifact(t *testing.T) {
	tests := []struct {
		name               string
		eksDReleaseChannel string
		assetName          string
		artifacts          map[string][]releasetypes.Artifact
		imageDigests       map[string]string
		expectedImage      anywherev1alpha1.Image
		expectError        bool
	}{
		{
			name:               "successful artifact retrieval",
			eksDReleaseChannel: "1-28",
			assetName:          "bottlerocket-bootstrap-vsphere-multi-network",
			artifacts: map[string][]releasetypes.Artifact{
				"bottlerocket-bootstrap-1-28": {
					{
						Image: &releasetypes.ImageArtifact{
							AssetName:       "bottlerocket-bootstrap-vsphere-multi-network",
							ReleaseImageURI: "public.ecr.aws/release-container-registry/bottlerocket-bootstrap-vsphere-multi-network:v1-28-63-eks-a-v0.0.0-dev-build.1",
							OS:              "linux",
							Arch:            []string{"amd64", "arm64"},
						},
					},
				},
			},
			imageDigests: map[string]string{
				"public.ecr.aws/release-container-registry/bottlerocket-bootstrap-vsphere-multi-network:v1-28-63-eks-a-v0.0.0-dev-build.1": "sha256:abcdef123456",
			},
			expectedImage: anywherev1alpha1.Image{
				Name:        "bottlerocket-bootstrap-vsphere-multi-network",
				Description: "Container image for bottlerocket-bootstrap-vsphere-multi-network image",
				OS:          "linux",
				Arch:        []string{"amd64", "arm64"},
				URI:         "public.ecr.aws/release-container-registry/bottlerocket-bootstrap-vsphere-multi-network:v1-28-63-eks-a-v0.0.0-dev-build.1",
				ImageDigest: "sha256:abcdef123456",
			},
			expectError: false,
		},
		{
			name:               "artifact not found",
			eksDReleaseChannel: "1-28",
			assetName:          "non-existent-artifact",
			artifacts: map[string][]releasetypes.Artifact{
				"bottlerocket-bootstrap-1-28": {
					{
						Image: &releasetypes.ImageArtifact{
							AssetName:       "bottlerocket-bootstrap",
							ReleaseImageURI: "public.ecr.aws/release-container-registry/bottlerocket-bootstrap:v1-28-63-eks-a-v0.0.0-dev-build.1",
							OS:              "linux",
							Arch:            []string{"amd64", "arm64"},
						},
					},
				},
			},
			imageDigests: map[string]string{},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock release config
			bundleArtifactsTable := releasetypes.ArtifactsTable{}
			for key, artifacts := range tt.artifacts {
				bundleArtifactsTable.Store(key, artifacts)
			}

			imageDigestsTable := releasetypes.ImageDigestsTable{}
			for uri, digest := range tt.imageDigests {
				imageDigestsTable.Store(uri, digest)
			}

			r := &releasetypes.ReleaseConfig{
				BundleArtifactsTable: bundleArtifactsTable,
			}

			// Test the function
			result, err := getBottlerocketBootstrapArtifact(r, tt.eksDReleaseChannel, imageDigestsTable, tt.assetName)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedImage, result)
			}
		})
	}
}
