package curatedpackages_test

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"oras.land/oras-go/v2/registry/remote"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/registry"
	registrymocks "github.com/aws/eks-anywhere/pkg/registry/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed testdata/image-manifest.json
var imageManifest []byte

//go:embed testdata/package-bundle.yaml
var packageBundle []byte

var desc = ocispec.Descriptor{}

type packageReaderTest struct {
	*WithT
	ctx           context.Context
	command       *curatedpackages.PackageReader
	storageClient *registrymocks.MockStorageClient
	registryName  string
	bundles       *releasev1.Bundles
}

func newPackageReaderTest(t *testing.T) *packageReaderTest {
	ctrl := gomock.NewController(t)
	registryName := "public.ecr.aws"
	sc := registrymocks.NewMockStorageClient(ctrl)
	cache := registry.NewCache()
	cache.Set(registryName, sc)
	credentialStore := registry.NewCredentialStore()
	bundles := releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.21",
					PackageController: releasev1.PackageBundle{
						Version: "test-version",
						Controller: releasev1.Image{
							URI: registryName + "/l0g8r8j6/ctrl:v1",
						},
					},
				},
			},
		},
	}

	return &packageReaderTest{
		WithT:         NewWithT(t),
		ctx:           context.Background(),
		registryName:  registryName,
		storageClient: sc,
		bundles:       &bundles,
		command:       curatedpackages.NewPackageReader(cache, credentialStore, "us-east-1"),
	}
}

func TestPackageReader_ReadImagesFromBundles(t *testing.T) {
	tt := newPackageReaderTest(t)
	repo, err := remote.NewRepository("owner/name")
	assert.NoError(t, err)
	tt.storageClient.EXPECT().GetStorage(tt.ctx, gomock.Any()).Return(repo, nil)
	tt.storageClient.EXPECT().FetchBytes(tt.ctx, gomock.Any(), gomock.Any()).Return(desc, imageManifest, nil)
	tt.storageClient.EXPECT().FetchBlob(tt.ctx, gomock.Any(), gomock.Any()).Return(packageBundle, nil)

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, tt.bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).NotTo(BeEmpty())
}

func TestPackageReader_ReadImagesFromBundlesProduction(t *testing.T) {
	tt := newPackageReaderTest(t)
	artifact := registry.NewArtifactFromURI("public.ecr.aws/eks-anywhere/eks-anywhere-packages-bundles:v1-21-latest")
	repo, err := remote.NewRepository("owner/name")
	assert.NoError(t, err)
	tt.storageClient.EXPECT().GetStorage(tt.ctx, gomock.Any()).Return(repo, nil)
	tt.storageClient.EXPECT().FetchBytes(tt.ctx, gomock.Any(), artifact).Return(desc, imageManifest, nil)
	tt.storageClient.EXPECT().FetchBlob(tt.ctx, gomock.Any(), gomock.Any()).Return(packageBundle, nil)
	tt.bundles.Spec.VersionsBundles[0].PackageController.Controller.URI = tt.registryName + "/eks-anywhere/ctrl:v1"

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

func TestPackageReader_ReadImagesFromBundlesBadRegistry(t *testing.T) {
	tt := newPackageReaderTest(t)
	tt.bundles.Spec.VersionsBundles[0].PackageController.Controller.URI = "!@#$/eks-anywhere/ctrl:v1"

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, tt.bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).To(BeEmpty())
}

func TestPackageReader_ReadImagesFromBundlesBadData(t *testing.T) {
	tt := newPackageReaderTest(t)
	repo, err := remote.NewRepository("owner/name")
	assert.NoError(t, err)
	tt.storageClient.EXPECT().GetStorage(tt.ctx, gomock.Any()).Return(repo, nil)
	tt.storageClient.EXPECT().FetchBytes(tt.ctx, gomock.Any(), gomock.Any()).Return(desc, []byte("wot?"), nil)

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, tt.bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).To(BeEmpty())
}

func TestPackageReader_ReadImagesFromBundlesBundlePullError(t *testing.T) {
	tt := newPackageReaderTest(t)
	repo, err := remote.NewRepository("owner/name")
	assert.NoError(t, err)
	tt.storageClient.EXPECT().GetStorage(tt.ctx, gomock.Any()).Return(repo, nil)
	tt.storageClient.EXPECT().FetchBytes(tt.ctx, gomock.Any(), gomock.Any()).Return(desc, []byte{}, fmt.Errorf("oops"))

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, tt.bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).To(BeEmpty())
}

func TestPackageReader_ReadChartsFromBundles(t *testing.T) {
	tt := newPackageReaderTest(t)
	artifact := registry.NewArtifactFromURI("public.ecr.aws/l0g8r8j6/eks-anywhere-packages-bundles:v1-21-latest")
	repo, err := remote.NewRepository("owner/name")
	assert.NoError(t, err)
	tt.storageClient.EXPECT().GetStorage(tt.ctx, gomock.Any()).Return(repo, nil)
	tt.storageClient.EXPECT().FetchBytes(tt.ctx, gomock.Any(), artifact).Return(desc, imageManifest, nil)
	tt.storageClient.EXPECT().FetchBlob(tt.ctx, gomock.Any(), gomock.Any()).Return(packageBundle, nil)

	images := tt.command.ReadChartsFromBundles(tt.ctx, tt.bundles)

	tt.Expect(images).NotTo(BeEmpty())
}

func TestPackageReader_ReadChartsFromBundlesProduction(t *testing.T) {
	tt := newPackageReaderTest(t)
	artifact := registry.NewArtifactFromURI("public.ecr.aws/eks-anywhere/eks-anywhere-packages-bundles:v1-21-latest")
	repo, err := remote.NewRepository("owner/name")
	assert.NoError(t, err)
	tt.storageClient.EXPECT().GetStorage(tt.ctx, gomock.Any()).Return(repo, nil)
	tt.storageClient.EXPECT().FetchBytes(tt.ctx, gomock.Any(), artifact).Return(desc, imageManifest, nil)
	tt.storageClient.EXPECT().FetchBlob(tt.ctx, gomock.Any(), gomock.Any()).Return(packageBundle, nil)
	tt.bundles.Spec.VersionsBundles[0].PackageController.Controller.URI = tt.registryName + "/eks-anywhere/ctrl:v1"

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
	repo, err := remote.NewRepository("owner/name")
	assert.NoError(t, err)
	tt.storageClient.EXPECT().GetStorage(tt.ctx, gomock.Any()).Return(repo, nil)
	tt.storageClient.EXPECT().FetchBytes(tt.ctx, gomock.Any(), gomock.Any()).Return(desc, []byte{}, fmt.Errorf("oops"))

	images := tt.command.ReadChartsFromBundles(tt.ctx, tt.bundles)

	tt.Expect(images).To(BeEmpty())
}
