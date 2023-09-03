package awsiamauth

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

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

// TemplateBuilder generates manifest files from templates.
type TemplateBuilder struct{}

// GenerateManifest generates a YAML Kubernetes manifest for deploying the AWS IAM Authenticator.
func (t *TemplateBuilder) GenerateManifest(clusterSpec *cluster.Spec, clusterID uuid.UUID) ([]byte, error) {
	// Give uuid.Nil semantics that result in no ConfigMap being generated containing the cluster ID
	var clusterIDValue string
	if clusterID == uuid.Nil {
		clusterIDValue = ""
	} else {
		clusterIDValue = clusterID.String()
	}

	bundle := clusterSpec.RootVersionsBundle()

	data := map[string]interface{}{
		"image":              bundle.KubeDistro.AwsIamAuthImage.VersionedImage(),
		"initContainerImage": bundle.Eksa.DiagnosticCollector.VersionedImage(),
		"awsRegion":          clusterSpec.AWSIamConfig.Spec.AWSRegion,
		"clusterID":          clusterIDValue,
		"backendMode":        strings.Join(clusterSpec.AWSIamConfig.Spec.BackendMode, ","),
		"partition":          clusterSpec.AWSIamConfig.Spec.Partition,
	}

	nodeSelector, err := t.setControlPlaneNodeSelector(clusterSpec.Cluster.Spec.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("setting control plane node selector: %v", err)
	}
	data["cpNodeSelector"] = nodeSelector

	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints != nil {
		data["controlPlaneTaints"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints
	}

	mapRoles, err := t.mapRolesToYaml(clusterSpec.AWSIamConfig.Spec.MapRoles)
	if err != nil {
		return nil, fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}
	data["mapRoles"] = mapRoles
	mapUsers, err := t.mapUsersToYaml(clusterSpec.AWSIamConfig.Spec.MapUsers)
	if err != nil {
		return nil, fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}
	data["mapUsers"] = mapUsers
	awsIamAuthManifest, err := templater.Execute(awsIamAuthTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}
	return awsIamAuthManifest, nil
}

func (t *TemplateBuilder) mapRolesToYaml(m []v1alpha1.MapRoles) (string, error) {
	if len(m) == 0 {
		return "", nil
	}
	b, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("marshalling AWSIamConfig MapRoles: %v", err)
	}
	s := string(b)
	s = strings.TrimSuffix(s, "\n")

	return s, nil
}

func (t *TemplateBuilder) mapUsersToYaml(m []v1alpha1.MapUsers) (string, error) {
	if len(m) == 0 {
		return "", nil
	}
	b, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("marshalling AWSIamConfig MapUsers: %v", err)
	}
	s := string(b)
	s = strings.TrimSuffix(s, "\n")

	return s, nil
}

func (t *TemplateBuilder) setControlPlaneNodeSelector(kubeVersion v1alpha1.KubernetesVersion) (string, error) {
	var nodeSelector string
	clusterKubeVersionSemver, err := v1alpha1.KubeVersionToSemver(kubeVersion)
	if err != nil {
		return "", fmt.Errorf("converting kubeVersion %v to semver %v", kubeVersion, err)
	}
	kube124Semver, err := v1alpha1.KubeVersionToSemver(v1alpha1.Kube124)
	if err != nil {
		return "", fmt.Errorf("converting kubeVersion %v to semver %v", v1alpha1.Kube124, err)
	}
	if clusterKubeVersionSemver.Compare(kube124Semver) != -1 {
		nodeSelector = `node-role.kubernetes.io/control-plane: ""`
	} else {
		nodeSelector = `node-role.kubernetes.io/master: ""`
	}

	return nodeSelector, nil
}

// GenerateCertKeyPairSecret generates a YAML Kubernetes Secret for deploying the AWS IAM Authenticator.
func (t *TemplateBuilder) GenerateCertKeyPairSecret(certgen crypto.CertificateGenerator, managementClusterName string) ([]byte, error) {
	certPemBytes, keyPemBytes, err := certgen.GenerateIamAuthSelfSignCertKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generating aws-iam-authenticator cert key pair secret: %v", err)
	}
	data := map[string]string{
		"name":         CASecretName(managementClusterName),
		"namespace":    constants.EksaSystemNamespace,
		"certPemBytes": base64.StdEncoding.EncodeToString(certPemBytes),
		"keyPemBytes":  base64.StdEncoding.EncodeToString(keyPemBytes),
	}
	awsIamAuthCaSecret, err := templater.Execute(awsIamAuthCaSecretTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("generating aws-iam-authenticator cert key pair secret: %v", err)
	}
	return awsIamAuthCaSecret, nil
}

// GenerateKubeconfig generates a Kubeconfig in yaml format to authenticate with AWS IAM Authenticator.
func (t *TemplateBuilder) GenerateKubeconfig(clusterSpec *cluster.Spec, clusterID uuid.UUID, serverURL, tlsCert string) ([]byte, error) {
	data := map[string]string{
		"clusterName": clusterSpec.Cluster.Name,
		"server":      serverURL,
		"cert":        tlsCert,
		"clusterID":   clusterID.String(),
	}
	awsIamAuthKubeconfig, err := templater.Execute(awsIamAuthKubeconfigTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("generating aws-iam-authenticator kubeconfig content: %v", err)
	}
	return awsIamAuthKubeconfig, nil
}
