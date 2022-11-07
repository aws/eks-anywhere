package workload

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflow/workflowcontext"
)

// CNIInstaller install a CNI in a given cluster.
type CNIInstaller interface {
	// Install configures a CNI for the first time in a kubernetes cluster
	Install(ctx context.Context, cluster *types.Cluster) error
}

// Create creates a Kubernetes conformant cluster.
type Create struct {
	CNI CNIInstaller
}

// RunTask satisfies workflow.Task.
func (t Create) RunTask(ctx context.Context) (context.Context, error) {
	// TODO: add provider create cluster, kubeconfig retrieve and write to disk

	workloadCluster := &types.Cluster{
		Name:           "my-cluster",      // TODO: use name of cluster created by provider
		KubeconfigFile: "fake.kubeconfig", // TODO: use real path
	}

	ctx = workflowcontext.WithWorkloadCluster(ctx, workloadCluster)

	if err := t.CNI.Install(ctx, workloadCluster); err != nil {
		return nil, fmt.Errorf("installing CNI in workload cluster: %v", err)
	}

	return ctx, nil
}
