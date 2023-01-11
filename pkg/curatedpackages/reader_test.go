package curatedpackages

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	registrymocks "github.com/aws/eks-anywhere/pkg/registry/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var ctx = context.Background()

func getURIs(images []releasev1.Image) []string {
	var result []string
	for _, image := range images {
		result = append(result, image.URI)
	}
	return result
}

func TestPackageReader_ReadChartsFromBundles(t *testing.T) {
	eksaImages := []releasev1.Image{
		//{
		//	URI: "public.ecr.aws/l0g8r8j6/eks-anywhere-packages:0.2.21-eks-a-v0.0.0-dev-build.5058",
		//},
	}
	b := releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.24",
					PackageController: releasev1.PackageBundle{
						Controller: releasev1.Image{
							URI: "public.ecr.aws/l0g8r8j6/eks-anywhere-packages:0.2.21-eks-a-v0.0.0-dev-build.5058",
						},
					},
				},
			},
		},
	}
	mockReader := mocks.NewMockManifestReader(gomock.NewController(t))
	mockReader.EXPECT().ReadChartsFromBundles(ctx, &b).Return(eksaImages)
	sut := NewPackageReader(mockReader)
	mockStorageClient := registrymocks.NewMockStorageClient(gomock.NewController(t))
	mockStorageClient.EXPECT().PullBytes(ctx, gomock.Any()).Return([]byte{}, nil)
	sut.storageClient = mockStorageClient

	images := sut.ReadChartsFromBundles(ctx, &b)
	expected := []string{
		"public.ecr.aws/l0g8r8j6/eks-anywhere-packages-bundles:v1-24-latest",
		"public.ecr.aws/l0g8r8j6/eks-anywhere-packages:0.2.21-eks-a-v0.0.0-dev-build.5058",
	}
	assert.Equal(t, expected, getURIs(images))
}
