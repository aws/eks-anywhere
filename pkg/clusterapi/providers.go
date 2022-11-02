package clusterapi

import (
	"context"

	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubeLister interface {
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
}

// GetProviders lists all installed CAPI providers across all namespaces from the kube-api server.
func GetProviders(ctx context.Context, client KubeLister) ([]clusterctlv1.Provider, error) {
	providersList := &clusterctlv1.ProviderList{}
	err := client.List(ctx, providersList)
	if err != nil {
		return nil, err
	}

	return providersList.Items, nil
}
