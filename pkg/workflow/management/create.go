package management

import (
	"github.com/aws/eks-anywhere/pkg/workflow"
)

const (
	PreCreateClusterTaskName  workflow.TaskName = "PreCreateManagementCluster"
	PostCreateClusterTaskName workflow.TaskName = "PostCreateManagementCluster"
)

// CreateClusterHookRegistrar is a Hook registrar that binds hooks to a create management cluster
// workflow.
type CreateClusterHookRegistrar interface {
	RegisterCreateManagementClusterHooks(workflow.HookBinder)
}

// CreateClusterConfig defines the configuration for a managment cluster creation workflow.
type CreateClusterBuilder struct {
	HookRegistrars []CreateClusterHookRegistrar
}

// WithHookRegistrar adds a hook registrar to the create cluster workflow builder.
func (b *CreateClusterBuilder) WithHookRegistrar(registrar CreateClusterHookRegistrar) *CreateClusterBuilder {
	b.HookRegistrars = append(b.HookRegistrars, registrar)
	return b
}

// Build builds the create cluster workflow.
func (cfg *CreateClusterBuilder) Build() (*workflow.Workflow, error) {
	wflw := workflow.New(workflow.Config{})

	for _, r := range cfg.HookRegistrars {
		r.RegisterCreateManagementClusterHooks(wflw)
	}

	// Construct and register tasks for a management cluster creation workflow.

	return wflw, nil
}
