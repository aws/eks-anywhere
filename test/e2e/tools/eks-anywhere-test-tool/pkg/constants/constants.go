package constants

const (
	AwsAccountRegion          = "us-west-2"
	BuildDescriptionFile      = "codebuild-description.json"
	EksATestCodebuildProject  = "aws-eks-anywhere-test"
	E2eIndividualTestLogGroup = "/eks-anywhere/test/e2e"
	E2eArtifactsBucketEnvVar  = "E2E_ARTIFACTS_BUCKET"
	FailedMessage             = "An e2e instance run has failed"
	SuccessMEssage            = "Ec2 instance tests completed successfully"
	FailedTestsFile           = "failed-tests.txt"
	LogOutputFile             = "codebuild-log.txt"
	CiProxyLogGroup           = "nginx-vcenter-proxy.log"
	CiProxyLogStream          = "nginx-vcenter-proxy.log"
)
