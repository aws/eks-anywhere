package helm

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	configcli "github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
	"github.com/aws/eks-anywhere/pkg/registrymirror"
)

// ClientBuilder builds a Helm Client.
type ClientBuilder interface {
	BuildHelm(...Opt) Client
}

// ClientFactory provides a helm client for a cluster.
type ClientFactory struct {
	client  client.Client
	builder ClientBuilder
}

// NewClientForClusterFactory returns a new helm ClientFactory.
func NewClientForClusterFactory(client client.Client, builder ClientBuilder) *ClientFactory {
	hf := &ClientFactory{
		client:  client,
		builder: builder,
	}
	return hf
}

// Get returns a new Helm client configured using information from the provided cluster's management cluster.
func (f *ClientFactory) Get(ctx context.Context, clus *anywherev1.Cluster) (Client, error) {
	managmentCluster := clus

	var err error
	if clus.IsManaged() {
		managmentCluster, err = clusters.FetchManagementEksaCluster(ctx, f.client, clus)
		if err != nil {
			return nil, err
		}
	}

	var rUsername, rPassword string
	if managmentCluster.RegistryAuth() {
		rUsername, rPassword, err = configcli.ReadCredentialsFromSecret(ctx, f.client)
		if err != nil {
			return nil, err
		}
	}

	r := registrymirror.FromCluster(managmentCluster)
	helmClient := f.builder.BuildHelm(WithRegistryMirror(r), WithInsecure())

	if r != nil && managmentCluster.RegistryAuth() {
		if err := helmClient.RegistryLogin(ctx, r.BaseRegistry, rUsername, rPassword); err != nil {
			return nil, err
		}
	}

	return helmClient, nil
}
