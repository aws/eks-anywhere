package nutanix

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"strconv"

	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	"github.com/nutanix-cloud-native/prism-go-client/converged"
	v4converged "github.com/nutanix-cloud-native/prism-go-client/converged/v4"
	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	envTypes "github.com/nutanix-cloud-native/prism-go-client/environment/types"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"

	clusterModels "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/config"
	subnetModels "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/config"
	prismModels "github.com/nutanix/ntnx-api-golang-clients/prism-go-client/v4/models/prism/v4/config"
	imageModels "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/content"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// v3Opt is a convenience alias for the v3 client option type.
type v3Opt = envTypes.ClientOption[v3.Client]

// ClientCache is a map of NutanixDatacenterConfig name to Nutanix client.
type ClientCache struct {
	clients map[string]Client
}

// NewClientCache returns a new ClientCache.
func NewClientCache() *ClientCache {
	return &ClientCache{
		clients: make(map[string]Client),
	}
}

// GetNutanixClient returns a Nutanix client for the given NutanixDatacenterConfig.
func (cb *ClientCache) GetNutanixClient(datacenterConfig *anywherev1.NutanixDatacenterConfig, creds credentials.BasicAuthCredential) (Client, error) {
	if client, ok := cb.clients[datacenterConfig.Name]; ok {
		return client, nil
	}

	if datacenterConfig.Spec.AdditionalTrustBundle != "" {
		block, _ := pem.Decode([]byte(datacenterConfig.Spec.AdditionalTrustBundle))
		certs, err := x509.ParseCertificates(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("unable to parse additional trust bundle %s: %v", datacenterConfig.Spec.AdditionalTrustBundle, err)
		}
		if len(certs) == 0 {
			return nil, fmt.Errorf("unable to extract certs from the addtional trust bundle %s", datacenterConfig.Spec.AdditionalTrustBundle)
		}
	}

	endpoint := datacenterConfig.Spec.Endpoint
	port := datacenterConfig.Spec.Port
	url := net.JoinHostPort(endpoint, strconv.Itoa(port))
	insecure := datacenterConfig.Spec.Insecure

	// v4 SDK doesn't support custom trust bundles yet. Match CAPX behaviour:
	// force Insecure=true when an AdditionalTrustBundle is provided so the
	// converged v4 client does not reject connections signed by private CAs.
	if datacenterConfig.Spec.AdditionalTrustBundle != "" {
		insecure = true
	}

	nutanixCreds := prismgoclient.Credentials{
		URL:      url,
		Username: creds.PrismCentral.Username,
		Password: creds.PrismCentral.Password,
		Endpoint: endpoint,
		Port:     fmt.Sprintf("%d", port),
		Insecure: insecure,
	}

	convergedClient, err := v4converged.NewClient(nutanixCreds)
	if err != nil {
		return nil, fmt.Errorf("error creating nutanix converged client: %v", err)
	}

	// v3 client for projects: pass the trust bundle via WithPEMEncodedCertBundle
	// so TLS verification works with private CAs.
	v3Opts := make([]v3Opt, 0)
	if datacenterConfig.Spec.AdditionalTrustBundle != "" {
		v3Opts = append(v3Opts, v3.WithPEMEncodedCertBundle([]byte(datacenterConfig.Spec.AdditionalTrustBundle)))
	}
	v3Client, err := v3.NewV3Client(nutanixCreds, v3Opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating nutanix v3 client: %v", err)
	}

	client := &convergedClientAdapter{client: convergedClient, v3Client: v3Client}
	cb.clients[datacenterConfig.Name] = client
	return client, nil
}

// convergedClientAdapter adapts the converged v4 client (and v3 for projects) to the Client interface.
type convergedClientAdapter struct {
	client   *v4converged.Client
	v3Client *v3.Client
}

func (a *convergedClientAdapter) GetSubnet(ctx context.Context, uuid string) (*subnetModels.Subnet, error) {
	return a.client.Subnets.Get(ctx, uuid)
}

func (a *convergedClientAdapter) ListSubnets(ctx context.Context, opts ...converged.ODataOption) ([]subnetModels.Subnet, error) {
	return a.client.Subnets.List(ctx, opts...)
}

func (a *convergedClientAdapter) GetImage(ctx context.Context, uuid string) (*imageModels.Image, error) {
	return a.client.Images.Get(ctx, uuid)
}

func (a *convergedClientAdapter) ListImages(ctx context.Context, opts ...converged.ODataOption) ([]imageModels.Image, error) {
	return a.client.Images.List(ctx, opts...)
}

func (a *convergedClientAdapter) GetCluster(ctx context.Context, uuid string) (*clusterModels.Cluster, error) {
	return a.client.Clusters.Get(ctx, uuid)
}

func (a *convergedClientAdapter) ListClusters(ctx context.Context, opts ...converged.ODataOption) ([]clusterModels.Cluster, error) {
	return a.client.Clusters.List(ctx, opts...)
}

func (a *convergedClientAdapter) ListAllHosts(ctx context.Context, opts ...converged.ODataOption) ([]clusterModels.Host, error) {
	return a.client.Clusters.ListAllHosts(ctx, opts...)
}

func (a *convergedClientAdapter) ListClusterPhysicalGPUs(ctx context.Context, clusterUUID string, opts ...converged.ODataOption) ([]clusterModels.PhysicalGpuProfile, error) {
	return a.client.Clusters.ListClusterPhysicalGPUs(ctx, clusterUUID, opts...)
}

func (a *convergedClientAdapter) ListClusterVirtualGPUs(ctx context.Context, clusterUUID string, opts ...converged.ODataOption) ([]clusterModels.VirtualGpuProfile, error) {
	return a.client.Clusters.ListClusterVirtualGPUs(ctx, clusterUUID, opts...)
}

func (a *convergedClientAdapter) GetCategory(ctx context.Context, uuid string) (*prismModels.Category, error) {
	return a.client.Categories.Get(ctx, uuid)
}

func (a *convergedClientAdapter) ListCategories(ctx context.Context, opts ...converged.ODataOption) ([]prismModels.Category, error) {
	return a.client.Categories.List(ctx, opts...)
}

func (a *convergedClientAdapter) ListAllProject(ctx context.Context, filter string) (*v3.ProjectListResponse, error) {
	return a.v3Client.V3.ListAllProject(ctx, filter)
}
