package cluster

import (
	"fmt"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type ManifestReader struct {
	*files.Reader
}

func NewManifestReader(opts ...files.ReaderOpt) *ManifestReader {
	return &ManifestReader{files.NewReader(opts...)}
}

func (m *ManifestReader) GetReleases(releasesManifest string) (*v1alpha1.Release, error) {
	logger.V(4).Info("Reading releases manifest", "url", releasesManifest)
	content, err := m.ReadFile(releasesManifest)
	if err != nil {
		return nil, err
	}

	releases := &v1alpha1.Release{}
	if err = yaml.Unmarshal(content, releases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the release manifest: %v", err)
	}

	return releases, nil
}

func (m *ManifestReader) GetEksdRelease(versionsBundle *v1alpha1.VersionsBundle) (*eksdv1alpha1.Release, error) {
	content, err := m.ReadFile(versionsBundle.EksD.EksDReleaseUrl)
	if err != nil {
		return nil, err
	}

	eksd := &eksdv1alpha1.Release{}
	if err = yaml.Unmarshal(content, eksd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal eksd manifest to build cluster spec: %v", err)
	}

	return eksd, nil
}

func (m *ManifestReader) GetBundles(bundlesURL string) (*v1alpha1.Bundles, error) {
	logger.V(4).Info("Reading bundles manifest", "url", bundlesURL)
	content, err := m.ReadFile(bundlesURL)
	if err != nil {
		return nil, err
	}

	bundles := &v1alpha1.Bundles{}
	if err = yaml.Unmarshal(content, bundles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundles manifest from [%s] to build cluster spec: %v", bundlesURL, err)
	}

	return bundles, nil
}
