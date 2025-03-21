package bundles

import (
	"fmt"

	eksdv1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests/eksd"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type Reader interface {
	ReadFile(url string) ([]byte, error)
}

func Read(reader Reader, url string) (*releasev1.Bundles, error) {
	logger.V(4).Info("Reading bundles manifest", "url", url)
	content, err := reader.ReadFile(url)
	if err != nil {
		return nil, err
	}

	bundles := &releasev1.Bundles{}
	if err = yaml.Unmarshal(content, bundles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundles manifest from [%s]: %v", url, err)
	}

	return bundles, nil
}

func ReadEKSD(reader Reader, versionsBundle releasev1.VersionsBundle) (*eksdv1.Release, error) {
	return eksd.ReadManifest(reader, versionsBundle.EksD.EksDReleaseURL)
}
