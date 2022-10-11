package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflow/workflowcontext"
)

// Cluster is an instance of a bootstrap cluster.
type Cluster interface {
	// CreateCluster creates the bootstrap cluster instance.
	Create(context.Context) error

	// DeleteCluster deletes the bootstrap cluster instance.
	Delete(context.Context) error

	// WriteKubeconfig writes the kubeconfig for the bootstrap cluster to the provided io.Writer.
	WriteKubeconfig(io.Writer) error
}

// FS is a filesystem abstraction.
type FS interface {
	// Create creates a new file called name at the instances configured root directory.
	Create(name string) (w io.WriteCloser, absPath string, err error)
}

// CreateCluster creates a functional Kubernetes cluster that can be used to faciliate
// EKS-A operations. The bootstrap cluster is populated in the context using
// workflow.WithBootstrapCluster for subsequent tasks.
type CreateCluster struct {
	// Spec is the spec to be used for bootstrapping the cluster.
	Spec *cluster.Spec

	// Bootstrapper is used to create the cluster.
	Cluster Cluster

	// FS is a filesystem abstraction with context of a root directory.
	FS FS
}

// RunTask satisfies workflow.Task.
func (t CreateCluster) RunTask(ctx context.Context) (context.Context, error) {
	if err := t.Cluster.Create(ctx); err != nil {
		return ctx, err
	}

	fh, fp, err := t.FS.Create(toKubeconfigFilename(t.Spec.Cluster.Name))
	if err != nil {
		return ctx, err
	}

	if err := t.Cluster.WriteKubeconfig(fh); err != nil {
		return ctx, err
	}

	cluster := &types.Cluster{
		Name:           t.Spec.Cluster.Name,
		KubeconfigFile: fp,
	}

	return workflowcontext.WithBootstrapCluster(ctx, cluster), nil
}

// DeleteCluster deletes a bootstrap cluster. It expects the bootstrap cluster to be
// populated in the context using workflow.WithBootstrapCluster.
type DeleteCluster struct {
	// Bootstrapper is used to delete the cluster.
	Cluster Cluster
}

// RunTask satisfies workflow.Task.
func (t DeleteCluster) RunTask(ctx context.Context) (context.Context, error) {
	if err := t.Cluster.Delete(ctx); err != nil {
		return ctx, err
	}

	cluster := workflowcontext.BootstrapCluster(ctx)
	if cluster == nil {
		return ctx, errors.New("bootstrap cluster not found in context")
	}

	if err := os.Remove(cluster.KubeconfigFile); err != nil {
		return ctx, fmt.Errorf("removing bootstrap kubeconfig file: %v", err)
	}

	return ctx, nil
}

func toKubeconfigFilename(clusterName string) string {
	return fmt.Sprintf("%s-kind.kubeconfig", clusterName)
}
