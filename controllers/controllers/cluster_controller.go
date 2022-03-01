package controllers

import (
	"context"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/aws/eks-anywhere/controllers/controllers/clusters"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

const clusterFinalizerName string = "clusters.anywhere.eks.amazonaws.com/finalizer"

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client    client.Client
	log       logr.Logger
	validator *vsphere.Validator
	defaulter *vsphere.Defaulter
	tracker   *remote.ClusterCacheTracker
}

func NewClusterReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme, govc *executables.Govc, tracker *remote.ClusterCacheTracker) *ClusterReconciler {
	validator := vsphere.NewValidator(govc, &networkutils.DefaultNetClient{})
	defaulter := vsphere.NewDefaulter(govc)

	return &ClusterReconciler{
		client:    client,
		log:       log,
		validator: validator,
		defaulter: defaulter,
		tracker:   tracker,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.Cluster{}).
		// Watches(&source.Kind{Type: &anywherev1.VSphereDatacenterConfig{}}, &handler.EnqueueRequestForObject{}).
		// Watches(&source.Kind{Type: &anywherev1.VSphereMachineConfig{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters;vspheredatacenterconfigs;vspheremachineconfigs;dockerdatacenterconfigs;bundles;awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=oidcconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/status;vspheredatacenterconfigs/status;vspheremachineconfigs/status;dockerdatacenterconfigs/status;bundles/status;awsiamconfigs/status,verbs=;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/finalizers;vspheredatacenterconfigs/finalizers;vspheremachineconfigs/finalizers;dockerdatacenterconfigs/finalizers;bundles/finalizers;awsiamconfigs/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=test,resources=test,verbs=get;list;watch;create;update;patch;delete;kill
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := r.log.WithValues("cluster", req.NamespacedName)
	// Fetch the Cluster object
	cluster := &anywherev1.Cluster{}
	log.Info("Reconciling cluster", "name", req.NamespacedName)
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

	if cluster.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(cluster, clusterFinalizerName) {
			controllerutil.AddFinalizer(cluster, clusterFinalizerName)
		}
	} else {
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

	result, err := r.reconcile(ctx, cluster, log)
	if err != nil {
		failureMessage := err.Error()
		cluster.Status.FailureMessage = &failureMessage
		log.Error(err, "Failed to reconcile Cluster")
	}
	return result, err
}

func (r *ClusterReconciler) reconcile(ctx context.Context, cluster *anywherev1.Cluster, log logr.Logger) (ctrl.Result, error) {
	clusterProviderReconciler, err := clusters.BuildProviderReconciler(cluster.Spec.DatacenterRef.Kind, r.client, r.log, r.validator, r.defaulter, r.tracker)
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
	capiCluster := &clusterv1.Cluster{}
	capiClusterName := types.NamespacedName{Namespace: "eksa-system", Name: cluster.Name}
	r.log.Info("Deleting", "name", cluster.Name)
	err := r.client.Get(ctx, capiClusterName, capiCluster)

	switch {
	case err == nil:
		r.log.Info("Deleting CAPI cluster", "name", capiCluster.Name)
		if err := r.client.Delete(ctx, capiCluster); err != nil {
			r.log.Info("Error deleting CAPI cluster", "name", capiCluster.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case !apierrors.IsNotFound(err):
		return ctrl.Result{}, err
	case apierrors.IsNotFound(err):
		r.log.Info("Deleting EKS Anywhere cluster", "name", capiCluster.Name, "cluster.DeletionTimestamp", cluster.DeletionTimestamp, "finalizer", cluster.Finalizers)

		// TODO delete GitOps,Datacenter and MachineConfig objects
		controllerutil.RemoveFinalizer(cluster, clusterFinalizerName)
		if err := r.client.Update(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}

	}
	return ctrl.Result{}, nil
}
