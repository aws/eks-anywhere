package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
)

type CNIReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *cluster.Spec) (controller.Result, error)
}

type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// IPValidator defines an interface for the methods to validate the control plane IP.
type IPValidator interface {
	ValidateControlPlaneIP(ctx context.Context, log logr.Logger, spec *cluster.Spec) (controller.Result, error)
}

type Reconciler struct {
	client               client.Client
	cniReconciler        CNIReconciler
	remoteClientRegistry RemoteClientRegistry
	ipValidator          IPValidator
	*serverside.ObjectApplier
}

// New initializes a new reconciler for the Snow provider.
func New(client client.Client, cniReconciler CNIReconciler, remoteClientRegistry RemoteClientRegistry, ipValidator IPValidator) *Reconciler {
	return &Reconciler{
		client:               client,
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		ipValidator:          ipValidator,
		ObjectApplier:        serverside.NewObjectApplier(client),
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, c *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "snow")
	clusterSpec, err := cluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), c)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner().Register(
		r.ipValidator.ValidateControlPlaneIP,
		r.ValidateMachineConfigs,
		clusters.CleanupStatusAfterValidate,
		r.ReconcileControlPlane,
		r.CheckControlPlaneReady,
		r.ReconcileCNI,
		r.ReconcileWorkers,
	).Run(ctx, log, clusterSpec)
}

// ReconcileWorkerNodes validates the cluster definition and reconciles the worker nodes
// to the desired state.
func (r *Reconciler) ReconcileWorkerNodes(ctx context.Context, log logr.Logger, c *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "vsphere", "reconcile type", "workers")
	clusterSpec, err := cluster.BuildSpec(ctx, clientutil.NewKubeClient(r.client), c)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner().Register(
		r.ValidateMachineConfigs,
		r.ReconcileWorkers,
	).Run(ctx, log, clusterSpec)
}

func (r *Reconciler) ValidateMachineConfigs(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "validateMachineConfigs")
	for _, machineConfig := range clusterSpec.SnowMachineConfigs {
		if !machineConfig.Status.SpecValid {
			if machineConfig.Status.FailureMessage != nil {
				failureMessage := fmt.Sprintf("Invalid %s SnowMachineConfig: %s", machineConfig.Name, *machineConfig.Status.FailureMessage)
				clusterSpec.Cluster.Status.FailureMessage = &failureMessage
				log.Error(errors.New(*machineConfig.Status.FailureMessage), "Invalid SnowMachineConfig", "machineConfig", klog.KObj(machineConfig))
			} else {
				log.Info("SnowMachineConfig hasn't been validated yet", "machineConfig", klog.KObj(machineConfig))
			}

			return controller.ResultWithReturn(), nil
		}
	}

	return controller.Result{}, nil
}

func (s *Reconciler) ReconcileControlPlane(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileControlPlane")
	log.Info("Applying control plane CAPI objects")

	return s.Apply(ctx, func() ([]kubernetes.Object, error) {
		return snow.ControlPlaneObjects(ctx, log, clusterSpec, clientutil.NewKubeClient(s.client))
	})
}

func (r *Reconciler) CheckControlPlaneReady(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "checkControlPlaneReady")
	return clusters.CheckControlPlaneReady(ctx, r.client, log, clusterSpec.Cluster)
}

func (s *Reconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileCNI")

	client, err := s.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(clusterSpec.Cluster))
	if err != nil {
		return controller.Result{}, err
	}

	return s.cniReconciler.Reconcile(ctx, log, client, clusterSpec)
}

func (s *Reconciler) ReconcileWorkers(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileWorkers")
	log.Info("Applying worker CAPI objects")

	w, err := snow.WorkersSpec(ctx, log, clusterSpec, clientutil.NewKubeClient(s.client))
	if err != nil {
		return controller.Result{}, err
	}

	return clusters.ReconcileWorkersForEKSA(ctx, log, s.client, clusterSpec.Cluster, clusters.ToWorkers(w))
}
