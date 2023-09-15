package reconciler

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/controller/serverside"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

// CNIReconciler is an interface for reconciling CNI in the VSphere cluster reconciler.
type CNIReconciler interface {
	Reconcile(ctx context.Context, logger logr.Logger, client client.Client, spec *c.Spec) (controller.Result, error)
}

// RemoteClientRegistry is an interface that defines methods for remote clients.
type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// IPValidator is an interface that defines methods to validate the control plane IP.
type IPValidator interface {
	ValidateControlPlaneIP(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error)
}

type Reconciler struct {
	client               client.Client
	validator            *vsphere.Validator
	defaulter            *vsphere.Defaulter
	cniReconciler        CNIReconciler
	remoteClientRegistry RemoteClientRegistry
	ipValidator          IPValidator
	*serverside.ObjectApplier
}

// New defines a new VSphere reconciler.
func New(client client.Client, validator *vsphere.Validator, defaulter *vsphere.Defaulter, cniReconciler CNIReconciler, remoteClientRegistry RemoteClientRegistry, ipValidator IPValidator) *Reconciler {
	return &Reconciler{
		client:               client,
		validator:            validator,
		defaulter:            defaulter,
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		ipValidator:          ipValidator,
		ObjectApplier:        serverside.NewObjectApplier(client),
	}
}

func VsphereCredentials(ctx context.Context, cli client.Client) (*apiv1.Secret, error) {
	secret := &apiv1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: "eksa-system",
		Name:      vsphere.CredentialsObjectName,
	}
	if err := cli.Get(ctx, secretKey, secret); err != nil {
		return nil, err
	}
	return secret, nil
}

func SetupEnvVars(ctx context.Context, vsphereDatacenter *anywherev1.VSphereDatacenterConfig, cli client.Client) error {
	secret, err := VsphereCredentials(ctx, cli)
	if err != nil {
		return fmt.Errorf("failed getting vsphere credentials secret: %v", err)
	}

	vsphereUsername := secret.Data["username"]
	vspherePassword := secret.Data["password"]

	if err := os.Setenv(config.EksavSphereUsernameKey, string(vsphereUsername)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", config.EksavSphereUsernameKey, err)
	}

	if err := os.Setenv(config.EksavSpherePasswordKey, string(vspherePassword)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", config.EksavSpherePasswordKey, err)
	}

	vsphereCPUsername := secret.Data["usernameCP"]
	vsphereCPPassword := secret.Data["passwordCP"]

	if err := os.Setenv(config.EksavSphereCPUsernameKey, string(vsphereCPUsername)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", config.EksavSphereCPUsernameKey, err)
	}

	if err := os.Setenv(config.EksavSphereCPPasswordKey, string(vsphereCPPassword)); err != nil {
		return fmt.Errorf("failed setting env %s: %v", config.EksavSphereCPPasswordKey, err)
	}

	if err := vsphere.SetupEnvVars(vsphereDatacenter); err != nil {
		return fmt.Errorf("failed setting env vars: %v", err)
	}

	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "vsphere")
	clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner[*c.Spec]().Register(
		r.ipValidator.ValidateControlPlaneIP,
		r.ValidateDatacenterConfig,
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
func (r *Reconciler) ReconcileWorkerNodes(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "vsphere", "reconcile type", "workers")
	clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner[*c.Spec]().Register(
		r.ValidateDatacenterConfig,
		r.ValidateMachineConfigs,
		r.ReconcileWorkers,
	).Run(ctx, log, clusterSpec)
}

// ValidateDatacenterConfig updates the cluster status if the VSphereDatacenter status indicates that the spec is invalid.
func (r *Reconciler) ValidateDatacenterConfig(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "validateDatacenterConfig")
	dataCenterConfig := clusterSpec.VSphereDatacenter

	if !dataCenterConfig.Status.SpecValid {
		if dataCenterConfig.Status.FailureMessage != nil {
			failureMessage := fmt.Sprintf("Invalid %s VSphereDatacenterConfig: %s", dataCenterConfig.Name, *dataCenterConfig.Status.FailureMessage)

			clusterSpec.Cluster.SetFailure(anywherev1.DatacenterConfigInvalidReason, failureMessage)
			log.Error(errors.New(*dataCenterConfig.Status.FailureMessage), "Invalid VSphereDatacenterConfig", "datacenterConfig", klog.KObj(dataCenterConfig))
		} else {
			log.Info("VSphereDatacenterConfig hasn't been validated yet", klog.KObj(dataCenterConfig))
		}

		return controller.ResultWithReturn(), nil
	}
	return controller.Result{}, nil
}

// ValidateMachineConfigs performs additional, context-aware validations on the machine configs.
func (r *Reconciler) ValidateMachineConfigs(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "validateMachineConfigs")
	datacenterConfig := clusterSpec.VSphereDatacenter

	// Set up env vars for executing Govc cmd
	if err := SetupEnvVars(ctx, datacenterConfig, r.client); err != nil {
		log.Error(err, "Failed to set up env vars for Govc")
		return controller.Result{}, err
	}

	vsphereClusterSpec := vsphere.NewSpec(clusterSpec)

	if err := r.validator.ValidateClusterMachineConfigs(ctx, vsphereClusterSpec); err != nil {
		log.Error(err, "Invalid VSphereMachineConfig")
		failureMessage := err.Error()
		clusterSpec.Cluster.SetFailure(anywherev1.MachineConfigInvalidReason, failureMessage)
		return controller.ResultWithReturn(), nil
	}
	return controller.Result{}, nil
}

// ReconcileControlPlane applies the control plane CAPI objects to the cluster.
func (r *Reconciler) ReconcileControlPlane(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileControlPlane")
	log.Info("Applying control plane CAPI objects")
	cp, err := vsphere.ControlPlaneSpec(ctx, log, clientutil.NewKubeClient(r.client), spec)
	if err != nil {
		return controller.Result{}, err
	}

	return clusters.ReconcileControlPlane(ctx, log, r.client, toClientControlPlane(cp))
}

// CheckControlPlaneReady checks whether the control plane for an eks-a cluster is ready or not.
// Requeues with the appropriate wait times whenever the cluster is not ready yet.
func (r *Reconciler) CheckControlPlaneReady(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "checkControlPlaneReady")
	return clusters.CheckControlPlaneReady(ctx, r.client, log, clusterSpec.Cluster)
}

// ReconcileCNI takes the Cilium CNI in a cluster to the desired state defined in a cluster spec.
func (r *Reconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileCNI")
	client, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(clusterSpec.Cluster))
	if err != nil {
		return controller.Result{}, err
	}

	return r.cniReconciler.Reconcile(ctx, log, client, clusterSpec)
}

// ReconcileWorkers applies the worker CAPI objects to the cluster.
func (r *Reconciler) ReconcileWorkers(ctx context.Context, log logr.Logger, spec *c.Spec) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileWorkers")
	log.Info("Applying worker CAPI objects")
	w, err := vsphere.WorkersSpec(ctx, log, clientutil.NewKubeClient(r.client), spec)
	if err != nil {
		return controller.Result{}, err
	}

	return clusters.ReconcileWorkersForEKSA(ctx, log, r.client, spec.Cluster, clusters.ToWorkers(w))
}

func toClientControlPlane(cp *vsphere.ControlPlane) *clusters.ControlPlane {
	other := make([]client.Object, 0, len(cp.ConfigMaps)+len(cp.Secrets)+len(cp.ClusterResourceSets)+1)
	for _, o := range cp.ClusterResourceSets {
		other = append(other, o)
	}
	for _, o := range cp.ConfigMaps {
		other = append(other, o)
	}
	for _, o := range cp.Secrets {
		other = append(other, o)
	}

	return &clusters.ControlPlane{
		Cluster:                     cp.Cluster,
		ProviderCluster:             cp.ProviderCluster,
		KubeadmControlPlane:         cp.KubeadmControlPlane,
		ControlPlaneMachineTemplate: cp.ControlPlaneMachineTemplate,
		EtcdCluster:                 cp.EtcdCluster,
		EtcdMachineTemplate:         cp.EtcdMachineTemplate,
		Other:                       other,
	}
}
