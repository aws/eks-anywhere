package common

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

//go:embed config/audit-policy.yaml
var auditPolicy string

var NoProxyDefaults = []string{
	"localhost",
	"127.0.0.1",
	".svc",
}

const (
	privateKeyFileName = "eks-a-id_rsa"
	publicKeyFileName  = "eks-a-id_rsa.pub"
)

func GetAuditPolicy() string {
	return auditPolicy
}

func BootstrapClusterOpts(serverEndpoint string, clusterConfig *v1alpha1.Cluster) ([]bootstrapper.BootstrapClusterOption, error) {
	env := map[string]string{}
	if clusterConfig.Spec.ProxyConfiguration != nil {
		noProxy := fmt.Sprintf("%s,%s", clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host, serverEndpoint)
		for _, s := range clusterConfig.Spec.ProxyConfiguration.NoProxy {
			if s != "" {
				noProxy += "," + s
			}
		}
		env["HTTP_PROXY"] = clusterConfig.Spec.ProxyConfiguration.HttpProxy
		env["HTTPS_PROXY"] = clusterConfig.Spec.ProxyConfiguration.HttpsProxy
		env["NO_PROXY"] = noProxy
	}
	return []bootstrapper.BootstrapClusterOption{bootstrapper.WithEnv(env)}, nil
}

func ParseSSHAuthKey(key *string) error {
	if len(*key) > 0 {
		// When public key is entered by user in provider config, it may contain email address (or any other comment) at the end. ssh-keygen allows users to add comments as suffixes to public key in
		// public key file. When CLI generates the key pair, no comments will be present. So we get rid of the comment from the public key to ensure unit tests that do string compare on the sshAuthorizedKey
		// will pass
		parts := strings.Fields(strings.TrimSpace(*key))
		if len(parts) >= 3 {
			*key = parts[0] + " " + parts[1]
		}
		_, _, _, _, err := ssh.ParseAuthorizedKey([]byte(*key))
		if err != nil {
			return fmt.Errorf("provided MachineConfig sshAuthorizedKey is invalid: %v", err)
		}
	}
	return nil
}

func GenerateSSHAuthKey(username string, writer filewriter.FileWriter) (string, error) {
	privateKeyPath, sshAuthorizedKeyBytes, err := crypto.NewSshKeyPairUsingFileWriter(writer, privateKeyFileName, publicKeyFileName)
	if err != nil {
		return "", fmt.Errorf("VSphereMachineConfig error generating sshAuthorizedKey: %v", err)
	}

	logger.Info(fmt.Sprintf(
		"DatacenterConfig private key saved to %[1]s. Use 'ssh -i %[1]s %s@<VM-IP-Address>' to login to your cluster VM",
		privateKeyPath,
		username,
	))

	key := string(sshAuthorizedKeyBytes)
	key = strings.TrimRight(key, "\n")

	return key, nil
}

func CPMachineTemplateName(clusterName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-control-plane-template-%d", clusterName, t)
}

func EtcdMachineTemplateName(clusterName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-etcd-template-%d", clusterName, t)
}

func WorkerMachineTemplateName(clusterName, workerNodeGroupName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%s-%d", clusterName, workerNodeGroupName, t)
}

func KubeadmConfigTemplateName(clusterName, workerNodeGroupName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%s-template-%d", clusterName, workerNodeGroupName, t)
}
