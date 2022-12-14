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
	Host        string
	Project     string
	CertFile    string
	Insecure    bool
	initialized bool
	registry    *remote.Registry
}

var _ StorageClient = (*OCIRegistryClient)(nil)

// NewOCIRegistry create an OCI registry client.
func NewOCIRegistry(host string, certFile string, insecure bool) *OCIRegistryClient {
	return &OCIRegistryClient{
		Host:     host,
		CertFile: certFile,
		Insecure: insecure,
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

	or.registry, err = remote.NewRegistry(or.Host)
	if err != nil {
		return fmt.Errorf("NewRegistry: %v", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:            certificates,
		InsecureSkipVerify: or.Insecure,
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
		return credentialStore.Credential(ctx, s)
	}
	or.registry.Client = authClient
	return nil
}

// SetProject for registry destination.
func (or *OCIRegistryClient) SetProject(project string) {
	or.Project = project
}

// Destination of this storage registry.
func (or *OCIRegistryClient) Destination(image releasev1.Image) string {
	return strings.Replace(image.VersionedImage(), image.Registry()+"/", or.Host+"/"+or.Project, 1)
}

// GetReference gets digest or tag version.
func (or *OCIRegistryClient) getCertificates() (certificates *x509.CertPool, err error) {
	if len(or.CertFile) < 1 {
		return nil, nil
	}
	fileContents, err := ioutil.ReadFile(or.CertFile)
	if err != nil {
		return nil, fmt.Errorf("error reading certificate file <%s>: %v", or.CertFile, err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(fileContents)

	return certPool, nil
}

// GetStorage object based on repository.
func (or *OCIRegistryClient) GetStorage(ctx context.Context, image releasev1.Image) (repo orasregistry.Repository, err error) {
	dstRepo := or.Project + image.Repository()
	repo, err = or.registry.Repository(ctx, dstRepo)
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
		return err
	}

	var desc ocispec.Descriptor
	or.registry.Reference.Reference = image.Version()
	desc, err = srcStorage.Resolve(ctx, or.registry.Reference.Reference)
	if err != nil {
		return err
	}

	dstStorage, err := dstClient.GetStorage(ctx, image)
	if err != nil {
		return err
	}

	fmt.Println(dstClient.Destination(image))
	err = oras.CopyGraph(ctx, srcStorage, dstStorage, desc, extendedCopyOptions.CopyGraphOptions)
	if err != nil {
		return fmt.Errorf("copyGraph: %v", err)
	}

	return nil
}
