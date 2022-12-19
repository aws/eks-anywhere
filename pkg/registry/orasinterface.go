package registry

import (
	"context"
	"oras.land/oras-go/v2/registry/remote"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	orasregistry "oras.land/oras-go/v2/registry"
)

// OrasInterface thin layer for oras.
type OrasInterface interface {
	Repository(ctx context.Context, reg *remote.Registry, name string) (orasregistry.Repository, error)
	Resolve(ctx context.Context, srcStorage orasregistry.Repository, reference string) (ocispec.Descriptor, error)
	CopyGraph(ctx context.Context, src content.ReadOnlyStorage, dst content.Storage, root ocispec.Descriptor, opts oras.CopyGraphOptions) error
}

// OrasImplementation thin wrapper on oras.
type OrasImplementation struct{}

var _ OrasInterface = (*OrasImplementation)(nil)

// Repository for the given registry.
func (oi *OrasImplementation) Repository(ctx context.Context, reg *remote.Registry, name string) (orasregistry.Repository, error) {
	result, err := reg.Repository(ctx, name)
	return result, err
}

// Resolve call oras Resolve.
func (oi *OrasImplementation) Resolve(ctx context.Context, srcStorage orasregistry.Repository, reference string) (ocispec.Descriptor, error) {
	return srcStorage.Resolve(ctx, reference)
}

// CopyGraph call oras CopyGraph.
func (oi *OrasImplementation) CopyGraph(ctx context.Context, src content.ReadOnlyStorage, dst content.Storage, root ocispec.Descriptor, opts oras.CopyGraphOptions) error {
	return oras.CopyGraph(ctx, src, dst, root, opts)
}
