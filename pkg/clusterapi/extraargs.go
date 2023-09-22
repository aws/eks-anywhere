package clusterapi

import (
	"fmt"
	"sort"
	"strconv"
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

// EtcdEncryptionExtraArgs takes a list of EtcdEncryption configs and returns the relevant API server extra args if it's not nil or empty.
func EtcdEncryptionExtraArgs(config *[]v1alpha1.EtcdEncryption) ExtraArgs {
	args := ExtraArgs{}
	if config == nil || len(*config) == 0 {
		return args
	}
	args.AddIfNotEmpty("encryption-provider-config", "/etc/kubernetes/enc/encryption-config.yaml")
	return args
}

// FeatureGatesExtraArgs takes a list of features with the value and returns it in the proper format
// Example FeatureGatesExtraArgs("ServiceLoadBalancerClass=true").
func FeatureGatesExtraArgs(features ...string) ExtraArgs {
	if len(features) == 0 {
		return nil
	}
	return ExtraArgs{
		"feature-gates": strings.Join(features[:], ","),
	}
}

func PodIAMAuthExtraArgs(podIAMConfig *v1alpha1.PodIAMConfig) ExtraArgs {
	if podIAMConfig == nil {
		return nil
	}
	args := ExtraArgs{}
	args.AddIfNotEmpty("service-account-issuer", podIAMConfig.ServiceAccountIssuer)
	return args
}

func NodeCIDRMaskExtraArgs(clusterNetwork *v1alpha1.ClusterNetwork) ExtraArgs {
	if clusterNetwork == nil || clusterNetwork.Nodes == nil || clusterNetwork.Nodes.CIDRMaskSize == nil {
		return nil
	}
	args := ExtraArgs{}
	args.AddIfNotEmpty("node-cidr-mask-size", strconv.Itoa(*clusterNetwork.Nodes.CIDRMaskSize))
	return args
}

func ResolvConfExtraArgs(resolvConf *v1alpha1.ResolvConf) ExtraArgs {
	if resolvConf == nil {
		return nil
	}
	args := ExtraArgs{}
	args.AddIfNotEmpty("resolv-conf", resolvConf.Path)
	return args
}

// We don't need to add these once the Kubernetes components default to using the secure cipher suites.
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

func WorkerNodeLabelsExtraArgs(wnc v1alpha1.WorkerNodeGroupConfiguration) ExtraArgs {
	return nodeLabelsExtraArgs(wnc.Labels)
}

func ControlPlaneNodeLabelsExtraArgs(cpc v1alpha1.ControlPlaneConfiguration) ExtraArgs {
	return nodeLabelsExtraArgs(cpc.Labels)
}

// CgroupDriverExtraArgs args added for kube versions below 1.24.
func CgroupDriverCgroupfsExtraArgs() ExtraArgs {
	args := ExtraArgs{}
	args.AddIfNotEmpty("cgroup-driver", "cgroupfs")
	return args
}

// CgroupDriverSystemdExtraArgs args added for kube versions 1.24 and above.
func CgroupDriverSystemdExtraArgs() ExtraArgs {
	args := ExtraArgs{}
	args.AddIfNotEmpty("cgroup-driver", "systemd")
	return args
}

func nodeLabelsExtraArgs(labels map[string]string) ExtraArgs {
	args := ExtraArgs{}
	args.AddIfNotEmpty("node-labels", labelsMapToArg(labels))
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

func labelsMapToArg(m map[string]string) string {
	labels := make([]string, 0, len(m))
	for k, v := range m {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}

	sort.Strings(labels)
	labelStr := strings.Join(labels, ",")
	return labelStr
}
