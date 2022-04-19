package eksd_test

import (
	"errors"
	"testing"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test/mocks"
	"github.com/aws/eks-anywhere/pkg/eksd"
)

func TestReadManifest(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	url := "url"

	manifest := `apiVersion: distro.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: kubernetes-1-19-eks-4`

	wantRelease := &eksdv1.Release{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "distro.eks.amazonaws.com/v1alpha1",
			Kind:       "Release",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubernetes-1-19-eks-4",
		},
	}

	reader.EXPECT().ReadFile(url).Return([]byte(manifest), nil)

	g.Expect(eksd.ReadManifest(reader, url)).To(Equal(wantRelease))
}

func TestReadManifestErrorReading(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	url := "url"

	reader.EXPECT().ReadFile(url).Return(nil, errors.New("error reading"))

	_, err := eksd.ReadManifest(reader, url)
	g.Expect(err).To(MatchError(ContainSubstring("error reading")))
}

func TestReadManifestErrorUnmarshaling(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	url := "url"

	manifest := `apiVersion: distro.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: {}`

	reader.EXPECT().ReadFile(url).Return([]byte(manifest), nil)

	_, err := eksd.ReadManifest(reader, url)
	g.Expect(err).To(MatchError(ContainSubstring("failed to unmarshal eksd manifest: error unmarshaling JSON:")))
}
