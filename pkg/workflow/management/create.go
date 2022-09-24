package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/workflow"
	"github.com/aws/eks-anywhere/pkg/workflow/task/bootstrap"
)

// Define tasks names for each task run as part of the create cluster workflow. To aid readability
// the order of task names should be representative of the order of execution.
const (
	CreateBootstrapCluster workflow.TaskName = "CreateBootstrapCluster"
	DeleteBootstrapCluster workflow.TaskName = "DeleteBootstrapCluster"
)

// Define tasks names for each task run as part of the create cluster workflow. To aid readability
// the order of task names should be representative of the order of execution.
const (
	CreateBootstrapCluster workflow.TaskName = "CreateBootstrapCluster"
	DeleteBootstrapCluster workflow.TaskName = "DeleteBootstrapCluster"
)

// CreateClusterHookRegistrar is a Hook registrar that binds hooks to a create management cluster
// workflow.
type CreateClusterHookRegistrar interface {
	RegisterCreateManagementClusterHooks(workflow.HookBinder)
}

// CreateCluster defines the configuration for a managment cluster creation workflow.
// It executes tasks in the following order:
//  1. CreateBootstrapCluster
//  2. DeleteBootstrapCluster
type CreateCluster struct {
	// The spec used to construcft all other dependencies.
	Spec *cluster.Spec

	// CreateBootstrapOptions supplies bootstrap cluster options for creating bootstrap clusters.
	CreateBootstrapOptions bootstrap.BootstrapOptionsRetriever

	// Bootstrapper creates and destroys bootstrap clusters.
	Bootstrapper bootstrap.Bootstrapper

	// hookRegistrars are data structures that wish to bind runtime hooks to the workflow.
	// They should be added via the WithHookRegistrar method.
	hookRegistrars []CreateClusterHookRegistrar
}

// WithHookRegistrar adds a hook registrar to the create cluster workflow builder.
func (b *CreateCluster) WithHookRegistrar(registrar CreateClusterHookRegistrar) *CreateCluster {
	b.hookRegistrars = append(b.hookRegistrars, registrar)
	return b
}

func (b CreateCluster) Run(ctx context.Context) error {
	wflw, err := b.build()
	if err != nil {
		return err
	}

	return wflw.Execute(ctx)
}

// Build builds the create cluster workflow.
func (cfg CreateCluster) build() (*workflow.Workflow, error) {
	wflw := workflow.New(workflow.Config{})

	for _, r := range cfg.hookRegistrars {
		r.RegisterCreateManagementClusterHooks(wflw)
	}

	err := wflw.AppendTask(CreateBootstrapCluster, bootstrap.CreateBootstrapCluster{
		Spec:         cfg.Spec,
		Options:      cfg.CreateBootstrapOptions,
		Bootstrapper: cfg.Bootstrapper,
	})
	if err != nil {
		return nil, err
	}

	err = wflw.AppendTask(DeleteBootstrapCluster, bootstrap.DeleteBootstrapCluster{
		Bootstrapper: cfg.Bootstrapper,
	})
	if err != nil {
		return nil, err
	}

	return wflw, nil
}
