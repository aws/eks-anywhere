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

func WithFluxConfigName(n string) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		c.Name = n
	}
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

func WithStringFromEnvVarFluxConfig(envVar string, opt func(string) FluxConfigOpt) FluxConfigOpt {
	return opt(os.Getenv(envVar))
}

type GitProviderOpt func(o *v1alpha1.GitProviderConfig)

func WithGenericGitProvider(opts ...GitProviderOpt) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		g := &v1alpha1.GitProviderConfig{}
		for _, opt := range opts {
			opt(g)
		}
		c.Spec.Git = g
	}
}

func WithGitRepositoryUrl(url string) GitProviderOpt {
	return func(c *v1alpha1.GitProviderConfig) {
		c.RepositoryUrl = url
	}
}

func WithStringFromEnvVarGenericGitProviderConfig(envVar string, opt func(string) GitProviderOpt) GitProviderOpt {
	return opt(os.Getenv(envVar))
}

type GithubProviderOpt func(o *v1alpha1.GithubProviderConfig)

func WithGithubProvider(opts ...GithubProviderOpt) FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		g := &v1alpha1.GithubProviderConfig{}
		for _, opt := range opts {
			opt(g)
		}
		c.Spec.Github = g
	}
}

func WithGithubOwner(owner string) GithubProviderOpt {
	return func(c *v1alpha1.GithubProviderConfig) {
		c.Owner = owner
	}
}

func WithGithubRepository(repository string) GithubProviderOpt {
	return func(c *v1alpha1.GithubProviderConfig) {
		c.Repository = repository
	}
}

func WithPersonalGithubRepository(personal bool) GithubProviderOpt {
	return func(c *v1alpha1.GithubProviderConfig) {
		c.Personal = personal
	}
}

func WithStringFromEnvVarGithubProviderConfig(envVar string, opt func(string) GithubProviderOpt) GithubProviderOpt {
	return opt(os.Getenv(envVar))
}
