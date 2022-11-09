package workload

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflow/workflowcontext"
)

// Cluster represents a workload cluster to be created.
type Cluster interface {
	// CreateAsync performs the necessary action that will eventually result in a cluster being
	// created. This likely includes applying CAPI manifests to the cluster but its not the only
	// thing that may be required.
	CreateAsync(_ context.Context, management *types.Cluster) error

	// GetName retrieves the cluster name.
	GetName() string

	// WriteKubeconfig writes the Kuberconfig for this cluster to the io.Writer.
	WriteKubeconfig(io.Writer) error

	// WaitUntilFirstControlPlaneReady blocks until the first control plane is ready node is ready.
	// The node is ready when its possible to interact with the Kube API server using
	// a Kubeconfig.
	WaitUntilFirstControlPlaneReady(_ context.Context, management *types.Cluster) error

	// WaitUntilReady blocks until all nodes within the cluster are ready. All nodes
	// are ready when its possible to deploy workloads to the cluster.
	WaitUntilReady(_ context.Context, management *types.Cluster) error
}

// CNIInstaller install a CNI in a given cluster.
type CNIInstaller interface {
	// Install configures a CNI for the first time in a kubernetes cluster
	Install(ctx context.Context, cluster *types.Cluster) error
}

// Create creates a Kubernetes conformant cluster that is immediately usable for simple workloads.
// It expects a management cluster configuration to be available in the context.
type Create struct {
	Cluster Cluster
	CNI     CNIInstaller
	Writer  filewriter.FileWriter
}

// RunTask satisfies workflow.Task.
func (t Create) RunTask(ctx context.Context) (context.Context, error) {
	managementCluster := workflowcontext.ManagementCluster(ctx)
	if managementCluster == nil {
		return nil, fmt.Errorf("no management cluster in context")
	}

	// Initiate the cluster creation process. This can take some time hence its an asyncronous
	// operation that we interrogate for progress as needed.
	if err := t.Cluster.CreateAsync(ctx, managementCluster); err != nil {
		return nil, err
	}

	// Wait for the first control plane to be ready. Once we have the first control plane we
	// assume we can write a Kubeconfig and install the CNI.
	//
	// Note we think this is important as the CNI is required for MachineHealthChecks to work.
	if err := t.Cluster.WaitUntilFirstControlPlaneReady(ctx, managementCluster); err != nil {
		return nil, err
	}

	kubeconfigFilename := kubeconfig.FormatWorkloadClusterKubeconfigFilename(t.Cluster.GetName())
	fh, path, err := t.Writer.Create(kubeconfigFilename)
	if err != nil {
		return nil, err
	}

	if err := t.Cluster.WriteKubeconfig(fh); err != nil {
		return nil, err
	}

	workloadCluster := &types.Cluster{
		Name:           t.Cluster.GetName(),
		KubeconfigFile: path,
	}

	ctx = workflowcontext.WithWorkloadCluster(ctx, workloadCluster)

	if err := t.CNI.Install(ctx, workloadCluster); err != nil {
		return nil, fmt.Errorf("installing CNI in workload cluster: %v", err)
	}

	if err := t.Cluster.WaitUntilReady(ctx, managementCluster); err != nil {
		return nil, err
	}

	return ctx, nil
}
