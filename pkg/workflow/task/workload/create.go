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

	// WriteKubeconfig writes the Kuberconfig for this cluster to the io.Writer.
	WriteKubeconfig(_ context.Context, _ io.Writer, management *types.Cluster) error

	// WaitUntilControlPlaneAvailable blocks until the first control plane is ready node is ready.
	// The node is ready when its possible to interact with the Kube API server using
	// a Kubeconfig.
	WaitUntilControlPlaneAvailable(_ context.Context, management *types.Cluster) error

	// WaitUntilReady blocks until all nodes within the cluster are ready. Nodes are ready when
	// they have joined the cluster and their Ready condition is true.
	WaitUntilReady(_ context.Context, management *types.Cluster) error // GetName retrieves the cluster name.
	GetName() string
}

// CNIInstaller install a CNI in a given cluster.
type CNIInstaller interface {
	// Install configures a CNI for the first time in a kubernetes cluster
	Install(ctx context.Context, cluster *types.Cluster) error
}

// Create creates a Kubernetes conformant cluster that is immediately usable for simple workloads.
// It expects a management cluster configuration to be available in the context.
type Create struct {
	// Cluster is an abstraction of a cluster that can be created.
	Cluster Cluster

	// CNI is an installer of a CNI. As per Kubernetes documentation, the CNI must be installed for
	// inter cluster communication.
	CNI CNIInstaller

	// FS is a file system abstraction providing file creation and write capabilities.
	FS filewriter.FileWriter
}

// RunTask satisfies workflow.Task.
func (t Create) RunTask(ctx context.Context) (context.Context, error) {
	management := workflowcontext.ManagementCluster(ctx)
	if management == nil {
		return nil, fmt.Errorf("no management cluster in context")
	}

	// Initiate the cluster creation process. This can take some time hence its an asyncronous
	// operation that we interrogate for progress as needed.
	if err := t.Cluster.CreateAsync(ctx, management); err != nil {
		return nil, err
	}

	// Wait for the first control plane to be ready. Once we have the first control plane we
	// assume we can write a Kubeconfig and install the CNI.
	//
	// Note we think this is important as the CNI is required for MachineHealthChecks to work.
	if err := t.Cluster.WaitUntilControlPlaneAvailable(ctx, management); err != nil {
		return nil, err
	}

	fh, path, err := t.FS.Create(
		kubeconfig.FormatWorkloadClusterKubeconfigFilename(t.Cluster.GetName()),
		filewriter.PersistentFile,
		filewriter.Permission0600,
	)
	if err != nil {
		return nil, err
	}

	if err := t.Cluster.WriteKubeconfig(ctx, fh, management); err != nil {
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

	// Ensure we block until the cluster is completely up. This is important as the Create task
	// should result in a usable cluster with all specified nodes ready.
	if err := t.Cluster.WaitUntilReady(ctx, management); err != nil {
		return nil, err
	}

	return ctx, nil
}
