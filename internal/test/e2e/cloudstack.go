package e2e

import (
	e2etests "github.com/aws/eks-anywhere/test/framework"
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	cloudstackRegex = `^.*CloudStack.*$`
)

func (e *E2ESession) setupCloudStackEnv(testRegex string) error {
	re := regexp.MustCompile(cloudstackRegex)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running CloudStack tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredCloudstackEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}
	return nil
}
