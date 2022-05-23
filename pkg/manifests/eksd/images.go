package eksd

import (
	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
)

func Images(release *eksdv1.Release) []eksdv1.Asset {
	images := []eksdv1.Asset{}
	for _, component := range release.Status.Components {
		for _, asset := range component.Assets {
			if asset.Image != nil {
				images = append(images, asset)
			}
		}
	}
	return images
}
