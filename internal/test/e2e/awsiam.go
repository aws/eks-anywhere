package e2e

import (
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupAwsIam(testRegex string) error {
	re := regexp.MustCompile(`^.*AWSIamAuth.*$`)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running AWSIamAuth tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredAWSIamEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	// Activate feature gate
	e.testEnvVars[features.AwsIamAuthenticatorEnvVar] = "true"
	logger.V(1).Info("Activated feature gate", features.AwsIamAuthenticatorEnvVar, "true")

	return nil
}
