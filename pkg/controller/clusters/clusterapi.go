package clusters

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/controller"
)

// CheckControlPlaneReady is a controller helper to check whether KCP object for
// the cluster is ready or not. This is intended to be used from cluster reconcilers
// due its signature and that it returns controller results with appropriate wait times whenever
// the cluster is not ready.
func CheckControlPlaneReady(ctx context.Context, client client.Client, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	kcp, err := controller.GetKubeadmControlPlane(ctx, client, cluster)
	if err != nil {
		return controller.Result{}, err
	}

	if kcp == nil {
		log.Info("KCP does not exist yet, requeuing")
		return controller.ResultWithRequeue(5 * time.Second), nil
	}

	// We make sure to check that the status is up to date before using it
	if kcp.Status.ObservedGeneration != kcp.ObjectMeta.Generation {
		log.Info("KCP information is outdated, requeing")
		return controller.ResultWithRequeue(5 * time.Second), nil
	}

	// CAPI reports kubelet versions, which can differ from EKS-D release labels.
	// Its current machine condition also guards unfinished version and template rollouts.
	machinesUpToDate := meta.FindStatusCondition(kcp.GetConditions(), controlplanev1beta2.KubeadmControlPlaneMachinesUpToDateCondition)
	if !meta.IsStatusConditionTrue(kcp.GetConditions(), clusterv1beta2.AvailableCondition) ||
		kcp.Status.Version == "" || machinesUpToDate == nil ||
		machinesUpToDate.Status != metav1.ConditionTrue || machinesUpToDate.ObservedGeneration != kcp.Generation {
		log.Info("KCP is not ready yet, requeing")
		return controller.ResultWithRequeue(30 * time.Second), nil
	}

	log.Info("KCP is ready")
	return controller.Result{}, nil
}
