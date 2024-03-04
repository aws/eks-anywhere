package controllers

import (
	"context"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/handlers"
)

const IAMConfigClusterFinalizerName = "clusters.anywhere.eks.amazonaws.com/iam-finalizer"

// IAMConfigClusterReconciler reconciles the IAM installation for Clusters.
type IAMConfigClusterReconciler struct {
	client     client.Client
	awsIamAuth AWSIamConfigReconciler
}

// AWSIamConfigReconciler manages aws-iam-authenticator installation and configuration for an eks-a cluster.
type AWSIamConfigReconciler interface {
	EnsureCASecret(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error)
	Reconcile(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error)
	ReconcileDelete(ctx context.Context, logger logr.Logger, cluster *anywherev1.Cluster) error
}

// SetupWithManager sets up the controller with the Manager.
func (r *IAMConfigClusterReconciler) SetupWithManager(mgr ctrl.Manager, log logr.Logger) error {
	childObjectHandler := handlers.ChildObjectToClusters(log)

	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.Cluster{}).
		Watches(
			&source.Kind{Type: &anywherev1.AWSIamConfig{}},
			handler.EnqueueRequestsFromMapFunc(childObjectHandler),
		).
		Complete(r)
}

// Reconcile installs AWS IAM components in a cluster. It implements controller-runtime Reconciler.
func (r *IAMConfigClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
	log := ctrl.LoggerFrom(ctx)
	cluster := &anywherev1.Cluster{}
	if err := r.client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if cluster.HasAWSIamConfig() {
		return ctrl.Result{}, nil
	}

	patchHelper, err := patch.NewHelper(cluster, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if err := patchCluster(ctx, patchHelper, cluster); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// AddFinalizer	is idempotent
	controllerutil.AddFinalizer(cluster, ClusterFinalizerName)

	if !cluster.DeletionTimestamp.IsZero() {
		result, reterr = r.reconcileDelete(ctx, log, cluster)
		return result, reterr
	}

	result, reterr = r.reconcile(ctx, log, cluster)
	return result, reterr
}

func (r *IAMConfigClusterReconciler) reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (ctrl.Result, error) {
	if result, err := r.awsIamAuth.EnsureCASecret(ctx, log, cluster); err != nil {
		return ctrl.Result{}, err
	} else if result.Return() {
		return result.ToCtrlResult(), nil
	}

	if conditions.IsFalse(cluster, anywherev1.ControlPlaneReadyCondition) {
		return ctrl.Result{}, nil
	}

	result, err := r.awsIamAuth.Reconcile(ctx, log, cluster)

	return result.ToCtrlResult(), err
}

func (r *IAMConfigClusterReconciler) reconcileDelete(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (ctrl.Result, error) {
	if err := r.awsIamAuth.ReconcileDelete(ctx, log, cluster); err != nil {
		return ctrl.Result{}, err
	}

	controllerutil.RemoveFinalizer(cluster, ClusterFinalizerName)

	return ctrl.Result{}, nil
}
