package bundles

import (
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/aws/eks-anywhere/pkg/manifests/eksd"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// ReadImages returns a list of all images included in the Bundles and the referenced
// EKS-D Releases, all of them filtered by kubernetes version. If not kubernetes versions
// all provided, all images are returned.
func ReadImages(reader Reader, bundles *releasev1.Bundles, kubeVersions ...string) ([]releasev1.Image, error) {
	var images []releasev1.Image
	for _, v := range bundles.Spec.VersionsBundles {
		if len(kubeVersions) > 0 && !slices.Contains(kubeVersions, v.KubeVersion) {
			continue
		}

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
