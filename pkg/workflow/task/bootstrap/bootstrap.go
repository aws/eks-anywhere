package bootstrap

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"

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

	// Force indicates the bootstrap cluster should be force created. In practice, this means
	// deleting any previously created bootstrap clusters that may contain important state.
	Force bool
}

// RunTask satisfies workflow.Task.
func (t CreateCluster) RunTask(ctx context.Context) (context.Context, error) {
	clusterName := t.Spec.Cluster.Name

	// Ensure the cluster doesn't already exist.
	exists, err := t.Client.ClusterExists(ctx, clusterName)
	if err != nil {
		return ctx, err
	}

	switch {
	case exists && !t.Force:
		return ctx, ErrClusterExists{clusterName}
	case exists && t.Force:
		t.Log.Info("Force deleting existing bootstrap cluster", clusterLogKey, clusterName)
		if err := t.Client.DeleteBootstrapCluster(ctx, clusterName); err != nil {
			return ctx, fmt.Errorf("force delete bootstrap cluster: %v", err)
		}
	}

	// Create the bootstrap cluster.
	opts := t.Options.GetCreateBootstrapClusterOptions(t.Spec)
	err = t.Client.CreateBootstrapCluster(ctx, clusterName, opts)
	if err != nil {
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

// KubernetesClient is used to interact with a Kubernetes cluster.
type KubernetesClient interface {
	Get(ctx context.Context, kubeconfig string, key apimachinerytypes.NamespacedName, obj runtime.Object) error
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

	// K8s is a kubernetes client used to check for cluster state.
	K8s KubernetesClient

	// Force indicates the bootstrap cluster should be force deleted irrespective of state it
	// may contain.
	Force bool
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

	// If the cluster contains anything we think is necessary for recovering the state of an
	// operation then we shouldn't delete the cluster. For example, if the cluster contains
	// CAPI objects that should no longer reside in the bootstrap cluster having been moved
	// in a previous task, we wouldn't want to remove the clsuter without an explicit opt-in
	// from the user.
	hasState, err := hasIrrecoverableState(ctx, t.K8s, *cluster)
	if err != nil {
		return ctx, err
	}

	if !t.Force && hasState {
		return ctx, ErrUnexpectedState{cluster.Name}
	}

	// Create a sane log statement for the various conditions.
	switch {
	case t.Force && hasState:
		t.Log.Info("Force deleting bootstrap cluster containing unexpected state", clusterLogKey, cluster.Name)
	case t.Force && !hasState:
		t.Log.Info("Force deleting bootstrap cluster", clusterLogKey, cluster.Name)
	}

	if err := t.Client.DeleteBootstrapCluster(ctx, cluster.Name); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func toKubeconfigFilename(clusterName string) string {
	return fmt.Sprintf("%s.kubeconfig", clusterName)
}

func hasIrrecoverableState(ctx context.Context, client KubernetesClient, cluster types.Cluster) (bool, error) {
	// TODO(chrisdoherty) Determine if there are CAPI objects that shouldn't exist on bootstrap
	// cluster deletion.
	return false, nil
}
