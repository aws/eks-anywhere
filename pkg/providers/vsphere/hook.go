package vsphere

import (
	"context"
	"errors"

	"github.com/aws/eks-anywhere/pkg/workflow"
	"github.com/aws/eks-anywhere/pkg/workflow/management"
	"github.com/aws/eks-anywhere/pkg/workflow/workflowcontext"
)

func (p *vsphereProvider) RegisterCreateManagementClusterHooks(binder workflow.HookBinder) {
	if !p.csiEnabled {
		return
	}

	binder.BindPostTaskHook(
		management.CreateWorkloadCluster,
		workflow.TaskFunc(func(ctx context.Context) (context.Context, error) {
			cluster := workflowcontext.WorkloadCluster(ctx)
			if cluster == nil {
				return ctx, errors.New("workload cluster not found in context")
			}

			return ctx, p.InstallStorageClass(ctx, cluster)
		}),
	)
}
