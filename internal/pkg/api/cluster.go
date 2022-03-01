package api

import (
	"fmt"
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
)

type ClusterFiller func(c *anywherev1.Cluster)

func AutoFillClusterFromFile(filename string, fillers ...ClusterFiller) ([]byte, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}

	return AutoFillClusterFromYaml(content, fillers...)
}

func AutoFillClusterFromYaml(yamlContent []byte, fillers ...ClusterFiller) ([]byte, error) {
	clusterConfig, err := anywherev1.GetClusterConfigFromContent(yamlContent)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster config from content: %v", err)
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

func WithKubernetesVersion(v anywherev1.KubernetesVersion) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.KubernetesVersion = v
	}
}

func WithClusterNamespace(ns string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Namespace = ns
	}
}

func WithControlPlaneCount(r int) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ControlPlaneConfiguration.Count = r
	}
}

func WithControlPlaneEndpointIP(value string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ControlPlaneConfiguration.Endpoint.Host = value
	}
}

func WithControlPlaneTaints(taints []corev1.Taint) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ControlPlaneConfiguration.Taints = taints
	}
}

func WithControlPlaneLabel(key string, val string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ControlPlaneConfiguration.Labels == nil {
			c.Spec.ControlPlaneConfiguration.Labels = map[string]string{}
		}
		c.Spec.ControlPlaneConfiguration.Labels[key] = val
	}
}

func WithPodCidr(podCidr string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ClusterNetwork.Pods.CidrBlocks = []string{podCidr}
	}
}

func WithServiceCidr(svcCidr string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ClusterNetwork.Services.CidrBlocks = []string{svcCidr}
	}
}

func WithWorkerNodeCount(r int) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if len(c.Spec.WorkerNodeGroupConfigurations) == 0 {
			c.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{{Count: 0}}
		}
		c.Spec.WorkerNodeGroupConfigurations[0].Count = r
	}
}

func WithOIDCIdentityProviderRef(name string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.IdentityProviderRefs = append(c.Spec.IdentityProviderRefs,
			anywherev1.Ref{Name: name, Kind: anywherev1.OIDCConfigKind})
	}
}

func WithGitOpsRef(name string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.GitOpsRef = &anywherev1.Ref{Name: name, Kind: anywherev1.GitOpsConfigKind}
	}
}

func WithExternalEtcdTopology(count int) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ExternalEtcdConfiguration == nil {
			c.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{}
		}
		c.Spec.ExternalEtcdConfiguration.Count = count
	}
}

func WithStackedEtcdTopology() ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ExternalEtcdConfiguration = nil
	}
}

func WithProxyConfig(httpProxy, httpsProxy string, noProxy []string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ProxyConfiguration == nil {
			c.Spec.ProxyConfiguration = &anywherev1.ProxyConfiguration{}
		}
		c.Spec.ProxyConfiguration.HttpProxy = httpProxy
		c.Spec.ProxyConfiguration.HttpsProxy = httpProxy
		c.Spec.ProxyConfiguration.NoProxy = noProxy
	}
}

func WithRegistryMirror(endpoint, caCert string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.RegistryMirrorConfiguration == nil {
			c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{}
		}
		c.Spec.RegistryMirrorConfiguration.Endpoint = endpoint
		c.Spec.RegistryMirrorConfiguration.CACertContent = caCert
	}
}

func WithManagementCluster(name string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ManagementCluster.Name = name
	}
}

func WithAWSIamIdentityProviderRef(name string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.IdentityProviderRefs = append(c.Spec.IdentityProviderRefs,
			anywherev1.Ref{Name: name, Kind: anywherev1.AWSIamConfigKind})
	}
}

func RemoveAllWorkerNodeGroups() ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.WorkerNodeGroupConfigurations = make([]anywherev1.WorkerNodeGroupConfiguration, 0)
	}
}

func RemoveWorkerNodeGroup(name string) ClusterFiller {
	logger.Info("removing", "name", name)
	return func(c *anywherev1.Cluster) {
		logger.Info("before deleting", "w", c.Spec.WorkerNodeGroupConfigurations)
		for i, w := range c.Spec.WorkerNodeGroupConfigurations {
			if w.Name == name {
				copy(c.Spec.WorkerNodeGroupConfigurations[i:], c.Spec.WorkerNodeGroupConfigurations[i+1:])
				c.Spec.WorkerNodeGroupConfigurations[len(c.Spec.WorkerNodeGroupConfigurations)-1] = anywherev1.WorkerNodeGroupConfiguration{}
				c.Spec.WorkerNodeGroupConfigurations = c.Spec.WorkerNodeGroupConfigurations[:len(c.Spec.WorkerNodeGroupConfigurations)-1]
				logger.Info("after deleting", "w", c.Spec.WorkerNodeGroupConfigurations)
				return
			}
		}
	}
}

func WithWorkerNodeGroup(name string, fillers ...WorkerNodeGroupFiller) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		var nodeGroup *anywherev1.WorkerNodeGroupConfiguration
		position := -1
		for i, w := range c.Spec.WorkerNodeGroupConfigurations {
			if w.Name == name {
				logger.Info("Updating worker node group", "name", name)
				nodeGroup = &w
				position = i
				break
			}
		}

		if nodeGroup == nil {
			logger.Info("Adding worker node group", "name", name)
			nodeGroup = &anywherev1.WorkerNodeGroupConfiguration{Name: name}
			c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations, *nodeGroup)
			position = len(c.Spec.WorkerNodeGroupConfigurations) - 1
		}

		FillWorkerNodeGroup(nodeGroup, fillers...)
		c.Spec.WorkerNodeGroupConfigurations[position] = *nodeGroup
	}
}
