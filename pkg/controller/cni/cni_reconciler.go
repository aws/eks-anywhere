package cni

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	eksacluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/controller"
)

type CNIReconciler interface {
	Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster, client client.Client, specWithBundles *eksacluster.Spec) (controller.Result, error)
}

func BuildCNIReconciler(cniName string) (CNIReconciler, error) {
	if cniName == "cilium" {
		return NewCiliumReconciler(), nil
	}
	return nil, fmt.Errorf("invalid CNI %s", cniName)
}
