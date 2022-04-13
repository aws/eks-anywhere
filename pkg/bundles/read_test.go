package bundles_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test/mocks"
	"github.com/aws/eks-anywhere/pkg/bundles"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestRead(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	url := "url"

	manifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
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

	reader.EXPECT().ReadFile(url).Return([]byte(manifest), nil)

	g.Expect(bundles.Read(reader, url)).To(Equal(wantBundles))
}

func TestReadErrorReading(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	url := "url"

	reader.EXPECT().ReadFile(url).Return(nil, errors.New("error reading"))

	_, err := bundles.Read(reader, url)
	g.Expect(err).To(MatchError(ContainSubstring("error reading")))
}

func TestReadErrorUnmarshaling(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	reader := mocks.NewMockReader(ctrl)
	url := "url"

	manifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Bundles
metadata:
  name: {}`

	reader.EXPECT().ReadFile(url).Return([]byte(manifest), nil)

	_, err := bundles.Read(reader, url)
	g.Expect(err).To(MatchError(ContainSubstring("failed to unmarshal bundles manifest from [url]:")))
}
