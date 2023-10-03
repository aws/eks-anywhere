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
	"github.com/aws/eks-anywhere/pkg/features"
)

// CheckControlPlaneReady is a controller helper to check whether KCP object for
// the cluster is ready or not. This is intended to be used from cluster reconcilers
// due its signature and that it returns controller results with appropriate wait times whenever
// the cluster is not ready.
func CheckControlPlaneReady(ctx context.Context, client client.Client, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	if features.IsActive(features.ExperimentalSelfManagedClusterUpgrade()) {
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

		if !conditions.IsTrue(kcp, clusterapi.ReadyCondition) {
			log.Info("KCP is not ready yet, requeing")
			return controller.ResultWithRequeue(30 * time.Second), nil
		}

		log.Info("KCP is ready")
		return controller.Result{}, nil
	}

	capiCluster, err := controller.GetCAPICluster(ctx, client, cluster)
	if err != nil {
		return controller.Result{}, err
	}

	if capiCluster == nil {
		log.Info("CAPI cluster does not exist yet, requeuing")
		return controller.ResultWithRequeue(5 * time.Second), nil
	}

	if !conditions.IsTrue(capiCluster, clusterapi.ControlPlaneReadyCondition) {
		log.Info("CAPI control plane is not ready yet, requeuing")
		// TODO: eventually this can be implemented with controller watches
		return controller.ResultWithRequeue(30 * time.Second), nil
	}

	log.Info("CAPI control plane is ready")
	return controller.Result{}, nil
}
