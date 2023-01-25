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
	"oras.land/oras-go/v2/content"
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
	onceFunc := func() {
		or.registry, err = remote.NewRegistry(or.host)
		if err != nil {
			err = fmt.Errorf("error with registry <%s>: %v", or.host, err)
			return
		}

		transport := http.DefaultTransport.(*http.Transport).Clone()
		{ // #nosec G402
			transport.TLSClientConfig = &tls.Config{
				RootCAs:            or.certificates,
				InsecureSkipVerify: or.insecure,
			}
		}
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
	return err
}

// GetHost for registry host.
func (or *OCIRegistryClient) GetHost() string {
	return or.host
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
func (or *OCIRegistryClient) GetStorage(ctx context.Context, artifact Artifact) (repo orasregistry.Repository, err error) {
	dstRepo := path.Join(or.project, artifact.Repository)
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

// FetchBytes a resource from the registry.
func (or *OCIRegistryClient) FetchBytes(ctx context.Context, srcStorage orasregistry.Repository, artifact Artifact) (ocispec.Descriptor, []byte, error) {
	return oras.FetchBytes(ctx, srcStorage, artifact.VersionedImage(), oras.DefaultFetchBytesOptions)
}

// FetchBlob get named blob.
func (or *OCIRegistryClient) FetchBlob(ctx context.Context, srcStorage orasregistry.Repository, descriptor ocispec.Descriptor) ([]byte, error) {
	return content.FetchAll(ctx, srcStorage, descriptor)
}

// CopyGraph copy manifest and all blobs to destination.
func (or *OCIRegistryClient) CopyGraph(ctx context.Context, srcStorage orasregistry.Repository, srcRef string, dstStorage orasregistry.Repository, dstRef string) (ocispec.Descriptor, error) {
	return oras.Copy(ctx, srcStorage, srcRef, dstStorage, dstRef, oras.CopyOptions{})
}

// Tag an image.
func (or *OCIRegistryClient) Tag(ctx context.Context, dstStorage orasregistry.Repository, desc ocispec.Descriptor, tag string) error {
	return dstStorage.Tag(ctx, desc, tag)
}
