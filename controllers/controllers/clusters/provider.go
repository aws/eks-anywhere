package clusters

import (
	"context"
	"fmt"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
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

type providerClusterReconciler struct {
	providerClient client.Client
}

func (p *providerClusterReconciler) eksdRelease(ctx context.Context, name, namespace string) (*eksdv1alpha1.Release, error) {
	eksd := &eksdv1alpha1.Release{}
	releaseName := types.NamespacedName{Namespace: namespace, Name: name}

	if err := p.providerClient.Get(ctx, releaseName, eksd); err != nil {
		return nil, err
	}

	return eksd, nil
}
