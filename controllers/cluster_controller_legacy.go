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

	"github.com/aws/eks-anywhere/controllers/resource"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// ClusterReconcilerLegacy reconciles a Cluster object.
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

//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters;snowmachineconfigs;snowippools;vspheredatacenterconfigs;vspheremachineconfigs;cloudstackdatacenterconfigs;cloudstackmachineconfigs;dockerdatacenterconfigs;nutanixdatacenterconfigs;nutanixmachineconfigs;tinkerbellmachineconfigs;tinkerbelldatacenterconfigs;bundles;awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=oidcconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/status;snowmachineconfigs/status;snowippools/status;vspheredatacenterconfigs/status;vspheremachineconfigs/status;cloudstackdatacenterconfigs/status;cloudstackmachineconfigs/status;dockerdatacenterconfigs/status;bundles/status;awsiamconfigs/status;tinkerbelldatacenterconfigs/status;tinkerbellmachineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/finalizers;snowmachineconfigs/finalizers;snowippools/finalizers;vspheredatacenterconfigs/finalizers;vspheremachineconfigs/finalizers;cloudstackdatacenterconfigs/finalizers;cloudstackmachineconfigs/finalizers;dockerdatacenterconfigs/finalizers;bundles/finalizers;awsiamconfigs/finalizers;tinkerbelldatacenterconfigs/finalizers;tinkerbellmachineconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups=distro.eks.amazonaws.com,resources=releases,verbs=get;list;watch
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=fluxconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=snowdatacenterconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=snowmachineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=gitopsconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=addons.cluster.x-k8s.io,resources=clusterresourcesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterctl.cluster.x-k8s.io,resources=providers,verbs=get;list;watch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awssnowclusters;awssnowmachinetemplates;awssnowippools;vsphereclusters;vspheremachinetemplates;dockerclusters;dockermachinetemplates;tinkerbellclusters;tinkerbellmachinetemplates;cloudstackclusters;cloudstackmachinetemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",namespace=eksa-system,resources=secrets,verbs=delete;
// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=bmc.tinkerbell.org,resources=machines;machines/status,verbs=get;list;watch
//
// For the full cluster lifecycle to support Curated Packages, the controller
// must be able to create, delete, update, and patch package bundle controller
// resources, which will trigger the curated packages controller to do the
// rest.
//
// +kubebuilder:rbac:groups=packages.eks.amazonaws.com,resources=packagebundlecontrollers,verbs=create;delete;get;list;patch;update;watch;

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
		Watches(&source.Kind{Type: &anywherev1.CloudStackDatacenterConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.CloudStackMachineConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.TinkerbellDatacenterConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.TinkerbellMachineConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.DockerDatacenterConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.AWSIamConfig{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &anywherev1.OIDCConfig{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
