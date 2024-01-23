package common

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
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
func GetCAPIBottlerocketSettingsConfig(config *v1alpha1.BottlerocketConfiguration) (string, error) {
	if config == nil {
		return "", nil
	}

	b := &v1beta1.BottlerocketSettings{}
	if config.Kubernetes != nil {
		b.Kubernetes = &v1beta1.BottlerocketKubernetesSettings{
			MaxPods: config.Kubernetes.MaxPods,
		}
		if len(config.Kubernetes.AllowedUnsafeSysctls) > 0 {
			b.Kubernetes.AllowedUnsafeSysctls = config.Kubernetes.AllowedUnsafeSysctls
		}
		if len(config.Kubernetes.ClusterDNSIPs) > 0 {
			b.Kubernetes.ClusterDNSIPs = config.Kubernetes.ClusterDNSIPs
		}
	}
	if config.Kernel != nil {
		if config.Kernel.SysctlSettings != nil {
			b.Kernel = &v1beta1.BottlerocketKernelSettings{
				SysctlSettings: config.Kernel.SysctlSettings,
			}
		}
	}
	if config.Boot != nil {
		if config.Boot.BootKernelParameters != nil {
			b.Boot = &v1beta1.BottlerocketBootSettings{
				BootKernelParameters: config.Boot.BootKernelParameters,
			}
		}
	}

	brMap := map[string]*v1beta1.BottlerocketSettings{
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
func GetExternalEtcdReleaseURL(clusterVersion string, versionBundle *cluster.VersionsBundle) (string, error) {
	clusterVersionSemVer, err := semver.New(clusterVersion)
	if err != nil {
		return "", fmt.Errorf("invalid semver for clusterVersion: %v", err)
	}
	minEksAVersionWithEtcdURL, err := semver.New(v1alpha1.MinEksAVersionWithEtcdURL)
	if err != nil {
		return "", fmt.Errorf("invalid semver for etcd url enabled clusterVersion: %v", err)
	}
	devEksaVersion, err := semver.New(v1alpha1.DevBuildVersion)
	if err != nil {
		return "", fmt.Errorf("invalid semver for dev eksa version: %v", err)
	}
	if clusterVersionSemVer.Equal(minEksAVersionWithEtcdURL) || clusterVersionSemVer.GreaterThan(minEksAVersionWithEtcdURL) ||
		clusterVersionSemVer.Equal(devEksaVersion) {
		return versionBundle.KubeDistro.EtcdURL, nil
	}
	logger.V(4).Info(fmt.Sprintf("Eks-a cluster version is less than version %s. Skip setting etcd url", v1alpha1.MinEksAVersionWithEtcdURL))
	return "", nil
}
