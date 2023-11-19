package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/pkg/errors"
)

const (
	upgradeScript        = "/foo/eksa-upgrades/scripts/upgrade.sh"
	defaultUpgraderImage = "public.ecr.aws/t0n3a9y4/aws/upgrader:v1.28.3-eks-1-28-9"
	controlPlaneLabel    = "node-role.kubernetes.io/control-plane"
)

// RemoteClientRegistry defines methods for remote cluster controller clients.
type RemoteClientRegistry interface {
	GetClient(ctx context.Context, cluster client.ObjectKey) (client.Client, error)
}

// NodeUpgradeReconciler reconciles a NodeUpgrade object.
type NodeUpgradeReconciler struct {
	client               client.Client
	remoteClientRegistry RemoteClientRegistry
}

// NewNodeUpgradeReconciler returns a new instance of NodeUpgradeReconciler.
func NewNodeUpgradeReconciler(client client.Client, remoteClientRegistry RemoteClientRegistry) *NodeUpgradeReconciler {
	return &NodeUpgradeReconciler{
		client:               client,
		remoteClientRegistry: remoteClientRegistry,
	}
}

//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=nodeupgrades,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=nodeupgrades/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=anywhere.eks.amazonaws.com,resources=nodeupgrades/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;delete

// Reconcile reconciles a NodeUpgrade object.
func (r *NodeUpgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	nodeUpgrade := &anywherev1.NodeUpgrade{}
	if err := r.client.Get(ctx, req.NamespacedName, nodeUpgrade); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	machineToBeUpgraded := &clusterv1.Machine{}
	if err := r.client.Get(ctx, getNamespacedNameType(nodeUpgrade.Spec.Machine.Name, nodeUpgrade.Spec.Machine.Namespace), machineToBeUpgraded); err != nil {
		log.Error(err, "machine not found", "Machine", nodeUpgrade.Spec.Machine.Name)
		return ctrl.Result{}, err
	}

	rClient, err := r.remoteClientRegistry.GetClient(ctx, getNamespacedNameType(nodeUpgrade.Spec.Cluster.Name, nodeUpgrade.Spec.Cluster.Namespace))
	if err != nil {
		return ctrl.Result{}, err
	}

	if machineToBeUpgraded.Status.NodeRef == nil {
		err := errors.New("Machine is missing nodeRef")
		log.Error(err, "nodeRef is not set for machine", "Machine", machineToBeUpgraded.Name)
	}

	node := &corev1.Node{}
	if err := rClient.Get(ctx, types.NamespacedName{Name: machineToBeUpgraded.Status.NodeRef.Name}, node); err != nil {
		return reconcile.Result{}, err
	}

	log.Info("Upgrading node", "Node", node.Name)
	if err := upgradeNode(ctx, node, nodeUpgrade, rClient); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&anywherev1.NodeUpgrade{}).
		Complete(r)
}

func upgradeNode(ctx context.Context, node *corev1.Node, nodeUpgrade *anywherev1.NodeUpgrade, remoteClient client.Client) error {
	var upgraderPod *corev1.Pod
	if isControlPlane(node) {
		upgraderPod = upgradeFirstControlPlanePod(node.Name, defaultUpgraderImage, nodeUpgrade.Spec.KubernetesVersion, *nodeUpgrade.Spec.EtcdVersion)
	} else {
		upgraderPod = upgradeWorkerPod(node.Name, defaultUpgraderImage)
	}

	if err := remoteClient.Create(ctx, upgraderPod); err != nil {
		return fmt.Errorf("failed to create the upgrader pod on node %s: %v", node.Name, err)
	}

	return nil
}

func upgradeFirstControlPlanePod(nodeName, image, kubernetesVersion, etcdVersion string) *corev1.Pod {
	p := upgraderPod(nodeName, image)
	p.Spec.InitContainers = containersForUpgrade(image, nodeName, "kubeadm_in_first_cp", kubernetesVersion, etcdVersion)
	p.Spec.Containers = []corev1.Container{printAndCleanupContainer(image)}

	return p
}

// func upgradeRestControlPlanePod(nodeName, image string) *corev1.Pod {
// 	p := upgraderPod("control-plane-upgrader", nodeName, image)
// 	p.Spec.InitContainers = containersForUpgrade(image, nodeName, "kubeadm_in_rest_cp")
// 	p.Spec.Containers = []corev1.Container{printAndCleanupContainer(image)}

// 	return p
// }

func upgradeWorkerPod(nodeName, image string) *corev1.Pod {
	p := upgraderPod(nodeName, image)
	p.Spec.InitContainers = containersForUpgrade(image, nodeName, "kubeadm_in_worker")
	p.Spec.Containers = []corev1.Container{printAndCleanupContainer(image)}
	return p
}

func upgraderPod(nodeName, image string) *corev1.Pod {
	dirOrCreate := corev1.HostPathDirectoryOrCreate
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-node-upgrader", nodeName),
			Namespace: "eksa-system",
			Labels: map[string]string{
				"ekd-d-upgrader": "true",
			},
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,
			HostPID:  true,
			Volumes: []corev1.Volume{
				{
					Name: "host-components",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/foo",
							Type: &dirOrCreate,
						},
					},
				},
			},
		},
	}
}

func containersForUpgrade(image, nodeName string, kubeadmUpgradeCommand ...string) []corev1.Container {
	return []corev1.Container{
		copierContainer(image),
		nsenterContainer(image, "containerd-upgrader", upgradeScript, "upgrade_containerd"),
		nsenterContainer(image, "cni-plugins-upgrader", upgradeScript, "cni_plugins"),
		nsenterContainer(image, "kubeadm-upgrader", append([]string{upgradeScript}, kubeadmUpgradeCommand...)...),
		// drainerContainer(image, nodeName),
		nsenterContainer(image, "kubelet-kubectl-upgrader", upgradeScript, "kubelet_and_kubectl"),
		// uncordonContainer(image, nodeName),
	}
}

func copierContainer(image string) corev1.Container {
	return corev1.Container{
		Name:    "components-copier",
		Image:   image,
		Command: []string{"cp"},
		Args:    []string{"-r", "/eksa-upgrades", "/usr/host"},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "host-components",
				MountPath: "/usr/host",
			},
		},
	}
}

func nsenterContainer(image, name string, extraArgs ...string) corev1.Container {
	args := []string{
		"--target",
		"1",
		"--mount",
		"--uts",
		"--ipc",
		"--net",
	}
	args = append(args, extraArgs...)

	return corev1.Container{
		Name:    name,
		Image:   image,
		Command: []string{"nsenter"},
		Args:    args,
		SecurityContext: &corev1.SecurityContext{
			Privileged: ptr.Bool(true),
		},
	}
}

// func drainerContainer(image, nodeName string) corev1.Container {
// 	return corev1.Container{
// 		Name:            "drain",
// 		Image:           image,
// 		Command:         []string{"/eksa-upgrades/binaries/kubernetes/usr/bin/kubectl"},
// 		Args:            []string{"drain", nodeName, "--ignore-daemonsets", "--pod-selector", "!ekd-d-upgrader"},
// 		ImagePullPolicy: corev1.PullAlways,
// 	}
// }

func uncordonContainer(image, nodeName string) corev1.Container {
	return corev1.Container{
		Name:            "uncordon",
		Image:           image,
		Command:         []string{"/eksa-upgrades/binaries/kubernetes/usr/bin/kubectl"},
		Args:            []string{"uncordon", nodeName},
		ImagePullPolicy: corev1.PullAlways,
	}
}

func isControlPlane(node *corev1.Node) bool {
	_, ok := node.Labels[controlPlaneLabel]
	return ok
}

func printAndCleanupContainer(image string) corev1.Container {
	return nsenterContainer(image, "post-upgrade-status", upgradeScript, "print_status_and_cleanup")
}

func getNamespacedNameType(name, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
}
