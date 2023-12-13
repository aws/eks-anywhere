package cluster

import (
	"context"
	"net"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// Helm is underlying executable used to perform Helm operations.
type Helm interface {
	Template(ctx context.Context, ociURI, version, namespace string, values interface{}, kubeVersion string) ([]byte, error)
	RegistryLogin(ctx context.Context, registry, username, password string) error
}

// HelmClient is responsible for perfoming helm registry operations for a cluster.
type HelmClient struct {
	helm     Helm
	cluster  *anywherev1.Cluster
	username string
	password string
}

// NewHelmClient returns a new HelmClient.
func NewHelmClient(helm Helm, cluster *anywherev1.Cluster, username string, password string) *HelmClient {
	return &HelmClient{
		helm:     helm,
		cluster:  cluster,
		username: username,
		password: password,
	}
}

// Template renders the helm chart templates by running the helm template command and returns the output.
func (hc *HelmClient) Template(ctx context.Context, ociURI, version, namespace string, values interface{}, kubeVersion string) ([]byte, error) {
	return hc.helm.Template(ctx, ociURI, version, namespace, values, kubeVersion)
}

// RegistryLoginIfNeeded authenticates the client to the Cluster's helm registry.
func (hc *HelmClient) RegistryLoginIfNeeded(ctx context.Context) error {
	if hc.cluster.Spec.RegistryMirrorConfiguration != nil && hc.cluster.Spec.RegistryMirrorConfiguration.Authenticate {
		registryMirror := hc.cluster.Spec.RegistryMirrorConfiguration
		endpoint := net.JoinHostPort(registryMirror.Endpoint, registryMirror.Port)
		return hc.helm.RegistryLogin(ctx, endpoint, hc.username, hc.password)
	}
	return nil
}
