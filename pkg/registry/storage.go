package registry

import (
	"context"

	orasregistry "oras.land/oras-go/v2/registry"
)

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

// VersionedImage returns full URI for image.
func (art *Artifact) VersionedImage() string {
	var version string
	if art.Tag != "" {
		version = ":" + art.Tag
	} else {
		version = "@" + art.Digest
	}
	return art.Registry + "/" + art.Repository + version
}

// StorageClient interface for general image storage client.
type StorageClient interface {
	Init() error
	Copy(ctx context.Context, image Artifact, dstClient StorageClient) error
	GetStorage(ctx context.Context, image Artifact) (repo orasregistry.Repository, err error)
	SetProject(project string)
	Destination(image Artifact) string
}
