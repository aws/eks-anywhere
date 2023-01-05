package registry

import (
	"context"

	orasregistry "oras.land/oras-go/v2/registry"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// Artifact to head release dependency.
type Artifact struct {
	releasev1.Image
}

// StorageClient interface for general image storage client.
type StorageClient interface {
	Init() error
	Copy(ctx context.Context, image Artifact, dstClient StorageClient) error
	GetStorage(ctx context.Context, image Artifact) (repo orasregistry.Repository, err error)
	SetProject(project string)
	Destination(image Artifact) string
}
