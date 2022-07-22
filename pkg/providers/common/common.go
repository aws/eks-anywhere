package common

import (
	_ "embed"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/bootstrapper"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

//go:embed config/audit-policy.yaml
var auditPolicy string

// TODO: Split out common into separate packages to avoid becoming a dumping ground

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

func CPMachineTemplateName(clusterName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s%s%d", clusterName, constants.ControlPlaneTemplatePrefix, t)
}

func EtcdMachineTemplateName(clusterName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s%s%d", clusterName, constants.EtcdTemplatePrefix, t)
}

func WorkerMachineTemplateName(clusterName, workerNodeGroupName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%s-%d", clusterName, workerNodeGroupName, t)
}

func KubeadmConfigTemplateName(clusterName, workerNodeGroupName string, now types.NowFunc) string {
	t := now().UnixNano() / int64(time.Millisecond)
	return fmt.Sprintf("%s-%s-template-%d", clusterName, workerNodeGroupName, t)
}
