package clusterapi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/logger"
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

func AwsIamAuthExtraArgs(awsiam *v1alpha1.AWSIamConfig) ExtraArgs {
	args := ExtraArgs{}
	if awsiam == nil {
		return args
	}
	args.AddIfNotEmpty("authentication-token-webhook-config-file", "/etc/kubernetes/aws-iam-authenticator/kubeconfig.yaml")

	return args
}

func PodIAMAuthExtraArgs(podIAMConfig *v1alpha1.PodIAMConfig) ExtraArgs {
	if podIAMConfig == nil {
		return nil
	}
	args := ExtraArgs{}
	args.AddIfNotEmpty("service-account-issuer", podIAMConfig.ServiceAccountIssuer)
	return args
}

// We don't need to add these once the Kubernetes components default to using the secure cipher suites
func SecureTlsCipherSuitesExtraArgs() ExtraArgs {
	args := ExtraArgs{}
	args.AddIfNotEmpty("tls-cipher-suites", crypto.SecureCipherSuitesString())
	return args
}

func SecureEtcdTlsCipherSuitesExtraArgs() ExtraArgs {
	args := ExtraArgs{}
	args.AddIfNotEmpty("cipher-suites", crypto.SecureCipherSuitesString())
	return args
}

func NodeLabelsExtraArgs(wnc v1alpha1.WorkerNodeGroupConfiguration) ExtraArgs {
	if wnc.Labels == nil {
		return nil
	}
	args := ExtraArgs{}
	args.AddIfNotEmpty("node-labels", labelsMapToArg(wnc))
	return args
}

func (e ExtraArgs) AddIfNotEmpty(k, v string) {
	if v != "" {
		logger.V(5).Info("Adding extraArgs", k, v)
		e[k] = v
	}
}

func (e ExtraArgs) Append(args ExtraArgs) ExtraArgs {
	for k, v := range args {
		e[k] = v
	}

	return e
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

func labelsMapToArg(n v1alpha1.WorkerNodeGroupConfiguration) string {
	var s []string
	for k, v := range n.Labels {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}

	sort.Strings(s)
	r := strings.Join(s, ",")
	return r
}
