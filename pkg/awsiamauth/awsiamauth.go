package awsiamauth

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/aws-iam-authenticator.yaml
var awsIamAuthTemplate string

//go:embed config/aws-iam-authenticator-ca-secret.yaml
var awsIamAuthCaSecretTemplate string

//go:embed config/aws-iam-authenticator-kubeconfig.yaml
var awsIamAuthKubeconfigTemplate string

type AwsIamAuth struct {
	certgen crypto.CertificateGenerator
}

func NewAwsIamAuth(certgen crypto.CertificateGenerator) *AwsIamAuth {
	return &AwsIamAuth{certgen: certgen}
}

func (a *AwsIamAuth) GenerateManifest(clusterSpec *cluster.Spec) ([]byte, error) {
	data := map[string]string{
		"image":       clusterSpec.VersionsBundle.KubeDistro.AwsIamAuthIamge.VersionedImage(),
		"awsRegion":   clusterSpec.AWSIamConfig.Spec.AWSRegion,
		"clusterID":   clusterSpec.AWSIamConfig.Spec.ClusterID,
		"backendMode": strings.Join(clusterSpec.AWSIamConfig.Spec.BackendMode, ","),
		"mapRoles":    clusterSpec.AWSIamConfig.Spec.MapRoles,
		"mapUsers":    clusterSpec.AWSIamConfig.Spec.MapUsers,
		"partition":   clusterSpec.AWSIamConfig.Spec.Partition,
	}
	awsIamAuthManifest, err := templater.Execute(awsIamAuthTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator manifest: %v", err)
	}
	return awsIamAuthManifest, nil
}

func (a *AwsIamAuth) GenerateCertKeyPairSecret() ([]byte, error) {
	certPemBytes, keyPemBytes, err := a.certgen.GenerateIamAuthSelfSignCertKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator cert key pair secret: %v", err)
	}
	data := map[string]string{
		"namespace":    constants.EksaSystemNamespace,
		"certPemBytes": base64.StdEncoding.EncodeToString(certPemBytes),
		"keyPemBytes":  base64.StdEncoding.EncodeToString(keyPemBytes),
	}
	awsIamAuthCaSecret, err := templater.Execute(awsIamAuthCaSecretTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator cert key pair secret: %v", err)
	}
	return awsIamAuthCaSecret, nil
}

func (a *AwsIamAuth) GenerateAwsIamAuthKubeconfig(clusterSpec *cluster.Spec, serverUrl, tlsCert string) ([]byte, error) {
	data := map[string]string{
		"clusterName": clusterSpec.Cluster.Name,
		"server":      serverUrl,
		"cert":        tlsCert,
		"clusterID":   clusterSpec.AWSIamConfig.Spec.ClusterID,
	}
	awsIamAuthKubeconfig, err := templater.Execute(awsIamAuthKubeconfigTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator kubeconfig content: %v", err)
	}
	return awsIamAuthKubeconfig, nil
}
