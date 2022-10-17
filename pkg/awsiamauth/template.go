package awsiamauth

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/templater"
)

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

	data := map[string]interface{}{
		"image":              clusterSpec.VersionsBundle.KubeDistro.AwsIamAuthImage.VersionedImage(),
		"initContainerImage": clusterSpec.VersionsBundle.Eksa.DiagnosticCollector.VersionedImage(),
		"awsRegion":          clusterSpec.AWSIamConfig.Spec.AWSRegion,
		"clusterID":          clusterIDValue,
		"backendMode":        strings.Join(clusterSpec.AWSIamConfig.Spec.BackendMode, ","),
		"partition":          clusterSpec.AWSIamConfig.Spec.Partition,
	}

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
