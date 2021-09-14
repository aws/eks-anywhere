package api

import (
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type ClusterFiller func(c *v1alpha1.Cluster)

func AutoFillCluster(filename string, fillers ...ClusterFiller) ([]byte, error) {
	clusterConfig, err := v1alpha1.GetAndValidateClusterConfig(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from file: %v", err)
	}

	for _, f := range fillers {
		f(clusterConfig)
	}

	clusterOutput, err := yaml.Marshal(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("error marshalling cluster config: %v", err)
	}

	return clusterOutput, nil
}

func WithKubernetesVersion(v v1alpha1.KubernetesVersion) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		c.Spec.KubernetesVersion = v
	}
}

func WithClusterNamespace(ns string) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		c.Namespace = ns
	}
}

func WithControlPlaneCount(r int) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		c.Spec.ControlPlaneConfiguration.Count = r
	}
}

func WithControlPlaneEndpointIP(value string) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		c.Spec.ControlPlaneConfiguration.Endpoint.Host = value
	}
}

func WithPodCidr(podCidr string) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		c.Spec.ClusterNetwork.Pods.CidrBlocks = []string{podCidr}
	}
}

func WithServiceCidr(svcCidr string) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		c.Spec.ClusterNetwork.Services.CidrBlocks = []string{svcCidr}
	}
}

func WithWorkerNodeCount(r int) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		if len(c.Spec.WorkerNodeGroupConfigurations) == 0 {
			c.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{{Count: 0}}
		}
		c.Spec.WorkerNodeGroupConfigurations[0].Count = r
	}
}

func WithOIDCIdentityProviderRef(name string) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		c.Spec.IdentityProviderRefs = append(c.Spec.IdentityProviderRefs,
			v1alpha1.Ref{Name: name, Kind: v1alpha1.OIDCConfigKind})
	}
}

func WithGitOpsRef(name string) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		c.Spec.GitOpsRef = &v1alpha1.Ref{Name: name, Kind: v1alpha1.GitOpsConfigKind}
	}
}

func WithExternalEtcdTopology(count int) ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		if c.Spec.ExternalEtcdConfiguration == nil {
			c.Spec.ExternalEtcdConfiguration = &v1alpha1.ExternalEtcdConfiguration{}
		}
		c.Spec.ExternalEtcdConfiguration.Count = count
	}
}

func WithStackedEtcdTopology() ClusterFiller {
	return func(c *v1alpha1.Cluster) {
		c.Spec.ExternalEtcdConfiguration = nil
	}
}
