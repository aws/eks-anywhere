package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
)

// Reconciler contains dependencies for a docker reconciler.
type Reconciler struct {
	client               client.Client
	cniReconciler        CNIReconciler
	remoteClientRegistry RemoteClientRegistry
	*serverside.ObjectApplier
}

// CNIReconciler is an interface for reconciling CNI in the Docker cluster reconciler.
type CNIReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (controller.Result, error)
}

// RemoteClientRegistry is an interface that defines methods for remote clients.
type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// New creates a new Docker provider reconciler.
func New(client client.Client, cniReconciler CNIReconciler, remoteClientRegistry RemoteClientRegistry) *Reconciler {
	return &Reconciler{
		client:               client,
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		ObjectApplier:        serverside.NewObjectApplier(client),
	}
}

// Reconcile brings the cluster to the desired state for the docker provider.
func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	return controller.Result{}, nil
}

// ReconcileCNI takes the Cilium CNI in a cluster to the desired state defined in a cluster spec.
func (r *Reconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileCNI")
	client, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(clusterSpec.Cluster))
	if err != nil {
		return controller.Result{}, err
	}

	return r.cniReconciler.Reconcile(ctx, log, client, clusterSpec)
}

// ReconcileWorkerNodes validates the cluster definition and reconciles the worker nodes
// to the desired state.
func (r *Reconciler) ReconcileWorkerNodes(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "docker", "reconcile type", "workers")
	clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner().Register(
		r.ReconcileWorkers,
	).Run(ctx, log, clusterSpec)
}

// ReconcileWorkers applies the worker CAPI objects to the cluster.
func (r *Reconciler) ReconcileWorkers(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileWorkers")
	log.Info("Applying worker CAPI objects")
	return r.Apply(ctx, func() ([]kubernetes.Object, error) {
		w, err := docker.WorkersSpec(ctx, log, clientutil.NewKubeClient(r.client), clusterSpec)
		if err != nil {
			return nil, err
		}
		for _, w := range w.WorkerObjects() {
			fmt.Printf("%v \n", w.GetName())
		}
		return w.WorkerObjects(), nil
	})
}
