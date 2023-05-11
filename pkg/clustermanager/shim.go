package clustermanager

import (
	"context"
	"fmt"
	"io"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/types"
)

// CreateClusterShim is a shim that implements the workload.Cluster interface. It leverages existing
// ClusterManager behavior to create a cluster for new workflows.
type CreateClusterShim struct {
	spec     *cluster.Spec
	manager  *ClusterManager
	provider providers.Provider
}

// NewCreateClusterShim returns a new CreateClusterShim instance.
func NewCreateClusterShim(
	spec *cluster.Spec,
	manager *ClusterManager,
	provider providers.Provider,
) *CreateClusterShim {
	return &CreateClusterShim{
		spec:    spec,
		manager: manager,
	}
}

// CreateAsync satisfies the workload.Cluster interface.
func (s CreateClusterShim) CreateAsync(ctx context.Context, management *types.Cluster) error {
	if err := s.manager.applyProviderManifests(ctx, s.spec, management, s.provider); err != nil {
		return fmt.Errorf("installing cluster creation manifests: %v", err)
	}

	if err := s.manager.InstallMachineHealthChecks(ctx, s.spec, management); err != nil {
		return fmt.Errorf("installing machine health checks: %v", err)
	}

	return nil
}

// GetName satisfies the workload.Cluster interface.
func (s CreateClusterShim) GetName() string {
	return s.spec.Cluster.Name
}

// WriteKubeconfig satisfies the workload.Cluster interface.
func (s CreateClusterShim) WriteKubeconfig(ctx context.Context, w io.Writer, management *types.Cluster) error {
	return s.manager.getWorkloadClusterKubeconfig(ctx, s.spec.Cluster.Name, management, w)
}

// WaitUntilControlPlaneAvailable satisfies the workload.Cluster interface.
func (s CreateClusterShim) WaitUntilControlPlaneAvailable(ctx context.Context, management *types.Cluster) error {
	return s.manager.waitUntilControlPlaneAvailable(ctx, s.spec, management)
}

// WaitUntilReady satisfies the workload.Cluster interface.
func (s CreateClusterShim) WaitUntilReady(ctx context.Context, management *types.Cluster) error {
	return s.manager.waitForNodesReady(
		ctx,
		management,
		s.spec.Cluster.Name,
		[]string{clusterv1.MachineControlPlaneNameLabel, clusterv1.MachineDeploymentNameLabel},
		types.WithNodeRef(),
	)
}
