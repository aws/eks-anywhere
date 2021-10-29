package clusterapi

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/types"
)

type ResourceSetManager struct {
	client Client
}

type Client interface{}

func (r *ResourceSetManager) ForceUpdate(ctx context.Context, name, namespace string, cluster *types.Cluster) error {
	return nil
}
