package curatedpackages_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	registrymocks "github.com/aws/eks-anywhere/pkg/registry/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

//go:embed testdata/bundle.yaml
var bundleData []byte

type packageReaderTest struct {
	*WithT
	ctx            context.Context
	command        *curatedpackages.PackageReader
	manifestReader *mocks.MockManifestReader
	storageClient  *registrymocks.MockStorageClient
	registry       string
}

func newPackageReaderTest(t *testing.T) *packageReaderTest {
	ctrl := gomock.NewController(t)
	r := mocks.NewMockManifestReader(ctrl)
	registry := "public.ecr.aws/l0g8r8j6"
	sc := registrymocks.NewMockStorageClient(ctrl)

	return &packageReaderTest{
		WithT:          NewWithT(t),
		ctx:            context.Background(),
		registry:       registry,
		manifestReader: r,
		storageClient:  sc,
		command:        curatedpackages.NewPackageReader(r, sc),
	}
}

func TestPackageReaderReadImagesFromBundlesSuccess(t *testing.T) {
	tt := newPackageReaderTest(t)
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.21",
					PackageController: releasev1.PackageBundle{
						Version: "test-version",
						Controller: releasev1.Image{
							URI: tt.registry + "/ctrl:v1",
						},
					},
				},
			},
		},
	}
	tt.manifestReader.EXPECT().ReadImagesFromBundles(tt.ctx, bundles).Return([]releasev1.Image{}, nil)
	tt.storageClient.EXPECT().PullBytes(tt.ctx, gomock.Any()).Return(bundleData, nil)

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).NotTo(BeEmpty())
}

func TestPackageReaderReadImagesFromBundlesFail(t *testing.T) {
	tt := newPackageReaderTest(t)
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1",
					PackageController: releasev1.PackageBundle{
						Version: "test-version",
						Controller: releasev1.Image{
							URI: tt.registry + "/ctrl:v1",
						},
					},
				},
			},
		},
	}
	tt.manifestReader.EXPECT().ReadImagesFromBundles(tt.ctx, bundles).Return([]releasev1.Image{}, nil)

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).To(BeEmpty())
}

func TestPackageReaderReadImagesFromBundlesFailWhenWrongBundle(t *testing.T) {
	tt := newPackageReaderTest(t)
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.21",
					PackageController: releasev1.PackageBundle{
						Version: "test-version",
						Controller: releasev1.Image{
							URI: "fake_registry/fake_env/ctrl:v1",
						},
					},
				},
			},
		},
	}
	tt.manifestReader.EXPECT().ReadImagesFromBundles(tt.ctx, bundles).Return([]releasev1.Image{}, nil)
	tt.storageClient.EXPECT().PullBytes(tt.ctx, gomock.Any()).Return([]byte{}, nil)

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).To(BeEmpty())
}

func TestPackageReaderReadChartsFromBundlesSuccess(t *testing.T) {
	tt := newPackageReaderTest(t)
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.21",
					PackageController: releasev1.PackageBundle{
						Version: "test-version",
						Controller: releasev1.Image{
							URI: tt.registry + "/ctrl:v1",
						},
					},
				},
			},
		},
	}
	tt.manifestReader.EXPECT().ReadChartsFromBundles(tt.ctx, bundles).Return([]releasev1.Image{})
	tt.storageClient.EXPECT().PullBytes(tt.ctx, gomock.Any()).Return(bundleData, nil)

	images := tt.command.ReadChartsFromBundles(tt.ctx, bundles)

	tt.Expect(images).NotTo(BeEmpty())
}

func TestPackageReaderReadChartsFromBundlesFail(t *testing.T) {
	tt := newPackageReaderTest(t)
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1",
					PackageController: releasev1.PackageBundle{
						Version: "test-version",
						Controller: releasev1.Image{
							URI: tt.registry + "/ctrl:v1",
						},
					},
				},
			},
		},
	}
	tt.manifestReader.EXPECT().ReadChartsFromBundles(tt.ctx, bundles).Return([]releasev1.Image{})

	images := tt.command.ReadChartsFromBundles(tt.ctx, bundles)

	tt.Expect(images).To(BeEmpty())
}

func TestPackageReaderReadChartsFromBundlesFailWhenWrongURI(t *testing.T) {
	tt := newPackageReaderTest(t)
	bundles := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.21",
					PackageController: releasev1.PackageBundle{
						Version: "test-version",
						Controller: releasev1.Image{
							URI: "fake_registry/fake_env/ctrl:v1",
						},
					},
				},
			},
		},
	}
	tt.manifestReader.EXPECT().ReadChartsFromBundles(tt.ctx, bundles).Return([]releasev1.Image{})
	tt.storageClient.EXPECT().PullBytes(tt.ctx, gomock.Any()).Return([]byte{}, nil)

	images := tt.command.ReadChartsFromBundles(tt.ctx, bundles)

	tt.Expect(images).To(BeEmpty())
}
