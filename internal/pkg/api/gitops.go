package api

import (
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type GitOpsConfigOpt func(o *v1alpha1.GitOpsConfig)

func NewGitOpsConfig(name string, opts ...GitOpsConfigOpt) *v1alpha1.GitOpsConfig {
	config := &v1alpha1.GitOpsConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
			Kind:       v1alpha1.GitOpsConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.GitOpsConfigSpec{},
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

func WithFluxOwner(username string) GitOpsConfigOpt {
	return func(c *v1alpha1.GitOpsConfig) {
		c.Spec.Flux.Github.Owner = username
	}
}

func WithFluxRepository(repository string) GitOpsConfigOpt {
	return func(c *v1alpha1.GitOpsConfig) {
		c.Spec.Flux.Github.Repository = repository
	}
}

func WithFluxConfigurationPath(configPath string) GitOpsConfigOpt {
	return func(c *v1alpha1.GitOpsConfig) {
		c.Spec.Flux.Github.ClusterConfigPath = configPath
	}
}

func WithFluxNamespace(namespace string) GitOpsConfigOpt {
	return func(c *v1alpha1.GitOpsConfig) {
		c.Spec.Flux.Github.FluxSystemNamespace = namespace
	}
}

func WithFluxBranch(branch string) GitOpsConfigOpt {
	return func(c *v1alpha1.GitOpsConfig) {
		c.Spec.Flux.Github.Branch = branch
	}
}

func WithPersonalFluxRepository(personal bool) GitOpsConfigOpt {
	return func(c *v1alpha1.GitOpsConfig) {
		c.Spec.Flux.Github.Personal = personal
	}
}

func WithStringFromEnvVarGitOpsConfig(envVar string, opt func(string) GitOpsConfigOpt) GitOpsConfigOpt {
	return opt(os.Getenv(envVar))
}
