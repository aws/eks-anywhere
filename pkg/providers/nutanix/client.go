package nutanix

import (
	"context"

	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
)

type Client interface {
	GetSubnet(ctx context.Context, uuid string) (*v3.SubnetIntentResponse, error)
	ListAllHost(ctx context.Context) (*v3.HostListResponse, error)
	ListAllSubnet(ctx context.Context, filter string, clientSideFilters []*prismgoclient.AdditionalFilter) (*v3.SubnetListIntentResponse, error)
	GetImage(ctx context.Context, uuid string) (*v3.ImageIntentResponse, error)
	ListAllImage(ctx context.Context, filter string) (*v3.ImageListIntentResponse, error)
	GetCluster(ctx context.Context, uuid string) (*v3.ClusterIntentResponse, error)
	ListAllCluster(ctx context.Context, filter string) (*v3.ClusterListIntentResponse, error)
	GetProject(ctx context.Context, uuid string) (*v3.Project, error)
	ListAllProject(ctx context.Context, filter string) (*v3.ProjectListResponse, error)
	GetCurrentLoggedInUser(ctx context.Context) (*v3.UserIntentResponse, error)
	ListCategories(ctx context.Context, getEntitiesRequest *v3.CategoryListMetadata) (*v3.CategoryKeyListResponse, error)
	GetCategoryKey(ctx context.Context, name string) (*v3.CategoryKeyStatus, error)
	ListCategoryValues(ctx context.Context, name string, getEntitiesRequest *v3.CategoryListMetadata) (*v3.CategoryValueListResponse, error)
	GetCategoryValue(ctx context.Context, name string, value string) (*v3.CategoryValueStatus, error)
	GetCategoryQuery(ctx context.Context, query *v3.CategoryQueryInput) (*v3.CategoryQueryResponse, error)
}
