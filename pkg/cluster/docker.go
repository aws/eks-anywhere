package cluster

import anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"

func dockerEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.DockerDatacenterKind: func() APIObject {
				return &anywherev1.DockerDatacenterConfig{}
			},
		},
		Processors: []ParsedProcessor{
			processDockerDatacenter,
		},
	}
}

func processDockerDatacenter(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.DatacenterRef.Kind == anywherev1.DockerDatacenterKind {
		datacenter := objects.GetFromRef(c.Cluster.APIVersion, c.Cluster.Spec.DatacenterRef)
		c.DockerDatacenter = datacenter.(*anywherev1.DockerDatacenterConfig)
	}
}
