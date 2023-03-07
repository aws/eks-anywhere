package reconciler

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	rufiov1alpha1 "github.com/tinkerbell/rufio/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

const (
	// ReconcileNewCluster means need to create nodes for a new cluster.
	ReconcileNewCluster clusterChange = "create-new-cluster"
	// NeedNewNode means node has k8s version change.
	NeedNewNode clusterChange = "k8s-version-change"
	// NeedMoreOrLessNode means node group scales up or scales down.
	NeedMoreOrLessNode clusterChange = "scale-update"
)

// CNIReconciler is an interface for reconciling CNI in the Tinkerbell cluster reconciler.
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

type clusterChange string

// Scope object for Tinkerbell reconciler.
type Scope struct {
	ClusterSpec   *c.Spec
	ClusterChange clusterChange
	ControlPlane  *tinkerbell.ControlPlane
	Workers       *tinkerbell.Workers
}

// NewScope creates a new Tinkerbell Reconciler Scope.
func NewScope(clusterSpec *c.Spec) *Scope {
	return &Scope{
		ClusterSpec: clusterSpec,
	}
}

// Reconciler for Tinkerbell.
type Reconciler struct {
	client               client.Client
	cniReconciler        CNIReconciler
	remoteClientRegistry RemoteClientRegistry
	ipValidator          IPValidator
}

// New defines a new Tinkerbell reconciler.
func New(client client.Client, cniReconciler CNIReconciler, remoteClientRegistry RemoteClientRegistry, ipValidator IPValidator) *Reconciler {
	return &Reconciler{
		client:               client,
		cniReconciler:        cniReconciler,
		remoteClientRegistry: remoteClientRegistry,
		ipValidator:          ipValidator,
	}
}

// Reconcile reconciles cluster to desired state.
func (r *Reconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	// Implement reconcile all here.
	// This would include validating machine and datacenter configs
	// and reconciling cp and worker nodes.
	log = log.WithValues("provider", "tinkerbell")
	clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, err
	}

	return controller.NewPhaseRunner[*Scope]().Register(
		r.ValidateControlPlaneIP,
		r.ValidateClusterSpec,
		r.GenerateSpec,
		r.DetectOperation,
		r.ValidateHardware,
		r.OmitMachineTemplate,
		r.ValidateDatacenterConfig,
		r.ValidateRufioMachines,
		r.ReconcileControlPlane,
		r.CheckControlPlaneReady,
		r.ReconcileCNI,
		r.ReconcileWorkers,
	).Run(ctx, log, NewScope(clusterSpec))
}

// ValidateControlPlaneIP passes the cluster spec from tinkerbellScope to the IP Validator.
func (r *Reconciler) ValidateControlPlaneIP(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	return r.ipValidator.ValidateControlPlaneIP(ctx, log, tinkerbellScope.ClusterSpec)
}

// ValidateClusterSpec performs a set of assertions on a cluster spec.
func (r *Reconciler) ValidateClusterSpec(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	clusterSpec := tinkerbellScope.ClusterSpec
	log = log.WithValues("phase", "validateClusterSpec")

	tinkerbellClusterSpec := tinkerbell.NewClusterSpec(clusterSpec, clusterSpec.Config.TinkerbellMachineConfigs, clusterSpec.Config.TinkerbellDatacenter)

	clusterSpecValidator := tinkerbell.NewClusterSpecValidator()

	if err := clusterSpecValidator.Validate(tinkerbellClusterSpec); err != nil {
		log.Error(err, "Invalid Tinkerbell Cluster spec")
		failureMessage := err.Error()
		clusterSpec.Cluster.Status.FailureMessage = &failureMessage
		return controller.ResultWithReturn(), nil
	}
	return controller.Result{}, nil
}

// GenerateSpec generates Tinkerbell control plane and workers spec.
func (r *Reconciler) GenerateSpec(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	spec := tinkerbellScope.ClusterSpec
	log = log.WithValues("phase", "GenerateSpec")
	log.Info("Generating tinkerbell control plane and workers spec")

	cp, err := tinkerbell.ControlPlaneSpec(ctx, log, clientutil.NewKubeClient(r.client), spec)
	if err != nil {
		return controller.Result{}, err
	}
	tinkerbellScope.ControlPlane = cp

	w, err := tinkerbell.WorkersSpec(ctx, log, clientutil.NewKubeClient(r.client), spec)
	if err != nil {
		return controller.Result{}, errors.Wrap(err, "generating workers spec")
	}
	tinkerbellScope.Workers = w

	return controller.Result{}, nil
}

// DetectOperation detects change type.
func (r *Reconciler) DetectOperation(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	log = log.WithValues("phase", "DetectOperation")
	log.Info("Detecting operation type")

	currentKCP, err := controller.GetKubeadmControlPlane(ctx, r.client, tinkerbellScope.ClusterSpec.Cluster)
	if err != nil {
		return controller.Result{}, err
	}
	if currentKCP == nil {
		tinkerbellScope.ClusterChange = ReconcileNewCluster
		return controller.Result{}, nil
	}

	tinkCurrentKCP := tinkerbell.KubeadmControlPlane{KubeadmControlPlane: currentKCP}
	tinkDesiredKCP := tinkerbell.KubeadmControlPlane{KubeadmControlPlane: tinkerbellScope.ControlPlane.KubeadmControlPlane}
	wantVersionChange := tinkerbell.HasKubernetesVersionChange(&tinkCurrentKCP, &tinkDesiredKCP)
	cpWantScaleChange := tinkerbell.ReplicasDiff(&tinkCurrentKCP, &tinkDesiredKCP)
	workerWantScaleChange := r.workerReplicasDiff(ctx, tinkerbellScope)

	if wantVersionChange {
		if cpWantScaleChange != 0 || workerWantScaleChange {
			return controller.Result{}, errors.Errorf("cannot perform scale up or down during k8s version change")
		}
		tinkerbellScope.ClusterChange = NeedNewNode
	} else if cpWantScaleChange != 0 || workerWantScaleChange {
		tinkerbellScope.ClusterChange = NeedMoreOrLessNode
	} else {
		return controller.Result{}, errors.Errorf("cannot detect operation type")
	}
	return controller.Result{}, nil
}

func (r *Reconciler) workerReplicasDiff(ctx context.Context, tinkerbellScope *Scope) bool {
	workerWantScaleChange := false
	for _, wnc := range tinkerbellScope.ClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		md := &clusterv1.MachineDeployment{}
		key := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: wnc.Name}
		err := r.client.Get(ctx, key, md)
		if err != nil {
			if apierrors.IsNotFound(err) {
				workerWantScaleChange = true
				break
			}
		}
		if int(*md.Spec.Replicas) != *wnc.Count {
			workerWantScaleChange = true
			break
		}
	}
	return workerWantScaleChange
}

// OmitMachineTemplate omits control plane and worker machine template on scaling update.
func (r *Reconciler) OmitMachineTemplate(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	log = log.WithValues("phase", "OmitMachineTemplate")
	log.Info("Omit machine template on scaling")
	if tinkerbellScope.ClusterChange == NeedMoreOrLessNode {
		tinkerbell.OmitTinkerbellCPMachineTemplate(tinkerbellScope.ControlPlane)
		tinkerbell.OmitTinkerbellWorkersMachineTemplate(tinkerbellScope.Workers)
	}
	return controller.Result{}, nil
}

// ReconcileControlPlane applies the control plane CAPI objects to the cluster.
func (r *Reconciler) ReconcileControlPlane(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	log = log.WithValues("phase", "reconcileControlPlane")
	log.Info("Applying control plane CAPI objects")

	return clusters.ReconcileControlPlane(ctx, r.client, toClientControlPlane(tinkerbellScope.ControlPlane))
}

// CheckControlPlaneReady checks whether the control plane for an eks-a cluster is ready or not.
// Requeues with the appropriate wait times whenever the cluster is not ready yet.
func (r *Reconciler) CheckControlPlaneReady(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	clusterSpec := tinkerbellScope.ClusterSpec
	log = log.WithValues("phase", "checkControlPlaneReady")
	return clusters.CheckControlPlaneReady(ctx, r.client, log, clusterSpec.Cluster)
}

// ReconcileWorkerNodes reconciles the worker nodes to the desired state.
func (r *Reconciler) ReconcileWorkerNodes(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	log = log.WithValues("provider", "tinkerbell", "reconcile type", "workers")
	clusterSpec, err := c.BuildSpec(ctx, clientutil.NewKubeClient(r.client), cluster)
	if err != nil {
		return controller.Result{}, errors.Wrap(err, "building cluster Spec for worker node reconcile")
	}

	return controller.NewPhaseRunner[*Scope]().Register(
		r.ValidateClusterSpec,
		r.GenerateSpec,
		r.DetectOperation,
		r.ValidateHardware,
		r.ValidateRufioMachines,
		r.OmitMachineTemplate,
		r.ReconcileWorkers,
	).Run(ctx, log, NewScope(clusterSpec))
}

// ReconcileWorkers applies the worker CAPI objects to the cluster.
func (r *Reconciler) ReconcileWorkers(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	spec := tinkerbellScope.ClusterSpec
	log = log.WithValues("phase", "reconcileWorkers")
	log.Info("Applying worker CAPI objects")

	return clusters.ReconcileWorkersForEKSA(ctx, log, r.client, spec.Cluster, clusters.ToWorkers(tinkerbellScope.Workers))
}

// ValidateDatacenterConfig updates the cluster status if the TinkerbellDatacenter status indicates that the spec is invalid.
func (r *Reconciler) ValidateDatacenterConfig(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	log = log.WithValues("phase", "validateDatacenterConfig")

	if err := r.validateTinkerbellIPMatch(ctx, tinkerbellScope); err != nil {
		log.Error(err, "Invalid TinkerbellDatacenterConfig")
		failureMessage := err.Error()
		tinkerbellScope.ClusterSpec.Cluster.Status.FailureMessage = &failureMessage
		return controller.ResultWithReturn(), nil
	}

	return controller.Result{}, nil
}

// ReconcileCNI reconciles the CNI to the desired state.
func (r *Reconciler) ReconcileCNI(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	clusterSpec := tinkerbellScope.ClusterSpec
	log = log.WithValues("phase", "reconcileCNI")

	client, err := r.remoteClientRegistry.GetClient(ctx, controller.CapiClusterObjectKey(clusterSpec.Cluster))
	if err != nil {
		return controller.Result{}, err
	}

	return r.cniReconciler.Reconcile(ctx, log, client, clusterSpec)
}

func (r *Reconciler) validateTinkerbellIPMatch(ctx context.Context, tinkerbellScope *Scope) error {
	clusterSpec := tinkerbellScope.ClusterSpec
	if clusterSpec.Cluster.IsManaged() {

		// for workload cluster tinkerbell IP must match management cluster tinkerbell IP
		managementClusterSpec := &anywherev1.Cluster{}

		err := r.client.Get(ctx, client.ObjectKey{
			Namespace: clusterSpec.Cluster.Namespace,
			Name:      clusterSpec.Cluster.Spec.ManagementCluster.Name,
		}, managementClusterSpec)
		if err != nil {
			return err
		}

		managementDatacenterConfig := &anywherev1.TinkerbellDatacenterConfig{}

		err = r.client.Get(ctx, client.ObjectKey{
			Namespace: clusterSpec.Cluster.Namespace,
			Name:      managementClusterSpec.Spec.DatacenterRef.Name,
		}, managementDatacenterConfig)
		if err != nil {
			return err
		}

		if clusterSpec.TinkerbellDatacenter.Spec.TinkerbellIP != managementDatacenterConfig.Spec.TinkerbellIP {
			return errors.New("workload cluster Tinkerbell IP must match managment cluster Tinkerbell IP")
		}
	}

	return nil
}

func toClientControlPlane(cp *tinkerbell.ControlPlane) *clusters.ControlPlane {
	other := make([]client.Object, 0, 1)
	if cp.Secrets != nil {
		other = append(other, cp.Secrets)
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

// ValidateHardware performs a set of validations on the tinkerbell hardware read from the cluster.
func (r *Reconciler) ValidateHardware(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	clusterSpec := tinkerbellScope.ClusterSpec
	log = log.WithValues("phase", "validateHardware")

	// We need a new reader each time so that the catalogue gets recreated.
	kubeReader := hardware.NewKubeReader(r.client)
	if err := kubeReader.LoadHardware(ctx); err != nil {
		log.Error(err, "Hardware validation failure")
		failureMessage := err.Error()
		clusterSpec.Cluster.Status.FailureMessage = &failureMessage
		return controller.ResultWithReturn(), nil
	}

	var v tinkerbell.ClusterSpecValidator
	v.Register(tinkerbell.HardwareSatisfiesOnlyOneSelectorAssertion(kubeReader.GetCatalogue()))

	// todo: add AssertionsForScaleUpDown
	switch tinkerbellScope.ClusterChange {
	case NeedNewNode:
		v.Register(tinkerbell.ExtraHardwareAvailableAssertionForRollingUpgrade(kubeReader.GetCatalogue()))
	case ReconcileNewCluster:
		v.Register(tinkerbell.MinimumHardwareAvailableAssertionForCreate(kubeReader.GetCatalogue()))
	}

	tinkClusterSpec := tinkerbell.NewClusterSpec(
		clusterSpec,
		clusterSpec.Config.TinkerbellMachineConfigs,
		clusterSpec.Config.TinkerbellDatacenter,
	)

	if err := v.Validate(tinkClusterSpec); err != nil {
		log.Error(err, "Hardware validation failure")
		failureMessage := err.Error()
		clusterSpec.Cluster.Status.FailureMessage = &failureMessage

		return controller.ResultWithReturn(), nil
	}

	return controller.Result{}, nil
}

// ValidateRufioMachines checks to ensure all the Rufio machines condition contactable is True.
func (r *Reconciler) ValidateRufioMachines(ctx context.Context, log logr.Logger, tinkerbellScope *Scope) (controller.Result, error) {
	clusterSpec := tinkerbellScope.ClusterSpec
	log = log.WithValues("phase", "validateRufioMachines")

	kubeReader := hardware.NewKubeReader(r.client)
	if err := kubeReader.LoadRufioMachines(ctx); err != nil {
		log.Error(err, "loading existing rufio machines from the cluster")
		failureMessage := err.Error()
		clusterSpec.Cluster.Status.FailureMessage = &failureMessage

		return controller.ResultWithReturn(), nil
	}

	for _, rm := range kubeReader.GetCatalogue().AllBMCs() {
		if err := r.checkContactable(rm); err != nil {
			log.Error(err, "rufio machine check failure")
			failureMessage := err.Error()
			clusterSpec.Cluster.Status.FailureMessage = &failureMessage

			return controller.ResultWithReturn(), nil
		}
	}

	return controller.Result{}, nil
}

func (r *Reconciler) checkContactable(rm *rufiov1alpha1.Machine) error {
	for _, c := range rm.Status.Conditions {
		if c.Type == rufiov1alpha1.Contactable {
			if c.Status == rufiov1alpha1.ConditionTrue {
				return nil
			}
			if c.Status == rufiov1alpha1.ConditionFalse {
				return errors.New(c.Message)
			}
		}
	}

	return nil
}
