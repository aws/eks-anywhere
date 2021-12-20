package internal

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func LoadManifest(clusterSpec *cluster.Spec, manifest releasev1.Manifest) ([]byte, error) {
	m, err := clusterSpec.LoadManifest(manifest)
	if err != nil {
		return nil, fmt.Errorf("can't load networking manifest [%s]: %v", manifest.URI, err)
	}

	return m.Content, nil
}
