package cluster

import (
	"context"
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
			setGitOpsDefaults,
			SetDefaultFluxGitHubConfigPath,
		},
		Validations: []Validation{
			validateGitOps,
			validateGitOpsNamespace,
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

		// GitOpsConfig will be deprecated.
		// During the deprecation window, FluxConfig will be used internally
		// GitOpsConfig will preserved it was in the original spec
		gitOpsConf := gitOps.(*anywherev1.GitOpsConfig)
		c.GitOpsConfig = gitOpsConf
		c.FluxConfig = gitOpsConf.ConvertToFluxConfig()
	}
}

func validateGitOps(c *Config) error {
	if c.GitOpsConfig != nil {
		return c.GitOpsConfig.Validate()
	}
	return nil
}

func validateGitOpsNamespace(c *Config) error {
	if c.GitOpsConfig != nil {
		if err := validateSameNamespace(c, c.GitOpsConfig); err != nil {
			return err
		}
	}
	return nil
}

func setGitOpsDefaults(c *Config) error {
	if c.GitOpsConfig != nil {
		c.GitOpsConfig.SetDefaults()
	}
	return nil
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

func getGitOps(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.GitOpsRef == nil || c.Cluster.Spec.GitOpsRef.Kind != anywherev1.GitOpsConfigKind {
		return nil
	}

	gitOps := &anywherev1.GitOpsConfig{}
	if err := client.Get(ctx, c.Cluster.Spec.GitOpsRef.Name, c.Cluster.Namespace, gitOps); err != nil {
		return err
	}

	c.GitOpsConfig = gitOps

	return nil
}
