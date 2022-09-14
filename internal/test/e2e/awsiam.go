package e2e

import (
	"os"
	"regexp"

	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupAwsIam(testRegex string) error {
	re := regexp.MustCompile(`^.*AWSIamAuth.*$`)
	if !re.MatchString(testRegex) {
		e.logger.V(2).Info("Not running AWSIamAuth tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredAWSIamEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	return nil
}
