package bundles_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test/mocks"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var eksdManifest = `apiVersion: distro.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  creationTimestamp: null
  name: kubernetes-1-20-eks-1
spec:
  channel: 1-20
  number: 1
status:
  components:
    - assets:
        - arch:
            - amd64
            - arm64
          description: node-driver-registrar container image
          image:
            uri: public.ecr.aws/eks-distro/kubernetes-csi/node-driver-registrar:v2.1.0-eks-1-20-1
          name: node-driver-registrar-image
          os: linux
          type: Image
      gitTag: v2.1.0
      name: node-driver-registrar
    - assets:
        - arch:
            - amd64
            - arm64
          description: csi-snapshotter container image
          image:
            uri: public.ecr.aws/eks-distro/kubernetes-csi/external-snapshotter/csi-snapshotter:v3.0.3-eks-1-20-1
          name: csi-snapshotter-image
          os: linux
          type: Image`

func TestReadImages(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)

	eksdURL := "eksdurl"
	b := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					EksD: releasev1.EksDRelease{
						EksDReleaseUrl: eksdURL,
					},
				},
			},
		},
	}

	driverImage := releasev1.Image{
		Name:        "csi-snapshotter-image",
		Description: "csi-snapshotter container image",
		OS:          "linux",
		OSName:      "",
		Arch:        []string{"amd64", "arm64"},
		URI:         "public.ecr.aws/eks-distro/kubernetes-csi/external-snapshotter/csi-snapshotter:v3.0.3-eks-1-20-1",
		ImageDigest: "",
	}

	csiImage := releasev1.Image{
		Name:        "csi-snapshotter-image",
		Description: "csi-snapshotter container image",
		OS:          "linux",
		OSName:      "",
		Arch:        []string{"amd64", "arm64"},
		URI:         "public.ecr.aws/eks-distro/kubernetes-csi/external-snapshotter/csi-snapshotter:v3.0.3-eks-1-20-1",
		ImageDigest: "",
	}

	reader.EXPECT().ReadFile(eksdURL).Return([]byte(eksdManifest), nil)

	gotImages, err := bundles.ReadImages(reader, b)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(gotImages)).To(BeNumerically(">", 2), "it should return more than the two images from eksd")
	g.Expect(gotImages).To(ContainElement(driverImage), "it should return the node drive registar image in the eksd manifest")
	g.Expect(gotImages).To(ContainElement(csiImage), "it should return the csi image in the eksd manifest")
}

func TestReadImagesWithKubeVersionFilter(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)

	eksdURL := "eksdurl"
	b := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					KubeVersion: "1.20",
					EksD: releasev1.EksDRelease{
						EksDReleaseUrl: eksdURL,
					},
				},
				{
					KubeVersion: "1.21",
					EksD: releasev1.EksDRelease{
						EksDReleaseUrl: "fake url",
					},
				},
			},
		},
	}

	driverImage := releasev1.Image{
		Name:        "csi-snapshotter-image",
		Description: "csi-snapshotter container image",
		OS:          "linux",
		OSName:      "",
		Arch:        []string{"amd64", "arm64"},
		URI:         "public.ecr.aws/eks-distro/kubernetes-csi/external-snapshotter/csi-snapshotter:v3.0.3-eks-1-20-1",
		ImageDigest: "",
	}

	csiImage := releasev1.Image{
		Name:        "csi-snapshotter-image",
		Description: "csi-snapshotter container image",
		OS:          "linux",
		OSName:      "",
		Arch:        []string{"amd64", "arm64"},
		URI:         "public.ecr.aws/eks-distro/kubernetes-csi/external-snapshotter/csi-snapshotter:v3.0.3-eks-1-20-1",
		ImageDigest: "",
	}

	reader.EXPECT().ReadFile(eksdURL).Return([]byte(eksdManifest), nil)

	gotImages, err := bundles.ReadImages(reader, b, "1.20")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(len(gotImages)).To(BeNumerically(">", 2), "it should return more than the two images from eksd")
	g.Expect(gotImages).To(ContainElement(driverImage), "it should return the node drive registar image in the eksd manifest")
	g.Expect(gotImages).To(ContainElement(csiImage), "it should return the csi image in the eksd manifest")
}

func TestReadImagesError(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	eksdURL := "eksdurl"
	b := &releasev1.Bundles{
		Spec: releasev1.BundlesSpec{
			VersionsBundles: []releasev1.VersionsBundle{
				{
					EksD: releasev1.EksDRelease{
						EksDReleaseUrl: eksdURL,
					},
				},
			},
		},
	}

	reader.EXPECT().ReadFile(eksdURL).Return(nil, errors.New("error reading eksd"))

	_, err := bundles.ReadImages(reader, b)
	g.Expect(err).To(MatchError(ContainSubstring("reading images from Bundle: error reading eksd")))
}
