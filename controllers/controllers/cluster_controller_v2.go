package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/controllers/remote"
	capicontrolplane "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// ClusterReconcilerV2 reconciles a Cluster object
type ClusterReconcilerV2 struct {
	Client  eksaClient
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Tracker *remote.ClusterCacheTracker
}

func NewClusterReconcilerV2(client client.Client, tracker *remote.ClusterCacheTracker, log logr.Logger, scheme *runtime.Scheme) *ClusterReconcilerV2 {
	return &ClusterReconcilerV2{
		Client:  eksaClient{Client: client},
		Log:     log,
		Scheme:  scheme,
		Tracker: tracker,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconcilerV2) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.Cluster{}). // TODO: add filtering to ignore self managed clusters
		Watches(&source.Kind{Type: &capi.Cluster{}},
			handler.EnqueueRequestsFromMapFunc(r.capiClusterToCluster),
			// TODO: add filtering for only capi clusters with the eks-a cluster name label
		).
		Watches(&source.Kind{Type: &capicontrolplane.KubeadmControlPlane{}},
			handler.EnqueueRequestsFromMapFunc(r.capiControlPlaneToCluster),
			// TODO: add filtering for only capi kubeadm control planes with the eks-a cluster name label
		).
		Watches(&source.Kind{Type: &capi.MachineDeployment{}},
			handler.EnqueueRequestsFromMapFunc(r.capiMachineDeploymentToCluster),
			// TODO: add filtering for only capi clusters with the eks-a cluster name label
		).
		Complete(r)
}

// TODO: review this and reduce to minimun set of permissions
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters;vspheredatacenterconfigs;vspheremachineconfigs;dockerdatacenterconfigs;bundles;awsiamconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/status;vspheredatacenterconfigs/status;vspheremachineconfigs/status;dockerdatacenterconfigs/status;bundles/status;awsiamconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=clusters/finalizers;vspheredatacenterconfigs/finalizers;vspheremachineconfigs/finalizers;dockerdatacenterconfigs/finalizers;bundles/finalizers;awsiamconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io;infrastructure.cluster.x-k8s.io;bootstrap.cluster.x-k8s.io;controlplane.cluster.x-k8s.io;etcdcluster.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ClusterReconcilerV2) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	_ = r.Log.WithValues("cluster", req.NamespacedName)
	// Fetch the Cluster instance.
	cluster := &anywherev1.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(cluster, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always attempt to patch the object and status after each reconciliation.
		if err := patchHelper.Patch(ctx, cluster); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}
	}()

	// Ignore deleted Clusters, this can happen when foregroundDeletion
	// is enabled
	if !cluster.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	// If the external object is paused, return without any further processing.
	if cluster.IsReconcilePaused() {
		r.Log.Info("eksa reconciliation is paused")
		return ctrl.Result{}, nil
	}

	if cluster.IsSelfManaged() {
		r.Log.Info("Ignoring self managed cluster", "cluster", cluster.Name)
		return ctrl.Result{}, nil
	}

	result, err := r.reconcile(ctx, cluster)
	if err != nil {
		r.Log.Error(err, "Failed to reconcile Cluster")
	}
	return result, err
}

func (r *ClusterReconcilerV2) reconcile(ctx context.Context, cluster *anywherev1.Cluster) (ctrl.Result, error) {
	r.Log.Info("Reconcile EKS-A Cluster")

	// Secrets reconcile
	// ignoring this for now, there has to be a better way to solve this
	// TODO: figure out what this is for and find a pattern that fits the controller

	// Etcd providers installed (only applicable to controller when we support management from managed clusters)
	// TODO: Ignoring this for now, we might be able to only do this in the cli

	// Core components reconcile (here the one that will always apply is Cilium, but we don’t have implemented yet)
	// TODO: Ignoring this for now

	// CP capi objects reconcile
	if err := r.reconcileControlPlane(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	capiCluster, err := r.Client.GetCAPICluster(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}
	if capiCluster == nil {
		// if capi cluster doesn't exist yet, requeue, we probably got here too fast
		r.Log.Info("CAPI cluster not found, requeing", "cluster", cluster.Name)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	// External etcd ready
	if cluster.Spec.ExternalEtcdConfiguration != nil && !IsTrue(capiCluster, "ManagedEtcdReady") {
		r.Log.Info("External etcd not ready", "cluster", cluster.Name)
		return ctrl.Result{}, nil
	}

	// Control plane ready
	if !IsTrue(capiCluster, "ControlPlaneReady") {
		r.Log.Info("Control plane not ready", "cluster", cluster.Name, "capiCluster", capiCluster.Name)
		return ctrl.Result{}, nil
	}

	// CP Machines ready (use label to distinguish)
	// TODO: Ignoring this for now. I'm not sure we need it in the controller. I believe this was added to prevent errors
	//  during the Move operation, because clusterctl runs a similar validation and will fail in that case
	//  Move is not a controller's responsability domain. That's a cli responsability, so such valiodation probably belongs there

	// Control plane ready (again)
	// TODO: do we need this twice?
	if !IsTrue(capiCluster, "ControlPlaneReady") {
		r.Log.Info("Control plane not ready", "cluster", cluster.Name, "capiCluster", capiCluster.Name)
		return ctrl.Result{}, nil
	}

	// CNI reconcile
	if err := r.reconcileCNI(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Control plane nodes ready
	capiControlPlane, err := r.Client.GetCAPIControlPlane(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}
	if capiControlPlane == nil {
		// if capi control plane doesn't exist yet, requeue, we probably got here too fast
		// TODO: figure out if there is a way to distinguish between race conditions and actual errors: Why wouldn't the kubeadmcontrolplane don't exist at this point?
		r.Log.Info("Kubeadm control plane not found, requeing", "cluster", cluster.Name)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}
	if !capiControlPlaneReady(capiControlPlane) {
		r.Log.Info("Kubeadm control plane replicas not ready", "cluster", cluster.Name, "capiControlPlane", capiControlPlane.Name)
		return ctrl.Result{}, nil
	}

	// Kubeconfig reconcile (do we need this?)
	// TODO: ignoring this for now, this should be probably a cli responsibility

	// Workers capi objects reconcile
	if err := r.reconcileWorkers(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// Worker machine deployments ready (assumes there is only one machine deployment)
	machineDeployments, err := r.Client.GetWorkerMachineDeployments(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}
	if machineDeployments == nil || len(machineDeployments.Items) != len(cluster.Spec.WorkerNodeGroupConfigurations) {
		// TODO: we should match WorkerNodeGroupConfigurations one by one to MachineDeployments
		// Not only all the MachineDeployments need to be ready, but also they all of them need to match the current status of the WorkerNodeGroupConfigurations
		// We should avoid counting old MachineDeployments

		// if no worker machine deployment don't exist yet, requeue, we probably got here too fast
		r.Log.Info("Worker MachineDeployments not found, requeing", "cluster", cluster.Name)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	for _, m := range machineDeployments.Items {
		if !machineDeploymentReady(&m) {
			r.Log.Info("Worker MachineDeployment not ready", "cluster", cluster.Name, "machineDeployment", m.Name)
			return ctrl.Result{}, nil
		}
	}

	// Worker nodes ready (use label to distinguish)
	// TODO: Ignoring this for now. I'm not sure we need it in the controller
	//  Same as for the control plane nodes

	// CAPI deployments ready (core CAPI, controlplane, kubeadm, etcdadm and providers)
	// TODO: what do we need this for?

	// "Post-upgrade" reconcile. This is logic specific for each provider. In VSphere, it “refreshes” the ClusterResourceSet and updates the image for a DaemonSet
	// TODO: ignoring this for now

	// Extra objects reconciled (this is basically whatever extra stuff we need for kubernetes to be up, in this case, some extra permissions for coredns)
	// TODO: should we move this up, just after control plane is ready?
	if err := r.reconcileExtraObjects(ctx, cluster); err != nil {
		return ctrl.Result{}, err
	}

	// AWS IAM reconcile
	// TODO: ignoring this for now

	// Storage class reconcile
	// TODO: ignoring this for now

	// CAPI providers reconcile
	// Ignoring this for now, most likely won't support this in the foreseeable future

	// Machine healths reconcile
	// TODO: ignoring this for now

	// GitOps reconcile
	// TODO: ignoring this for now

	r.Log.Info("Cluster reconciled", "cluster", cluster.Name)
	return ctrl.Result{}, nil
}

func (r *ClusterReconcilerV2) capiClusterToCluster(o client.Object) []ctrl.Request {
	capiCluster, ok := o.(*capi.Cluster)
	if !ok {
		panic(fmt.Sprintf("Expected a CAPI Cluster but got a %T", o))
	}

	return r.objectWithClusterLabelNameToCluster(capiCluster)
}

func (r *ClusterReconcilerV2) capiControlPlaneToCluster(o client.Object) []ctrl.Request {
	capiControlPlane, ok := o.(*capicontrolplane.KubeadmControlPlane)
	if !ok {
		panic(fmt.Sprintf("Expected a CAPI Cluster but got a %T", o))
	}

	return r.objectWithClusterLabelNameToCluster(capiControlPlane)
}

func (r *ClusterReconcilerV2) capiMachineDeploymentToCluster(o client.Object) []ctrl.Request {
	capiControlPlane, ok := o.(*capi.MachineDeployment)
	if !ok {
		panic(fmt.Sprintf("Expected a CAPI MachineDeployment but got a %T", o))
	}

	return r.objectWithClusterLabelNameToCluster(capiControlPlane)
}

func (r *ClusterReconcilerV2) objectWithClusterLabelNameToCluster(obj client.Object) []ctrl.Request {
	labels := obj.GetLabels()
	clusterName, ok := labels[ClusterLabelName]
	if !ok {
		// Object not managed by a eks-a Cluster, don't enqueue
		// We could also use ownership for this
		r.Log.Info("Object not managed by an eks-a Cluster, ignoring", "type", fmt.Sprintf("%T", obj), "name", obj.GetName())
		return nil
	}

	return []ctrl.Request{{
		NamespacedName: types.NamespacedName{
			Namespace: obj.GetNamespace(),
			Name:      clusterName,
		},
	}}
}
