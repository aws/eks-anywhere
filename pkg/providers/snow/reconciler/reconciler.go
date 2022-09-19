package reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

type Reconciler struct {
	client               client.Client
	cniReconciler        CNIReconciler
	remoteClientRegistry RemoteClientRegistry
	*serverside.ObjectApplier
}

func New(client client.Client, cniReconciler CNIReconciler, remoteClientRegistry RemoteClientRegistry) *Reconciler {
	return &Reconciler{
		client:               client,
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
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
		r.ValidateMachineConfigs,
		r.ReconcileControlPlane,
		r.CheckControlPlaneReady,
		r.ReconcileCNI,
		r.ReconcileWorkers,
	).Run(ctx, log, clusterSpec)
}

func (r *Reconciler) ValidateMachineConfigs(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "validateMachineConfigs")
	for _, machineConfig := range clusterSpec.SnowMachineConfigs {
		if !machineConfig.Status.SpecValid {
			failureMessage := fmt.Sprintf("SnowMachineConfig %s is invalid", machineConfig.Name)
			if machineConfig.Status.FailureMessage != nil {
				failureMessage += ": " + *machineConfig.Status.FailureMessage
			}

			log.Error(nil, failureMessage)
			clusterSpec.Cluster.Status.FailureMessage = &failureMessage
			return controller.Result{}, nil
		}
	}

	return controller.Result{}, nil
}

func (r *Reconciler) ReconcileControlPlane(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileControlPlane")
	log.Info("Applying control plane CAPI objects")

	// Ensure that the control plane machine config exists before reconciling
	cpMachineConfig := types.NamespacedName{Name: clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name, Namespace: clusterSpec.Cluster.Namespace}
	if err := r.getSnowMachineConfig(ctx, cpMachineConfig); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Snow machine config does not exist yet, requeuing", "machine config", cpMachineConfig.Name)
			return controller.ResultWithRequeue(5 * time.Second), nil
		}
	}

	return r.Apply(ctx, func() ([]kubernetes.Object, error) {
		return snow.ControlPlaneObjects(ctx, clusterSpec, clientutil.NewKubeClient(r.client))
	})
}

func (r *Reconciler) CheckControlPlaneReady(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "checkControlPlaneReady")
	return clusters.CheckControlPlaneReady(ctx, r.client, log, clusterSpec.Cluster)
}

func (r *Reconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileCNI")

	client, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(clusterSpec.Cluster))
	if err != nil {
		return controller.Result{}, err
	}

	return r.cniReconciler.Reconcile(ctx, log, client, clusterSpec)
}

func (r *Reconciler) ReconcileWorkers(ctx context.Context, log logr.Logger, clusterSpec *cluster.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileWorkers")
	log.Info("Applying worker CAPI objects")

	// Ensure that the worker node machine configs exist before reconciling
	for _, workerNodeGroupConfig := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		wnMachineConfig := types.NamespacedName{Name: workerNodeGroupConfig.MachineGroupRef.Name, Namespace: clusterSpec.Cluster.Namespace}
		if err := r.getSnowMachineConfig(ctx, wnMachineConfig); err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("Snow machine config does not exist yet, requeuing", "machine config", wnMachineConfig.Name)
				return controller.ResultWithRequeue(5 * time.Second), nil
			}
		}
	}

	return r.Apply(ctx, func() ([]kubernetes.Object, error) {
		return snow.WorkersObjects(ctx, clusterSpec, clientutil.NewKubeClient(r.client))
	})
}

func (r *Reconciler) getSnowMachineConfig(ctx context.Context, machineConfigName types.NamespacedName) error {
	snowMachineConfig := &anywherev1.SnowMachineConfig{}
	if err := r.client.Get(ctx, machineConfigName, snowMachineConfig); err != nil {
		return err
	}
	return nil
}
