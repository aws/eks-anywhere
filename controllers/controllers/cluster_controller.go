package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/aws/eks-anywhere/controllers/controllers/clients"
	"github.com/aws/eks-anywhere/controllers/controllers/clusters"
	"github.com/aws/eks-anywhere/controllers/controllers/utils/handlerutil"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

const (
	defaultRequeueTime   = time.Minute
	clusterFinalizerName = "clusters.anywhere.eks.amazonaws.com/finalizer"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client                  client.Client
	log                     logr.Logger
	validator               *vsphere.Validator
	defaulter               *vsphere.Defaulter
	tracker                 *remote.ClusterCacheTracker
	buildProviderReconciler BuildProviderReconciler
}

// TODO: this is not ideal and will need a refactor. I will follow up but for now this
// allows us to decouple the cluster reconciler main logic from provider specific logic
type BuildProviderReconciler func(datacenterKind string, client client.Client, log logr.Logger, validator *vsphere.Validator, defaulter *vsphere.Defaulter, tracker *remote.ClusterCacheTracker) (clusters.ProviderClusterReconciler, error)

func NewClusterReconciler(client client.Client, log logr.Logger, scheme *runtime.Scheme, govc *executables.Govc, tracker *remote.ClusterCacheTracker, buildProviderReconciler BuildProviderReconciler) *ClusterReconciler {
	validator := vsphere.NewValidator(govc, &networkutils.DefaultNetClient{})
	defaulter := vsphere.NewDefaulter(govc)

	return &ClusterReconciler{
		client:                  client,
		log:                     log,
		validator:               validator,
		defaulter:               defaulter,
		tracker:                 tracker,
		buildProviderReconciler: buildProviderReconciler,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	childObjectHandler := handlerutil.ChildObjectToClusters(r.log)

	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.Cluster{}).
		Watches(
			&source.Kind{Type: &anywherev1.OIDCConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.AWSIamConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.GitOpsConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.FluxConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.VSphereDatacenterConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.VSphereMachineConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.SnowDatacenterConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.SnowMachineConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Complete(r)
}

// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters;vspheredatacenterconfigs;vspheremachineconfigs;dockerdatacenterconfigs;bundles;awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=oidcconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/status;vspheredatacenterconfigs/status;vspheremachineconfigs/status;dockerdatacenterconfigs/status;bundles/status;awsiamconfigs/status,verbs=;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/finalizers;vspheredatacenterconfigs/finalizers;vspheremachineconfigs/finalizers;dockerdatacenterconfigs/finalizers;bundles/finalizers;awsiamconfigs/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=test,resources=test,verbs=get;list;watch;create;update;patch;delete;kill
// +kubebuilder:rbac:groups=distro.eks.amazonaws.com,resources=releases,verbs=get;list;watch
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := r.log.WithValues("cluster", req.NamespacedName)
	// Fetch the Cluster object
	cluster := &anywherev1.Cluster{}
	log.Info("Reconciling cluster", "name", req.NamespacedName)
	if err := r.client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
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
		return r.reconcileDelete(ctx, cluster)
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

	if err = r.ensureClusterOwnerReferences(ctx, cluster); err != nil {
		return ctrl.Result{}, err
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
	clusterProviderReconciler, err := r.buildProviderReconciler(cluster.Spec.DatacenterRef.Kind, r.client, r.log, r.validator, r.defaulter, r.tracker)
	if err != nil {
		return ctrl.Result{}, err
	}

	reconcileResult, err := clusterProviderReconciler.Reconcile(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}
	return reconcileResult.ToCtrlResult(), nil
}

func (r *ClusterReconciler) reconcileDelete(ctx context.Context, cluster *anywherev1.Cluster) (ctrl.Result, error) {
	capiCluster := &clusterv1.Cluster{}
	capiClusterName := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: cluster.Name}
	r.log.Info("Deleting", "name", cluster.Name)
	err := r.client.Get(ctx, capiClusterName, capiCluster)

	switch {
	case err == nil:
		r.log.Info("Deleting CAPI cluster", "name", capiCluster.Name)
		if err := r.client.Delete(ctx, capiCluster); err != nil {
			r.log.Info("Error deleting CAPI cluster", "name", capiCluster.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	case apierrors.IsNotFound(err):
		r.log.Info("Deleting EKS Anywhere cluster", "name", capiCluster.Name, "cluster.DeletionTimestamp", cluster.DeletionTimestamp, "finalizer", cluster.Finalizers)

		// TODO delete GitOps,Datacenter and MachineConfig objects
		controllerutil.RemoveFinalizer(cluster, clusterFinalizerName)
	default:
		return ctrl.Result{}, err

	}
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) ensureClusterOwnerReferences(ctx context.Context, clus *anywherev1.Cluster) error {
	builder := cluster.NewDefaultConfigClientBuilder()
	config, err := builder.Build(ctx, clients.NewKubeClient(r.client), clus)
	if err != nil {
		return err
	}

	childObjs := config.ChildObjects()
	for _, obj := range childObjs {
		numberOfOwnerReferences := len(obj.GetOwnerReferences())
		if err = controllerutil.SetOwnerReference(clus, obj, r.client.Scheme()); err != nil {
			return errors.Wrapf(err, "setting cluster owner reference for %s", obj.GetObjectKind())
		}

		if numberOfOwnerReferences == len(obj.GetOwnerReferences()) {
			// obj already had the owner reference
			continue
		}

		if err = r.client.Update(ctx, obj); err != nil {
			return errors.Wrapf(err, "updating object (%s) with cluster owner reference", obj.GetObjectKind())
		}
	}

	return nil
}
