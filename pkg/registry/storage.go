package registry

import (
	"context"

	orasregistry "oras.land/oras-go/v2/registry"
)

// StorageClient interface for general image storage client.
type StorageClient interface {
	Init() error
	Copy(ctx context.Context, image Artifact, dstClient StorageClient) error
	GetStorage(ctx context.Context, image Artifact) (repo orasregistry.Repository, err error)
	SetProject(project string)
	Destination(image Artifact) string
}
