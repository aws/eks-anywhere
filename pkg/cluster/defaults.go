package cluster

import (
	"path"
	"reflect"
)

type configDefaulter func(*Config)

var configDefaulters = []configDefaulter{
	SetDefaultFluxGitHubConfigPath,
}

func SetConfigDefaults(c *Config) {
	for _, d := range getDefaultables(c) {
		d.SetDefaults()
	}

	for _, d := range configDefaulters {
		d(c)
	}
}

type defaultable interface {
	SetDefaults()
}

func getDefaultables(c *Config) []defaultable {
	d := make([]defaultable, 0, 1)
	d = appendDefaultablesIfNotNil(d, c.Cluster, c.VSphereDatacenter, c.GitOpsConfig)

	for _, e := range c.AWSIAMConfigs {
		d = appendDefaultablesIfNotNil(d, e)
	}

	return d
}

func appendDefaultablesIfNotNil(defaultables []defaultable, elems ...defaultable) []defaultable {
	for _, e := range elems {
		// Since we receive interfaces, these will never be nil since they contain
		// the type of the original implementing struct
		// I can't find another clean option of doing this
		if !reflect.ValueOf(e).IsNil() {
			defaultables = append(defaultables, e)
		}
	}

	return defaultables
}

func SetDefaultFluxGitHubConfigPath(c *Config) {
	if c.GitOpsConfig == nil {
		return
	}

	gitops := c.GitOpsConfig
	if gitops.Spec.Flux.Github.ClusterConfigPath != "" {
		return
	}

	if c.Cluster.IsSelfManaged() {
		gitops.Spec.Flux.Github.ClusterConfigPath = path.Join("clusters", c.Cluster.Name)
	} else {
		gitops.Spec.Flux.Github.ClusterConfigPath = path.Join("clusters", c.Cluster.ManagedBy())
	}
}
