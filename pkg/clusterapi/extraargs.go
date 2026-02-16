package clusterapi

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/logger"
)

type ExtraArgs map[string]string

// ToArgs converts ExtraArgs (map[string]string) to the v1beta2 []bootstrapv1.Arg format.
// The output is sorted by name for deterministic ordering.
// SortArgs sorts a slice of bootstrapv1beta2.Arg by name for deterministic ordering.
func SortArgs(args []bootstrapv1beta2.Arg) {
	sort.Slice(args, func(i, j int) bool {
		return args[i].Name < args[j].Name
	})
}

func (e ExtraArgs) ToArgs() []bootstrapv1beta2.Arg {
	if len(e) == 0 {
		return nil
	}
	args := make([]bootstrapv1beta2.Arg, 0, len(e))
	for k, v := range e {
		v := v // copy for pointer
		args = append(args, bootstrapv1beta2.Arg{Name: k, Value: &v})
	}
	sort.Slice(args, func(i, j int) bool {
		return args[i].Name < args[j].Name
	})
	return args
}

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

// APIServerExtraArgs takes a map of API Server extra args and returns the relevant API server extra args if it's not nil or empty.
func APIServerExtraArgs(apiServerExtraArgs map[string]string) ExtraArgs {
	args := ExtraArgs{}
	for k, v := range apiServerExtraArgs {
		args.AddIfNotEmpty(k, v)
	}
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

// SetPodIAMAuthExtraArgs sets the api server extra args for the podIAMConfig.
func SetPodIAMAuthExtraArgs(podIAMConfig *v1alpha1.PodIAMConfig, apiServerExtraArgs map[string]string) {
	if podIAMFlags := PodIAMAuthExtraArgs(podIAMConfig); podIAMFlags != nil {
		if v, has := apiServerExtraArgs["service-account-issuer"]; has {
			apiServerExtraArgs["service-account-issuer"] = strings.Join([]string{v, podIAMFlags["service-account-issuer"]}, ",")
		} else {
			apiServerExtraArgs["service-account-issuer"] = podIAMFlags["service-account-issuer"]
		}
	}
}

// SetPodIAMAuthInArgs merges PodIAM service-account-issuer into an existing []Arg slice.
// If service-account-issuer already exists, it concatenates the values with a comma
// (matching the behavior of SetPodIAMAuthExtraArgs for map[string]string).
func SetPodIAMAuthInArgs(podIAMConfig *v1alpha1.PodIAMConfig, args []bootstrapv1beta2.Arg) []bootstrapv1beta2.Arg {
	podIAMFlags := PodIAMAuthExtraArgs(podIAMConfig)
	if podIAMFlags == nil {
		return args
	}

	issuerValue := podIAMFlags["service-account-issuer"]
	for i, arg := range args {
		if arg.Name == "service-account-issuer" && arg.Value != nil {
			merged := strings.Join([]string{*arg.Value, issuerValue}, ",")
			args[i].Value = &merged
			return args
		}
	}
	return append(args, bootstrapv1beta2.Arg{Name: "service-account-issuer", Value: &issuerValue})
}

// ToYaml outputs ExtraArgs as v1beta2 []Arg YAML format (list of name/value pairs).
// Output is sorted by name for deterministic ordering.
func (e ExtraArgs) ToYaml() string {
	if len(e) == 0 {
		return ""
	}
	// Sort keys for deterministic output
	keys := make([]string, 0, len(e))
	for k := range e {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		v := e[k]
		fmt.Fprintf(&b, "- name: %s\n  value: \"%s\"\n", k, v)
	}
	return strings.TrimRight(b.String(), "\n")
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
