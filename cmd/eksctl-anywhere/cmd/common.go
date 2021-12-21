package cmd

import (
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func getImages(spec *cluster.Spec) ([]v1alpha1.Image, error) {
	bundle := spec.VersionsBundle
	images := append(bundle.Images(), spec.KubeDistroImages()...)
	return images, nil
}
