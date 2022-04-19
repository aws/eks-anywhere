package bundles

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/eksd"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func ReadImages(reader Reader, bundles *releasev1.Bundles) ([]releasev1.Image, error) {
	var images []releasev1.Image
	for _, v := range bundles.Spec.VersionsBundles {
		images = append(images, v.Images()...)

		eksdRelease, err := ReadEKSD(reader, v)
		if err != nil {
			return nil, fmt.Errorf("reading images from Bundle: %v", err)
		}

		for _, i := range eksd.Images(eksdRelease) {
			images = append(images, releasev1.Image{
				Name:        i.Name,
				Description: i.Description,
				ImageDigest: i.Image.ImageDigest,
				URI:         i.Image.URI,
				OS:          i.OS,
				Arch:        i.Arch,
			})
		}
	}

	return images, nil
}
