package awsiamauth

import (
	"context"
	"errors"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/workflow"
	"github.com/aws/eks-anywhere/pkg/workflow/management"
	"github.com/aws/eks-anywhere/pkg/workflow/workflowcontext"
)

// HookRegistrar is responsible for binding AWS IAM Auth hooks to workflows so it can be
// installed.
type HookRegistrar struct {
	*Installer

	// Spec is the configuration for the cluster we're trying to create.
	spec *cluster.Spec
}

// NewHookRegistrar creates a HookRegistrar instance.
func NewHookRegistrar(installer *Installer, spec *cluster.Spec) HookRegistrar {
	return HookRegistrar{
		Installer: installer,
		spec:      spec,
	}
}

func (r HookRegistrar) RegisterCreateManagementClusterHooks(binder workflow.HookBinder) {
	// We need to generate and install a CA certificate to the bootstrap cluster. The secret is
	// used to populate KubeadmControlPlane objects. The names used for the secret are
	// hard coded in the KubeadmControlPlane object template files.
	binder.BindPostTaskHook(
		management.CreateBootstrapCluster,
		workflow.TaskFunc(func(ctx context.Context) (context.Context, error) {
			cluster := workflowcontext.BootstrapCluster(ctx)
			if cluster == nil {
				return ctx, errors.New("cluster not found in context")
			}

			return ctx, r.CreateAndInstallAWSIAMAuthCASecret(ctx, cluster, r.spec.Cluster.Name)
		}),
	)

	// Bind a hook to install AWS IAM Authenticator into the permanent workload cluster.
	binder.BindPostTaskHook(
		management.CreateWorkloadCluster,
		workflow.TaskFunc(func(ctx context.Context) (context.Context, error) {
			management := workflowcontext.BootstrapCluster(ctx)
			if management == nil {
				return ctx, errors.New("management cluster not found in context")
			}

			workload := workflowcontext.WorkloadCluster(ctx)
			if workload == nil {
				return ctx, errors.New("workload cluster not found in context")
			}

			return ctx, r.InstallAWSIAMAuth(ctx, management, workload, r.spec)
		}),
	)
}
