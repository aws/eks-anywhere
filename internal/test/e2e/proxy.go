package e2e

import (
	"os"
	"regexp"

	"github.com/aws/eks-anywhere/pkg/logger"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

func (e *E2ESession) setupProxyEnv(testRegex string) error {
	re := regexp.MustCompile(`^.*Proxy.*$`)
	if !re.MatchString(testRegex) {
		logger.V(2).Info("Not running Proxy tests, skipping Env variable setup")
		return nil
	}

	requiredEnvVars := e2etests.RequiredProxyEnvVars()
	for _, eVar := range requiredEnvVars {
		if val, ok := os.LookupEnv(eVar); ok {
			e.testEnvVars[eVar] = val
		}
	}

	return nil
}
