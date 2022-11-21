package management

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/workflow"
	"github.com/aws/eks-anywhere/pkg/workflow/task/bootstrap"
	"github.com/aws/eks-anywhere/pkg/workflow/task/workload"
)

// Define tasks names for each task run as part of the create cluster workflow. To aid readability
// the order of task names should be representative of the order of execution.
const (
	CreateBootstrapCluster workflow.TaskName = "CreateBootstrapCluster"
	CreateWorkloadCluster  workflow.TaskName = "CreateWorkloadCluster"
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
	CreateBootstrapClusterOptions bootstrap.OptionsRetriever

	// Bootstrapper creates and destroys bootstrap clusters.
	Bootstrapper bootstrap.Bootstrapper

	// Cluster represents a logical cluster to be created.
	Cluster workload.Cluster

	// CNIInstaller installs a CNI in a Kubernetes cluster
	CNIInstaller workload.CNIInstaller

	// FS is a file system abstraction used to write files.
	FS filewriter.FileWriter

	// hookRegistrars are data structures that wish to bind runtime hooks to the workflow.
	// They should be added via the WithHookRegistrar method.
	hookRegistrars []CreateClusterHookRegistrar
}

// WithHookRegistrar adds a hook registrar to the create cluster workflow builder.
func (c *CreateCluster) WithHookRegistrar(registrar CreateClusterHookRegistrar) *CreateCluster {
	c.hookRegistrars = append(c.hookRegistrars, registrar)
	return c
}

// Run runs the create cluster workflow.
func (c CreateCluster) Run(ctx context.Context) error {
	wflw, err := c.build()
	if err != nil {
		return err
	}

	return wflw.Execute(ctx)
}

func (c CreateCluster) build() (*workflow.Workflow, error) {
	wflw := workflow.New(workflow.Config{})

	for _, r := range c.hookRegistrars {
		r.RegisterCreateManagementClusterHooks(wflw)
	}

	err := wflw.AppendTask(CreateBootstrapCluster, bootstrap.CreateCluster{
		Spec:         c.Spec,
		Options:      c.CreateBootstrapClusterOptions,
		Bootstrapper: c.Bootstrapper,
	})
	if err != nil {
		return nil, err
	}

	err = wflw.AppendTask(CreateWorkloadCluster, workload.Create{
		Cluster: c.Cluster,
		CNI:     c.CNIInstaller,
		FS:      c.FS,
	})
	if err != nil {
		return nil, err
	}

	err = wflw.AppendTask(DeleteBootstrapCluster, bootstrap.DeleteCluster{
		Bootstrapper: c.Bootstrapper,
	})
	if err != nil {
		return nil, err
	}

	return wflw, nil
}
