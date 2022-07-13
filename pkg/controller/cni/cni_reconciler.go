package cni

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	eksacluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
)

type CNIReconciler interface {
	Reconcile(ctx context.Context, cluster *anywherev1.Cluster, capiCluster *clusterv1.Cluster, specWithBundles *eksacluster.Spec) (controller.Result, error)
}

func BuildCNIReconciler(cniName string, client client.Client, log logr.Logger, tracker *remote.ClusterCacheTracker) (CNIReconciler, error) {
	if cniName == "cilium" {
		return NewCiliumReconciler(client, log, tracker), nil
	}
	return nil, fmt.Errorf("invalid CNI %s", cniName)
}
