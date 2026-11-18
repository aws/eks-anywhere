package nutanix

import (
	"context"
	"fmt"
	"strings"

	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	"github.com/nutanix-cloud-native/prism-go-client/converged"
	v4converged "github.com/nutanix-cloud-native/prism-go-client/converged/v4"
	clusterModels "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/models/clustermgmt/v4/config"

	"github.com/aws/eks-anywhere/pkg/providers/nutanix"
)

// PrismClient is a interface that provides useful functions for performing Prism operations.
type PrismClient interface {
	GetImageUUIDFromName(ctx context.Context, imageName string) (*string, error)
	GetClusterUUIDFromName(ctx context.Context, clusterName string) (*string, error)
	GetSubnetUUIDFromName(ctx context.Context, subnetName string) (*string, error)
}

type client struct {
	convergedClient *v4converged.Client
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
	pclient, err := v4converged.NewClient(nutanixCreds)
	if err != nil {
		return nil, err
	}

	return &client{convergedClient: pclient}, nil
}

// GetImageUUIDFromName retrieves the image uuid from the given image name.
func (c *client) GetImageUUIDFromName(ctx context.Context, imageName string) (*string, error) {
	images, err := c.convergedClient.Images.List(ctx, converged.WithFilter(fmt.Sprintf("name eq '%s'", imageName)))
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	var matched []*string
	for _, image := range images {
		if image.Name != nil && strings.EqualFold(*image.Name, imageName) {
			matched = append(matched, image.ExtId)
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("failed to find image by name %q", imageName)
	}
	if len(matched) > 1 {
		return nil, fmt.Errorf("found more than one (%v) image with name %q", len(matched), imageName)
	}
	return matched[0], nil
}

// hasPEClusterServiceEnabled checks if a cluster has the AOS (PE) cluster function.
func hasPEClusterServiceEnabled(peCluster *clusterModels.Cluster) bool {
	if peCluster.Config == nil || peCluster.Config.ClusterFunction == nil {
		return false
	}
	for _, s := range peCluster.Config.ClusterFunction {
		if strings.ToUpper(string(s.GetName())) == clusterModels.CLUSTERFUNCTIONREF_AOS.GetName() {
			return true
		}
	}
	return false
}

// GetClusterUUIDFromName retrieves the cluster uuid from the given cluster name.
func (c *client) GetClusterUUIDFromName(ctx context.Context, clusterName string) (*string, error) {
	clusters, err := c.convergedClient.Clusters.List(ctx, converged.WithFilter(fmt.Sprintf("name eq '%s'", clusterName)))
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %v", err)
	}

	entities := make([]clusterModels.Cluster, 0)
	for _, entity := range clusters {
		if strings.EqualFold(*entity.Name, clusterName) && hasPEClusterServiceEnabled(&entity) {
			entities = append(entities, entity)
		}
	}

	if len(entities) == 0 {
		return nil, fmt.Errorf("failed to find cluster by name %q", clusterName)
	}
	if len(entities) > 1 {
		return nil, fmt.Errorf("found more than one (%v) cluster with name %q", len(entities), clusterName)
	}
	return entities[0].ExtId, nil
}

// GetSubnetUUIDFromName retrieves the subnet uuid from the given subnet name.
func (c *client) GetSubnetUUIDFromName(ctx context.Context, subnetName string) (*string, error) {
	subnets, err := c.convergedClient.Subnets.List(ctx, converged.WithFilter(fmt.Sprintf("name eq '%s'", subnetName)))
	if err != nil {
		return nil, fmt.Errorf("failed to list subnets: %v", err)
	}

	var matched []*string
	for _, subnet := range subnets {
		if subnet.Name != nil && strings.EqualFold(*subnet.Name, subnetName) {
			matched = append(matched, subnet.ExtId)
		}
	}

	if len(matched) == 0 {
		return nil, fmt.Errorf("failed to find subnet by name %q", subnetName)
	}
	if len(matched) > 1 {
		return nil, fmt.Errorf("found more than one (%v) subnet with name %q", len(matched), subnetName)
	}
	return matched[0], nil
}
