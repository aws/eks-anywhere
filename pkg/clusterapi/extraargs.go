package clusterapi

import (
	"fmt"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/templater"
)

type ExtraArgs map[string]string

func OIDCToExtraArgs(oidc *v1alpha1.OIDCConfig) ExtraArgs {
	args := ExtraArgs{}
	if oidc == nil {
		return args
	}

	args.AddIfNotEmpty("oidc-client-id", oidc.Spec.ClientId)
	args.AddIfNotEmpty("oidc-groups-claim", oidc.Spec.GroupsClaim)
	args.AddIfNotEmpty("oidc-groups-prefix", oidc.Spec.GroupsPrefix)
	args.AddIfNotEmpty("oidc-issuer-url", oidc.Spec.IssuerUrl)
	if len(oidc.Spec.RequiredClaims) > 0 {
		args.AddIfNotEmpty("oidc-required-claim", requiredClaimToArg(&oidc.Spec.RequiredClaims[0]))
	}
	args.AddIfNotEmpty("oidc-username-claim", oidc.Spec.UsernameClaim)
	args.AddIfNotEmpty("oidc-username-prefix", oidc.Spec.UsernamePrefix)

	return args
}

func (e ExtraArgs) AddIfNotEmpty(k, v string) {
	if v != "" {
		e[k] = v
	}
}

func (e ExtraArgs) ToPartialYaml() templater.PartialYaml {
	p := templater.PartialYaml{}
	for k, v := range e {
		p.AddIfNotZero(k, v)
	}
	return p
}

func requiredClaimToArg(r *v1alpha1.OIDCConfigRequiredClaim) string {
	if r == nil || r.Claim == "" {
		return ""
	}

	return fmt.Sprintf("%s=%s", r.Claim, r.Value)
}
