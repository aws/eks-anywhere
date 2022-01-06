package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/controllers/controllers/clusters"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client client.Client
	log    logr.Logger
}

func NewClusterReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme) *ClusterReconciler {
	return &ClusterReconciler{
		client: client,
		log:    log,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.Cluster{}).
		Complete(r)
}

// TODO: add here kubebuilder permissions as neeeded
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := r.log.WithValues("cluster", req.NamespacedName)

	log.Info("Cluster reconciliation initializing")

	// Fetch the Cluster object
	cluster := &anywherev1.Cluster{}
	if err := r.client.Get(ctx, req.NamespacedName, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(cluster, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		if err := patchHelper.Patch(ctx, cluster); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	if !cluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, cluster, log)
	}

	// If the cluster is paused, return without any further processing.
	if cluster.IsReconcilePaused() {
		log.Info("Cluster reconciliation is paused")
		return ctrl.Result{}, nil
	}

	if cluster.IsSelfManaged() {
		log.Info("Ignoring self managed cluster")
		return ctrl.Result{}, nil
	}

	// Fetch the VsphereDatacenter object
	dc := &anywherev1.VSphereDatacenterConfig{}
	dcName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.DatacenterRef.Name}
	if err := r.client.Get(ctx, dcName, dc); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("Using datacenter config config %v", dc)

	if !dc.Status.SpecValid {
		log.Info("Skipping cluster reconciliation because datacenter config is invalid %v", dc)
		return ctrl.Result{}, nil
	}

	// Fetch the VsphereDatacenter object
	mc := &anywherev1.VSphereMachineConfig{}
	mcName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name}
	if err := r.client.Get(ctx, mcName, mc); err != nil {
		return ctrl.Result{}, err
	}
	log.Info("Using machine config %v", mc)

	//Validate machine config for etcd

	// repeat the process for other machine config objects

	result, err := r.reconcile(ctx, cluster, log)
	if err != nil {
		log.Error(err, "Failed to reconcile Cluster")
	}
	return result, err
}

func (r *ClusterReconciler) reconcile(ctx context.Context, cluster *anywherev1.Cluster, log logr.Logger) (ctrl.Result, error) {
	clusterProviderReconciler, err := clusters.BuildProviderReconciler(cluster.Spec.DatacenterRef.Kind)
	if err != nil {
		return ctrl.Result{}, err
	}

	reconcileResult, err := clusterProviderReconciler.Reconcile(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}
	return reconcileResult.ToCtrlResult(), nil
}

func (r *ClusterReconciler) reconcileDelete(ctx context.Context, cluster *anywherev1.Cluster, log logr.Logger) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
