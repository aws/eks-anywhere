package clusters

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/controllers/controllers/reconciler"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
)

type ProviderClusterReconciler interface {
	Reconcile(ctx context.Context, cluster *anywherev1.Cluster) (reconciler.Result, error)
}

func BuildProviderReconciler(datacenterKind string, client client.Client, log logr.Logger, validator *vsphere.Validator, defaulter *vsphere.Defaulter, tracker *remote.ClusterCacheTracker) (ProviderClusterReconciler, error) {
	switch datacenterKind {
	case anywherev1.VSphereDatacenterKind:
		return NewVSphereReconciler(client, log, validator, defaulter, tracker), nil
	}
	return nil, fmt.Errorf("invalid data center type %s", datacenterKind)
}

type providerClusterReconciler struct{}
