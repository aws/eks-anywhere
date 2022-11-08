package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/controller/handlers"
)

const (
	defaultRequeueTime   = time.Minute
	clusterFinalizerName = "clusters.anywhere.eks.amazonaws.com/finalizer"
)

// ClusterReconciler reconciles a Cluster object.
type ClusterReconciler struct {
	client                     client.Client
	log                        logr.Logger
	providerReconcilerRegistry ProviderClusterReconcilerRegistry
}

type ProviderClusterReconcilerRegistry interface {
	Get(datacenterKind string) clusters.ProviderClusterReconciler
}

func NewClusterReconciler(client client.Client, log logr.Logger, registry ProviderClusterReconcilerRegistry) *ClusterReconciler {
	return &ClusterReconciler{
		client:                     client,
		log:                        log,
		providerReconcilerRegistry: registry,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	childObjectHandler := handlers.ChildObjectToClusters(r.log)

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

// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters;snowmachineconfigs;vspheredatacenterconfigs;vspheremachineconfigs;dockerdatacenterconfigs;bundles;awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=oidcconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/status;snowmachineconfigs/status;vspheredatacenterconfigs/status;vspheremachineconfigs/status;dockerdatacenterconfigs/status;bundles/status;awsiamconfigs/status,verbs=;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/finalizers;snowmachineconfigs/finalizers;vspheredatacenterconfigs/finalizers;vspheremachineconfigs/finalizers;dockerdatacenterconfigs/finalizers;bundles/finalizers;awsiamconfigs/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=test,resources=test,verbs=get;list;watch;create;update;patch;delete;kill
// +kubebuilder:rbac:groups=distro.eks.amazonaws.com,resources=releases,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awssnowclusters;awssnowmachinetemplates,verbs=get;list;watch;create;update;patch;delete
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

	if cluster.Spec.BundlesRef == nil {
		if err = r.setBundlesRef(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err = r.ensureClusterOwnerReferences(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	return r.reconcile(ctx, cluster, log)
}

func (r *ClusterReconciler) reconcile(ctx context.Context, cluster *anywherev1.Cluster, log logr.Logger) (ctrl.Result, error) {
	clusterProviderReconciler := r.providerReconcilerRegistry.Get(cluster.Spec.DatacenterRef.Kind)

	var reconcileResult controller.Result
	var err error
	if cluster.IsSelfManaged() {
		// self-managed clusters should only reconcile worker nodes to avoid control plane instability
		reconcileResult, err = clusterProviderReconciler.ReconcileWorkerNodes(ctx, log, cluster)
	} else {
		reconcileResult, err = clusterProviderReconciler.Reconcile(ctx, log, cluster)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	return reconcileResult.ToCtrlResult(), nil
}

func (r *ClusterReconciler) reconcileDelete(ctx context.Context, cluster *anywherev1.Cluster) (ctrl.Result, error) {
	if cluster.IsSelfManaged() {
		return ctrl.Result{}, errors.New("deleting self-managed clusters is not supported")
	}

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
	config, err := builder.Build(ctx, clientutil.NewKubeClient(r.client), clus)
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

func (r *ClusterReconciler) setBundlesRef(ctx context.Context, clus *anywherev1.Cluster) error {
	mgmtCluster := &anywherev1.Cluster{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: clus.ManagedBy(), Namespace: clus.Namespace}, mgmtCluster); err != nil {
		return err
	}
	clus.Spec.BundlesRef = mgmtCluster.Spec.BundlesRef
	return nil
}
