package awsiamauth

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
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
	certgen         crypto.CertificateGenerator
	templateBuilder *AwsIamAuthTemplateBuilder
	clusterId       uuid.UUID
}

func NewAwsIamAuth(certgen crypto.CertificateGenerator, clusterId uuid.UUID) *AwsIamAuth {
	return &AwsIamAuth{
		certgen:         certgen,
		templateBuilder: &AwsIamAuthTemplateBuilder{},
		clusterId:       clusterId,
	}
}

type AwsIamAuthTemplateBuilder struct{}

func NewAwsIamAuthTemplateBuilder() *AwsIamAuthTemplateBuilder {
	return &AwsIamAuthTemplateBuilder{}
}

func (a *AwsIamAuthTemplateBuilder) GenerateManifest(clusterSpec *cluster.Spec, clusterId uuid.UUID) ([]byte, error) {
	data := map[string]interface{}{
		"image":       clusterSpec.VersionsBundle.KubeDistro.AwsIamAuthIamge.VersionedImage(),
		"awsRegion":   clusterSpec.AWSIamConfig.Spec.AWSRegion,
		"clusterID":   clusterId.String(),
		"backendMode": strings.Join(clusterSpec.AWSIamConfig.Spec.BackendMode, ","),
		"partition":   clusterSpec.AWSIamConfig.Spec.Partition,
	}

	if clusterSpec.Spec.ControlPlaneConfiguration.Taints != nil {
		data["controlPlaneTaints"] = clusterSpec.Spec.ControlPlaneConfiguration.Taints
	}

	mapRoles, err := a.mapRolesToYaml(clusterSpec.AWSIamConfig.Spec.MapRoles)
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator manifest: %v", err)
	}
	data["mapRoles"] = mapRoles
	mapUsers, err := a.mapUsersToYaml(clusterSpec.AWSIamConfig.Spec.MapUsers)
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator manifest: %v", err)
	}
	data["mapUsers"] = mapUsers
	awsIamAuthManifest, err := templater.Execute(awsIamAuthTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator manifest: %v", err)
	}
	return awsIamAuthManifest, nil
}

func (a *AwsIamAuth) GenerateManifest(clusterSpec *cluster.Spec) ([]byte, error) {
	return a.templateBuilder.GenerateManifest(clusterSpec, a.clusterId)
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
		"clusterID":   a.clusterId.String(),
	}
	awsIamAuthKubeconfig, err := templater.Execute(awsIamAuthKubeconfigTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("error generating aws-iam-authenticator kubeconfig content: %v", err)
	}
	return awsIamAuthKubeconfig, nil
}

func (a *AwsIamAuthTemplateBuilder) mapRolesToYaml(m []v1alpha1.MapRoles) (string, error) {
	if len(m) == 0 {
		return "", nil
	}
	b, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("error marshalling AWSIamConfig MapRoles: %v", err)
	}
	s := string(b)
	s = strings.TrimSuffix(s, "\n")

	return s, nil
}

func (a *AwsIamAuthTemplateBuilder) mapUsersToYaml(m []v1alpha1.MapUsers) (string, error) {
	if len(m) == 0 {
		return "", nil
	}
	b, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("error marshalling AWSIamConfig MapUsers: %v", err)
	}
	s := string(b)
	s = strings.TrimSuffix(s, "\n")

	return s, nil
}
