package curatedpackages_test

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/aws/eks-anywhere/pkg/registry"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	registrymocks "github.com/aws/eks-anywhere/pkg/registry/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed testdata/packsages-bundle.yaml
var bundleData []byte

var bundleOnly = []releasev1.Image{
	{
		URI: "public.ecr.aws/l0g8r8j6/eks-anywhere-packages-bundles:v1-21-latest",
	},
}

type packageReaderTest struct {
	*WithT
	ctx            context.Context
	command        *curatedpackages.PackageReader
	manifestReader *mocks.MockManifestReader
	storageClient  *registrymocks.MockStorageClient
	registryName   string
	bundles        *releasev1.Bundles
}

func newPackageReaderTest(t *testing.T) *packageReaderTest {
	ctrl := gomock.NewController(t)
	r := mocks.NewMockManifestReader(ctrl)
	registryName := "public.ecr.aws/l0g8r8j6"
	sc := registrymocks.NewMockStorageClient(ctrl)
	cache := registry.NewCache()
	cache.Set(registryName, sc)
	bundles := releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.21",
					PackageController: releasev1.PackageBundle{
						Version: "test-version",
						Controller: releasev1.Image{
							URI: registryName + "/ctrl:v1",
						},
					},
				},
			},
		},
	}

	return &packageReaderTest{
		WithT:          NewWithT(t),
		ctx:            context.Background(),
		registryName:   registryName,
		manifestReader: r,
		storageClient:  sc,
		bundles:        &bundles,
		command:        curatedpackages.NewPackageReader(r, cache),
	}
}

func TestPackageReader_ReadImagesFromBundles(t *testing.T) {
	tt := newPackageReaderTest(t)
	tt.storageClient.EXPECT().PullBytes(tt.ctx, gomock.Any()).Return(bundleData, nil)

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, tt.bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).NotTo(BeEmpty())
}

func TestPackageReader_ReadImagesFromBundlesBadKubeVersion(t *testing.T) {
	tt := newPackageReaderTest(t)
	bundles := tt.bundles.DeepCopy()
	bundles.Spec.VersionsBundles[0].KubeVersion = "1"

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).To(BeEmpty())
}

func TestPackageReader_ReadImagesFromBundlesBundlePullError(t *testing.T) {
	tt := newPackageReaderTest(t)
	tt.storageClient.EXPECT().PullBytes(tt.ctx, gomock.Any()).Return([]byte{}, fmt.Errorf("oops"))

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, tt.bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).To(BeEmpty())
}

func TestPackageReader_ReadChartsFromBundles(t *testing.T) {
	tt := newPackageReaderTest(t)
	artifact := registry.NewArtifactFromURI("public.ecr.aws/l0g8r8j6/eks-anywhere-packages-bundles:v1-21-latest")
	tt.storageClient.EXPECT().PullBytes(tt.ctx, artifact).Return(bundleData, nil)

	images := tt.command.ReadChartsFromBundles(tt.ctx, tt.bundles)

	tt.Expect(images).NotTo(BeEmpty())
}

func TestPackageReader_ReadChartsFromBundlesBadKubeVersion(t *testing.T) {
	tt := newPackageReaderTest(t)
	bundles := tt.bundles.DeepCopy()
	bundles.Spec.VersionsBundles[0].KubeVersion = "1"

	images := tt.command.ReadChartsFromBundles(tt.ctx, bundles)

	tt.Expect(images).To(BeEmpty())
}

func TestPackageReader_ReadChartsFromBundlesBundlePullError(t *testing.T) {
	tt := newPackageReaderTest(t)
	tt.storageClient.EXPECT().PullBytes(tt.ctx, gomock.Any()).Return([]byte{}, fmt.Errorf("oops"))

	images := tt.command.ReadChartsFromBundles(tt.ctx, tt.bundles)

	tt.Expect(images).To(Equal(bundleOnly))
}
