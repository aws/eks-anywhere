package manifests_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test/mocks"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/manifests/releases"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestReaderReadBundlesForVersion(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)

	releasesURL := releases.ManifestURL()

	releasesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: release-1
spec:
  releases:
    - bundleManifestUrl: "https://bundles/bundles.yaml"
      version: v0.0.1`
	reader.EXPECT().ReadFile(releasesURL).Return([]byte(releasesManifest), nil)

	bundlesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Bundles
metadata:
  name: bundles-1`

	wantBundles := &releasev1.Bundles{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
			Kind:       "Bundles",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "bundles-1",
		},
	}

	reader.EXPECT().ReadFile("https://bundles/bundles.yaml").Return([]byte(bundlesManifest), nil)

	r := manifests.NewReader(reader)
	g.Expect(r.ReadBundlesForVersion("v0.0.1")).To(Equal(wantBundles))
}

func TestReaderReadBundlesForVersionErrorVersionNotSupported(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)

	releasesURL := releases.ManifestURL()

	releasesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: release-1
spec:
  releases:
    - bundleManifestUrl: "https://bundles/bundles.yaml"
      version: v0.0.1`
	reader.EXPECT().ReadFile(releasesURL).Return([]byte(releasesManifest), nil)

	r := manifests.NewReader(reader)
	_, err := r.ReadBundlesForVersion("v0.0.2")
	g.Expect(err).To(MatchError(ContainSubstring("invalid version v0.0.2, no matching release found")))
}

func TestReaderReadBundleForVersionNotExists(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)

	releasesURL := releases.ManifestURL()

	reader.EXPECT().ReadFile(releasesURL).Return(nil, errors.New("reading Releases file"))
	r := manifests.NewReader(reader)
	_, err := r.ReadBundlesForVersion("")
	g.Expect(err).To(MatchError(ContainSubstring("reading Releases file")))
}

func TestReaderReadEKSD(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)

	releasesURL := releases.ManifestURL()

	releasesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: release-1
spec:
  releases:
    - bundleManifestUrl: "https://bundles/bundles.yaml"
      version: v0.0.1`
	reader.EXPECT().ReadFile(releasesURL).Return([]byte(releasesManifest), nil)

	bundlesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Bundles
metadata:
  name: bundles-1
spec:
  versionsBundles:
  - kubeVersion: "1.21"
    eksD:
      channel: 1-21
      kubeVersion: v1.21.5
      manifestUrl: https://distro.eks.amazonaws.com/kubernetes-1-21/kubernetes-1-21-eks-7.yaml`

	reader.EXPECT().ReadFile("https://bundles/bundles.yaml").Return([]byte(bundlesManifest), nil)
	reader.EXPECT().ReadFile("https://distro.eks.amazonaws.com/kubernetes-1-21/kubernetes-1-21-eks-7.yaml").Return([]byte(bundlesManifest), nil)

	r := manifests.NewReader(reader)
	_, err := r.ReadEKSD("v0.0.1", "1.21")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestReaderReadEKSDUnsupportedKubeVersion(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)

	releasesURL := releases.ManifestURL()

	releasesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: release-1
spec:
  releases:
    - bundleManifestUrl: "https://bundles/bundles.yaml"
      version: v0.0.1`
	reader.EXPECT().ReadFile(releasesURL).Return([]byte(releasesManifest), nil)

	bundlesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Bundles
metadata:
  name: bundles-1
spec:
  versionsBundles:
  - kubeVersion: "1.21"
    eksD:
      channel: 1-21
      kubeVersion: v1.21.5
      manifestUrl: https://distro.eks.amazonaws.com/kubernetes-1-21/kubernetes-1-21-eks-7.yaml`

	reader.EXPECT().ReadFile("https://bundles/bundles.yaml").Return([]byte(bundlesManifest), nil)

	r := manifests.NewReader(reader)
	_, err := r.ReadEKSD("v0.0.1", "1.22")
	g.Expect(err).To(MatchError(ContainSubstring("kubernetes version 1.22 is not supported by bundles manifest 0")))
}

func TestReaderReadImages(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)

	releasesURL := releases.ManifestURL()

	releasesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: release-1
spec:
  releases:
    - bundleManifestUrl: "https://bundles/bundles.yaml"
      version: v0.0.1`
	reader.EXPECT().ReadFile(releasesURL).Return([]byte(releasesManifest), nil)

	bundlesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Bundles
metadata:
  name: bundles-1
spec:
  versionsBundles:
  - kubeVersion: "1.21"
    eksD:
      channel: 1-21
      kubeVersion: v1.21.5
      manifestUrl: https://distro.eks.amazonaws.com/kubernetes-1-21/kubernetes-1-21-eks-7.yaml`

	reader.EXPECT().ReadFile("https://bundles/bundles.yaml").Return([]byte(bundlesManifest), nil)
	reader.EXPECT().ReadFile("https://distro.eks.amazonaws.com/kubernetes-1-21/kubernetes-1-21-eks-7.yaml").Return([]byte(bundlesManifest), nil)

	r := manifests.NewReader(reader)
	_, err := r.ReadImages("v0.0.1")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestReaderReadCharts(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)

	releasesURL := releases.ManifestURL()

	releasesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: release-1
spec:
  releases:
    - bundleManifestUrl: "https://bundles/bundles.yaml"
      version: v0.0.1`
	reader.EXPECT().ReadFile(releasesURL).Return([]byte(releasesManifest), nil)

	bundlesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Bundles
metadata:
  name: bundles-1`

	reader.EXPECT().ReadFile("https://bundles/bundles.yaml").Return([]byte(bundlesManifest), nil)

	r := manifests.NewReader(reader)
	charts, err := r.ReadCharts("v0.0.1")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(charts).To(BeEmpty())
}
