package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/controller/handlers"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	defaultRequeueTime = 10 * time.Second
	// ClusterFinalizerName is the finalizer added to clusters to handle deletion.
	ClusterFinalizerName = "clusters.anywhere.eks.amazonaws.com/finalizer"
)

// ClusterReconciler reconciles a Cluster object.
type ClusterReconciler struct {
	client                     client.Client
	providerReconcilerRegistry ProviderClusterReconcilerRegistry
	awsIamAuth                 AWSIamConfigReconciler
	clusterValidator           ClusterValidator
	packagesClient             PackagesClient

	// experimentalSelfManagedUpgrade enables management cluster full upgrades.
	// The default behavior for management cluster only reconciles the worker nodes.
	// When this is enabled, the controller will handle management clusters in the same
	// way as workload clusters: it will reconcile CP, etcd and workers.
	// Only intended for internal testing.
	experimentalSelfManagedUpgrade bool
}

// PackagesClient handles curated packages operations from within the cluster
// controller.
type PackagesClient interface {
	EnableFullLifecycle(ctx context.Context, log logr.Logger, clusterName, kubeConfig string, chart *v1alpha1.Image, registry *registrymirror.RegistryMirror, options ...curatedpackages.PackageControllerClientOpt) error
	ReconcileDelete(context.Context, logr.Logger, curatedpackages.KubeDeleter, *anywherev1.Cluster) error
	Reconcile(context.Context, logr.Logger, client.Client, *anywherev1.Cluster) error
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

// ClusterReconcilerOption allows to configure the ClusterReconciler.
type ClusterReconcilerOption func(*ClusterReconciler)

// NewClusterReconciler constructs a new ClusterReconciler.
func NewClusterReconciler(client client.Client, registry ProviderClusterReconcilerRegistry, awsIamAuth AWSIamConfigReconciler, clusterValidator ClusterValidator, pkgs PackagesClient, opts ...ClusterReconcilerOption) *ClusterReconciler {
	c := &ClusterReconciler{
		client:                     client,
		providerReconcilerRegistry: registry,
		awsIamAuth:                 awsIamAuth,
		clusterValidator:           clusterValidator,
		packagesClient:             pkgs,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithExperimentalSelfManagedClusterUpgrades allows to enable experimental upgrades for self
// managed clusters.
func WithExperimentalSelfManagedClusterUpgrades(exp bool) ClusterReconcilerOption {
	return func(c *ClusterReconciler) {
		c.experimentalSelfManagedUpgrade = exp
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
		Watches(
			&source.Kind{Type: &anywherev1.CloudStackDatacenterConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.CloudStackMachineConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.NutanixDatacenterConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Watches(
			&source.Kind{Type: &anywherev1.NutanixMachineConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Complete(r)
}

// Reconcile reconciles a cluster object.
// nolint:gocyclo //TODO: Reduce high cycomatic complexity.
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters;snowmachineconfigs;snowippools;vspheredatacenterconfigs;vspheremachineconfigs;dockerdatacenterconfigs;tinkerbellmachineconfigs;tinkerbelldatacenterconfigs;cloudstackdatacenterconfigs;cloudstackmachineconfigs;nutanixdatacenterconfigs;nutanixmachineconfigs;bundles;awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=oidcconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/status;snowmachineconfigs/status;snowippools/status;vspheredatacenterconfigs/status;vspheremachineconfigs/status;dockerdatacenterconfigs/status;cloudstackdatacenterconfigs/status;cloudstackmachineconfigs/status;nutanixdatacenterconfigs/status;nutanixmachineconfigs/status;bundles/status;awsiamconfigs/status;tinkerbelldatacenterconfigs/status;tinkerbellmachineconfigs/status,verbs=;get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/finalizers;snowmachineconfigs/finalizers;snowippools/finalizers;vspheredatacenterconfigs/finalizers;vspheremachineconfigs/finalizers;dockerdatacenterconfigs/finalizers;cloudstackdatacenterconfigs/finalizers;cloudstackmachineconfigs/finalizers;nutanixdatacenterconfigs/finalizers;nutanixmachineconfigs/finalizers;bundles/finalizers;awsiamconfigs/finalizers;tinkerbelldatacenterconfigs/finalizers;tinkerbellmachineconfigs/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=addons.cluster.x-k8s.io,resources=clusterresourcesets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=test,resources=test,verbs=get;list;watch;create;update;patch;delete;kill
// +kubebuilder:rbac:groups=distro.eks.amazonaws.com,resources=releases,verbs=get;list;watch
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=awssnowclusters;awssnowmachinetemplates;awssnowippools;vsphereclusters;vspheremachinetemplates;dockerclusters;dockermachinetemplates;tinkerbellclusters;tinkerbellmachinetemplates;cloudstackclusters;cloudstackmachinetemplates;nutanixclusters;nutanixmachinetemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",namespace=eksa-system,resources=secrets,verbs=delete;
// +kubebuilder:rbac:groups=tinkerbell.org,resources=hardware;hardware/status,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=bmc.tinkerbell.org,resources=machines;machines/status,verbs=get;list;watch
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
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
		err := r.updateStatus(ctx, log, cluster)
		if err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		args := make([]string, 0, len(cluster.Status.Conditions))
		for _, condition := range cluster.Status.Conditions {
			args = append(args, fmt.Sprintf("%s=%s %s", condition.Type, condition.Status, condition.Reason))
		}

		log.Info("Current conditions", "conditions", strings.Join(args, ", "))

		// Always attempt to patch the object and status after each reconciliation.
		patchOpts := []patch.Option{}

		// Patch ObservedGeneration only if the reconciliation completed without error
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchCluster(ctx, patchHelper, cluster, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Only requeue if we are not already re-queueing and the Cluster ready condition is false
		if reterr == nil && !result.Requeue && result.RequeueAfter <= 0 && conditions.IsFalse(cluster, anywherev1.ReadyCondition) {
			result = ctrl.Result{RequeueAfter: defaultRequeueTime}
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

	config, err := r.buildClusterConfig(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err = r.ensureClusterOwnerReferences(ctx, cluster, config); err != nil {
		return ctrl.Result{}, err
	}

	aggregatedGeneration := aggregatedGeneration(config)

	// If there is no difference between the aggregated generation and childrenReconciledGeneration,
	// and there is no difference in the reconciled generation and .metadata.generation of the cluster,
	// then return without any further processing.
	if aggregatedGeneration == cluster.Status.ChildrenReconciledGeneration && cluster.Status.ReconciledGeneration == cluster.Generation {
		log.Info("Generation and aggregated generation match reconciled generations for cluster and child objects, skipping reconciliation.")
		return ctrl.Result{}, nil
	}

	return r.reconcile(ctx, log, cluster, aggregatedGeneration)
}

func (r *ClusterReconciler) reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster, aggregatedGeneration int64) (ctrl.Result, error) {
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

	if cluster.IsSelfManaged() && !r.experimentalSelfManagedUpgrade {
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

	reconcileResult, err = r.postClusterProviderReconcile(ctx, log, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	if reconcileResult.Return() {
		return reconcileResult.ToCtrlResult(), nil
	}

	// At the end of the reconciliation, if there have been no requeues or errors, we update the cluster's status.
	// NOTE: This update must be the last step in the reconciliation process to denote the complete reconciliation.
	// No other mutating changes or reconciliations must happen in this loop after this step, so all such changes must
	// be placed above this line.
	cluster.Status.ReconciledGeneration = cluster.Generation
	cluster.Status.ChildrenReconciledGeneration = aggregatedGeneration

	return ctrl.Result{}, nil
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

	if cluster.RegistryAuth() {
		rUsername, rPassword, err := config.ReadCredentialsFromSecret(ctx, r.client)
		if err != nil {
			return controller.Result{}, err
		}

		if err := config.SetCredentialsEnv(rUsername, rPassword); err != nil {
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

	// Self-managed clusters can support curated packages, but that support
	// comes from the CLI at this time.
	if cluster.IsManaged() && cluster.IsPackagesEnabled() {
		if err := r.packagesClient.Reconcile(ctx, log, r.client, cluster); err != nil {
			return controller.Result{}, err
		}
	}

	return controller.Result{}, nil
}

func (r *ClusterReconciler) updateStatus(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	// An EKS-A Cluster managed by the CLI is deleted after the cluster finalizer is removed, leaving the CAPI cluster behind. In this case, we do not want to update
	// the status. If we do, patching will throw a 404 NotFound error when trying to update the status of the already deleted
	// Cluster object.
	if !cluster.DeletionTimestamp.IsZero() && metav1.HasAnnotation(cluster.ObjectMeta, anywherev1.ManagedByCLIAnnotation) {
		log.Info("Cluster managed by CLI has been deleted but CAPI cluster may remain, skipping updating cluster status")
		return nil
	}

	// ObservedGeneration represents the last observed generation, and consequently represents that the current status is up to date.
	// Then if observedGeneration is equal to generation AND the Cluster's "Ready" condition is "True", then we can skip another status update
	if cluster.Status.ObservedGeneration == cluster.Generation && conditions.IsTrue(cluster, anywherev1.ReadyCondition) {
		log.Info("Generation matches observedGeneration and the cluster is ready, skipping status update")
		return nil
	}

	log.Info("Updating cluster status")
	defer func() {
		// Always update the readyCondition by summarizing the state of other conditions.
		conditions.SetSummary(cluster,
			conditions.WithConditions(
				anywherev1.ControlPlaneReadyCondition,
				anywherev1.WorkersReadyConditon,
			),
		)
	}()

	if err := r.updateControlPlaneStatus(ctx, log, cluster); err != nil {
		return errors.Wrap(err, "updating controlplane status")
	}

	if err := r.updateWorkersStatus(ctx, log, cluster); err != nil {
		return errors.Wrap(err, "updating workers status")
	}

	updateCNIStatus(log, cluster)

	return nil
}

func (r *ClusterReconciler) updateControlPlaneStatus(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	log.Info("Updating control plane status")
	kcp, err := controller.GetKubeadmControlPlane(ctx, r.client, cluster)
	if err != nil {
		return errors.Wrapf(err, "getting kubeadmcontrolplane")
	}

	if err = clusters.UpdateControlPlaneInitializedCondition(cluster, kcp); err != nil {
		return errors.Wrap(err, "updating control plane initialized condition")
	}

	if err = clusters.UpdateControlPlaneReadyCondition(cluster, kcp); err != nil {
		return errors.Wrap(err, "updating control plane ready condition")
	}

	return nil
}

func (r *ClusterReconciler) updateWorkersStatus(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) error {
	log.Info("Updating workers status")
	machineDeployments, err := controller.GetMachineDeployments(ctx, r.client, cluster)
	if err != nil {
		return errors.Wrap(err, "getting machine deployments")
	}

	if err = clusters.UpdateWorkersReadyCondition(cluster, machineDeployments); err != nil {
		return errors.Wrap(err, "updating workers ready condition")
	}

	return nil
}

func updateCNIStatus(log logr.Logger, cluster *anywherev1.Cluster) {
	// Initialize DefaultCNIConfiguredCondition condition, so that it appears in the status sooner rather than later
	// The CNI reconciler handles the rest of the logic for determining the condition and updating the status.
	if conditions.Get(cluster, anywherev1.DefaultCNIConfiguredCondition) == nil {
		log.Info("Initializing CNI status")

		conditions.MarkFalse(
			cluster,
			anywherev1.DefaultCNIConfiguredCondition,
			anywherev1.WaitingForDefaultCNIConfiguredReason,
			clusterv1.ConditionSeverityInfo,
			"Waiting for default CNI to be configured",
		)
	}

	// TODO: Remove after self-managed clusters are created with the controller in the CLI
	// Self managed clusters do not use the CNI reconciler, so this status would never get resolved.
	if cluster.IsSelfManaged() {
		log.Info("Updating CNI status")
		clusters.UpdateSelfManagedClusterDefaultCNIConfiguredCondition(cluster)
	}
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

	if cluster.IsManaged() {
		if err := r.packagesClient.ReconcileDelete(ctx, log, r.client, cluster); err != nil {
			return ctrl.Result{}, fmt.Errorf("deleting packages for cluster %q: %w", cluster.Name, err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) buildClusterConfig(ctx context.Context, clus *anywherev1.Cluster) (*cluster.Config, error) {
	builder := cluster.NewDefaultConfigClientBuilder()
	config, err := builder.Build(ctx, clientutil.NewKubeClient(r.client), clus)
	if err != nil {
		var notFound apierrors.APIStatus
		if apierrors.IsNotFound(err) && errors.As(err, &notFound) {
			clus.Status.FailureMessage = ptr.String(fmt.Sprintf("Dependent cluster objects don't exist: %s", notFound))
		}
		return nil, err
	}

	return config, nil
}

func (r *ClusterReconciler) ensureClusterOwnerReferences(ctx context.Context, clus *anywherev1.Cluster, config *cluster.Config) error {
	for _, obj := range config.ChildObjects() {
		numberOfOwnerReferences := len(obj.GetOwnerReferences())
		if err := controllerutil.SetOwnerReference(clus, obj, r.client.Scheme()); err != nil {
			return errors.Wrapf(err, "setting cluster owner reference for %s", obj.GetObjectKind())
		}

		if numberOfOwnerReferences == len(obj.GetOwnerReferences()) {
			// obj already had the owner reference
			continue
		}

		if err := r.client.Update(ctx, obj); err != nil {
			return errors.Wrapf(err, "updating object (%s) with cluster owner reference", obj.GetObjectKind())
		}
	}

	return nil
}

func patchCluster(ctx context.Context, patchHelper *patch.Helper, cluster *anywherev1.Cluster, patchOpts ...patch.Option) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	options := append([]patch.Option{
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			// TODO: Add each condition her that the controller should ignored conflicts for.
			anywherev1.ReadyCondition,
			anywherev1.ControlPlaneInitializedCondition,
			anywherev1.ControlPlaneReadyCondition,
			anywherev1.WorkersReadyConditon,
			anywherev1.DefaultCNIConfiguredCondition,
		}},
	}, patchOpts...)

	// Always attempt to patch the object and status after each reconciliation.
	return patchHelper.Patch(ctx, cluster, options...)
}

// aggregatedGeneration computes the combined generation of the resources linked
// by the cluster by summing up the .metadata.generation value for all the child
// objects of this cluster.
func aggregatedGeneration(config *cluster.Config) int64 {
	var aggregatedGeneration int64
	for _, obj := range config.ChildObjects() {
		aggregatedGeneration += obj.GetGeneration()
	}

	return aggregatedGeneration
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
