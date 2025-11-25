package clusters

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
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

	// Checking for version as well to avoid race condition of status not being updated in time at least for Kubernetes version upgrades
	if !conditions.IsTrue(kcp, clusterapi.ReadyCondition) ||
		kcp.Status.Version == nil || kcp.Spec.Version != *kcp.Status.Version {
		log.Info("KCP is not ready yet, requeing")
		return controller.ResultWithRequeue(30 * time.Second), nil
	}

	log.Info("KCP is ready")
	return controller.Result{}, nil
}
