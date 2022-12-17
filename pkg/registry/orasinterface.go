package registry

import (
	"context"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	orasregistry "oras.land/oras-go/v2/registry"
)

// OrasInterface thin layer for oras.
type OrasInterface interface {
	Resolve(ctx context.Context, srcStorage orasregistry.Repository, reference string) (ocispec.Descriptor, error)
	CopyGraph(ctx context.Context, src content.ReadOnlyStorage, dst content.Storage, root ocispec.Descriptor, opts oras.CopyGraphOptions) error
}

// OrasImplementation thin wrapper on oras.
type OrasImplementation struct{}

var _ OrasInterface = (*OrasImplementation)(nil)

// Resolve call oras Resolve.
func (oi *OrasImplementation) Resolve(ctx context.Context, srcStorage orasregistry.Repository, reference string) (ocispec.Descriptor, error) {
	return srcStorage.Resolve(ctx, reference)
}

// CopyGraph call oras CopyGraph.
func (oi *OrasImplementation) CopyGraph(ctx context.Context, src content.ReadOnlyStorage, dst content.Storage, root ocispec.Descriptor, opts oras.CopyGraphOptions) error {
	return oras.CopyGraph(ctx, src, dst, root, opts)
}
