package bundles

import (
	"path/filepath"

	"github.com/pkg/errors"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// Manifest holds the data of a manifest referenced in the Bundles.
type Manifest struct {
	Filename string
	Content  []byte
}

// ReadManifest reads the content of a [releasev1.Manifest].
func ReadManifest(reader Reader, manifest releasev1.Manifest) (*Manifest, error) {
	content, err := reader.ReadFile(manifest.URI)
	if err != nil {
		return nil, errors.Errorf("reading manifest %s: %v", manifest.URI, err)
	}

	return &Manifest{
		Filename: filepath.Base(manifest.URI),
		Content:  content,
	}, nil
}
