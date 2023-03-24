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
	kubeadmControlPlaneName := clusterapi.KubeadmControlPlaneName(cluster)

	kubeadmControlPlane := &controlplanev1.KubeadmControlPlane{}
	key := types.NamespacedName{Namespace: constants.EksaSystemNamespace, Name: kubeadmControlPlaneName}

	err := client.Get(ctx, key, kubeadmControlPlane)
	if apierrors.IsNotFound(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	return kubeadmControlPlane, nil
}
