package clusters

import (
	"context"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProviderClusterReconciler struct {
	providerClient client.Client
}

func NewProviderClusterReconciler(client client.Client) *ProviderClusterReconciler {
	return &ProviderClusterReconciler{
		providerClient: client,
	}
}

func (p *ProviderClusterReconciler) GetEksdRelease(ctx context.Context, name, namespace string) (*eksdv1alpha1.Release, error) {
	eksd := &eksdv1alpha1.Release{}
	releaseName := types.NamespacedName{Namespace: namespace, Name: name}

	if err := p.providerClient.Get(ctx, releaseName, eksd); err != nil {
		return nil, err
	}

	return eksd, nil
}
