package controllers

import (
	"context"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HelmFactory is responsible for creating and owning instances of Helm client,
// configured using information from the current state of the cluster using a k8s client.
type HelmFactory struct {
	client                client.Client
	dependencyHelmFactory dependencies.HelmFactory
}

// NewHelmFactory returns a new HelmFactory.
func NewHelmFactory(client client.Client, dependencyHelmFactory *dependencies.HelmFactory) *HelmFactory {
	return &HelmFactory{
		client: client,
	}
}

// GetClientForCluster returns a new Helm client configured using information from the provided cluster.
func (f *HelmFactory) GetClientForCluster(ctx context.Context, clusterName string) (*executables.Helm, error) {
	cluster := &anywherev1.Cluster{}
	namespacedNamed := types.NamespacedName{
		Name:      clusterName,
		Namespace: constants.EksaSystemNamespace,
	}

	if err := f.client.Get(ctx, namespacedNamed, cluster); err != nil {
		return nil, err
	}

	r := registrymirror.FromCluster(cluster)
	return f.dependencyHelmFactory.GetClient(executables.WithRegistryMirror(r)), nil

}
