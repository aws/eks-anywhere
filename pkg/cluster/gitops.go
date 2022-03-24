package cluster

import (
	"path"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func gitOpsEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			anywherev1.GitOpsConfigKind: func() APIObject {
				return &anywherev1.GitOpsConfig{}
			},
		},
		Processors: []ParsedProcessor{processGitOps},
		Defaulters: []Defaulter{
			func(c *Config) error {
				if c.GitOpsConfig != nil {
					c.GitOpsConfig.SetDefaults()
				}
				return nil
			},
			SetDefaultFluxGitHubConfigPath,
		},
		Validations: []Validation{
			func(c *Config) error {
				if c.GitOpsConfig != nil {
					return c.GitOpsConfig.Validate()
				}
				return nil
			},
			func(c *Config) error {
				if c.GitOpsConfig != nil {
					if err := validateSameNamespace(c, c.GitOpsConfig); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}

func processGitOps(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.GitOpsRef == nil {
		return
	}

	if c.Cluster.Spec.GitOpsRef.Kind == anywherev1.GitOpsConfigKind {
		gitOps := objects.GetFromRef(c.Cluster.APIVersion, *c.Cluster.Spec.GitOpsRef)
		if gitOps == nil {
			return
		}

		c.GitOpsConfig = gitOps.(*anywherev1.GitOpsConfig)
	}
}

func SetDefaultFluxGitHubConfigPath(c *Config) error {
	if c.GitOpsConfig == nil {
		return nil
	}

	gitops := c.GitOpsConfig
	if gitops.Spec.Flux.Github.ClusterConfigPath != "" {
		return nil
	}

	if c.Cluster.IsSelfManaged() {
		gitops.Spec.Flux.Github.ClusterConfigPath = path.Join("clusters", c.Cluster.Name)
	} else {
		gitops.Spec.Flux.Github.ClusterConfigPath = path.Join("clusters", c.Cluster.ManagedBy())
	}
	return nil
}
