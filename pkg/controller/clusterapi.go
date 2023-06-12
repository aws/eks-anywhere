package controller

import (
	"context"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
)

// GetCAPICluster reads a cluster-api Cluster for an eks-a cluster using a kube client
// If the CAPI cluster is not found, the method returns (nil, nil).
func GetCAPICluster(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) (*clusterv1.Cluster, error) {
	capiClusterName := clusterapi.ClusterName(cluster)

	capiCluster := &clusterv1.Cluster{}
	key := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: capiClusterName}

	err := client.Get(ctx, key, capiCluster)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return capiCluster, nil
}

// CapiClusterObjectKey generates an ObjectKey for the CAPI cluster owned by
// the provided eks-a cluster.
func CapiClusterObjectKey(cluster *anywherev1.Cluster) client.ObjectKey {
	// TODO: we should consider storing a reference to the CAPI cluster in the eksa cluster status
	return client.ObjectKey{
		Name:      clusterapi.ClusterName(cluster),
		Namespace: constants.EksaSystemNamespace,
	}
}

// GetKubeadmControlPlane reads a cluster-api KubeadmControlPlane for an eks-a cluster using a kube client
// If the KubeadmControlPlane is not found, the method returns (nil, nil).
func GetKubeadmControlPlane(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) (*controlplanev1.KubeadmControlPlane, error) {
	kubeadmControlPlane, err := KubeadmControlPlane(ctx, client, cluster)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return kubeadmControlPlane, nil
}

// KubeadmControlPlane reads a cluster-api KubeadmControlPlane for an eks-a cluster using a kube client.
func KubeadmControlPlane(ctx context.Context, client client.Client, cluster *anywherev1.Cluster) (*controlplanev1.KubeadmControlPlane, error) {
	kubeadmControlPlane := &controlplanev1.KubeadmControlPlane{}
	if err := client.Get(ctx, CAPIKubeadmControlPlaneKey(cluster), kubeadmControlPlane); err != nil {
		return nil, err
	}
	return kubeadmControlPlane, nil
}

// CAPIKubeadmControlPlaneKey generates an ObjectKey for the CAPI Kubeadm control plane owned by
// the provided eks-a cluster.
func CAPIKubeadmControlPlaneKey(cluster *anywherev1.Cluster) client.ObjectKey {
	return client.ObjectKey{
		Name:      clusterapi.KubeadmControlPlaneName(cluster),
		Namespace: constants.EksaSystemNamespace,
	}
}

// GetMachineDeployment reads a cluster-api MachineDeployment for an eks-a cluster using a kube client.
// If the MachineDeployment is not found, the method returns (nil, nil).
func GetMachineDeployment(ctx context.Context, client client.Client, machineDeploymentName string) (*clusterv1.MachineDeployment, error) {
	machineDeployment := &clusterv1.MachineDeployment{}
	key := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: machineDeploymentName}

	err := client.Get(ctx, key, machineDeployment)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return machineDeployment, nil
}

// GetMachines reads a list cluster-api Machines for an eks-a cluster using a kube client.
func GetMachines(ctx context.Context, c client.Client, cluster *anywherev1.Cluster) ([]clusterv1.Machine, error) {
	// "get", capiMachinesType, "-o", "json", "--kubeconfig", cluster.KubeconfigFile,
	// "--selector=cluster.x-k8s.io/cluster-name=" + clusterName,
	// "--namespace", constants.EksaSystemNamespace,
	machines := &clusterv1.MachineList{}
	err := c.List(ctx, machines, client.MatchingLabels{clusterv1.ClusterNameLabel: cluster.Name}, client.InNamespace(constants.EksaSystemNamespace))

	if err != nil {
		return nil, err
	}

	return machines.Items, nil
}

// ControlPlaneMachines takes a list of cluster-api Machines and filters for those with the control plane name label.
func ControlPlaneMachines(machines []clusterv1.Machine) []clusterv1.Machine {
	filteredMachines := []clusterv1.Machine{}
	for _, m := range machines {
		if _, ok := m.ObjectMeta.Labels[clusterv1.MachineControlPlaneNameLabel]; ok {
			filteredMachines = append(filteredMachines, m)
		}
	}
	return filteredMachines
}

// WorkerNodeMachines takes a list of cluster-api Machines and filters for those with the machine deployment name label.
func WorkerNodeMachines(machines []clusterv1.Machine) []clusterv1.Machine {
	filteredMachines := []clusterv1.Machine{}
	for _, m := range machines {
		if _, ok := m.ObjectMeta.Labels[clusterv1.MachineDeploymentNameLabel]; ok {
			filteredMachines = append(filteredMachines, m)
		}
	}
	return filteredMachines
}

type NodeReadyChecker func(status clusterv1.MachineStatus) bool

func WithNodeRef() NodeReadyChecker {
	return func(status clusterv1.MachineStatus) bool {
		return status.NodeRef != nil
	}
}

func WithNodeHealthy() NodeReadyChecker {
	return func(status clusterv1.MachineStatus) bool {
		for _, c := range status.Conditions {
			if c.Type == clusterv1.MachineNodeHealthyCondition {
				return c.Status == "True"
			}
		}
		return false
	}
}

func WithConditionReady() NodeReadyChecker {
	return func(status clusterv1.MachineStatus) bool {
		for _, c := range status.Conditions {
			if c.Type == clusterv1.ReadyCondition {
				return c.Status == "True"
			}
		}
		return false
	}
}

func WithK8sVersion(kubeVersion anywherev1.KubernetesVersion) NodeReadyChecker {
	return func(status clusterv1.MachineStatus) bool {
		return strings.Contains(status.NodeInfo.KubeletVersion, string(kubeVersion))
	}
}
