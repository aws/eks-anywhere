package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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

// RegistryContext describes aspects of a registry.
type RegistryContext struct {
	host            string
	project         string
	credentialStore CredentialStore
	certificates    *x509.CertPool
	insecure        bool
}

// NewRegistryContext create registry context.
func NewRegistryContext(host string, credentialStore CredentialStore, certificates *x509.CertPool, insecure bool) RegistryContext {
	return RegistryContext{
		host:            host,
		credentialStore: credentialStore,
		certificates:    certificates,
		insecure:        insecure,
	}
}

// OCIRegistryClient storage client for an OCI registry.
type OCIRegistryClient struct {
	RegistryContext
	initialized sync.Once
	registry    *remote.Registry
}

var _ StorageClient = (*OCIRegistryClient)(nil)

// NewOCIRegistry create an OCI registry client.
func NewOCIRegistry(context RegistryContext) *OCIRegistryClient {
	return &OCIRegistryClient{
		RegistryContext: context,
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

// Copy an image from a source to a destination.
func (or *OCIRegistryClient) Copy(ctx context.Context, image Artifact, dstClient StorageClient) (err error) {
	srcStorage, err := or.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("registry copy source: %v", err)
	}

	dstStorage, err := dstClient.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("registry copy destination: %v", err)
	}

	var desc ocispec.Descriptor
	or.registry.Reference.Reference = image.VersionedImage()
	desc, err = srcStorage.Resolve(ctx, or.registry.Reference.Reference)
	if err != nil {
		return fmt.Errorf("registry source resolve: %v", err)
	}

	extendedCopyOptions := oras.DefaultExtendedCopyOptions
	err = oras.CopyGraph(ctx, srcStorage, dstStorage, desc, extendedCopyOptions.CopyGraphOptions)
	if err != nil {
		return fmt.Errorf("registry copy: %v", err)
	}

	return nil
}
