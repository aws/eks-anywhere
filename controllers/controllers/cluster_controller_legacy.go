package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/aws/eks-anywhere/controllers/controllers/resource"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// ClusterReconcilerLegacy reconciles a Cluster object
type ClusterReconcilerLegacy struct {
	client.Client
	Log             logr.Logger
	Scheme          *runtime.Scheme
	reconcilers     []resource.Reconciler
	resourceFetcher resource.ResourceFetcher
}

func NewClusterReconcilerLegacy(client client.Client, log logr.Logger, scheme *runtime.Scheme) *ClusterReconcilerLegacy {
	return &ClusterReconcilerLegacy{
		Client: client,
		Log:    log,
		Scheme: scheme,
		reconcilers: []resource.Reconciler{
			resource.NewClusterReconciler(
				resource.NewCAPIResourceFetcher(client, log),
				resource.NewCAPIResourceUpdater(client, log),
				time.Now,
				log),
		},
		resourceFetcher: resource.NewCAPIResourceFetcher(client, log),
	}
}

//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters;vspheredatacenterconfigs;vspheremachineconfigs;cloudstackdeploymentconfigs;cloudstackmachineconfigs;dockerdatacenterconfigs;bundles;awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=oidcconfigs,verbs=get;list
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/status;vspheredatacenterconfigs/status;vspheremachineconfigs/status;cloudstackdeploymentconfigs/status;cloudstackmachineconfigs/status;dockerdatacenterconfigs/status;bundles/status;awsiamconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/finalizers;vspheredatacenterconfigs/finalizers;vspheremachineconfigs/finalizers;cloudstackdeploymentconfigs/finalizers;cloudstackmachineconfigs/finalizers;dockerdatacenterconfigs/finalizers;bundles/finalizers;awsiamconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterReconcilerLegacy) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	_ = r.Log.WithValues("cluster", req.NamespacedName)
	// Fetch the Cluster instance.
	cluster, err := r.resourceFetcher.FetchCluster(ctx, req.NamespacedName)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(cluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		if err := patchHelper.Patch(ctx, cluster); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Ignore deleted Clusters, this can happen when foregroundDeletion
	// is enabled
	if !cluster.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// If the external object is paused, return without any further processing.
	if cluster.IsReconcilePaused() {
		r.Log.Info("eksa reconciliation is paused")
		return ctrl.Result{}, nil
	}

	// dry run
	result, err := r.reconcile(ctx, req.NamespacedName, true)
	if err != nil {
		r.Log.Error(err, "Dry run failed to reconcile Cluster")
		return result, err
	}
	// non dry run
	result, err = r.reconcile(ctx, req.NamespacedName, false)
	if err != nil {
		r.Log.Error(err, "Failed to reconcile Cluster")
	}
	return result, err
}

func (r *ClusterReconcilerLegacy) reconcile(ctx context.Context, objectKey types.NamespacedName, dryRun bool) (ctrl.Result, error) {
	r.Log.Info("Reconcile EKS-A Cluster", "dryRun", dryRun)
	errs := []error{}
	for _, phase := range r.reconcilers {
		err := phase.Reconcile(ctx, objectKey, dryRun)
		if err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			continue
		}
	}
	return ctrl.Result{}, kerrors.NewAggregate(errs)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconcilerLegacy) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.Cluster{}).
		Watches(&source.Kind{Type: &anywherev1.VSphereDatacenterConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.VSphereMachineConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.CloudStackDeploymentConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.CloudStackMachineConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.DockerDatacenterConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.AWSIamConfig{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
