package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	orasregistry "oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// OCIRegistryClient storage client for an OCI registry.
type OCIRegistryClient struct {
	host        string
	project     string
	certFile    string
	insecure    bool
	dryRun      bool
	initialized bool
	OI          OrasInterface
	registry    *remote.Registry
}

var _ StorageClient = (*OCIRegistryClient)(nil)

// NewOCIRegistry create an OCI registry client.
func NewOCIRegistry(host string, certFile string, insecure bool) *OCIRegistryClient {
	return &OCIRegistryClient{
		host:     host,
		certFile: certFile,
		insecure: insecure,
		OI:       &OrasImplementation{},
	}
}

// Init registry configuration.
func (or *OCIRegistryClient) Init() error {
	if or.initialized {
		return nil
	}
	or.initialized = true

	credentialStore := NewCredentialStore()
	err := credentialStore.Init()
	if err != nil {
		return err
	}

	certificates, err := or.getCertificates()
	if err != nil {
		return err
	}

	or.registry, err = remote.NewRegistry(or.host)
	if err != nil {
		return fmt.Errorf("NewRegistry: %v", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:            certificates,
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
		return credentialStore.Credential(s)
	}
	or.registry.Client = authClient
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

// GetCertFile for registry certificate file.
func (or *OCIRegistryClient) GetCertFile() string {
	return or.certFile
}

// IsInsecure insecure TLS connection.
func (or *OCIRegistryClient) IsInsecure() bool {
	return or.insecure
}

// SetProject for registry destination.
func (or *OCIRegistryClient) SetProject(project string) {
	or.project = project
}

// SetDryRun dry run validates read, but not write.
func (or *OCIRegistryClient) SetDryRun(value bool) {
	or.dryRun = value
}

// Destination of this storage registry.
func (or *OCIRegistryClient) Destination(image releasev1.Image) string {
	return strings.Replace(image.VersionedImage(), image.Registry()+"/", or.host+"/"+or.project, 1)
}

// GetReference gets digest or tag version.
func (or *OCIRegistryClient) getCertificates() (certificates *x509.CertPool, err error) {
	if len(or.certFile) < 1 {
		return nil, nil
	}
	fileContents, err := ioutil.ReadFile(or.certFile)
	if err != nil {
		return nil, fmt.Errorf("error reading certificate file <%s>: %v", or.certFile, err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(fileContents)

	return certPool, nil
}

// GetStorage object based on repository.
func (or *OCIRegistryClient) GetStorage(ctx context.Context, image releasev1.Image) (repo orasregistry.Repository, err error) {
	dstRepo := or.project + image.Repository()
	repo, err = or.OI.Repository(ctx, or.registry, dstRepo)
	if err != nil {
		return nil, fmt.Errorf("error creating repository %s: %v", dstRepo, err)
	}
	return repo, nil
}

// Copy a image from a source to a destination.
func (or *OCIRegistryClient) Copy(ctx context.Context, image releasev1.Image, dstClient StorageClient) (err error) {
	committed := &sync.Map{}
	extendedCopyOptions := oras.DefaultExtendedCopyOptions
	extendedCopyOptions.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		committed.Store(desc.Digest.String(), desc.Annotations[ocispec.AnnotationTitle])
		return nil
	}
	extendedCopyOptions.OnCopySkipped = func(ctx context.Context, desc ocispec.Descriptor) error {
		committed.Store(desc.Digest.String(), desc.Annotations[ocispec.AnnotationTitle])
		return nil
	}

	srcStorage, err := or.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("registry copy source: %v", err)
	}

	var desc ocispec.Descriptor
	or.registry.Reference.Reference = image.Version()
	desc, err = or.OI.Resolve(ctx, srcStorage, or.registry.Reference.Reference)
	if err != nil {
		return fmt.Errorf("registry copy destination: %v", err)
	}

	dstStorage, err := dstClient.GetStorage(ctx, image)
	if err != nil {
		return fmt.Errorf("registry copy destination: %v", err)
	}

	fmt.Println(dstClient.Destination(image))
	if or.dryRun {
		return nil
	}
	err = or.OI.CopyGraph(ctx, srcStorage, dstStorage, desc, extendedCopyOptions.CopyGraphOptions)
	if err != nil {
		return fmt.Errorf("registry copy: %v", err)
	}

	return nil
}
