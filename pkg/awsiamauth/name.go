package awsiamauth

import "fmt"

const (
	awsIamAuthCaSecretSuffix = "aws-iam-authenticator-ca"

	// AwsIamAuthConfigMapName is the name of AWS IAM Authenticator configuration.
	AwsIamAuthConfigMapName = "aws-iam-authenticator"

	// AwsAuthConfigMapName is the name of IAM roles and users mapping for AWS IAM Authenticator.
	AwsAuthConfigMapName = "aws-auth"
)

// GetAwsIamAuthCaSecretName returns the name of AWS IAM Authenticator secret for the cluster.
func GetAwsIamAuthCaSecretName(clusterName string) string {
	return fmt.Sprintf("%s-%s", clusterName, awsIamAuthCaSecretSuffix)
}
