package api

import (
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type OIDCConfigOpt func(o *v1alpha1.OIDCConfig)

func NewOIDCConfig(name string, opts ...OIDCConfigOpt) *v1alpha1.OIDCConfig {
	config := &v1alpha1.OIDCConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
			Kind:       v1alpha1.OIDCConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.OIDCConfigSpec{},
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

func WithOIDCClientId(id string) OIDCConfigOpt {
	return func(o *v1alpha1.OIDCConfig) {
		o.Spec.ClientId = id
	}
}

func WithOIDCIssuerUrl(url string) OIDCConfigOpt {
	return func(o *v1alpha1.OIDCConfig) {
		o.Spec.IssuerUrl = url
	}
}

func WithOIDCUsernameClaim(claim string) OIDCConfigOpt {
	return func(o *v1alpha1.OIDCConfig) {
		o.Spec.UsernameClaim = claim
	}
}

func WithOIDCUsernamePrefix(prefix string) OIDCConfigOpt {
	return func(o *v1alpha1.OIDCConfig) {
		o.Spec.UsernamePrefix = prefix
	}
}

func WithOIDCGroupsClaim(claim string) OIDCConfigOpt {
	return func(o *v1alpha1.OIDCConfig) {
		o.Spec.GroupsClaim = claim
	}
}

func WithOIDCGroupsPrefix(prefix string) OIDCConfigOpt {
	return func(o *v1alpha1.OIDCConfig) {
		o.Spec.GroupsPrefix = prefix
	}
}

func WithOIDCRequiredClaims(claim, value string) OIDCConfigOpt {
	return func(o *v1alpha1.OIDCConfig) {
		o.Spec.RequiredClaims = []v1alpha1.OIDCConfigRequiredClaim{{
			Claim: claim,
			Value: value,
		}}
	}
}

func WithStringFromEnvVarOIDCConfig(envVar string, opt func(string) OIDCConfigOpt) OIDCConfigOpt {
	return opt(os.Getenv(envVar))
}
