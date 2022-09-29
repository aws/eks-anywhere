package curatedpackages_test

import (
	"context"
	"errors"
	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
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
	bundlePuller   *mocks.MockBundlePuller
	registry       string
	workingBundle  []byte
}

func newPackageReaderTest(t *testing.T) *packageReaderTest {
	ctrl := gomock.NewController(t)
	r := mocks.NewMockManifestReader(ctrl)
	bp := mocks.NewMockBundlePuller(ctrl)
	registry := "test_registry/local"
	goodBundle := packagesv1.PackageBundle{
		Spec: packagesv1.PackageBundleSpec{
			Packages: []packagesv1.BundlePackage{
				{
					Name: "harbor",
					Source: packagesv1.BundlePackageSource{
						Versions: []packagesv1.SourceVersion{
							{
								Images: []packagesv1.VersionImages{
									{
										Repository: "harbor/harbor",
										Digest:     "sha256:v1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	workingBundle := convertJsonToBytes(goodBundle)

	return &packageReaderTest{
		WithT:          NewWithT(t),
		ctx:            context.Background(),
		registry:       registry,
		manifestReader: r,
		bundlePuller:   bp,
		workingBundle:  workingBundle.Bytes(),
		command:        curatedpackages.NewPackageReader(r, bp),
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

	expectedArtifact := "test_registry/local/eks-anywhere-packages-bundles:v1-21-latest"
	tt.manifestReader.EXPECT().ReadImagesFromBundles(tt.ctx, bundles).Return([]releasev1.Image{}, nil)
	tt.bundlePuller.EXPECT().PullLatestBundle(tt.ctx, expectedArtifact).Return(tt.workingBundle, nil)

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

	expectedArtifact := "fake_registry/fake_env/eks-anywhere-packages-bundles:v1-21-latest"
	tt.manifestReader.EXPECT().ReadImagesFromBundles(tt.ctx, bundles).Return([]releasev1.Image{}, nil)
	tt.bundlePuller.EXPECT().PullLatestBundle(tt.ctx, expectedArtifact).Return(nil, errors.New("error pulling bundle"))

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

	expectedArtifact := "test_registry/local/eks-anywhere-packages-bundles:v1-21-latest"
	tt.manifestReader.EXPECT().ReadChartsFromBundles(tt.ctx, bundles).Return([]releasev1.Image{})
	tt.bundlePuller.EXPECT().PullLatestBundle(tt.ctx, expectedArtifact).Return(tt.workingBundle, nil)

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

	expectedArtifact := "fake_registry/fake_env/eks-anywhere-packages-bundles:v1-21-latest"
	tt.manifestReader.EXPECT().ReadChartsFromBundles(tt.ctx, bundles).Return([]releasev1.Image{})
	tt.bundlePuller.EXPECT().PullLatestBundle(tt.ctx, expectedArtifact).Return(nil, errors.New("error pulling bundle"))

	images := tt.command.ReadChartsFromBundles(tt.ctx, bundles)

	tt.Expect(images).To(BeEmpty())
}
