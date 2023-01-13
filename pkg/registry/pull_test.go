package registry_test

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry"
	"github.com/aws/eks-anywhere/pkg/registry/mocks"
)

//go:embed testdata/image-manifest.json
var imageManifest []byte

//go:embed testdata/bad-image-manifest.json
var badImageManifest []byte

//go:embed testdata/no-layer-image-manifest.json
var noLayersImageManifest []byte

//go:embed testdata/package-bundle.yaml
var packageBundle []byte

func TestPull(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	desc := ocispec.Descriptor{
		Digest: "sha256:8bc5f46db8c98aedfba4ade0d7ebbdecd8e4130e172d3d62871fc3258c40a910",
	}
	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	srcClient.EXPECT().FetchBytes(ctx, &mockSrcRepo, srcArtifact).Return(desc, imageManifest, nil)
	srcClient.EXPECT().FetchBlob(ctx, &mockSrcRepo, gomock.Any()).Return(packageBundle, nil)

	result, err := registry.PullBytes(ctx, srcClient, srcArtifact)
	assert.NotEmpty(t, result)
	assert.NoError(t, err)
}

func TestPullFetchBlobFail(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	desc := ocispec.Descriptor{
		Digest: "sha256:8bc5f46db8c98aedfba4ade0d7ebbdecd8e4130e172d3d62871fc3258c40a910",
	}
	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	srcClient.EXPECT().FetchBytes(ctx, &mockSrcRepo, srcArtifact).Return(desc, imageManifest, nil)
	srcClient.EXPECT().FetchBlob(ctx, &mockSrcRepo, gomock.Any()).Return(packageBundle, fmt.Errorf("oops"))

	result, err := registry.PullBytes(ctx, srcClient, srcArtifact)
	assert.Nil(t, result)
	assert.EqualError(t, err, "fetch blob: oops")
}

func TestPullUnmarshalFail(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	desc := ocispec.Descriptor{
		Digest: "sha256:8bc5f46db8c98aedfba4ade0d7ebbdecd8e4130e172d3d62871fc3258c40a910",
	}
	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	srcClient.EXPECT().FetchBytes(ctx, &mockSrcRepo, srcArtifact).Return(desc, badImageManifest, nil)

	result, err := registry.PullBytes(ctx, srcClient, srcArtifact)
	assert.Nil(t, result)
	assert.EqualError(t, err, "unmarshal manifest: unexpected end of JSON input")
}

func TestPullNoLayerFail(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	desc := ocispec.Descriptor{
		Digest: "sha256:8bc5f46db8c98aedfba4ade0d7ebbdecd8e4130e172d3d62871fc3258c40a910",
	}
	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	srcClient.EXPECT().FetchBytes(ctx, &mockSrcRepo, srcArtifact).Return(desc, noLayersImageManifest, nil)

	result, err := registry.PullBytes(ctx, srcClient, srcArtifact)
	assert.Nil(t, result)
	assert.EqualError(t, err, "missing layer")
}

func TestPullFetchBytesFail(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	desc := ocispec.Descriptor{
		Digest: "sha256:8bc5f46db8c98aedfba4ade0d7ebbdecd8e4130e172d3d62871fc3258c40a910",
	}
	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, nil)
	srcClient.EXPECT().FetchBytes(ctx, &mockSrcRepo, srcArtifact).Return(desc, imageManifest, fmt.Errorf("oops"))

	result, err := registry.PullBytes(ctx, srcClient, srcArtifact)
	assert.Nil(t, result)
	assert.EqualError(t, err, "fetch manifest: oops")
}

func TestPullGetStorageFail(t *testing.T) {
	srcClient := mocks.NewMockStorageClient(gomock.NewController(t))

	mockSrcRepo := *mocks.NewMockRepository(gomock.NewController(t))
	srcClient.EXPECT().GetStorage(ctx, srcArtifact).Return(&mockSrcRepo, fmt.Errorf("oops"))

	result, err := registry.PullBytes(ctx, srcClient, srcArtifact)
	assert.Nil(t, result)
	assert.EqualError(t, err, "repository source: oops")
}
