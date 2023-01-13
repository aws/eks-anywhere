package registry_test

import (
	_ "embed"
	"testing"

	"github.com/golang/mock/gomock"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/registry"
	"github.com/aws/eks-anywhere/pkg/registry/mocks"
)

//go:embed testdata/image-manifest.json
var imageManifest []byte

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
