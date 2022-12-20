package management

import "github.com/aws/eks-anywhere/pkg/workflow"

const (
	PreUpgradeClusterTaskName  workflow.TaskName = "PreUpgradeManagementCluster"
	PostUpgradeClusterTaskName workflow.TaskName = "PostUpgradeManagementCluster"
)

// UpgradeClusterHookRegistrar is a Hook registrar that binds hooks to an upgrade management
// cluster workflow.
type UpgradeClusterHookRegistrar interface {
	RegisterUpgradeManagementClusterHooks(workflow.HookBinder)
}

// UpgradeClusterBuilder defines the configuration for a management cluster upgrade workflow.
type UpgradeClusterBuilder struct {
	HookRegistrars []UpgradeClusterHookRegistrar
}

// WithHookRegistrar adds a hook registrar to the upgrade cluster workflow builder.
func (b *UpgradeClusterBuilder) WithHookRegistrar(registrar UpgradeClusterHookRegistrar) *UpgradeClusterBuilder {
	b.HookRegistrars = append(b.HookRegistrars, registrar)
	return b
}

// Build builds the upgrade cluster workflow.
func (cfg *UpgradeClusterBuilder) Build() (*workflow.Workflow, error) {
	wflw := workflow.New(workflow.Config{})

	for _, r := range cfg.HookRegistrars {
		r.RegisterUpgradeManagementClusterHooks(wflw)
	}

	// Construct and register tasks for a management cluster upgrade workflow.

	return wflw, nil
}
