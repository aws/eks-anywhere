package common

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	bootstrapv1 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/semver"
	"github.com/aws/eks-anywhere/pkg/types"
)

//go:embed config/audit-policy.yaml
var auditPolicy string

// TODO: Split out common into separate packages to avoid becoming a dumping ground

const (
	privateKeyFileName = "eks-a-id_rsa"
	publicKeyFileName  = "eks-a-id_rsa.pub"
)

func BootstrapClusterOpts(clusterConfig *v1alpha1.Cluster, serverEndpoints ...string) ([]bootstrapper.BootstrapClusterOption, error) {
	env := map[string]string{}
	if clusterConfig.Spec.ProxyConfiguration != nil {
		noProxyes := append([]string{}, serverEndpoints...)
		noProxyes = append(noProxyes, clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host)
		for _, s := range clusterConfig.Spec.ProxyConfiguration.NoProxy {
			if s != "" {
				noProxyes = append(noProxyes, s)
			}
		}
		env["HTTP_PROXY"] = clusterConfig.Spec.ProxyConfiguration.HttpProxy
		env["HTTPS_PROXY"] = clusterConfig.Spec.ProxyConfiguration.HttpsProxy
		env["NO_PROXY"] = strings.Join(noProxyes, ",")
	}
	return []bootstrapper.BootstrapClusterOption{bootstrapper.WithEnv(env)}, nil
}

func StripSshAuthorizedKeyComment(key string) (string, error) {
	public, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key))
	if err != nil {
		return "", err
	}
	// ssh.MarshalAuthorizedKey returns the key with a trailing newline, which we want to remove
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(public))), nil
}

func GenerateSSHAuthKey(writer filewriter.FileWriter) (string, error) {
	privateKeyPath, sshAuthorizedKeyBytes, err := crypto.NewSshKeyPairUsingFileWriter(writer, privateKeyFileName, publicKeyFileName)
	if err != nil {
		return "", fmt.Errorf("generating ssh key pair: %v", err)
	}

	logger.Info(fmt.Sprintf(
		"Private key saved to %[1]s. Use 'ssh -i %[1]s <username>@<Node-IP-Address>' to login to your cluster node",
		privateKeyPath,
	))

	key := string(sshAuthorizedKeyBytes)
	key = strings.TrimRight(key, "\n")

	return key, nil
}

func CPMachineTemplateBase(clusterName string) string {
	return fmt.Sprintf("%s-control-plane-template", clusterName)
}

func EtcdMachineTemplateBase(clusterName string) string {
	return fmt.Sprintf("%s-etcd-template", clusterName)
}

func WorkerMachineTemplateBase(clusterName, workerNodeGroupName string) string {
	return fmt.Sprintf("%s-%s", clusterName, workerNodeGroupName)
}

func CPMachineTemplateName(clusterName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%d", CPMachineTemplateBase(clusterName), t)
}

func EtcdMachineTemplateName(clusterName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%d", EtcdMachineTemplateBase(clusterName), t)
}

func WorkerMachineTemplateName(clusterName, workerNodeGroupName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%d", WorkerMachineTemplateBase(clusterName, workerNodeGroupName), t)
}

func KubeadmConfigTemplateName(clusterName, workerNodeGroupName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%s-template-%d", clusterName, workerNodeGroupName, t)
}

// GetCAPIBottlerocketSettingsConfig returns the formatted CAPI Bottlerocket settings config as a YAML marshaled string.
func GetCAPIBottlerocketSettingsConfig(config *v1alpha1.HostOSConfiguration, brKubeSettings *bootstrapv1.BottlerocketKubernetesSettings) (string, error) {
	if (config == nil || config.BottlerocketConfiguration == nil) && brKubeSettings == nil {
		return "", nil
	}

	b := &bootstrapv1.BottlerocketSettings{}
	if brKubeSettings != nil {
		b.Kubernetes = copyBottlerocketKubernetesSettings(brKubeSettings)
	}

	if config != nil {
		if config.BottlerocketConfiguration != nil {

			if config.BottlerocketConfiguration.Kernel != nil {
				if config.BottlerocketConfiguration.Kernel.SysctlSettings != nil {
					b.Kernel = &bootstrapv1.BottlerocketKernelSettings{
						SysctlSettings: config.BottlerocketConfiguration.Kernel.SysctlSettings,
					}
				}
			}

			if config.BottlerocketConfiguration.Boot != nil {
				if config.BottlerocketConfiguration.Boot.BootKernelParameters != nil {
					b.Boot = &bootstrapv1.BottlerocketBootSettings{
						BootKernelParameters: config.BottlerocketConfiguration.Boot.BootKernelParameters,
					}
				}
			}
		}
	}

	return getCAPIConfig(b)
}

func getCAPIConfig(b *bootstrapv1.BottlerocketSettings) (string, error) {
	brMap := map[string]*bootstrapv1.BottlerocketSettings{
		"bottlerocket": b,
	}

	marshaledConfig, err := yaml.Marshal(brMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal bottlerocket config: %v", err)
	}

	return strings.Trim(string(marshaledConfig), "\n"), nil
}

// GetExternalEtcdReleaseURL returns a valid etcd URL  from version bundles if the eksaVersion is greater than
// MinEksAVersionWithEtcdURL. Return "" if eksaVersion < MinEksAVersionWithEtcdURL to prevent etcd node rolled out.
func GetExternalEtcdReleaseURL(clusterVersion *v1alpha1.EksaVersion, versionBundle *cluster.VersionsBundle) (string, error) {
	if clusterVersion == nil {
		logger.V(4).Info("Eks-a cluster version is not specified. Skip setting etcd url")
		return "", nil
	}
	clusterVersionSemVer, err := semver.New(string(*clusterVersion))
	if err != nil {
		return "", fmt.Errorf("invalid semver for clusterVersion: %v", err)
	}
	minEksAVersionWithEtcdURL, err := semver.New(v1alpha1.MinEksAVersionWithEtcdURL)
	if err != nil {
		return "", fmt.Errorf("invalid semver for etcd url enabled clusterVersion: %v", err)
	}
	if err != nil {
		return "", fmt.Errorf("invalid semver for dev eksa version: %v", err)
	}
	if !(clusterVersionSemVer.LessThan(minEksAVersionWithEtcdURL)) {
		return versionBundle.KubeDistro.EtcdURL, nil
	}
	logger.V(4).Info(fmt.Sprintf("Eks-a cluster version is less than version %s. Skip setting etcd url", v1alpha1.MinEksAVersionWithEtcdURL))
	return "", nil
}

// ConvertToBottlerocketKubernetesSettings converts an unstructured object into a Bottlerocket
// Kubernetes settings object.
func ConvertToBottlerocketKubernetesSettings(kubeletConfig *unstructured.Unstructured) (*bootstrapv1.BottlerocketKubernetesSettings, error) {
	if kubeletConfig == nil {
		return nil, nil
	}
	kubeletConfigCopy, err := copyObject(kubeletConfig)
	if err != nil {
		return nil, err
	}

	delete(kubeletConfigCopy.Object, "kind")
	delete(kubeletConfigCopy.Object, "apiVersion")
	kcString, err := yaml.Marshal(kubeletConfigCopy)
	if err != nil {
		return nil, err
	}

	_, err = yaml.YAMLToJSONStrict([]byte(kcString))
	if err != nil {
		return nil, fmt.Errorf("unmarshaling the yaml, malformed yaml %v", err)
	}

	var bottlerocketKC *bootstrapv1.BottlerocketKubernetesSettings
	err = yaml.UnmarshalStrict(kcString, &bottlerocketKC)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling KubeletConfiguration for %v", err)
	}

	return bottlerocketKC, nil
}

func copyObject(kubeletConfig *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	var kubeletConfigBackup *unstructured.Unstructured

	kcString, err := yaml.Marshal(kubeletConfig)
	if err != nil {
		return nil, err
	}

	err = yaml.UnmarshalStrict(kcString, &kubeletConfigBackup)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling KubeletConfiguration for %v", err)
	}

	return kubeletConfigBackup, nil
}

func copyBottlerocketKubernetesSettings(config *bootstrapv1.BottlerocketKubernetesSettings) *bootstrapv1.BottlerocketKubernetesSettings {
	b := &bootstrapv1.BottlerocketKubernetesSettings{}
	if config != nil {
		b = &bootstrapv1.BottlerocketKubernetesSettings{
			ClusterDomain:               config.ClusterDomain,
			ContainerLogMaxFiles:        config.ContainerLogMaxFiles,
			ContainerLogMaxSize:         config.ContainerLogMaxSize,
			CPUManagerPolicy:            config.CPUManagerPolicy,
			CPUManagerPolicyOptions:     copyBottlerocketMaps(config.CPUManagerPolicyOptions),
			EventBurst:                  config.EventBurst,
			EventRecordQPS:              config.EventRecordQPS,
			EvictionHard:                copyBottlerocketMaps(config.EvictionHard),
			EvictionMaxPodGracePeriod:   config.EvictionMaxPodGracePeriod,
			EvictionSoft:                copyBottlerocketMaps(config.EvictionSoft),
			EvictionSoftGracePeriod:     copyBottlerocketMaps(config.EvictionSoftGracePeriod),
			ImageGCHighThresholdPercent: config.ImageGCHighThresholdPercent,
			ImageGCLowThresholdPercent:  config.ImageGCLowThresholdPercent,
			KubeAPIBurst:                config.KubeAPIBurst,
			KubeAPIQPS:                  config.KubeAPIQPS,
			KubeReserved:                copyBottlerocketMaps(config.KubeReserved),
			MaxPods:                     config.MaxPods,
			MemoryManagerPolicy:         config.MemoryManagerPolicy,
			PodPidsLimit:                config.PodPidsLimit,
			RegistryBurst:               config.RegistryBurst,
			RegistryPullQPS:             config.RegistryPullQPS,
			SystemReserved:              copyBottlerocketMaps(config.SystemReserved),
			TopologyManagerPolicy:       config.TopologyManagerPolicy,
			TopologyManagerScope:        config.TopologyManagerScope,
		}

		if len(config.AllowedUnsafeSysctls) > 0 {
			b.AllowedUnsafeSysctls = config.AllowedUnsafeSysctls
		}
		if len(config.ClusterDNSIPs) > 0 {
			b.ClusterDNSIPs = config.ClusterDNSIPs
		}
		if config.CPUCFSQuota != nil {
			b.CPUCFSQuota = config.CPUCFSQuota
		}
		if config.CPUManagerReconcilePeriod != nil {
			b.CPUManagerReconcilePeriod = config.CPUManagerReconcilePeriod
		}
		if config.ShutdownGracePeriod != nil {
			b.ShutdownGracePeriod = config.ShutdownGracePeriod
		}
		if config.ShutdownGracePeriodCriticalPods != nil {
			b.ShutdownGracePeriodCriticalPods = config.ShutdownGracePeriodCriticalPods
		}
	}

	return b
}

func copyBottlerocketMaps(source map[string]string) map[string]string {
	dst := make(map[string]string)
	for key, value := range source {
		dst[key] = value
	}
	return dst
}
