package bootstrap

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

// ClientOptionsRetriever supplies bootstrap cluster options to be used in bootstrap cluster
// creation.
type ClientOptionsRetriever interface {
	GetCreateBootstrapClusterOptions(*cluster.Spec) CreateBootstrapClusterOptions
}

// Client provides behavior for interacting with a bootstrap cluster.
type Client interface {
	// CreateBootstrapCluster creates a bootstrap cluster with the given name and options.
	CreateBootstrapCluster(ctx context.Context, name string) error

	// DeleteBootstrapCluster deletes a bootstrap cluster identified with the given name.
	DeleteBootstrapCluster(ctx context.Context, name string) error

	// GetKubeconfig retrieves an in-memory kubeconfig for the bootstrap cluster identified by
	// name.
	GetKubeconfig(ctx context.Context, name string) ([]byte, error)

	// ClusterExists determines if a bootstrap cluster identified by name exists.
	ClusterExists(ctx context.Context, name string) (bool, error)
}
