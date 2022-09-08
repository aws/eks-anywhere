package management

import "github.com/aws/eks-anywhere/pkg/workflow"

const (
	PreDeleteClusterTaskName  workflow.TaskName = "PreDeleteManagementCluster"
	PostDeleteClusterTaskName workflow.TaskName = "PostDeleteManagementCluster"
)

// DeleteClusterHookRegistrar is a Hook registrar that binds hooks to a delete management cluster
// workflow.
type DeleteClusterHookRegistrar interface {
	RegisterDeleteManagementClusterHooks(workflow.HookBinder)
}

// DeleteClusterBuilder defines the configuration for a management cluster deletion workflow.
type DeleteClusterBuilder struct {
	HookRegistrars []DeleteClusterHookRegistrar
}

// WithHookRegistrar adds a hook registrar to the delete cluster workflow builder.
func (b *DeleteClusterBuilder) WithHookRegistrar(registrar DeleteClusterHookRegistrar) *DeleteClusterBuilder {
	b.HookRegistrars = append(b.HookRegistrars, registrar)
	return b
}

// Build builds the delete cluster workflow.
func (cfg *DeleteClusterBuilder) Build() (*workflow.Workflow, error) {
	wflw := workflow.New(workflow.Config{})

	for _, r := range cfg.HookRegistrars {
		r.RegisterDeleteManagementClusterHooks(wflw)
	}

	// Construct and register tasks for a management cluster deletion workflow.

	return wflw, nil
}
