package registry

import (
	"context"

	orasregistry "oras.land/oras-go/v2/registry"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// StorageClient interface for general image storage client.
type StorageClient interface {
	Init() error
	Copy(ctx context.Context, image releasev1.Image, dstClient StorageClient) error
	GetStorage(ctx context.Context, image releasev1.Image) (repo orasregistry.Repository, err error)
	SetProject(project string)
	Destination(image releasev1.Image) string
}
