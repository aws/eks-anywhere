package api

import (
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type FluxConfigOpt func(o *v1alpha1.FluxConfig)

func NewFluxConfig(name string, opts ...FluxConfigOpt) *v1alpha1.FluxConfig {
	config := &v1alpha1.FluxConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
			Kind:       v1alpha1.FluxConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.FluxConfigSpec{},
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

func WithFluxConfigNamespace(ns string) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Namespace = ns
	}
}

func WithBranch(branch string) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Spec.Branch = branch
	}
}

func WithClusterConfigPath(configPath string) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Spec.ClusterConfigPath = configPath
	}
}

func WithSystemNamespace(namespace string) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Spec.SystemNamespace = namespace
	}
}

func WithGithubOwner(owner string) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Spec.Github.Owner = owner
	}
}

func WithGithubRepository(repository string) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Spec.Github.Repository = repository
	}
}

func WithPersonalGithubRepository(personal bool) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Spec.Github.Personal = personal
	}
}

func WithGitUsername(username string) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Spec.Git.Username = username
	}
}

func WithGitRepositoryUrl(url string) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Spec.Git.RepositoryUrl = url
	}
}

func WithStringFromEnvVarFluxConfig(envVar string, opt func(string) FluxConfigOpt) FluxConfigOpt {
	return opt(os.Getenv(envVar))
}
