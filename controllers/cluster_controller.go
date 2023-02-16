package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const (
	defaultRequeueTime = time.Minute
	// ClusterFinalizerName is the finalizer added to clusters to handle deletion.
	ClusterFinalizerName = "clusters.anywhere.eks.amazonaws.com/finalizer"
)

// ClusterReconciler reconciles a Cluster object.
type ClusterReconciler struct {
	client                     client.Client
	providerReconcilerRegistry ProviderClusterReconcilerRegistry
	awsIamAuth                 AWSIamConfigReconciler
	clusterValidator           ClusterValidator
}

type ProviderClusterReconcilerRegistry interface {
	Get(datacenterKind string) clusters.ProviderClusterReconciler
}

// AWSIamConfigReconciler manages aws-iam-authenticator installation and configuration for an eks-a cluster.
type AWSIamConfigReconciler interface {
	EnsureCASecret(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error)
	Reconcile(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error)
	ReconcileDelete(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) error
}

// ClusterValidator runs cluster level preflight validations before it goes to provider reconciler.
type ClusterValidator interface {
	ValidateManagementClusterName(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error
}

// NewClusterReconciler constructs a new ClusterReconciler.
func NewClusterReconciler(client client.Client, registry ProviderClusterReconcilerRegistry, awsIamAuth AWSIamConfigReconciler, clusterValidator ClusterValidator) *ClusterReconciler {
	return &ClusterReconciler{
		client:                     client,
		providerReconcilerRegistry: registry,
		awsIamAuth:                 awsIamAuth,
		clusterValidator:           clusterValidator,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager, log logr.Logger) error {
	childObjectHandler := handlers.ChildObjectToClusters(log)

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
		Watches(
			&source.Kind{Type: &anywherev1.TinkerbellDatacenterConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.TinkerbellMachineConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.DockerDatacenterConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Complete(r)
}

// Reconcile reconciles a cluster object.
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters;snowmachineconfigs;snowippools;vspheredatacenterconfigs;vspheremachineconfigs;dockerdatacenterconfigs;tinkerbellmachineconfigs;tinkerbelldatacenterconfigs;bundles;awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=oidcconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/status;snowmachineconfigs/status;snowippools/status;vspheredatacenterconfigs/status;vspheremachineconfigs/status;dockerdatacenterconfigs/status;bundles/status;awsiamconfigs/status;tinkerbelldatacenterconfigs/status;tinkerbellmachineconfigs/status,verbs=;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/finalizers;snowmachineconfigs/finalizers;snowippools/finalizers;vspheredatacenterconfigs/finalizers;vspheremachineconfigs/finalizers;dockerdatacenterconfigs/finalizers;bundles/finalizers;awsiamconfigs/finalizers;tinkerbelldatacenterconfigs/finalizers;tinkerbellmachineconfigs/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=addons.cluster.x-k8s.io,resources=clusterresourcesets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=test,resources=test,verbs=get;list;watch;create;update;patch;delete;kill
// +kubebuilder:rbac:groups=distro.eks.amazonaws.com,resources=releases,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awssnowclusters;awssnowmachinetemplates;awssnowippools;vsphereclusters;vspheremachinetemplates;dockerclusters;dockermachinetemplates;tinkerbellclusters;tinkerbellmachinetemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",namespace=eksa-system,resources=secrets,verbs=delete;
// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch;update;patch
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)
	// Fetch the Cluster object
	cluster := &anywherev1.Cluster{}
	log.Info("Reconciling cluster")
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

	if !cluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, cluster)
	}

	// If the cluster is paused, return without any further processing.
	if cluster.IsReconcilePaused() {
		log.Info("Cluster reconciliation is paused")
		return ctrl.Result{}, nil
	}

	// AddFinalizer	is idempotent
	controllerutil.AddFinalizer(cluster, ClusterFinalizerName)

	if cluster.Spec.BundlesRef == nil {
		if err = r.setBundlesRef(ctx, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err = r.ensureClusterOwnerReferences(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	return r.reconcile(ctx, log, cluster)
}

func (r *ClusterReconciler) reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (ctrl.Result, error) {
	clusterProviderReconciler := r.providerReconcilerRegistry.Get(cluster.Spec.DatacenterRef.Kind)

	var reconcileResult controller.Result
	var err error

	reconcileResult, err = r.preClusterProviderReconcile(ctx, log, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if reconcileResult.Return() {
		return reconcileResult.ToCtrlResult(), nil
	}

	if cluster.IsSelfManaged() {
		// self-managed clusters should only reconcile worker nodes to avoid control plane instability
		reconcileResult, err = clusterProviderReconciler.ReconcileWorkerNodes(ctx, log, cluster)
	} else {
		reconcileResult, err = clusterProviderReconciler.Reconcile(ctx, log, cluster)
	}

	if err != nil {
		return ctrl.Result{}, err
	}

	if reconcileResult.Return() {
		return reconcileResult.ToCtrlResult(), nil
	}

	if reconcileResult, err = r.postClusterProviderReconcile(ctx, log, cluster); err != nil {
		return ctrl.Result{}, err
	}

	return reconcileResult.ToCtrlResult(), nil
}

func (r *ClusterReconciler) preClusterProviderReconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	// Run some preflight validations that can't be checked in webhook
	if cluster.HasAWSIamConfig() {
		if result, err := r.awsIamAuth.EnsureCASecret(ctx, log, cluster); err != nil {
			return controller.Result{}, err
		} else if result.Return() {
			return result, nil
		}
	}
	if cluster.IsManaged() {
		if err := r.clusterValidator.ValidateManagementClusterName(ctx, log, cluster); err != nil {
			cluster.Status.FailureMessage = ptr.String(err.Error())
			return controller.Result{}, err
		}
	}

	return controller.Result{}, nil
}

func (r *ClusterReconciler) postClusterProviderReconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	if cluster.HasAWSIamConfig() {
		if result, err := r.awsIamAuth.Reconcile(ctx, log, cluster); err != nil {
			return controller.Result{}, err
		} else if result.Return() {
			return result, nil
		}
	}

	return controller.Result{}, nil
}

func (r *ClusterReconciler) reconcileDelete(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (ctrl.Result, error) {
	if cluster.IsSelfManaged() {
		return ctrl.Result{}, errors.New("deleting self-managed clusters is not supported")
	}

	if metav1.HasAnnotation(cluster.ObjectMeta, anywherev1.ManagedByCLIAnnotation) {
		log.Info("Clusters is managed by CLI, removing finalizer")
		controllerutil.RemoveFinalizer(cluster, ClusterFinalizerName)
		return ctrl.Result{}, nil
	}

	if cluster.IsReconcilePaused() {
		log.Info("Cluster reconciliation is paused, won't process cluster deletion")
		return ctrl.Result{}, nil
	}

	capiCluster := &clusterv1.Cluster{}
	capiClusterName := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: cluster.Name}
	log.Info("Deleting", "name", cluster.Name)
	err := r.client.Get(ctx, capiClusterName, capiCluster)

	switch {
	case err == nil:
		log.Info("Deleting CAPI cluster", "name", capiCluster.Name)
		if err := r.client.Delete(ctx, capiCluster); err != nil {
			log.Info("Error deleting CAPI cluster", "name", capiCluster.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
	case apierrors.IsNotFound(err):
		log.Info("Deleting EKS Anywhere cluster", "name", capiCluster.Name, "cluster.DeletionTimestamp", cluster.DeletionTimestamp, "finalizer", cluster.Finalizers)

		// TODO delete GitOps,Datacenter and MachineConfig objects
		controllerutil.RemoveFinalizer(cluster, ClusterFinalizerName)
	default:
		return ctrl.Result{}, err

	}

	if cluster.HasAWSIamConfig() {
		if err := r.awsIamAuth.ReconcileDelete(ctx, log, cluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) ensureClusterOwnerReferences(ctx context.Context, clus *anywherev1.Cluster) error {
	builder := cluster.NewDefaultConfigClientBuilder()
	config, err := builder.Build(ctx, clientutil.NewKubeClient(r.client), clus)
	if err != nil {
		var notFound apierrors.APIStatus
		if apierrors.IsNotFound(err) && errors.As(err, &notFound) {
			clus.Status.FailureMessage = ptr.String(fmt.Sprintf("Dependent cluster objects don't exist: %s", notFound))
		}
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
		if apierrors.IsNotFound(err) {
			clus.Status.FailureMessage = ptr.String(fmt.Sprintf("Management cluster %s does not exist", clus.Spec.ManagementCluster.Name))
		}
		return err
	}
	clus.Spec.BundlesRef = mgmtCluster.Spec.BundlesRef
	return nil
}
