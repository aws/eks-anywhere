package reconciler

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
)

// CNIReconciler is an interface for reconciling CNI in the Tinkerbell cluster reconciler.
type CNIReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *c.Spec) (controller.Result, error)
}

// RemoteClientRegistry is an interface that defines methods for remote clients.
type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// IPValidator is an interface that defines methods to validate the control plane IP.
type IPValidator interface {
	ValidateControlPlaneIP(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error)
}

// Reconciler for Tinkerbell.
type Reconciler struct {
	client               client.Client
	cniReconciler        CNIReconciler
	remoteClientRegistry RemoteClientRegistry
	ipValidator          IPValidator
}

// New defines a new Tinkerbell reconciler.
func New(client client.Client, cniReconciler CNIReconciler, remoteClientRegistry RemoteClientRegistry, ipValidator IPValidator) *Reconciler {
	return &Reconciler{
		client:               client,
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		ipValidator:          ipValidator,
	}
}

// Reconcile reconciles cluster to desired state.
func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	// Implement reconcile all here.
	// This would include validating machine and datacenter configs
	// and reconciling cp and worker nodes.
	log = log.WithValues("provider", "tinkerbell")
	clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner().Register(
		r.ipValidator.ValidateControlPlaneIP,
		r.ValidateClusterSpec,
		r.ReconcileControlPlane,
		r.CheckControlPlaneReady,
		r.ReconcileCNI,
	).Run(ctx, log, clusterSpec)
}

// ValidateClusterSpec performs a set of assertions on a cluster spec.
func (r *Reconciler) ValidateClusterSpec(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "validateClusterSpec")

	tinkerbellClusterSpec := tinkerbell.NewClusterSpec(clusterSpec, clusterSpec.Config.TinkerbellMachineConfigs, clusterSpec.Config.TinkerbellDatacenter)

	clusterSpecValidator := tinkerbell.NewClusterSpecValidator()

	if err := clusterSpecValidator.Validate(tinkerbellClusterSpec); err != nil {
		log.Error(err, "Invalid Tinkerbell Cluster spec")
		failureMessage := err.Error()
		clusterSpec.Cluster.Status.FailureMessage = &failureMessage
		return controller.ResultWithReturn(), nil
	}
	return controller.Result{}, nil
}

// ReconcileControlPlane applies the control plane CAPI objects to the cluster.
func (r *Reconciler) ReconcileControlPlane(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileControlPlane")
	log.Info("Applying control plane CAPI objects")
	cp, err := tinkerbell.ControlPlaneSpec(ctx, log, clientutil.NewKubeClient(r.client), spec)
	if err != nil {
		return controller.Result{}, err
	}

	return clusters.ReconcileControlPlane(ctx, r.client, toClientControlPlane(cp))
}

// CheckControlPlaneReady checks whether the control plane for an eks-a cluster is ready or not.
// Requeues with the appropriate wait times whenever the cluster is not ready yet.
func (r *Reconciler) CheckControlPlaneReady(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "checkControlPlaneReady")
	return clusters.CheckControlPlaneReady(ctx, r.client, log, clusterSpec.Cluster)
}

// ReconcileWorkerNodes validates the cluster definition and reconciles the worker nodes
// to the desired state.
func (r *Reconciler) ReconcileWorkerNodes(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	// Implement reconcile worker nodes here

	return controller.Result{}, nil
}

// ReconcileCNI reconciles the CNI to the desired state.
func (r *Reconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileCNI")

	client, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(clusterSpec.Cluster))
	if err != nil {
		return controller.Result{}, err
	}

	return r.cniReconciler.Reconcile(ctx, log, client, clusterSpec)
}

func toClientControlPlane(cp *tinkerbell.ControlPlane) *clusters.ControlPlane {
	other := make([]client.Object, 0, 1)
	if cp.Secrets != nil {
		other = append(other, cp.Secrets)
	}
	return &clusters.ControlPlane{
		Cluster:                     cp.Cluster,
		ProviderCluster:             cp.ProviderCluster,
		KubeadmControlPlane:         cp.KubeadmControlPlane,
		ControlPlaneMachineTemplate: cp.ControlPlaneMachineTemplate,
		EtcdCluster:                 cp.EtcdCluster,
		EtcdMachineTemplate:         cp.EtcdMachineTemplate,
		Other:                       other,
	}
}
