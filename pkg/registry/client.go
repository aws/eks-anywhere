package registry

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"path"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	orasregistry "oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// OCIRegistryClient storage client for an OCI registry.
type OCIRegistryClient struct {
	StorageContext
	initialized sync.Once
	registry    *remote.Registry
}

var _ StorageClient = (*OCIRegistryClient)(nil)

// NewOCIRegistry create an OCI registry client.
func NewOCIRegistry(context StorageContext) *OCIRegistryClient {
	return &OCIRegistryClient{
		StorageContext: context,
	}
}

// Init registry configuration.
func (or *OCIRegistryClient) Init() error {
	var err error
	or.registry, err = remote.NewRegistry(or.host)
	if err != nil {
		return fmt.Errorf("error with registry <%s>: %v", or.host, err)
	}

	onceFunc := func() {
		tlsConfig := &tls.Config{
			RootCAs:            or.certificates,
			InsecureSkipVerify: or.insecure,
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = tlsConfig
		authClient := &auth.Client{
			Client: &http.Client{
				Transport: transport,
			},
			Cache: auth.NewCache(),
		}
		authClient.SetUserAgent("eksa")
		authClient.Credential = func(ctx context.Context, s string) (auth.Credential, error) {
			return or.credentialStore.Credential(s)
		}
		or.registry.Client = authClient
	}
	or.initialized.Do(onceFunc)
	return nil
}

// GetHost for registry host.
func (or *OCIRegistryClient) GetHost() string {
	return or.host
}

// GetProject for registry project.
func (or *OCIRegistryClient) GetProject() string {
	return or.project
}

// IsInsecure insecure TLS connection.
func (or *OCIRegistryClient) IsInsecure() bool {
	return or.insecure
}

// SetProject for registry destination.
func (or *OCIRegistryClient) SetProject(project string) {
	or.project = project
}

// Destination of this storage registry.
func (or *OCIRegistryClient) Destination(image Artifact) string {
	return path.Join(or.host, or.project, image.Repository) + image.Version()
}

// GetStorage object based on repository.
func (or *OCIRegistryClient) GetStorage(ctx context.Context, image Artifact) (repo orasregistry.Repository, err error) {
	dstRepo := or.project + image.Repository
	repo, err = or.registry.Repository(ctx, dstRepo)
	if err != nil {
		return nil, fmt.Errorf("error creating repository %s: %v", dstRepo, err)
	}
	return repo, nil
}

// Resolve the location of the source repository given the image.
func (or *OCIRegistryClient) Resolve(ctx context.Context, srcStorage orasregistry.Repository, versionedImage string) (desc ocispec.Descriptor, err error) {
	or.registry.Reference.Reference = versionedImage
	return srcStorage.Resolve(ctx, or.registry.Reference.Reference)
}

// CopyGraph copy manifest and all blobs to destination.
func (or *OCIRegistryClient) CopyGraph(ctx context.Context, srcStorage orasregistry.Repository, dstStorage orasregistry.Repository, desc ocispec.Descriptor) error {
	extendedCopyOptions := oras.DefaultExtendedCopyOptions
	return oras.CopyGraph(ctx, srcStorage, dstStorage, desc, extendedCopyOptions.CopyGraphOptions)
}
