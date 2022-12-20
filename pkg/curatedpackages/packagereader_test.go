package curatedpackages_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type packageReaderTest struct {
	*WithT
	ctx            context.Context
	command        *curatedpackages.PackageReader
	manifestReader *mocks.MockManifestReader
	registry       string
}

func newPackageReaderTest(t *testing.T) *packageReaderTest {
	ctrl := gomock.NewController(t)
	r := mocks.NewMockManifestReader(ctrl)
	registry := "public.ecr.aws/l0g8r8j6"

	return &packageReaderTest{
		WithT:          NewWithT(t),
		ctx:            context.Background(),
		registry:       registry,
		manifestReader: r,
		command:        curatedpackages.NewPackageReader(r),
	}
}

func TestPackageReaderReadImagesFromBundlesSuccess(t *testing.T) {
	t.Skip("Test consistently fails locally as it attempts to download unreachable artifacts (https://github.com/aws/eks-anywhere/issues/3881)")
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

	images, err := tt.command.ReadImagesFromBundles(tt.ctx, bundles)

	tt.Expect(err).To(BeNil())
	tt.Expect(images).To(BeEmpty())
}

func TestPackageReaderReadChartsFromBundlesSuccess(t *testing.T) {
	t.Skip("Test consistently fails locally as it attempts to download unreachable artifacts (https://github.com/aws/eks-anywhere/issues/3881)")
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

	images := tt.command.ReadChartsFromBundles(tt.ctx, bundles)

	tt.Expect(images).To(BeEmpty())
}
