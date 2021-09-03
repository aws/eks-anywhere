package awsiamauth

import (
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/templater"
)

//go:embed config/aws-iam-authenticator.yaml
var awsIamAuthTemplate []byte

//go:embed config/aws-iam-authenticator-ca-secret.yaml
var awsIamAuthCaSecretTemplate []byte

type AwsIamAuth struct {
	certgen crypto.CertificateGenerator
}

func NewAwsIamAuth(certgen crypto.CertificateGenerator) *AwsIamAuth {
	return &AwsIamAuth{certgen: certgen}
}

func (a *AwsIamAuth) GenerateManifest(clusterSpec *cluster.Spec) ([]byte, error) {
	data := map[string]string{
		"image":       clusterSpec.VersionsBundle.KubeDistro.AwsIamAuthIamge.VersionedImage(),
		"clusterID":   clusterSpec.AWSIamConfig.Spec.ClusterID,
		"backendMode": clusterSpec.AWSIamConfig.Spec.BackendMode,
		"config":      clusterSpec.AWSIamConfig.Spec.Data,
	}
	awsIamAuthManifest, err := templater.Execute(string(awsIamAuthTemplate), data)
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator manifest: %v", err)
	}
	return awsIamAuthManifest, nil
}

func (a *AwsIamAuth) GenerateCertKeyPairSecret() ([]byte, error) {
	certPemBytes, keyPemBytes, err := a.certgen.GenerateSelfSignCertKeyPair()
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator cert key pair secret: %v", err)
	}
	data := map[string]string{
		"namespace":    constants.EksaSystemNamespace,
		"certPemBytes": base64.StdEncoding.EncodeToString(certPemBytes),
		"keyPemBytes":  base64.StdEncoding.EncodeToString(keyPemBytes),
	}
	awsIamAuthCaSecret, err := templater.Execute(string(awsIamAuthCaSecretTemplate), data)
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator cert key pair secret: %v", err)
	}
	return awsIamAuthCaSecret, nil
}
