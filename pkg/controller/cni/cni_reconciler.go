package cni

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	eksacluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
)

type CNIReconciler interface {
	Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster, capiCluster *clusterv1.Cluster, specWithBundles *eksacluster.Spec) (controller.Result, error)
}

func BuildCNIReconciler(cniName string, tracker *remote.ClusterCacheTracker) (CNIReconciler, error) {
	if cniName == "cilium" {
		return NewCiliumReconciler(tracker), nil
	}
	return nil, fmt.Errorf("invalid CNI %s", cniName)
}
