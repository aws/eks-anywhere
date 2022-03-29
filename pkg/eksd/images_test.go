package eksd_test

import (
	"testing"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/eksd"
)

func TestImages(t *testing.T) {
	g := NewWithT(t)

	image1 := eksdv1.Asset{Name: "image1", Image: &eksdv1.AssetImage{}}
	image2 := eksdv1.Asset{Name: "image2", Image: &eksdv1.AssetImage{}}
	image3 := eksdv1.Asset{Name: "image3", Image: &eksdv1.AssetImage{}}
	wantImages := []eksdv1.Asset{image1, image2, image3}

	r := &eksdv1.Release{
		Status: eksdv1.ReleaseStatus{
			Components: []eksdv1.Component{
				{
					Assets: []eksdv1.Asset{image1, image2},
				},
				{
					Assets: []eksdv1.Asset{
						image3,
						{Name: "artifact", Archive: &eksdv1.AssetArchive{}},
					},
				},
			},
		},
	}

	g.Expect(eksd.Images(r)).To(Equal(wantImages))
}
