package nutanix

import (
	"context"

	"github.com/nutanix-cloud-native/prism-go-client/converged"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
	clusterModels "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/config"
	subnetModels "github.com/nutanix/ntnx-api-golang-clients/networking-go-client/v4/models/networking/v4/config"
	prismModels "github.com/nutanix/ntnx-api-golang-clients/prism-go-client/v4/models/prism/v4/config"
	imageModels "github.com/nutanix/ntnx-api-golang-clients/vmm-go-client/v4/models/vmm/v4/content"
)

// Client defines the interface for interacting with Nutanix Prism Central.
// Most methods use the v4 converged API; project methods still use v3.
type Client interface {
	GetSubnet(ctx context.Context, uuid string) (*subnetModels.Subnet, error)
	ListSubnets(ctx context.Context, opts ...converged.ODataOption) ([]subnetModels.Subnet, error)
	GetImage(ctx context.Context, uuid string) (*imageModels.Image, error)
	ListImages(ctx context.Context, opts ...converged.ODataOption) ([]imageModels.Image, error)
	GetCluster(ctx context.Context, uuid string) (*clusterModels.Cluster, error)
	ListClusters(ctx context.Context, opts ...converged.ODataOption) ([]clusterModels.Cluster, error)
	ListAllHosts(ctx context.Context, opts ...converged.ODataOption) ([]clusterModels.Host, error)
	ListClusterPhysicalGPUs(ctx context.Context, clusterUUID string, opts ...converged.ODataOption) ([]clusterModels.PhysicalGpuProfile, error)
	ListClusterVirtualGPUs(ctx context.Context, clusterUUID string, opts ...converged.ODataOption) ([]clusterModels.VirtualGpuProfile, error)
	GetCategory(ctx context.Context, uuid string) (*prismModels.Category, error)
	ListCategories(ctx context.Context, opts ...converged.ODataOption) ([]prismModels.Category, error)

	// v3-only: no v4 equivalent available yet
	ListAllProject(ctx context.Context, filter string) (*v3.ProjectListResponse, error)
}
