package nutanix

import (
	"context"

	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
)

type Client interface {
	GetSubnet(ctx context.Context, uuid string) (*v3.SubnetIntentResponse, error)
	ListSubnet(ctx context.Context, getEntitiesRequest *v3.DSMetadata) (*v3.SubnetListIntentResponse, error)

	GetImage(ctx context.Context, uuid string) (*v3.ImageIntentResponse, error)
	ListImage(ctx context.Context, getEntitiesRequest *v3.DSMetadata) (*v3.ImageListIntentResponse, error)

	GetCluster(ctx context.Context, uuid string) (*v3.ClusterIntentResponse, error)
	ListCluster(ctx context.Context, getEntitiesRequest *v3.DSMetadata) (*v3.ClusterListIntentResponse, error)

	GetCurrentLoggedInUser(ctx context.Context) (*v3.UserIntentResponse, error)
}
