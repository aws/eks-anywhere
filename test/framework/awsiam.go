package framework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/awsiam"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/version"
)

const (
	AWSIamRoleArn = "T_AWS_IAM_ROLE_ARN"
	AWSIamUserArn = "T_AWS_IAM_USER_ARN"
)

var awsIamRequiredEnvVars = []string{
	AWSIamRoleArn,
}

func RequiredAWSIamEnvVars() []string {
	return awsIamRequiredEnvVars
}

func WithAWSIam() E2ETestOpt {
	return func(e *E2ETest) {
		checkRequiredEnvVars(e.T, awsIamRequiredEnvVars)
		e.AWSIamConfig = api.NewAWSIamConfig(defaultClusterName,
			api.WithAWSIamAWSRegion("us-west-1"),
			api.WithAWSIamClusterID(defaultClusterName),
			api.WithAWSIamPartition("aws"),
			api.WithAWSIamBackendMode("EKSConfigMap"),
			api.AddAWSIamMapRoles(api.WithAWSIamRole(withArnFromEnv(AWSIamRoleArn), "kubernetes-admin", []string{"system:masters"})),
		)
		e.clusterFillers = append(e.clusterFillers,
			api.WithAWSIamIdentityProviderRef(defaultClusterName),
		)
	}
}

func withArnFromEnv(envVar string) string {
	return os.Getenv(envVar)
}

func (e *E2ETest) ValidateAWSIamAuth() {
	ctx := context.Background()
	e.T.Log("Downloading aws-iam-authenticator client")
	err := e.downloadAwsIamAuthClient()
	if err != nil {
		e.T.Fatalf("Error downloading aws-iam-authenticator client: %v", err)
	}
	e.T.Log("Getting pods with aws-iam-authenticator kubeconfig")
	pods, err := awsiam.GetAllPods(ctx, "--kubeconfig", e.iamAuthKubeconfigFilePath())
	if err != nil {
		e.T.Fatalf("Error getting pods: %v", err)
	}
	if len(pods) > 0 {
		e.T.Log("Successfully got pods with aws-iam-authenticator authentication")
	}
}

func (e *E2ETest) downloadAwsIamAuthClient() error {
	eksdRelease, err := e.getEksdReleaseManifest()
	if err != nil {
		return err
	}
	err = awsiam.DownloadAwsIamAuthClient(eksdRelease)
	if err != nil {
		return err
	}
	return nil
}

func (e *E2ETest) getEksdReleaseManifest() (*eksdv1alpha1.Release, error) {
	c := e.clusterConfig()
	_, eksdRelease, err := cluster.GetEksdRelease(version.Get(), c)
	if err != nil {
		return nil, fmt.Errorf("error getting EKS-D release spec from bundle: %v", err)
	}
	return eksdRelease, nil
}

func (e *E2ETest) iamAuthKubeconfigFilePath() string {
	return filepath.Join(e.ClusterName, fmt.Sprintf("%s-aws.kubeconfig", e.ClusterName))
}
