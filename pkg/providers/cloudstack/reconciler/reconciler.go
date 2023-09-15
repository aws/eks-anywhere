package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack"
)

// IPValidator is an interface that defines methods to validate the control plane IP.
type IPValidator interface {
	ValidateControlPlaneIP(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error)
}

// CNIReconciler is an interface for reconciling CNI in the CloudStack cluster reconciler.
type CNIReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *c.Spec) (controller.Result, error)
}

// RemoteClientRegistry is an interface that defines methods for remote clients.
type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// Reconciler for CloudStack.
type Reconciler struct {
	client               client.Client
	ipValidator          IPValidator
	cniReconciler        CNIReconciler
	remoteClientRegistry RemoteClientRegistry
	validatorRegistry    cloudstack.ValidatorRegistry
}

// New defines a new CloudStack reconciler.
func New(client client.Client, ipValidator IPValidator, cniReconciler CNIReconciler, remoteClientRegistry RemoteClientRegistry, validatorRegistry cloudstack.ValidatorRegistry) *Reconciler {
	return &Reconciler{
		client:               client,
		ipValidator:          ipValidator,
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		validatorRegistry:    validatorRegistry,
	}
}

// Reconcile reconciles cluster to desired state.
func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "cloudstack")
	clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner[*c.Spec]().Register(
		r.ipValidator.ValidateControlPlaneIP,
		r.ValidateDatacenterConfig,
		r.ValidateMachineConfig,
		clusters.CleanupStatusAfterValidate,
		r.ReconcileControlPlane,
		r.CheckControlPlaneReady,
		r.ReconcileCNI,
		r.ReconcileWorkers,
	).Run(ctx, log, clusterSpec)
}

// ValidateDatacenterConfig updates the cluster status if the CloudStackDatacenter status indicates that the spec is invalid.
func (r *Reconciler) ValidateDatacenterConfig(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "validateDatacenterConfig")
	log.Info("Validating datacenter config")
	dataCenterConfig := spec.CloudStackDatacenter

	if dataCenterConfig.Status.SpecValid {
		return controller.Result{}, nil
	}
	if dataCenterConfig.Status.FailureMessage != nil {
		failureMessage := fmt.Sprintf("Invalid %s CloudStackDatacenterConfig: %s", dataCenterConfig.Name, *dataCenterConfig.Status.FailureMessage)
		spec.Cluster.SetFailure(anywherev1.DatacenterConfigInvalidReason, failureMessage)

		log.Error(errors.New(*dataCenterConfig.Status.FailureMessage), "Invalid CloudStackDatacenterConfig", "datacenterConfig", klog.KObj(dataCenterConfig))
	} else {
		log.Info("CloudStackDatacenterConfig hasn't been validated yet", klog.KObj(dataCenterConfig))
	}
	return controller.ResultWithReturn(), nil
}

// ValidateMachineConfig performs additional, context-aware validations on the machine configs.
func (r *Reconciler) ValidateMachineConfig(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "validateMachineConfigs")
	log.Info("Validating machine config")

	datacenterConfig := spec.CloudStackDatacenter
	execConfig, err := cloudstack.GetCloudstackExecConfig(ctx, r.client, datacenterConfig)
	if err != nil {
		return controller.Result{}, err
	}
	validator, err := r.validatorRegistry.Get(execConfig)
	if err != nil {
		return controller.Result{}, err
	}
	if err = validator.ValidateClusterMachineConfigs(ctx, spec); err != nil {
		log.Error(err, "Invalid CloudStackMachineConfig")
		failureMessage := err.Error()
		spec.Cluster.SetFailure(anywherev1.MachineConfigInvalidReason, failureMessage)

		return controller.ResultWithReturn(), nil
	}
	return controller.Result{}, nil
}

// ReconcileControlPlane applies the control plane CAPI objects to the cluster.
func (r *Reconciler) ReconcileControlPlane(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileControlPlane")
	log.Info("Applying control plane CAPI objects")
	cp, err := cloudstack.ControlPlaneSpec(ctx, log, clientutil.NewKubeClient(r.client), spec)
	if err != nil {
		return controller.Result{}, err
	}

	return clusters.ReconcileControlPlane(ctx, log, r.client, &clusters.ControlPlane{
		Cluster:                     cp.Cluster,
		ProviderCluster:             cp.ProviderCluster,
		KubeadmControlPlane:         cp.KubeadmControlPlane,
		ControlPlaneMachineTemplate: cp.ControlPlaneMachineTemplate,
		EtcdCluster:                 cp.EtcdCluster,
		EtcdMachineTemplate:         cp.EtcdMachineTemplate,
	})
}

// CheckControlPlaneReady checks whether the control plane for an eks-a cluster is ready or not.
// Requeues with the appropriate wait times whenever the control plane is not ready yet.
func (r *Reconciler) CheckControlPlaneReady(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "checkControlPlaneReady")
	return clusters.CheckControlPlaneReady(ctx, r.client, log, spec.Cluster)
}

// ReconcileWorkerNodes validates the cluster definition and reconciles the worker nodes
// to the desired state.
func (r *Reconciler) ReconcileWorkerNodes(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "cloudstack", "reconcile type", "workers")
	clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, errors.Wrap(err, "building cluster Spec for worker node reconcile")
	}

	return controller.NewPhaseRunner[*c.Spec]().Register(
		r.ReconcileWorkers,
	).Run(ctx, log, clusterSpec)
}

// ReconcileWorkers applies the worker CAPI objects to the cluster.
func (r *Reconciler) ReconcileWorkers(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileWorkers")
	log.Info("Applying worker CAPI objects")

	w, err := cloudstack.WorkersSpec(ctx, log, clientutil.NewKubeClient(r.client), clusterSpec)
	if err != nil {
		return controller.Result{}, errors.Wrap(err, "Generate worker node CAPI spec")
	}

	return clusters.ReconcileWorkersForEKSA(ctx, log, r.client, clusterSpec.Cluster, clusters.ToWorkers(w))
}

// ReconcileCNI reconciles the CNI to the desired state.
func (r *Reconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileCNI")
	client, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(clusterSpec.Cluster))
	if err != nil {
		return controller.Result{}, err
	}

	return r.cniReconciler.Reconcile(ctx, log, client, clusterSpec)
}
