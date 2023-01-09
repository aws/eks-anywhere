package registry

import (
	"context"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	orasregistry "oras.land/oras-go/v2/registry"
)

// StorageClient interface for general image storage client.
type StorageClient interface {
	Init() error
	Resolve(ctx context.Context, srcStorage orasregistry.Repository, versionedImage string) (desc ocispec.Descriptor, err error)
	GetStorage(ctx context.Context, image Artifact) (repo orasregistry.Repository, err error)
	SetProject(project string)
	Destination(image Artifact) string
}
