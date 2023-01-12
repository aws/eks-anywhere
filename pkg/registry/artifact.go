package registry

import releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"

// Artifact to head release dependency.
type Artifact struct {
	Registry   string
	Repository string
	Tag        string
	Digest     string
}

// NewArtifact creates a new artifact object.
func NewArtifact(registry, repository, tag, digest string) Artifact {
	return Artifact{
		Registry:   registry,
		Repository: repository,
		Tag:        tag,
		Digest:     digest,
	}
}

// NewArtifactFromURI creates a new artifact object from a URI.
func NewArtifactFromURI(uri string) Artifact {
	image := releasev1.Image{
		URI: uri,
	}
	return Artifact{
		Registry:   image.Registry(),
		Repository: image.Repository(),
		Tag:        image.Tag(),
		Digest:     image.Digest(),
	}
}

// Version returns tag or digest.
func (art *Artifact) Version() string {
	if art.Digest != "" {
		return "@" + art.Digest
	}
	return ":" + art.Tag
}

// VersionedImage returns full URI for image.
func (art *Artifact) VersionedImage() string {
	version := art.Version()
	return art.Registry + "/" + art.Repository + version
}
