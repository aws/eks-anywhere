package api

import (
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

// OIDCConfigOpt updates an OIDC config.
type OIDCConfigOpt func(o *anywherev1.OIDCConfig)

// WithOIDCConfig builds a ClusterConfigFiller that adds a OIDCConfig with the
// given name and spec to the cluster config.
func WithOIDCConfig(name string, opts ...OIDCConfigOpt) ClusterConfigFiller {
	return func(c *cluster.Config) {
		if c.OIDCConfigs == nil {
			c.OIDCConfigs = make(map[string]*anywherev1.OIDCConfig, 1)
		}

		oidc, ok := c.OIDCConfigs[name]
		if !ok {
			oidc = &anywherev1.OIDCConfig{
				TypeMeta: metav1.TypeMeta{
					APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
					Kind:       anywherev1.OIDCConfigKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
			}
			c.OIDCConfigs[name] = oidc
		}

		for _, opt := range opts {
			opt(oidc)
		}
	}
}

func WithOIDCClientId(id string) OIDCConfigOpt {
	return func(o *anywherev1.OIDCConfig) {
		o.Spec.ClientId = id
	}
}

func WithOIDCIssuerUrl(url string) OIDCConfigOpt {
	return func(o *anywherev1.OIDCConfig) {
		o.Spec.IssuerUrl = url
	}
}

func WithOIDCUsernameClaim(claim string) OIDCConfigOpt {
	return func(o *anywherev1.OIDCConfig) {
		o.Spec.UsernameClaim = claim
	}
}

func WithOIDCUsernamePrefix(prefix string) OIDCConfigOpt {
	return func(o *anywherev1.OIDCConfig) {
		o.Spec.UsernamePrefix = prefix
	}
}

func WithOIDCGroupsClaim(claim string) OIDCConfigOpt {
	return func(o *anywherev1.OIDCConfig) {
		o.Spec.GroupsClaim = claim
	}
}

func WithOIDCGroupsPrefix(prefix string) OIDCConfigOpt {
	return func(o *anywherev1.OIDCConfig) {
		o.Spec.GroupsPrefix = prefix
	}
}

func WithOIDCRequiredClaims(claim, value string) OIDCConfigOpt {
	return func(o *anywherev1.OIDCConfig) {
		o.Spec.RequiredClaims = []anywherev1.OIDCConfigRequiredClaim{{
			Claim: claim,
			Value: value,
		}}
	}
}

func WithStringFromEnvVarOIDCConfig(envVar string, opt func(string) OIDCConfigOpt) OIDCConfigOpt {
	return opt(os.Getenv(envVar))
}
