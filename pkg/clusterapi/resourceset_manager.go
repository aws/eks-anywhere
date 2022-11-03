package clusterapi

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	addons "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

// ResourceSetManager helps managing capi ClusterResourceSet's
// It doesn't implement the complete ClusterResourceSet specification so there might be some
// configurations that are not supported. JsonLists as content in resources are not supported.
type ResourceSetManager struct {
	client Client
}

func NewResourceSetManager(client Client) *ResourceSetManager {
	return &ResourceSetManager{
		client: client,
	}
}

type Client interface {
	GetClusterResourceSet(ctx context.Context, kubeconfigFile, name, namespace string) (*addons.ClusterResourceSet, error)
	GetConfigMap(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.ConfigMap, error)
	GetSecretFromNamespace(ctx context.Context, kubeconfigFile, name, namespace string) (*corev1.Secret, error)
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
}

func (r *ResourceSetManager) ForceUpdate(ctx context.Context, name, namespace string, managementCluster, workloadCluster *types.Cluster) error {
	logger.V(3).Info("Reconciling ClusterResourceSet", "name", name)
	resourceSet, err := r.client.GetClusterResourceSet(ctx, managementCluster.KubeconfigFile, name, namespace)
	if err != nil {
		return fmt.Errorf("failed getting resourceset to update it: %v", err)
	}

	for _, resource := range resourceSet.Spec.Resources {
		logger.V(3).Info("Reconciling CRS managed resource", "name", resource.Name)
		objects, err := r.getResources(ctx, resource, namespace, managementCluster)
		if err != nil {
			return err
		}

		for _, object := range objects {
			if err := r.client.ApplyKubeSpecFromBytes(ctx, workloadCluster, object); err != nil {
				return fmt.Errorf("failed force updating object from ClusterResourceSet: %v", err)
			}
		}
	}

	return nil
}

type unstructuredObject []byte

func (r *ResourceSetManager) getResources(ctx context.Context, resource addons.ResourceRef, namespace string, cluster *types.Cluster) ([]unstructuredObject, error) {
	switch addons.ClusterResourceSetResourceKind(resource.Kind) {
	case addons.SecretClusterResourceSetResourceKind:
		return r.getResourcesFromSecret(ctx, resource.Name, namespace, cluster)
	case addons.ConfigMapClusterResourceSetResourceKind:
		return r.getResourcesFromConfigMap(ctx, resource.Name, namespace, cluster)
	default:
		return nil, fmt.Errorf("invalid type [%s] for resource in ClusterResourceSet", resource.Kind)
	}
}

func (r *ResourceSetManager) getResourcesFromConfigMap(ctx context.Context, name, namespace string, cluster *types.Cluster) ([]unstructuredObject, error) {
	logger.V(4).Info("Getting objects from CRS managed configmap", "name", name)
	configMap, err := r.client.GetConfigMap(ctx, cluster.KubeconfigFile, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed getting config map from resource set: %v", err)
	}

	return extractResourcesFromData(toMapOfBytes(configMap.Data)), nil
}

func (r *ResourceSetManager) getResourcesFromSecret(ctx context.Context, name, namespace string, cluster *types.Cluster) ([]unstructuredObject, error) {
	logger.V(4).Info("Getting objects from CRS managed secret", "name", name)
	secret, err := r.client.GetSecretFromNamespace(ctx, cluster.KubeconfigFile, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed getting secret from resource set: %v", err)
	}

	// Secret's content don't need to be decoded if they are unmarshalled into a corev1.Secret
	return extractResourcesFromData(secret.Data), nil
}

func toMapOfBytes(data map[string]string) map[string][]byte {
	d := make(map[string][]byte, len(data))
	for k, v := range data {
		d[k] = []byte(v)
	}

	return d
}

func extractResourcesFromData(data map[string][]byte) []unstructuredObject {
	resources := make([]unstructuredObject, 0, len(data))
	for _, r := range data {
		resources = append(resources, unstructuredObject(r))
	}

	return resources
}
