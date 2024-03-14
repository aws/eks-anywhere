package clustermanager

import (
	"context"
	"io"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

// CreateClusterShim is a shim that implements the workload.Cluster interface. It leverages existing
// ClusterManager behavior to create a cluster for new workflows.
type CreateClusterShim struct {
	spec     *cluster.Spec
	provider providers.Provider
}

// NewCreateClusterShim returns a new CreateClusterShim instance.
func NewCreateClusterShim(
	spec *cluster.Spec,
	provider providers.Provider,
) *CreateClusterShim {
	return &CreateClusterShim{
		spec: spec,
	}
}

// CreateAsync satisfies the workload.Cluster interface.
func (s CreateClusterShim) CreateAsync(ctx context.Context, management *types.Cluster) error {
	// TODO: implement reusing the apply logic from clustermanager.Applier
	return nil
}

// GetName satisfies the workload.Cluster interface.
func (s CreateClusterShim) GetName() string {
	return s.spec.Cluster.Name
}

// WriteKubeconfig satisfies the workload.Cluster interface.
func (s CreateClusterShim) WriteKubeconfig(ctx context.Context, w io.Writer, management *types.Cluster) error {
	// TODO: reuse logic from clustermanager.ClusterCreator
	return nil
}

// WaitUntilControlPlaneAvailable satisfies the workload.Cluster interface.
func (s CreateClusterShim) WaitUntilControlPlaneAvailable(ctx context.Context, management *types.Cluster) error {
	// TODO: implement reusing the wait logic from clustermanager.Applier
	return nil
}

// WaitUntilReady satisfies the workload.Cluster interface.
func (s CreateClusterShim) WaitUntilReady(ctx context.Context, management *types.Cluster) error {
	// TODO: implement reusing the wait logic from clustermanager.Applier
	return nil
}
