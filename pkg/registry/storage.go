package registry

import (
	"context"
	"crypto/x509"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orasregistry "oras.land/oras-go/v2/registry"
)

// StorageContext describes aspects of a registry.
type StorageContext struct {
	host            string
	project         string
	credentialStore CredentialStore
	certificates    *x509.CertPool
	insecure        bool
}

// NewStorageContext create registry context.
func NewStorageContext(host string, credentialStore CredentialStore, certificates *x509.CertPool, insecure bool) StorageContext {
	return StorageContext{
		host:            host,
		credentialStore: credentialStore,
		certificates:    certificates,
		insecure:        insecure,
	}
}

// StorageClient interface for general image storage client.
type StorageClient interface {
	Init() error
	Resolve(ctx context.Context, srcStorage orasregistry.Repository, versionedImage string) (desc ocispec.Descriptor, err error)
	GetStorage(ctx context.Context, image Artifact) (repo orasregistry.Repository, err error)
	SetProject(project string)
	Destination(image Artifact) string
	PullBytes(ctx context.Context, artifact Artifact) (data []byte, err error)
	CopyGraph(ctx context.Context, srcStorage orasregistry.Repository, dstStorage orasregistry.Repository, desc ocispec.Descriptor) error
}
