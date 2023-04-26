package nutanix

import (
	"context"
	"fmt"
	"strings"

	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	"github.com/nutanix-cloud-native/prism-go-client/utils"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"

	"github.com/aws/eks-anywhere/pkg/providers/nutanix"
)

// PrismClient is a interface that provides useful functions for performing Prism operations.
type PrismClient interface {
	GetImageUUIDFromName(ctx context.Context, imageName string) (*string, error)
	GetClusterUUIDFromName(ctx context.Context, clusterName string) (*string, error)
	GetSubnetUUIDFromName(ctx context.Context, subnetName string) (*string, error)
}

type client struct {
	v3.Client
}

// NewPrismClient returns an implementation of the PrismClient interface.
func NewPrismClient(endpoint, port string, insecure bool) (PrismClient, error) {
	creds := nutanix.GetCredsFromEnv()
	nutanixCreds := prismgoclient.Credentials{
		URL:      fmt.Sprintf("%s:%s", endpoint, port),
		Username: creds.PrismCentral.Username,
		Password: creds.PrismCentral.Password,
		Endpoint: endpoint,
		Port:     port,
		Insecure: insecure,
	}
	pclient, err := v3.NewV3Client(nutanixCreds)
	if err != nil {
		return nil, err
	}

	return &client{*pclient}, nil
}

// GetImageUUIDFromName retrieves the image uuid from the given image name.
func (c *client) GetImageUUIDFromName(ctx context.Context, imageName string) (*string, error) {
	res, err := c.V3.ListImage(ctx, &v3.DSMetadata{
		Filter: utils.StringPtr(fmt.Sprintf("name==%s", imageName)),
	})
	if err != nil || len(res.Entities) == 0 {
		return nil, fmt.Errorf("failed to find image by name %q: %v", imageName, err)
	}

	if len(res.Entities) > 1 {
		return nil, fmt.Errorf("found more than one (%v) image with name %q", len(res.Entities), imageName)
	}

	return res.Entities[0].Metadata.UUID, nil
}

// GetClusterUUIDFromName retrieves the cluster uuid from the given cluster name.
//
//nolint:gocyclo
func (c *client) GetClusterUUIDFromName(ctx context.Context, clusterName string) (*string, error) {
	res, err := c.V3.ListCluster(ctx, &v3.DSMetadata{
		Filter: utils.StringPtr(fmt.Sprintf("name==%s", clusterName)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find cluster by name %q: %v", clusterName, err)
	}
	entities := make([]*v3.ClusterIntentResponse, 0)
	for _, entity := range res.Entities {
		if entity.Status != nil && entity.Status.Resources != nil && entity.Status.Resources.Config != nil {
			serviceList := entity.Status.Resources.Config.ServiceList
			isPrismCentral := false
			for _, svc := range serviceList {
				// Prism Central is also internally a cluster, but we filter that out here as we only care about prism element clusters
				if svc != nil && strings.ToUpper(*svc) == "PRISM_CENTRAL" {
					isPrismCentral = true
				}
			}
			if !isPrismCentral && *entity.Spec.Name == clusterName {
				entities = append(entities, entity)
			}
		}
	}
	if len(entities) == 0 {
		return nil, fmt.Errorf("failed to find cluster by name %q: %v", clusterName, err)
	}

	if len(entities) > 1 {
		return nil, fmt.Errorf("found more than one (%v) cluster with name %q", len(entities), clusterName)
	}

	return entities[0].Metadata.UUID, nil
}

// GetSubnetUUIDFromName retrieves the subnet uuid from the given subnet name.
func (c *client) GetSubnetUUIDFromName(ctx context.Context, subnetName string) (*string, error) {
	res, err := c.V3.ListSubnet(ctx, &v3.DSMetadata{
		Filter: utils.StringPtr(fmt.Sprintf("name==%s", subnetName)),
	})
	if err != nil || len(res.Entities) == 0 {
		return nil, fmt.Errorf("failed to find subnet by name %q: %v", subnetName, err)
	}

	if len(res.Entities) > 1 {
		return nil, fmt.Errorf("found more than one (%v) subnet with name %q", len(res.Entities), subnetName)
	}

	return res.Entities[0].Metadata.UUID, nil
}
