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
	CreateBootstrapCluster(ctx context.Context, name string, opts CreateBootstrapClusterOptions) error

	// DeleteBootstrapCluster deletes a bootstrap cluster identified with the given name.
	DeleteBootstrapCluster(ctx context.Context, name string) error

	// GetKubeconfig retrieves an in-memory kubeconfig for the bootstrap cluster identified by
	// name.
	GetKubeconfig(ctx context.Context, name string) ([]byte, error)

	// ClusterExists determines if a bootstrap cluster identified by name exists.
	ClusterExists(ctx context.Context, name string) (bool, error)
}

// CreateBootstrapClusterOptions defines the options available for bootstrap cluster creation.
type CreateBootstrapClusterOptions struct {
	// Mounts
	// No idea what this does yet.
	Mounts []string

	// Ports defines a list of ports that must be exposed from the cluster.
	Ports []int

	// Swap this out for a proxy configuration, maybe.
	Env map[string]string

	// DefaultCNIDisabled determines if the default CNI should be disabled.
	// This can also do with deleting.
	DefaultCNIDisabled bool
}
