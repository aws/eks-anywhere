package controller

import (
	"context"

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

// GetMachineDeployments reads all of cluster-api MachineDeployment for an eks-a cluster using a kube client.
func GetMachineDeployments(ctx context.Context, c client.Client, cluster *anywherev1.Cluster) ([]clusterv1.MachineDeployment, error) {
	machineDeployments := &clusterv1.MachineDeploymentList{}

	err := c.List(ctx, machineDeployments, client.MatchingLabels{clusterv1.ClusterNameLabel: cluster.Name}, client.InNamespace(constants.EksaSystemNamespace))
	if err != nil {
		return nil, err
	}
	return machineDeployments.Items, nil
}
