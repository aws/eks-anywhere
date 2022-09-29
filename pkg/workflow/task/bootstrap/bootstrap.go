package bootstrap

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflow/contextutil"
)

const clusterLogKey = "cluster"

// CreateClusters creates a functional Kubernetes cluster that can be used to faciliate
// EKS-A operations. The bootstrap cluster is populated in the context using
// contextutil.WithBootstrapCluster for subsequent tasks.
// If a cluster with the same name derived from the Spec already exists, it returns ErrClusterExists.
// If the Force option is set, it will recreate the cluster should it exist.
type CreateCluster struct {
	// Log contains a contextual logger.
	Log logr.Logger

	// Spec is the spec to be used for bootstrapping the cluster.
	Spec *cluster.Spec

	// Options supplies bootstrap cluster creation options.
	Options ClientOptionsRetriever

	// Bootstrapper is used to create the cluster.
	Client Client

	// File writes files to a specific directory context.
	File filewriter.FileWriter
}

// RunTask satisfies workflow.Task.
func (t CreateCluster) RunTask(ctx context.Context) (context.Context, error) {
	clusterName := t.Spec.Cluster.Name

	// Create the bootstrap cluster.
	opts := t.Options.GetCreateBootstrapClusterOptions(t.Spec)
	if err := t.Client.CreateBootstrapCluster(ctx, clusterName, opts); err != nil {
		return ctx, err
	}

	// Retrieve the kubeconfig and write it to non-volatile storage for later use.
	kubeconfig, err := t.Client.GetKubeconfig(ctx, clusterName)
	if err != nil {
		return ctx, err
	}

	kubeconfigPath, err := t.File.Write(toKubeconfigFilename(clusterName), kubeconfig)
	if err != nil {
		return ctx, err
	}
	t.Log.Info(fmt.Sprintf("Bootstrap cluster kubeconfig written to %v", kubeconfigPath))

	ctx = contextutil.WithBootstrapCluster(ctx, &types.Cluster{
		Name:           clusterName,
		KubeconfigFile: kubeconfigPath,
	})

	return ctx, nil
}

// DeleteCluster deletes a bootstrap cluster. It expects the bootstrap cluster to be populated
// in the context using contextutil.WithBootstrapCluster (typically done by CreateCluster).
// If the bootstrap cluster contains irrecoverable state it returns ErrUnexpectedState. If the Force
// option is set, the bootstrap cluster will be deleted regardless of state it may contain.
type DeleteCluster struct {
	// Log contains a contextual logger.
	Log logr.Logger

	// Client is used to delete the cluster.
	Client Client
}

// RunTask satisfies workflow.Task.
func (t DeleteCluster) RunTask(ctx context.Context) (context.Context, error) {
	cluster := contextutil.BootstrapCluster(ctx)
	if cluster == nil {
		return ctx, fmt.Errorf("bootstrap cluster not found in context")
	}

	exists, err := t.Client.ClusterExists(ctx, cluster.Name)
	if err != nil {
		return ctx, err
	}
	if !exists {
		t.Log.Info("Bootstrap cluster not found, skipping deletion", clusterLogKey, cluster.Name)
		return ctx, nil
	}

	if err := t.Client.DeleteBootstrapCluster(ctx, cluster.Name); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func toKubeconfigFilename(clusterName string) string {
	return fmt.Sprintf("%s.kubeconfig", clusterName)
}
