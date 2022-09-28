package awsiamauth

import (
	"context"
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
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
)

type KubernetesClient interface {
	GetApiServerUrl(ctx context.Context, cluster *types.Cluster) (string, error)
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	GetClusterCATlsCert(
		ctx context.Context,
		clusterName string,
		cluster *types.Cluster,
		namespace string,
	) ([]byte, error)
}

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
	k8s             KubernetesClient
	writer          filewriter.FileWriter
}

func NewAwsIamAuth(
	certgen crypto.CertificateGenerator,
	clusterId uuid.UUID,
	k8s KubernetesClient,
	writer filewriter.FileWriter,
) *AwsIamAuth {
	return &AwsIamAuth{
		certgen:         certgen,
		templateBuilder: &AwsIamAuthTemplateBuilder{},
		clusterId:       clusterId,
		k8s:             k8s,
		writer:          writer,
	}
}

type AwsIamAuthTemplateBuilder struct{}

func NewAwsIamAuthTemplateBuilder() *AwsIamAuthTemplateBuilder {
	return &AwsIamAuthTemplateBuilder{}
}

func (a *AwsIamAuthTemplateBuilder) GenerateManifest(clusterSpec *cluster.Spec, clusterId uuid.UUID) ([]byte, error) {
	var clusterIdValue string
	// If clusterId is Nil; set value as empty string.
	if clusterId == uuid.Nil {
		clusterIdValue = ""
	} else {
		clusterIdValue = clusterId.String()
	}

	data := map[string]interface{}{
		"image":              clusterSpec.VersionsBundle.KubeDistro.AwsIamAuthImage.VersionedImage(),
		"initContainerImage": clusterSpec.VersionsBundle.Eksa.DiagnosticCollector.VersionedImage(),
		"awsRegion":          clusterSpec.AWSIamConfig.Spec.AWSRegion,
		"clusterID":          clusterIdValue,
		"backendMode":        strings.Join(clusterSpec.AWSIamConfig.Spec.BackendMode, ","),
		"partition":          clusterSpec.AWSIamConfig.Spec.Partition,
	}

	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints != nil {
		data["controlPlaneTaints"] = clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Taints
	}

	mapRoles, err := a.mapRolesToYaml(clusterSpec.AWSIamConfig.Spec.MapRoles)
	if err != nil {
		return nil, fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}
	data["mapRoles"] = mapRoles
	mapUsers, err := a.mapUsersToYaml(clusterSpec.AWSIamConfig.Spec.MapUsers)
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

func (a *AwsIamAuth) GenerateManifest(clusterSpec *cluster.Spec) ([]byte, error) {
	return a.templateBuilder.GenerateManifest(clusterSpec, a.clusterId)
}

func (a *AwsIamAuth) GenerateManifestForUpgrade(clusterSpec *cluster.Spec) ([]byte, error) {
	return a.templateBuilder.GenerateManifest(clusterSpec, uuid.Nil)
}

func (a *AwsIamAuth) GenerateCertKeyPairSecret(managementClusterName string) ([]byte, error) {
	certPemBytes, keyPemBytes, err := a.certgen.GenerateIamAuthSelfSignCertKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generating aws-iam-authenticator cert key pair secret: %v", err)
	}
	data := map[string]string{
		"clusterName":  managementClusterName,
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

func (a *AwsIamAuth) generateAwsIamAuthKubeconfig(clusterSpec *cluster.Spec, serverUrl, tlsCert string) ([]byte, error) {
	data := map[string]string{
		"clusterName": clusterSpec.Cluster.Name,
		"server":      serverUrl,
		"cert":        tlsCert,
		"clusterID":   a.clusterId.String(),
	}
	awsIamAuthKubeconfig, err := templater.Execute(awsIamAuthKubeconfigTemplate, data)
	if err != nil {
		return nil, fmt.Errorf("generating aws-iam-authenticator kubeconfig content: %v", err)
	}
	return awsIamAuthKubeconfig, nil
}

func (a *AwsIamAuthTemplateBuilder) mapRolesToYaml(m []v1alpha1.MapRoles) (string, error) {
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

func (a *AwsIamAuthTemplateBuilder) mapUsersToYaml(m []v1alpha1.MapUsers) (string, error) {
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

func (a *AwsIamAuth) InstallAWSIAMAuth(
	ctx context.Context,
	management, workload *types.Cluster,
	spec *cluster.Spec,
) error {
	manifest, err := a.GenerateManifest(spec)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}

	if err = a.k8s.ApplyKubeSpecFromBytes(ctx, workload, manifest); err != nil {
		return fmt.Errorf("applying aws-iam-authenticator manifest: %v", err)
	}

	if err = a.generateKubeconfig(ctx, management, workload, spec); err != nil {
		return err
	}
	return nil
}

func (a *AwsIamAuth) generateKubeconfig(
	ctx context.Context,
	management, workload *types.Cluster,
	spec *cluster.Spec,
) error {
	fileName := fmt.Sprintf("%s-aws.kubeconfig", workload.Name)

	serverUrl, err := a.k8s.GetApiServerUrl(ctx, workload)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}

	tlsCert, err := a.k8s.GetClusterCATlsCert(
		ctx,
		workload.Name,
		management,
		constants.EksaSystemNamespace,
	)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}

	awsIamAuthKubeconfigContent, err := a.generateAwsIamAuthKubeconfig(spec, serverUrl, string(tlsCert))
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}

	writtenFile, err := a.writer.Write(
		fileName,
		awsIamAuthKubeconfigContent,
		filewriter.PersistentFile,
		filewriter.Permission0600,
	)
	if err != nil {
		return fmt.Errorf("writing aws-iam-authenticator kubeconfig to %s: %v", writtenFile, err)
	}

	logger.V(3).Info("Generated aws-iam-authenticator kubeconfig", "kubeconfig", writtenFile)

	return nil
}

// CreateAndInstallAWSIAMAuthCASecret creates a Kubernetes Secret in cluster containing a
// self-signed certificate and key for a cluster identified by clusterName.
func (a *AwsIamAuth) CreateAndInstallAWSIAMAuthCASecret(ctx context.Context, cluster *types.Cluster, clusterName string) error {
	secret, err := a.GenerateCertKeyPairSecret(clusterName)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator ca secret: %v", err)
	}

	if err = a.k8s.ApplyKubeSpecFromBytes(ctx, cluster, secret); err != nil {
		return fmt.Errorf("applying aws-iam-authenticator ca secret: %v", err)
	}

	return nil
}

func (a *AwsIamAuth) UpgradeAWSIAMAuth(ctx context.Context, cluster *types.Cluster, spec *cluster.Spec) error {
	awsIamAuthManifest, err := a.GenerateManifestForUpgrade(spec)
	if err != nil {
		return fmt.Errorf("generating manifest: %v", err)
	}

	err = a.k8s.ApplyKubeSpecFromBytes(ctx, cluster, awsIamAuthManifest)
	if err != nil {
		return fmt.Errorf("applying manifest: %v", err)
	}

	return nil
}
