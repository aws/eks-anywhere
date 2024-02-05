package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	upgrader "github.com/aws/eks-anywhere/pkg/nodeupgrader"
)

const (
	controlPlaneLabel = "node-role.kubernetes.io/control-plane"
	podDNEMessage     = "Upgrader pod does not exist"

	// nodeUpgradeFinalizerName is the finalizer added to NodeUpgrade objects to handle deletion.
	nodeUpgradeFinalizerName = "nodeupgrades.anywhere.eks.amazonaws.com/finalizer"
)

// RemoteClientRegistry defines methods for remote cluster controller clients.
type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// NodeUpgradeReconciler reconciles a NodeUpgrade object.
type NodeUpgradeReconciler struct {
	client               client.Client
	log                  logr.Logger
	remoteClientRegistry RemoteClientRegistry
}

// NewNodeUpgradeReconciler returns a new instance of NodeUpgradeReconciler.
func NewNodeUpgradeReconciler(client client.Client, remoteClientRegistry RemoteClientRegistry) *NodeUpgradeReconciler {
	return &NodeUpgradeReconciler{
		client:               client,
		remoteClientRegistry: remoteClientRegistry,
		log:                  ctrl.Log.WithName("NodeUpgradeController"),
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.NodeUpgrade{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=nodeupgrades,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=nodeupgrades/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=nodeupgrades/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups="cluster.x-k8s.io",resources=machines,verbs=list;watch;get;patch;update

// Reconcile reconciles a NodeUpgrade object.
// nolint:gocyclo
func (r *NodeUpgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, reterr error) {
	// TODO(in-place): Add validating webhook to block updating the nodeUpgrade object.
	// It should be immutable. If it needs to be changed, a new spec should be applied.

	log := r.log.WithValues("NodeUpgrade", req.NamespacedName)

	log.Info("Reconciling NodeUpgrade object")
	nodeUpgrade := &anywherev1.NodeUpgrade{}
	if err := r.client.Get(ctx, req.NamespacedName, nodeUpgrade); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	machineToBeUpgraded := &clusterv1.Machine{}
	if err := r.client.Get(ctx, GetNamespacedNameType(nodeUpgrade.Spec.Machine.Name, nodeUpgrade.Spec.Machine.Namespace), machineToBeUpgraded); err != nil {
		return ctrl.Result{}, err
	}

	rClient, err := r.remoteClientRegistry.GetClient(ctx, GetNamespacedNameType(machineToBeUpgraded.Spec.ClusterName, machineToBeUpgraded.Namespace))
	if err != nil {
		return ctrl.Result{}, err
	}

	if machineToBeUpgraded.Status.NodeRef == nil {
		return ctrl.Result{}, fmt.Errorf("machine %s is missing nodeRef", machineToBeUpgraded.Name)
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(nodeUpgrade, r.client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		err := r.updateStatus(ctx, log, rClient, nodeUpgrade, machineToBeUpgraded.Status.NodeRef.Name)
		if err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Always attempt to patch the object and status after each reconciliation.
		patchOpts := []patch.Option{}

		// We want the observedGeneration to indicate, that the status shown is up-to-date given the desired spec of the same generation.
		// However, if there is an error while updating the status, we may get a partial status update, In this case,
		// a partially updated status is not considered up to date, so we should not update the observedGeneration

		// Patch ObservedGeneration only if the reconciliation completed without error
		if reterr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchNodeUpgrade(ctx, patchHelper, *nodeUpgrade, patchOpts...); err != nil {
			reterr = kerrors.NewAggregate([]error{reterr, err})
		}

		// Only requeue if we are not already re-queueing and the NodeUpgrade ready condition is false.
		// We do this to be able to update the status continuously until the NodeUpgrade becomes ready,
		// since there might be changes in state of the world that don't trigger reconciliation requests

		if reterr == nil && !result.Requeue && result.RequeueAfter <= 0 && conditions.IsFalse(nodeUpgrade, anywherev1.ReadyCondition) {
			result = ctrl.Result{RequeueAfter: 10 * time.Second}
		}
	}()

	// Reconcile the NodeUpgrade deletion if the DeletionTimestamp is set.
	if !nodeUpgrade.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, nodeUpgrade, machineToBeUpgraded.Status.NodeRef.Name, rClient)
	}

	controllerutil.AddFinalizer(nodeUpgrade, nodeUpgradeFinalizerName)

	return r.reconcile(ctx, log, machineToBeUpgraded, nodeUpgrade, rClient)
}

func (r *NodeUpgradeReconciler) reconcile(ctx context.Context, log logr.Logger, machineToBeUpgraded *clusterv1.Machine, nodeUpgrade *anywherev1.NodeUpgrade, remoteClient client.Client) (ctrl.Result, error) {
	node := &corev1.Node{}
	if err := remoteClient.Get(ctx, types.NamespacedName{Name: machineToBeUpgraded.Status.NodeRef.Name}, node); err != nil {
		return reconcile.Result{}, err
	}

	// return early if node upgrade is already complete.
	if nodeUpgrade.Status.Completed {
		log.Info("Node is upgraded", "Node", node.Name)
		return ctrl.Result{}, nil
	}

	if err := namespaceOrCreate(ctx, remoteClient, log, constants.EksaSystemNamespace); err != nil {
		return ctrl.Result{}, nil
	}

	log.Info("Upgrading node", "Node", node.Name)
	upgraderPod := &corev1.Pod{}
	if conditions.IsTrue(nodeUpgrade, anywherev1.UpgraderPodCreated) || upgraderPodExists(ctx, remoteClient, node.Name) {
		log.Info("Upgrader pod already exists, skipping creation of the pod", "Pod", upgrader.PodName(node.Name))
		return ctrl.Result{}, nil
	}

	configMap := &corev1.ConfigMap{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: constants.UpgraderConfigMapName, Namespace: constants.EksaSystemNamespace}, configMap); err != nil {
		return ctrl.Result{}, err
	}
	if configMap.Data == nil {
		return ctrl.Result{}, errors.New("upgrader config map is empty")
	}
	upgraderImage, ok := configMap.Data[nodeUpgrade.Spec.KubernetesVersion]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("upgrader image corresponding to EKS Distro version %s not found in the config map", nodeUpgrade.Spec.KubernetesVersion)
	}

	if isControlPlane(node) {
		if nodeUpgrade.Spec.FirstNodeToBeUpgraded {
			upgraderPod = upgrader.UpgradeFirstControlPlanePod(node.Name, upgraderImage, nodeUpgrade.Spec.KubernetesVersion, *nodeUpgrade.Spec.EtcdVersion)
		} else {
			upgraderPod = upgrader.UpgradeSecondaryControlPlanePod(node.Name, upgraderImage)
		}
	} else {
		upgraderPod = upgrader.UpgradeWorkerPod(node.Name, upgraderImage)
	}

	if err := remoteClient.Create(ctx, upgraderPod); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create the upgrader pod on node %s: %v", node.Name, err)
	}

	return ctrl.Result{}, nil
}

// namespaceOrCreate creates a namespace if it doesn't already exist.
func namespaceOrCreate(ctx context.Context, client client.Client, log logr.Logger, namespace string) error {
	ns := &corev1.Namespace{}
	if err := client.Get(ctx, types.NamespacedName{Name: namespace}, ns); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Creating namespace on the remote cluster", "Namespace", namespace)
			ns := &corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: namespace,
				},
			}
			if err := client.Create(ctx, ns); err != nil {
				return fmt.Errorf("creating namespace %s on cluster: %v", namespace, err)
			}
		} else {
			return fmt.Errorf("getting namespace %s on cluster: %v", namespace, err)
		}
	}
	return nil
}

func (r *NodeUpgradeReconciler) reconcileDelete(ctx context.Context, log logr.Logger, nodeUpgrade *anywherev1.NodeUpgrade, nodeName string, remoteClient client.Client) (ctrl.Result, error) {
	log.Info("Reconcile NodeUpgrade deletion")

	pod, err := getUpgraderPod(ctx, remoteClient, nodeName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Upgrader pod not found, skipping pod deletion")
		} else {
			return ctrl.Result{}, fmt.Errorf("getting upgrader pod: %v", err)
		}
	} else {
		// TODO(in-place): Make pod deletion logic more robust by checking if the pod is still running.
		// If it is still running and not errored out, then wait before deleting the pod.
		log.Info("Deleting upgrader pod", "Pod", pod.Name, "Namespace", pod.Namespace)
		if err := remoteClient.Delete(ctx, pod); err != nil {
			return ctrl.Result{}, fmt.Errorf("deleting upgrader pod: %v", err)
		}
	}

	// Remove the finalizer from NodeUpgrade object
	controllerutil.RemoveFinalizer(nodeUpgrade, nodeUpgradeFinalizerName)
	return ctrl.Result{}, nil
}

func (r *NodeUpgradeReconciler) updateStatus(ctx context.Context, log logr.Logger, remoteClient client.Client, nodeUpgrade *anywherev1.NodeUpgrade, nodeName string) error {
	// When NodeUpgrade is fully deleted, we do not need to update the status. Without this check
	// the subsequent patch operations would fail if the status is updated after it is fully deleted.
	if !nodeUpgrade.DeletionTimestamp.IsZero() && len(nodeUpgrade.GetFinalizers()) == 0 {
		log.Info("NodeUpgrade is deleted, skipping status update")
		return nil
	}

	log.Info("Updating NodeUpgrade status")

	pod, err := getUpgraderPod(ctx, remoteClient, nodeName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			markAllConditionsFalse(nodeUpgrade, podDNEMessage, clusterv1.ConditionSeverityInfo)
		} else {
			markAllConditionsFalse(nodeUpgrade, err.Error(), clusterv1.ConditionSeverityError)
		}
		return fmt.Errorf("getting upgrader pod: %v", err)
	}

	conditions.MarkTrue(nodeUpgrade, anywherev1.UpgraderPodCreated)
	updateComponentsConditions(pod, nodeUpgrade)

	// Always update the readyCondition by summarizing the state of other conditions.
	conditions.SetSummary(nodeUpgrade,
		conditions.WithConditions(
			anywherev1.UpgraderPodCreated,
			anywherev1.BinariesCopied,
			anywherev1.ContainerdUpgraded,
			anywherev1.CNIPluginsUpgraded,
			anywherev1.KubeadmUpgraded,
			anywherev1.KubeletUpgraded,
			anywherev1.PostUpgradeCleanupCompleted,
		),
	)
	return nil
}

func updateComponentsConditions(pod *corev1.Pod, nodeUpgrade *anywherev1.NodeUpgrade) {
	containersMap := []struct {
		name      string
		condition clusterv1.ConditionType
	}{
		{
			name:      upgrader.CopierContainerName,
			condition: anywherev1.BinariesCopied,
		},
		{
			name:      upgrader.ContainerdUpgraderContainerName,
			condition: anywherev1.ContainerdUpgraded,
		},
		{
			name:      upgrader.CNIPluginsUpgraderContainerName,
			condition: anywherev1.CNIPluginsUpgraded,
		},
		{
			name:      upgrader.KubeadmUpgraderContainerName,
			condition: anywherev1.KubeadmUpgraded,
		},
		{
			name:      upgrader.KubeletUpgradeContainerName,
			condition: anywherev1.KubeletUpgraded,
		},
		{
			name:      upgrader.PostUpgradeContainerName,
			condition: anywherev1.PostUpgradeCleanupCompleted,
		},
	}

	completed := true
	for _, container := range containersMap {
		status, err := getContainerStatus(pod, container.name)
		if err != nil {
			conditions.MarkFalse(nodeUpgrade, container.condition, "Container status not available yet", clusterv1.ConditionSeverityWarning, "")
			completed = false
		} else {
			if status.State.Waiting != nil {
				conditions.MarkFalse(nodeUpgrade, container.condition, "Container is waiting to be initialized", clusterv1.ConditionSeverityInfo, "")
				completed = false
			} else if status.State.Running != nil {
				conditions.MarkFalse(nodeUpgrade, container.condition, "Container is still running", clusterv1.ConditionSeverityInfo, "")
				completed = false
			} else if status.State.Terminated != nil {
				if status.State.Terminated.ExitCode != 0 {
					conditions.MarkFalse(nodeUpgrade, container.condition, fmt.Sprintf("Container exited with a non-zero exit code, reason: %s", status.State.Terminated.Reason), clusterv1.ConditionSeverityError, "")
					completed = false
				} else {
					conditions.MarkTrue(nodeUpgrade, container.condition)
				}
			} else {
				// this should not happen
				conditions.MarkFalse(nodeUpgrade, container.condition, "Container state is unknown", clusterv1.ConditionSeverityWarning, "")
				completed = false
			}
		}
	}
	nodeUpgrade.Status.Completed = completed
}

func getContainerStatus(pod *corev1.Pod, containerName string) (*corev1.ContainerStatus, error) {
	for _, status := range pod.Status.InitContainerStatuses {
		if status.Name == containerName {
			return &status, nil
		}
	}
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
			return &status, nil
		}
	}
	return nil, fmt.Errorf("status not found for container %s in pod %s", containerName, pod.Name)
}

func markAllConditionsFalse(nodeUpgrade *anywherev1.NodeUpgrade, message string, severity clusterv1.ConditionSeverity) {
	conditions.MarkFalse(nodeUpgrade, anywherev1.UpgraderPodCreated, message, clusterv1.ConditionSeverityError, "")
	conditions.MarkFalse(nodeUpgrade, anywherev1.BinariesCopied, message, clusterv1.ConditionSeverityError, "")
	conditions.MarkFalse(nodeUpgrade, anywherev1.ContainerdUpgraded, message, clusterv1.ConditionSeverityError, "")
	conditions.MarkFalse(nodeUpgrade, anywherev1.CNIPluginsUpgraded, message, clusterv1.ConditionSeverityError, "")
	conditions.MarkFalse(nodeUpgrade, anywherev1.KubeadmUpgraded, message, clusterv1.ConditionSeverityError, "")
	conditions.MarkFalse(nodeUpgrade, anywherev1.KubeletUpgraded, message, clusterv1.ConditionSeverityError, "")
}

func isControlPlane(node *corev1.Node) bool {
	_, ok := node.Labels[controlPlaneLabel]
	return ok
}

// GetNamespacedNameType takes name and namespace and returns NamespacedName in namespace/name format.
func GetNamespacedNameType(name, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}

func patchNodeUpgrade(ctx context.Context, patchHelper *patch.Helper, nodeUpgrade anywherev1.NodeUpgrade, patchOpts ...patch.Option) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	options := append([]patch.Option{
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			// Add each condition her that the controller should ignored conflicts for.
			anywherev1.UpgraderPodCreated,
			anywherev1.BinariesCopied,
			anywherev1.ContainerdUpgraded,
			anywherev1.CNIPluginsUpgraded,
			anywherev1.KubeadmUpgraded,
			anywherev1.KubeletUpgraded,
		}},
	}, patchOpts...)

	// Always attempt to patch the object and status after each reconciliation.
	return patchHelper.Patch(ctx, &nodeUpgrade, options...)
}

func upgraderPodExists(ctx context.Context, remoteClient client.Client, nodeName string) bool {
	_, err := getUpgraderPod(ctx, remoteClient, nodeName)
	return err == nil
}

func getUpgraderPod(ctx context.Context, remoteClient client.Client, nodeName string) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	if err := remoteClient.Get(ctx, GetNamespacedNameType(upgrader.PodName(nodeName), constants.EksaSystemNamespace), pod); err != nil {
		return nil, err
	}
	return pod, nil
}

func getNodeUpgrade(ctx context.Context, remoteClient client.Client, nodeUpgradeName string) (*anywherev1.NodeUpgrade, error) {
	n := &anywherev1.NodeUpgrade{}
	if err := remoteClient.Get(ctx, GetNamespacedNameType(nodeUpgradeName, constants.EksaSystemNamespace), n); err != nil {
		return nil, err
	}
	return n, nil
}

// nodeUpgradeName returns the name of the node upgrade object based on the machine reference.
func nodeUpgraderName(machineRefName string) string {
	return fmt.Sprintf("%s-node-upgrader", machineRefName)
}
