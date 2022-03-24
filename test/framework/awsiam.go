package framework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/awsiam"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/version"
)

const (
	AWSIamRoleArn = "T_AWS_IAM_ROLE_ARN"
)

var awsIamRequiredEnvVars = []string{
	AWSIamRoleArn,
}

func RequiredAWSIamEnvVars() []string {
	return awsIamRequiredEnvVars
}

func WithAWSIam() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, awsIamRequiredEnvVars)
		e.AWSIamConfig = api.NewAWSIamConfig(defaultClusterName,
			api.WithAWSIamAWSRegion("us-west-1"),
			api.WithAWSIamPartition("aws"),
			api.WithAWSIamBackendMode("EKSConfigMap"),
			api.WithAWSIamMapRoles(api.AddAWSIamRole(withArnFromEnv(AWSIamRoleArn), "kubernetes-admin", []string{"system:masters"})),
		)
		e.clusterFillers = append(e.clusterFillers,
			api.WithAWSIamIdentityProviderRef(defaultClusterName),
		)
	}
}

func withArnFromEnv(envVar string) string {
	return os.Getenv(envVar)
}

func (e *ClusterE2ETest) ValidateAWSIamAuth() {
	ctx := context.Background()
	e.T.Log("Downloading aws-iam-authenticator client")
	err := e.downloadAwsIamAuthClient()
	if err != nil {
		e.T.Fatalf("Error downloading aws-iam-authenticator client: %v", err)
	}
	e.T.Log("Setting aws-iam-authenticator client in env PATH")
	err = e.setIamAuthClientPATH()
	if err != nil {
		e.T.Fatalf("Error updating PATH: %v", err)
	}
	e.T.Log("Getting pods with aws-iam-authenticator kubeconfig")
	kubectlClient := buildLocalKubectl()
	pods, err := kubectlClient.GetPods(ctx,
		executables.WithAllNamespaces(),
		executables.WithKubeconfig(e.iamAuthKubeconfigFilePath()),
	)
	if err != nil {
		e.T.Fatalf("Error getting pods: %v", err)
	}
	if len(pods) > 0 {
		e.T.Log("Successfully got pods with aws-iam-authenticator authentication")
	}
}

func (e *ClusterE2ETest) downloadAwsIamAuthClient() error {
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

func (e *ClusterE2ETest) setIamAuthClientPATH() error {
	envPath := os.Getenv("PATH")
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("finding current working directory: %v", err)
	}
	iamAuthClientPath := fmt.Sprintf("%s/bin", workDir)
	if strings.Contains(envPath, iamAuthClientPath) {
		return nil
	}
	err = os.Setenv("PATH", fmt.Sprintf("%s:%s", iamAuthClientPath, envPath))
	if err != nil {
		return fmt.Errorf("setting %s to PATH: %v", iamAuthClientPath, err)
	}
	return nil
}

func (e *ClusterE2ETest) getEksdReleaseManifest() (*eksdv1alpha1.Release, error) {
	c := e.clusterConfig()
	_, eksdRelease, err := cluster.GetEksdRelease(version.Get(), c)
	if err != nil {
		return nil, fmt.Errorf("getting EKS-D release spec from bundle: %v", err)
	}
	return eksdRelease, nil
}

func (e *ClusterE2ETest) iamAuthKubeconfigFilePath() string {
	return filepath.Join(e.ClusterName, fmt.Sprintf("%s-aws.kubeconfig", e.ClusterName))
}
