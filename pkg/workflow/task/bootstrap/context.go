package bootstrap

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/types"
)

// contextKey is used to create collisionless context keys.
type contextKey string

func (c contextKey) String() string {
	return string(c)
}

const clusterKey contextKey = "bootstrap-cluster"

// WithCluster returns a context with a bootstrap cluster set.
func WithCluster(ctx context.Context, cluster *types.Cluster) context.Context {
	return context.WithValue(ctx, clusterKey, cluster)
}

// FromContext retrieves the bootstrap cluster from the context that was previously set
// with WithCluster. It returns nil if no cluster was configured in ctx.
func FromContext(ctx context.Context) *types.Cluster {
	return ctx.Value(clusterKey).(*types.Cluster)
}
